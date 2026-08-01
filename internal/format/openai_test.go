package format

import (
	"strings"
	"testing"

	"github.com/ikhsan3adi/gemini-web2api/internal/models"
)

func TestParseToolCalls(t *testing.T) {
	input := "Here is a call:\n```tool_call\n{\"name\": \"get_weather\", \"arguments\": {\"city\": \"Jakarta\"}}\n```\nDone."
	clean, calls := ParseToolCalls(input)

	if len(calls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(calls))
	}

	if calls[0].Function.Name != "get_weather" {
		t.Errorf("Expected name get_weather, got %s", calls[0].Function.Name)
	}

	if !strings.Contains(clean, "Here is a call:") || !strings.Contains(clean, "Done.") {
		t.Errorf("Unexpected clean text: %q", clean)
	}
}

func TestMessagesToPrompt(t *testing.T) {
	req := models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{
			{Role: "system", Content: "Be helpful."},
			{Role: "user", Content: "Hello!"},
		},
		ToolChoice: "auto",
	}

	prompt, err := MessagesToPrompt(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(prompt, "[System instruction]: Be helpful.") {
		t.Errorf("Prompt missing system instruction: %q", prompt)
	}
	if !strings.Contains(prompt, "Hello!") {
		t.Errorf("Prompt missing user message: %q", prompt)
	}
}

func TestResponsesInputToMessagesMultiPartText(t *testing.T) {
	input := []any{
		map[string]any{
			"role": "user",
			"content": []any{
				map[string]any{"type": "text", "text": "Hello"},
				map[string]any{"type": "text", "text": "world"},
			},
		},
	}

	messages, err := ResponsesInputToMessages(input, "System instructions")
	if err != nil {
		t.Fatalf("ResponsesInputToMessages error: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages (system + user), got %d", len(messages))
	}

	userContent, _ := messages[1]["content"].(string)
	expected := "Hello world"
	if userContent != expected {
		t.Errorf("Got content %q, want %q", userContent, expected)
	}
}
