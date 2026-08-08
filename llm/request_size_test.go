package llm

import (
	"strings"
	"testing"

	"github.com/DarkCaster/Perpetual/langchaingo/llms"
)

func TestCalculateRequestSize(t *testing.T) {
	testCases := []struct {
		name     string
		messages []llms.MessageContent
		expected int
	}{
		{
			name:     "empty message list",
			messages: nil,
			expected: 0,
		},
		{
			name: "ASCII text",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "hello"},
					},
				},
			},
			expected: 5,
		},
		{
			name: "multibyte UTF-8 text is counted as runes",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "hé🙂世界"},
					},
				},
			},
			expected: 5,
		},
		{
			name: "system and conversation text is aggregated",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "system"},
					},
				},
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "user"},
						llms.TextContent{Text: "🙂"},
					},
				},
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "answer"},
					},
				},
			},
			expected: 17,
		},
		{
			name: "non-text content is excluded",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "text"},
						llms.ImageURLContent{URL: "https://example.invalid/image.png"},
						llms.BinaryContent{},
					},
				},
			},
			expected: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := calculateRequestSize(tc.messages)
			if actual != tc.expected {
				t.Fatalf("calculateRequestSize() = %d, want %d", actual, tc.expected)
			}
		})
	}
}

func TestCalculateRequestSizeFromRenderedMessages(t *testing.T) {
	sourceMessages := []Message{
		AddPlainTextFragment(NewMessage(UserRequest), "hé🙂"),
		SetRawResponse(NewMessage(SimulatedAIResponse), "答"),
	}

	rendered, _, err := renderMessagesToGenericAILangChainFormat(
		nil,
		sourceMessages,
		"",
		"",
	)
	if err != nil {
		t.Fatalf("renderMessagesToGenericAILangChainFormat() returned error: %v", err)
	}

	messages := append(
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "sys"},
				},
			},
		},
		rendered...,
	)

	// "sys" = 3, rendered "hé🙂\n" = 4, raw "答" = 1.
	const expected = 8
	if actual := calculateRequestSize(messages); actual != expected {
		t.Fatalf("calculateRequestSize() = %d, want %d", actual, expected)
	}
}

func TestValidateRequestSize(t *testing.T) {
	testCases := []struct {
		name        string
		requestSize int
		limit       int
		wantError   bool
	}{
		{
			name:        "disabled limit",
			requestSize: 1000,
			limit:       0,
			wantError:   false,
		},
		{
			name:        "below limit",
			requestSize: 4,
			limit:       5,
			wantError:   false,
		},
		{
			name:        "exact limit is rejected",
			requestSize: 5,
			limit:       5,
			wantError:   true,
		},
		{
			name:        "over limit is rejected",
			requestSize: 6,
			limit:       5,
			wantError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRequestSize(tc.requestSize, tc.limit)

			if tc.wantError && err == nil {
				t.Fatalf(
					"validateRequestSize(%d, %d) returned nil, want error",
					tc.requestSize,
					tc.limit,
				)
			}
			if !tc.wantError && err != nil {
				t.Fatalf(
					"validateRequestSize(%d, %d) returned error %v, want nil",
					tc.requestSize,
					tc.limit,
					err,
				)
			}
		})
	}
}

func TestValidateRequestSizeErrorIncludesSizeAndLimit(t *testing.T) {
	err := validateRequestSize(12, 10)
	if err == nil {
		t.Fatal("validateRequestSize() returned nil, want error")
	}

	message := err.Error()
	if !strings.Contains(message, "12") {
		t.Errorf("error %q does not contain request size", message)
	}
	if !strings.Contains(message, "10") {
		t.Errorf("error %q does not contain request size limit", message)
	}
}
