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

func (a *App) handleChat(w http.ResponseWriter, r *http.Request) {
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

	var messages []map[string]any
	if msgsRaw, ok := req["messages"].([]any); ok {
		for _, m := range msgsRaw {
			if mm, ok := m.(map[string]any); ok {
				messages = append(messages, mm)
			}
		}
	}

	prompt, err := format.MessagesToPrompt(messages, tools, toolChoice)
	if err != nil || strings.TrimSpace(prompt) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "empty prompt"}})
		return
	}

	stream, _ := req["stream"].(bool)
	cid := fmt.Sprintf("chatcmpl-%s", randHex(12))

	strChoice, isStr := toolChoice.(string)
	isToolNone := isStr && strChoice == "none"

	if stream && (len(tools) == 0 || isToolNone) {
		if !startSSE(w) {
			return
		}

		emitErr := a.Gem.GenerateStream(prompt, resolved.Mode, resolved.Think, nil, resolved.Extra, func(delta string) error {
			chunk := map[string]any{
				"id":      cid,
				"object":  "chat.completion.chunk",
				"created": time.Now().Unix(),
				"model":   resolved.Name,
				"choices": []map[string]any{
					{
						"index": 0,
						"delta": map[string]any{
							"content": delta,
						},
						"finish_reason": nil,
					},
				},
			}
			return writeSSEData(w, chunk)
		})

		if emitErr == nil {
			endChunk := map[string]any{
				"id":      cid,
				"object":  "chat.completion.chunk",
				"created": time.Now().Unix(),
				"model":   resolved.Name,
				"choices": []map[string]any{
					{
						"index":         0,
						"delta":         map[string]any{},
						"finish_reason": "stop",
					},
				},
			}
			_ = writeSSEData(w, endChunk)
			_ = writeSSEDone(w)
		} else {
			a.Logf("Chat stream error: %v", emitErr)
		}
		return
	}

	text, err := a.Gem.Generate(prompt, resolved.Mode, resolved.Think, nil, resolved.Extra)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": map[string]any{"message": fmt.Sprintf("upstream error: %v", err)}})
		return
	}

	var toolCalls []format.ToolCall
	if len(tools) > 0 && text != "" && !isToolNone {
		text, toolCalls = format.ParseToolCalls(text)
	}

	msg := map[string]any{
		"role":    "assistant",
		"content": text,
	}
	if text == "" {
		msg["content"] = nil
	}

	if len(toolCalls) > 0 {
		var tcList []map[string]any
		for _, tc := range toolCalls {
			tcList = append(tcList, map[string]any{
				"id":   tc.ID,
				"type": tc.Type,
				"function": map[string]any{
					"name":      tc.Function.Name,
					"arguments": tc.Function.Arguments,
				},
			})
		}
		msg["tool_calls"] = tcList
	}

	finish := "stop"
	if len(toolCalls) > 0 {
		finish = "tool_calls"
	}

	if stream {
		if !startSSE(w) {
			return
		}
		chunk := map[string]any{
			"id":      cid,
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   resolved.Name,
			"choices": []map[string]any{
				{
					"index":         0,
					"delta":         msg,
					"finish_reason": finish,
				},
			},
		}
		_ = writeSSEData(w, chunk)
		_ = writeSSEDone(w)
	} else {
		promptTokens := len(prompt) / 4
		completionTokens := len(text) / 4
		writeJSON(w, http.StatusOK, map[string]any{
			"id":      cid,
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"model":   resolved.Name,
			"choices": []map[string]any{
				{
					"index":         0,
					"message":       msg,
					"finish_reason": finish,
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     promptTokens,
				"completion_tokens": completionTokens,
				"total_tokens":      promptTokens + completionTokens,
			},
		})
	}
}
