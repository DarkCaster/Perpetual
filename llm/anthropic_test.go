package llm

import (
	"errors"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

// Tests for renderMessagesToLangChainFormat
func TestRenderMessagesToAnthropicLangChainFormat(t *testing.T) {
	testCases := []struct {
		name     string
		messages []Message
		expected []llms.MessageContent
		err      error
	}{
		{
			name:     "Empty messages",
			messages: []Message{},
			expected: []llms.MessageContent{},
			err:      errors.New("no messages was generated"),
		},
		{
			name: "User request message",
			messages: []Message{
				NewMessage(UserRequest),
			},
			expected: []llms.MessageContent{
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: ""}}},
			},
			err: nil,
		},
		{
			name: "AI response message",
			messages: []Message{
				NewMessage(SimulatedAIResponse),
			},
			expected: []llms.MessageContent{
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: ""}}},
			},
			err: nil,
		},
		{
			name: "Real AI response with raw response",
			messages: []Message{
				SetRawResponse(NewMessage(RealAIResponse), "This is a raw response."),
			},
			expected: []llms.MessageContent{},
			err:      errors.New("cannot process real ai response, sending such message types are not supported for now"),
		},
		{
			name: "Multiple messages with different fragments",
			messages: []Message{
				AddPlainTextFragment(NewMessage(UserRequest), "Hello"),
				AddPlainTextFragment(AddPlainTextFragment(NewMessage(UserRequest), "Hello"), "World"),
				AddIndexFragment(AddPlainTextFragment(NewMessage(SimulatedAIResponse), "World"), "main.go"),
				AddFileFragment(AddPlainTextFragment(NewMessage(SimulatedAIResponse), "This is a file content."), "file.go", "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n"),
				AddFileFragment(AddPlainTextFragment(NewMessage(SimulatedAIResponse), "This is a file content."), "file.go", "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}"),
				AddFileFragment(NewMessage(SimulatedAIResponse), "file.go", "file"),
				AddFileFragment(NewMessage(SimulatedAIResponse), "file.go", "\nfile\n"),
				AddFileFragment(NewMessage(SimulatedAIResponse), "file.go", "\n\nfile\n\n"),
				AddFileFragment(NewMessage(SimulatedAIResponse), "file.go", ""),
				AddTaggedFragment(NewMessage(UserRequest), "Tagged text", []string{"[", "]"}),
				AddTaggedFragment(AddTaggedFragment(AddPlainTextFragment(NewMessage(UserRequest), "Hello"), "Tagged text", []string{"[", "]"}), "Tagged text", []string{"<tag>", "</tag>"}),
				SetRawResponse(NewMessage(SimulatedAIResponse), "this is raw response"),
				AddMultilineTaggedFragment(NewMessage(SimulatedAIResponse), "Hello", []string{"[", "]"}),
				AddMultilineTaggedFragment(NewMessage(SimulatedAIResponse), "\n\nHello\n", []string{"[", "]"}),
				AddMultilineTaggedFragment(NewMessage(SimulatedAIResponse), "", []string{"[", "]"}),
				AddMultilineTaggedFragment(NewMessage(SimulatedAIResponse), "\n", []string{"[", "]"}),
				AddMultilineTaggedFragment(NewMessage(SimulatedAIResponse), "\n\n", []string{"[", "]"}),
			},
			expected: []llms.MessageContent{
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "Hello\n"}}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "Hello\n\nWorld\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "World\n\n# File: main.go\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "This is a file content.\n\n<content filename=\"file.go\">\npackage main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n</content>\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "This is a file content.\n\n<content filename=\"file.go\">\npackage main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n</content>\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "<content filename=\"file.go\">\nfile\n</content>\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "<content filename=\"file.go\">\n\nfile\n</content>\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "<content filename=\"file.go\">\n\n\nfile\n\n</content>\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "<content filename=\"file.go\">\n</content>\n"}}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "[Tagged text]\n"}}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "Hello\n\n[Tagged text]\n\n<tag>Tagged text</tag>\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "this is raw response"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "[\nHello\n]\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "[\n\n\nHello\n]\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "[\n]\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "[\n\n]\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "[\n\n\n]\n"}}},
			},
			err: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := renderMessagesToAnthropicLangChainFormat(tc.messages)
			if err != nil && tc.err == nil || err == nil && tc.err != nil || (err != nil && tc.err != nil && err.Error() != tc.err.Error()) {
				t.Errorf("Unexpected error: got %v, want %v", err, tc.err)
			}
			if !equalMessageContents(result, tc.expected) {
				t.Errorf("Unexpected result: got %v, want %v", result, tc.expected)
			}
		})
	}
}

func equalMessageContents(a, b []llms.MessageContent) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Role != b[i].Role || len(a[i].Parts) != len(b[i].Parts) {
			return false
		}
		for j := range a[i].Parts {
			aPart := a[i].Parts[j].(llms.TextContent)
			bPart := b[i].Parts[j].(llms.TextContent)
			if aPart.Text != bPart.Text {
				return false
			}
		}
	}
	return true
}
