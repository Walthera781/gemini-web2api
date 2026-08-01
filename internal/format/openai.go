package format

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type ToolDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func randHex(n int) string {
	bytes := make([]byte, (n+1)/2)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)[:n]
}

func BuildToolChoiceInstruction(toolChoice any) string {
	if strChoice, ok := toolChoice.(string); ok {
		if strChoice == "none" {
			return "\n\nIMPORTANT: Do NOT call any tools. Respond with text only."
		}
		if strChoice == "required" {
			return "\n\nIMPORTANT: You MUST call at least one tool. Do not respond with text only."
		}
	} else if mapChoice, ok := toolChoice.(map[string]any); ok {
		if fn, ok := mapChoice["function"].(map[string]any); ok {
			if name, ok := fn["name"].(string); ok && name != "" {
				return fmt.Sprintf("\n\nIMPORTANT: You MUST call the tool \"%s\". Do not call other tools.", name)
			}
		}
	}
	return ""
}

func MessagesToPrompt(messages []map[string]any, tools []map[string]any, toolChoice any) (string, error) {
	var parts []string

	strChoice, isStr := toolChoice.(string)
	if !(isStr && strChoice == "none") && len(tools) > 0 {
		var toolDefs []ToolDef
		for _, tool := range tools {
			var fn map[string]any
			if tType, ok := tool["type"].(string); ok && tType == "function" {
				if f, ok := tool["function"].(map[string]any); ok {
					fn = f
				} else {
					fn = tool
				}
			} else {
				fn = tool
			}

			name, _ := fn["name"].(string)
			if name == "" {
				name, _ = tool["name"].(string)
			}
			desc, _ := fn["description"].(string)
			if desc == "" {
				desc, _ = tool["description"].(string)
			}
			params := fn["parameters"]
			if params == nil {
				params = tool["parameters"]
			}
			if params == nil {
				params = map[string]any{}
			}

			toolDefs = append(toolDefs, ToolDef{
				Name:        name,
				Description: desc,
				Parameters:  params,
			})
		}

		if len(toolDefs) > 0 {
			constraint := BuildToolChoiceInstruction(toolChoice)
			defsJSON, _ := json.MarshalIndent(toolDefs, "", "  ")
			parts = append(parts, fmt.Sprintf(
				"# Tool Use\n\n"+
					"You can call the following tools. Call format:\n"+
					"```tool_call\n{\"name\": \"func_name\", \"arguments\": {...}}\n```\n"+
					"When calling tools, output ONLY the tool_call block(s).\n\n"+
					"Available tools:\n%s%s",
				string(defsJSON),
				constraint,
			))
		}
	}

	for _, msg := range messages {
		role, _ := msg["role"].(string)
		if role == "" {
			role = "user"
		}
		rawContent := msg["content"]

		var contentStr string
		if contentList, ok := rawContent.([]any); ok {
			var textParts []string
			for _, item := range contentList {
				if mapItem, ok := item.(map[string]any); ok {
					iType, _ := mapItem["type"].(string)
					if iType == "text" || iType == "input_text" {
						if t, ok := mapItem["text"].(string); ok {
							textParts = append(textParts, t)
						}
					} else if iType == "image_url" || iType == "image" {
						textParts = append(textParts, "[Note: Image input not supported in this API. Please describe the image in text.]")
					}
				}
			}
			contentStr = strings.Join(textParts, " ")
		} else if strContent, ok := rawContent.(string); ok {
			contentStr = strContent
		}

		switch role {
		case "system":
			parts = append(parts, fmt.Sprintf("[System instruction]: %s", contentStr))
		case "assistant":
			toolCalls, _ := msg["tool_calls"].([]any)
			if len(toolCalls) > 0 {
				var tcStrs []string
				for _, tc := range toolCalls {
					if tcMap, ok := tc.(map[string]any); ok {
						fn, _ := tcMap["function"].(map[string]any)
						name, _ := fn["name"].(string)
						argsRaw := fn["arguments"]
						var argsStr string
						if str, ok := argsRaw.(string); ok {
							argsStr = str
						} else {
							b, _ := json.Marshal(argsRaw)
							argsStr = string(b)
						}
						if argsStr == "" {
							argsStr = "{}"
						}
						tcStrs = append(tcStrs, fmt.Sprintf("```tool_call\n{\"name\": \"%s\", \"arguments\": %s}\n```", name, argsStr))
					}
				}
				parts = append(parts, fmt.Sprintf("[Assistant]: %s\n%s", contentStr, strings.Join(tcStrs, "\n")))
			} else {
				parts = append(parts, fmt.Sprintf("[Assistant]: %s", contentStr))
			}
		case "tool":
			name, _ := msg["name"].(string)
			parts = append(parts, fmt.Sprintf("[Tool result for %s]: %s", name, contentStr))
		default:
			if contentStr != "" {
				parts = append(parts, contentStr)
			}
		}
	}

	return strings.Join(parts, "\n\n"), nil
}

var reToolCall = regexp.MustCompile(`(?s)\x60\x60\x60tool_call\s*\n(.*?)\n\x60\x60\x60`)

func ParseToolCalls(text string) (string, []ToolCall) {
	var toolCalls []ToolCall
	matches := reToolCall.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return text, nil
	}

	var cleanParts []string
	lastEnd := 0

	for _, m := range matches {
		cleanParts = append(cleanParts, text[lastEnd:m[0]])
		lastEnd = m[1]

		content := strings.TrimSpace(text[m[2]:m[3]])
		var data struct {
			Name      string `json:"name"`
			Arguments any    `json:"arguments"`
		}
		if err := json.Unmarshal([]byte(content), &data); err == nil && data.Name != "" {
			var argsStr string
			if str, ok := data.Arguments.(string); ok {
				argsStr = str
			} else if data.Arguments != nil {
				b, _ := json.Marshal(data.Arguments)
				argsStr = string(b)
			} else {
				argsStr = "{}"
			}

			toolCalls = append(toolCalls, ToolCall{
				ID:   fmt.Sprintf("call_%s", randHex(8)),
				Type: "function",
				Function: ToolCallFunction{
					Name:      data.Name,
					Arguments: argsStr,
				},
			})
		}
	}

	cleanParts = append(cleanParts, text[lastEnd:])
	clean := strings.TrimSpace(strings.Join(cleanParts, ""))
	return clean, toolCalls
}
