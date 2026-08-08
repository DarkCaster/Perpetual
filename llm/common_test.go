package llm

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/DarkCaster/Perpetual/langchaingo/llms"
	"github.com/DarkCaster/Perpetual/utils"
)

// Tests for renderMessagesToLangChainFormat
func TestRenderMessagesToGenericLangChainFormat(t *testing.T) {
	testCases := []struct {
		name     string
		prefix   string
		suffix   string
		messages []Message
		expected []llms.MessageContent
		err      error
	}{
		{
			name:     "Empty messages",
			prefix:   "",
			suffix:   "",
			messages: []Message{},
			expected: []llms.MessageContent{},
			err:      errors.New("no messages was generated"),
		},
		{
			name:     "Empty messages with prefix and suffix",
			prefix:   "a",
			suffix:   "b",
			messages: []Message{},
			expected: []llms.MessageContent{},
			err:      errors.New("no messages was generated"),
		},
		{
			name:   "User request message",
			prefix: "",
			suffix: "",
			messages: []Message{
				NewMessage(UserRequest),
			},
			expected: []llms.MessageContent{
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: ""}}},
			},
			err: nil,
		},
		{
			name:   "User request message with prefix and suffix",
			prefix: "a",
			suffix: "b",
			messages: []Message{
				NewMessage(UserRequest),
			},
			expected: []llms.MessageContent{
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "ab"}}},
			},
			err: nil,
		},
		{
			name:   "AI response message",
			prefix: "",
			suffix: "",
			messages: []Message{
				NewMessage(SimulatedAIResponse),
			},
			expected: []llms.MessageContent{
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: ""}}},
			},
			err: nil,
		},
		{
			name:   "AI response message with prefix and suffix",
			prefix: "a",
			suffix: "b",
			messages: []Message{
				NewMessage(SimulatedAIResponse),
			},
			expected: []llms.MessageContent{
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "ab"}}},
			},
			err: nil,
		},
		{
			name:   "Real AI response with raw response",
			prefix: "",
			suffix: "",
			messages: []Message{
				SetRawResponse(NewMessage(RealAIResponse), "This is a raw response."),
			},
			expected: []llms.MessageContent{},
			err:      errors.New("cannot process real ai response, sending such message types are not supported for now"),
		},
		{
			name:   "Real AI response with raw response with prefix and suffix",
			prefix: "a",
			suffix: "b",
			messages: []Message{
				SetRawResponse(NewMessage(RealAIResponse), "This is a raw response."),
			},
			expected: []llms.MessageContent{},
			err:      errors.New("cannot process real ai response, sending such message types are not supported for now"),
		},
		{
			name:   "Multiple messages with different fragment types, prefix and suffix",
			prefix: "a",
			suffix: "b",
			messages: []Message{
				AddPlainTextFragment(NewMessage(UserRequest), "Hello"),
				AddPlainTextFragment(AddPlainTextFragment(NewMessage(UserRequest), "Hello"), "World"),
				AddIndexFragment(AddPlainTextFragment(NewMessage(SimulatedAIResponse), "World"), "main.go", utils.TagPair{Left: "<filename>", Right: "</filename>"}),
				AddFileFragment(AddPlainTextFragment(NewMessage(SimulatedAIResponse), "This is a file content."), "file.go", "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n", utils.TagPair{Left: "<filename>", Right: "</filename>"}),
				AddFileFragment(AddPlainTextFragment(NewMessage(SimulatedAIResponse), "This is a file content."), "file.go", "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}", utils.TagPair{Left: "<filename>", Right: "</filename>"}),
				AddFileFragment(NewMessage(SimulatedAIResponse), "file.go", "file", utils.TagPair{Left: "<filename>", Right: "</filename>"}),
				AddFileFragment(NewMessage(SimulatedAIResponse), "file.go", "\nfile\n", utils.TagPair{Left: "<filename>", Right: "</filename>"}),
				AddFileFragment(NewMessage(SimulatedAIResponse), "file.go", "\n\nfile\n\n", utils.TagPair{Left: "<filename>", Right: "</filename>"}),
				AddFileFragment(NewMessage(SimulatedAIResponse), "file.go", "", utils.TagPair{Left: "<filename>", Right: "</filename>"}),
				AddTaggedFragment(NewMessage(UserRequest), "Tagged text", utils.TagPair{Left: "[", Right: "]"}),
				AddTaggedFragment(AddTaggedFragment(AddPlainTextFragment(NewMessage(UserRequest), "Hello"), "Tagged text", utils.TagPair{Left: "[", Right: "]"}), "Tagged text", utils.TagPair{Left: "<tag>", Right: "</tag>"}),
				SetRawResponse(NewMessage(SimulatedAIResponse), "this is raw response"),
				AddMultilineTaggedFragment(NewMessage(SimulatedAIResponse), "Hello", utils.TagPair{Left: "[", Right: "]"}),
				AddMultilineTaggedFragment(NewMessage(SimulatedAIResponse), "\n\nHello\n", utils.TagPair{Left: "[", Right: "]"}),
				AddMultilineTaggedFragment(NewMessage(SimulatedAIResponse), "", utils.TagPair{Left: "[", Right: "]"}),
				AddMultilineTaggedFragment(NewMessage(SimulatedAIResponse), "\n", utils.TagPair{Left: "[", Right: "]"}),
				AddMultilineTaggedFragment(NewMessage(SimulatedAIResponse), "\n\n", utils.TagPair{Left: "[", Right: "]"}),
				AddMultilineTaggedFragment(NewMessage(SimulatedAIResponse), "\n\n", utils.TagPair{Left: "[", Right: "]"}),
			},
			expected: []llms.MessageContent{
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "Hello\n"}}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "Hello\n\nWorld\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "World\n\n<filename>main.go</filename>\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "This is a file content.\n\n<filename>file.go</filename>\n```go\npackage main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n```\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "This is a file content.\n\n<filename>file.go</filename>\n```go\npackage main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n```\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "<filename>file.go</filename>\n```go\nfile\n```\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "<filename>file.go</filename>\n```go\n\nfile\n```\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "<filename>file.go</filename>\n```go\n\n\nfile\n\n```\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "<filename>file.go</filename>\n```go\n```\n"}}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "[Tagged text]\n"}}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "Hello\n\n[Tagged text]\n\n<tag>Tagged text</tag>\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "this is raw response"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "[\nHello\n]\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "[\n\n\nHello\n]\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "[\n]\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "[\n\n]\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "[\n\n\n]\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "a[\n\n\n]\nb"}}},
			},
			err: nil,
		},

		{
			name:   "Typical message queue with prefix and suffix",
			prefix: "a",
			suffix: "b",
			messages: []Message{
				AddPlainTextFragment(NewMessage(UserRequest), "Hello"),
				AddMultilineTaggedFragment(NewMessage(SimulatedAIResponse), "Hello", utils.TagPair{Left: "[", Right: "]"}),
				AddPlainTextFragment(NewMessage(UserRequest), "Test"),
			},
			expected: []llms.MessageContent{
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "Hello\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "[\nHello\n]\n"}}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "aTest\nb"}}},
			},
			err: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := renderMessagesToGenericAILangChainFormat(nil, tc.messages, tc.prefix, tc.suffix)
			if err != nil && tc.err == nil || err == nil && tc.err != nil || (err != nil && tc.err != nil && err.Error() != tc.err.Error()) {
				t.Errorf("Unexpected error: got %v, want %v", err, tc.err)
			}
			if !equalMessageContents(result, tc.expected) {
				t.Errorf("Unexpected result: got %v, want %v", result, tc.expected)
			}
		})
	}
}

