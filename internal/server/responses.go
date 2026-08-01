package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ikhsan3adi/gemini-web2api/internal/format"
	"github.com/ikhsan3adi/gemini-web2api/internal/models"
)

func (a *App) handleResponses(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON"}})
		return
	}

	var req map[string]any
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON"}})
		return
	}

	modelStr, _ := req["model"].(string)
	if modelStr == "" {
		modelStr = a.Cfg.DefaultModel
	}

	resolved, err := models.Resolve(modelStr, a.Cfg.DefaultModel)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": err.Error()}})
		return
	}

	inputRaw := req["input"]
	instructions, _ := req["instructions"].(string)

	messages, err := format.ResponsesInputToMessages(inputRaw, instructions)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid input structure"}})
		return
	}

	var tools []map[string]any
	if toolsRaw, ok := req["tools"].([]any); ok {
		for _, t := range toolsRaw {
			if tm, ok := t.(map[string]any); ok {
				tools = append(tools, tm)
			}
		}
	}

	toolChoice := req["tool_choice"]
	if toolChoice == nil {
		toolChoice = "auto"
	}

	prompt, err := format.MessagesToPrompt(messages, tools, toolChoice)
	if err != nil || strings.TrimSpace(prompt) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "empty input"}})
		return
	}

	text, err := a.Gem.Generate(prompt, resolved.Mode, resolved.Think, nil, resolved.Extra)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": map[string]any{"message": fmt.Sprintf("upstream error: %v", err)}})
		return
	}

	strChoice, isStr := toolChoice.(string)
	isToolNone := isStr && strChoice == "none"

	var toolCalls []format.ToolCall
	if len(tools) > 0 && text != "" && !isToolNone {
		text, toolCalls = format.ParseToolCalls(text)
	}

	rid := fmt.Sprintf("resp_%s", randHex(16))
	mid := fmt.Sprintf("msg_%s", randHex(12))

	outputItems := format.BuildResponseOutput(text, toolCalls, mid)

	stream, _ := req["stream"].(bool)
	promptTokens := len(prompt) / 4
	outputTokens := len(text) / 4

	if stream {
		if !startSSE(w) {
			return
		}

		evCreated := map[string]any{
			"type": "response.created",
			"response": map[string]any{
				"id":     rid,
				"object": "response",
				"status": "in_progress",
				"model":  resolved.Name,
				"output": []any{},
			},
		}
		_ = writeSSEEvent(w, "response.created", evCreated)

		for _, item := range outputItems {
			iType, _ := item["type"].(string)
			if iType == "function_call" {
				evDone := map[string]any{
					"type":      "response.function_call_arguments.done",
					"item_id":   item["id"],
					"call_id":   item["call_id"],
					"name":      item["name"],
					"arguments": item["arguments"],
				}
				_ = writeSSEEvent(w, "response.function_call_arguments.done", evDone)
			} else if iType == "message" {
				if content, ok := item["content"].([]map[string]any); ok {
					for ci, cp := range content {
						evDone := map[string]any{
							"type":          "response.output_text.done",
							"item_id":       item["id"],
							"content_index": ci,
							"text":          cp["text"],
						}
						_ = writeSSEEvent(w, "response.output_text.done", evDone)
					}
				}
			}
		}

		respObj := map[string]any{
			"id":     rid,
			"object": "response",
			"status": "completed",
			"model":  resolved.Name,
			"output": outputItems,
			"usage": map[string]any{
				"input_tokens":  promptTokens,
				"output_tokens": outputTokens,
				"total_tokens":  promptTokens + outputTokens,
			},
		}
		_ = writeSSEEvent(w, "response.completed", map[string]any{
			"type":     "response.completed",
			"response": respObj,
		})
	} else {
		writeJSON(w, http.StatusOK, map[string]any{
			"id":         rid,
			"object":     "response",
			"created_at": time.Now().Unix(),
			"status":     "completed",
			"model":      resolved.Name,
			"output":     outputItems,
			"usage": map[string]any{
				"input_tokens":  promptTokens,
				"output_tokens": outputTokens,
				"total_tokens":  promptTokens + outputTokens,
			},
		})
	}
}
