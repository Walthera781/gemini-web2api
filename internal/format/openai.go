package format

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ikhsan3adi/gemini-web2api/internal/models"
)

func RandHex(n int) string {
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

func MessagesToPrompt(req models.OpenAIChatRequest) (string, error) {
	var parts []string

	strChoice, isStr := req.ToolChoice.(string)
	if !(isStr && strChoice == "none") && len(req.Tools) > 0 {
		var toolDefs []models.OpenAIFunction
		for _, tool := range req.Tools {
			fn := tool.Function
			if fn.Name == "" {
				fn.Name = tool.Name
			}
			if fn.Description == "" {
				fn.Description = tool.Description
			}
			if fn.Parameters == nil {
				fn.Parameters = tool.Parameters
			}
			if fn.Parameters == nil {
				fn.Parameters = map[string]any{}
			}

			toolDefs = append(toolDefs, fn)
		}

		if len(toolDefs) > 0 {
			constraint := BuildToolChoiceInstruction(req.ToolChoice)
			defsJSON, _ := json.Marshal(toolDefs)
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

	for _, msg := range req.Messages {
		role := msg.Role
		if role == "" {
			role = "user"
		}

		var contentStr string
		if strContent, ok := msg.Content.(string); ok {
			contentStr = strContent
		} else if contentList, ok := msg.Content.([]any); ok {
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
		}

		switch role {
		case "system":
			parts = append(parts, fmt.Sprintf("[System instruction]: %s", contentStr))
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				var tcStrs []string
				for _, tc := range msg.ToolCalls {
					argsStr := tc.Function.Arguments
					if argsStr == "" {
						argsStr = "{}"
					}
					tcStrs = append(tcStrs, fmt.Sprintf("```tool_call\n{\"name\": \"%s\", \"arguments\": %s}\n```", tc.Function.Name, argsStr))
				}
				parts = append(parts, fmt.Sprintf("[Assistant]: %s\n%s", contentStr, strings.Join(tcStrs, "\n")))
			} else {
				parts = append(parts, fmt.Sprintf("[Assistant]: %s", contentStr))
			}
		case "tool":
			parts = append(parts, fmt.Sprintf("[Tool result for %s]: %s", msg.Name, contentStr))
		default:
			if contentStr != "" {
				parts = append(parts, contentStr)
			}
		}
	}

	return strings.Join(parts, "\n\n"), nil
}

var reToolCall = regexp.MustCompile(`(?s)\x60\x60\x60tool_call\s*\n(.*?)\n\x60\x60\x60`)

// ParseToolCalls extracts tool calls from the raw string output of the model.
// The model is instructed to output tool calls in markdown blocks e.g. ```tool_call\n{...}\n```.
// This function parses those blocks, removes them from the text, and returns the cleaned text along with structured ToolCalls.
func ParseToolCalls(text string) (string, []models.OpenAIToolCall) {
	var toolCalls []models.OpenAIToolCall
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

			toolCalls = append(toolCalls, models.OpenAIToolCall{
				ID:   fmt.Sprintf("call_%s", RandHex(8)),
				Type: "function",
				Function: models.OpenAIToolCallFunction{
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