func createFileToMdMappings(data [][]string) utils.TextMatcher[string] {
	byteData, _ := json.Marshal(data)
	var decodedData any
	json.Unmarshal(byteData, &decodedData)
	result, _ := utils.NewRxMatcher[string](1, decodedData)
	return result
}

func TestRenderMessagesWithMappings(t *testing.T) {
	testCases := []struct {
		name     string
		mappings utils.TextMatcher[string]
		messages []Message
		expected []llms.MessageContent
		err      error
	}{
		{
			name:     "Empty messages",
			mappings: createFileToMdMappings([][]string{{"(?i)^.*\\.(frm|cls|bas)$", "vb"}}),
			messages: []Message{},
			expected: []llms.MessageContent{},
			err:      errors.New("no messages was generated"),
		},
		{
			name:     "User request message",
			mappings: createFileToMdMappings([][]string{{"(?i)^.*\\.(frm|cls|bas)$", "vb"}}),
			messages: []Message{
				NewMessage(UserRequest),
			},
			expected: []llms.MessageContent{
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: ""}}},
			},
			err: nil,
		},
		{
			name:     "AI response message",
			mappings: createFileToMdMappings([][]string{{"(?i)^.*\\.(frm|cls|bas)$", "vb"}}),
			messages: []Message{
				NewMessage(SimulatedAIResponse),
			},
			expected: []llms.MessageContent{
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: ""}}},
			},
			err: nil,
		},
		{
			name:     "Real AI response with raw response",
			mappings: createFileToMdMappings([][]string{{"(?i)^.*\\.(frm|cls|bas)$", "vb"}}),
			messages: []Message{
				SetRawResponse(NewMessage(RealAIResponse), "This is a raw response."),
			},
			expected: []llms.MessageContent{},
			err:      errors.New("cannot process real ai response, sending such message types are not supported for now"),
		},
		{
			name:     "Multiple messages with different fragments",
			mappings: createFileToMdMappings([][]string{{"(?i)^.*\\.(frm|cls|bas)$", "vb"}}),
			messages: []Message{
				AddFileFragment(AddPlainTextFragment(NewMessage(SimulatedAIResponse), "This is a file content."), "file.bas", "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n", utils.TagPair{Left: "<filename>", Right: "</filename>"}),
				AddFileFragment(AddPlainTextFragment(NewMessage(SimulatedAIResponse), "This is a file content."), "File.BAS", "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}", utils.TagPair{Left: "<filename>", Right: "</filename>"}),
				AddFileFragment(NewMessage(SimulatedAIResponse), "file.clS", "file", utils.TagPair{Left: "<filename>", Right: "</filename>"}),
				AddFileFragment(NewMessage(SimulatedAIResponse), "file.FRM", "\nfile\n", utils.TagPair{Left: "<filename>", Right: "</filename>"}),
				AddFileFragment(NewMessage(SimulatedAIResponse), "file.bAs", "\n\nfile\n\n", utils.TagPair{Left: "<filename>", Right: "</filename>"}),
				AddFileFragment(NewMessage(SimulatedAIResponse), "file.xxx", "", utils.TagPair{Left: "<filename>", Right: "</filename>"}),
			},
			expected: []llms.MessageContent{
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "This is a file content.\n\n<filename>file.bas</filename>\n```vb\npackage main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n```\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "This is a file content.\n\n<filename>File.BAS</filename>\n```vb\npackage main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n```\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "<filename>file.clS</filename>\n```vb\nfile\n```\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "<filename>file.FRM</filename>\n```vb\n\nfile\n```\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "<filename>file.bAs</filename>\n```vb\n\n\nfile\n\n```\n"}}},
				{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: "<filename>file.xxx</filename>\n```text\n```\n"}}},
			},
			err: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := renderMessagesToGenericAILangChainFormat(tc.mappings, tc.messages, "", "")
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

func TestProviderMaxRequestSizeConfiguration(t *testing.T) {
	const operation = "REQUEST_SIZE_TEST"

	type providerTest struct {
		name       string
		prefix     string
		subprofile string
		setup      func(t *testing.T, prefix string)
		create     func(subprofile string) (int, error)
	}

	providers := []providerTest{
		{
			name:       "Anthropic",
			prefix:     "ANTHROPIC99101",
			subprofile: "99101",
			setup: func(t *testing.T, prefix string) {
				t.Helper()
				t.Setenv(prefix+"_API_KEY", "test-key")
				t.Setenv(prefix+"_MODEL", "test-model")
				t.Setenv(prefix+"_MAX_TOKENS", "1024")
			},
			create: func(subprofile string) (int, error) {
				connector, err := NewAnthropicLLMConnectorFromEnv(
					subprofile,
					operation,
					"system prompt",
					nil,
					nil,
				)
				if err != nil {
					return 0, err
				}
				return connector.MaxRequestSize, nil
			},
		},
		{
			name:       "OpenAI",
			prefix:     "OPENAI99102",
			subprofile: "99102",
			setup: func(t *testing.T, prefix string) {
				t.Helper()
				t.Setenv(prefix+"_API_KEY", "test-key")
				t.Setenv(prefix+"_MODEL", "test-model")
			},
			create: func(subprofile string) (int, error) {
				connector, err := NewOpenAILLMConnectorFromEnv(
					subprofile,
					operation,
					"system prompt",
					"acknowledgement",
					nil,
					nil,
				)
				if err != nil {
					return 0, err
				}
				return connector.MaxRequestSize, nil
			},
		},
		{
			name:       "Ollama",
			prefix:     "OLLAMA99103",
			subprofile: "99103",
			setup: func(t *testing.T, prefix string) {
				t.Helper()
				t.Setenv(prefix+"_MODEL", "test-model")
				t.Setenv(prefix+"_MAX_TOKENS", "1024")
			},
			create: func(subprofile string) (int, error) {
				connector, err := NewOllamaLLMConnectorFromEnv(
					subprofile,
					operation,
					"system prompt",
					"acknowledgement",
					nil,
					nil,
				)
				if err != nil {
					return 0, err
				}
				return connector.MaxRequestSize, nil
			},
		},
		{
			name:       "Generic",
			prefix:     "GENERIC99104",
			subprofile: "99104",
			setup: func(t *testing.T, prefix string) {
				t.Helper()
				t.Setenv(prefix+"_MODEL", "test-model")
				t.Setenv(prefix+"_BASE_URL", "https://example.invalid/v1")
			},
			create: func(subprofile string) (int, error) {
				connector, err := NewGenericLLMConnectorFromEnv(
					subprofile,
					operation,
					"system prompt",
					"acknowledgement",
					nil,
					nil,
				)
				if err != nil {
					return 0, err
				}
				return connector.MaxRequestSize, nil
			},
		},
	}

	testCases := []struct {
		name           string
		providerValue  string
		operationValue string
		setProvider    bool
		setOperation   bool
		expected       int
		wantError      bool
	}{
		{
			name:           "operation-specific value takes precedence",
			providerValue:  "1000",
			operationValue: "500",
			setProvider:    true,
			setOperation:   true,
			expected:       500,
		},
		{
			name:          "provider-wide value is used as fallback",
			providerValue: "1000",
			setProvider:   true,
			expected:      1000,
		},
		{
			name:           "operation-specific zero disables provider-wide limit",
			providerValue:  "1000",
			operationValue: "0",
			setProvider:    true,
			setOperation:   true,
			expected:       0,
		},
		{
			name:     "unset limit is disabled",
			expected: 0,
		},
		{
			name:           "invalid operation-specific value disables unset limit",
			operationValue: "invalid",
			setOperation:   true,
			expected:       0,
		},
		{
			name:           "invalid operation-specific value falls back to provider-wide value",
			providerValue:  "1000",
			operationValue: "invalid",
			setProvider:    true,
			setOperation:   true,
			expected:       1000,
		},
		{
			name:          "invalid provider-wide value disables limit",
			providerValue: "invalid",
			setProvider:   true,
			expected:      0,
		},
		{
			name:           "negative operation-specific value is rejected",
			operationValue: "-1",
			setOperation:   true,
			wantError:      true,
		},
		{
			name:          "negative provider-wide value is rejected",
			providerValue: "-1",
			setProvider:   true,
			wantError:     true,
		},
	}

	for _, provider := range providers {
		for _, tc := range testCases {
			t.Run(provider.name+"/"+tc.name, func(t *testing.T) {
				provider.setup(t, provider.prefix)

				if tc.setProvider {
					t.Setenv(provider.prefix+"_MAX_REQUEST_SIZE", tc.providerValue)
				}
				if tc.setOperation {
					t.Setenv(
						provider.prefix+"_MAX_REQUEST_SIZE_OP_"+operation,
						tc.operationValue,
					)
				}

				actual, err := provider.create(provider.subprofile)
				if tc.wantError {
					if err == nil {
						t.Fatalf("connector creation returned nil error, want error")
					}
					return
				}
				if err != nil {
					t.Fatalf("connector creation returned error: %v", err)
				}
				if actual != tc.expected {
					t.Fatalf("MaxRequestSize = %d, want %d", actual, tc.expected)
				}
			})
		}
	}
}
