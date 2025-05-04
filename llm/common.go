package llm

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

type llmDebug struct {
	Values [][2]any
}

func (p *llmDebug) Add(key, value any) {
	kv := [...]any{key, value}
	p.Values = append(p.Values, kv)
}

func (p *llmDebug) Format() string {
	var sb strings.Builder
	for i, kv := range p.Values {
		if i > 0 {
			sb.WriteString(" ") // Add space between pairs
		}
		sb.WriteString(fmt.Sprint("[", kv[0], ":", kv[1], "]"))
	}
	return sb.String()
}

func getMarkdownCodeBlockType(filesToMdLangMappings [][]string, fileName string) string {
	for _, mapping := range filesToMdLangMappings {
		matched, err := regexp.MatchString(mapping[0], fileName)
		if err == nil && matched {
			return mapping[1]
		}
	}
	ext := filepath.Ext(fileName)
	switch strings.ToLower(ext) {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".ts":
		return "javascript"
	case ".java":
		return "java"
	case ".bat", ".cmd":
		return "bat"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp":
		return "cpp"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".pl":
		return "perl"
	case ".swift":
		return "swift"
	case ".rs":
		return "rust"
	case ".cs":
		return "csharp"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".md", ".markdown":
		return "markdown"
	case ".json":
		return "json"
	case ".yml", ".yaml":
		return "yaml"
	case ".xml":
		return "xml"
	case ".sh", ".bash", ".sh.in", ".bash.in":
		return "bash"
	case ".sql":
		return "sql"
	default:
		return "text"
	}
}

func renderMessagesToGenericAILangChainFormat(filesToMdLangMappings [][]string, messages []Message, msgPrefix, msgSuffix string) ([]llms.MessageContent, error) {
	var result []llms.MessageContent
	for i, message := range messages {
		lastMessage := (i >= len(messages)-1)
		var llmMessage llms.MessageContent
		// Convert message type
		switch message.Type {
		case UserRequest:
			llmMessage.Role = llms.ChatMessageTypeHuman
		case RealAIResponse:
			return result, errors.New("cannot process real ai response, sending such message types are not supported for now")
		case SimulatedAIResponse:
			llmMessage.Role = llms.ChatMessageTypeAI
		default:
			return result, fmt.Errorf("invalid message type: %d", message.Type)
		}
		if message.Type == SimulatedAIResponse && message.RawText != "" {
			llmMessage.Parts = []llms.ContentPart{llms.TextContent{Text: message.RawText}}
		} else {
			// Convert message content
			var builder strings.Builder
			if lastMessage {
				builder.WriteString(msgPrefix)
			}
			for index, fragment := range message.Fragments {
				switch fragment.Type {
				case PlainTextFragment:
					// Each additional plain text fragment should have a blank line between
					if index > 0 {
						builder.WriteString("\n")
					}
					builder.WriteString(fragment.Contents)
					// Add extra new line to the end of the fragment if missing
					if fragment.Contents != "" && fragment.Contents[len(fragment.Contents)-1] != '\n' {
						builder.WriteString("\n")
					}
				case IndexFragment:
					// Each index fragment should have a blank line between it and previous text
					if index > 0 {
						builder.WriteString("\n")
					}
					var tags []string
					err := json.Unmarshal([]byte(fragment.FileNameTags), &tags)
					if err != nil {
						return result, err
					}
					// Placing filenames in a such way between tags will serve as example for LLM how to deal with filenames in responses
					builder.WriteString(tags[0] + fragment.FileName + tags[1])
					builder.WriteString("\n")
				case FileFragment:
					// Each file fragment must have a blank line between it and previous text
					if index > 0 {
						builder.WriteString("\n")
					}
					var tags []string
					err := json.Unmarshal([]byte(fragment.FileNameTags), &tags)
					if err != nil {
						return result, err
					}
					// Following formatting will also show LLM how to deal with filenames and file contens in responses
					builder.WriteString(tags[0] + fragment.FileName + tags[1])
					builder.WriteString("\n")
					builder.WriteString("```" + getMarkdownCodeBlockType(filesToMdLangMappings, fragment.FileName))
					builder.WriteString("\n")
					builder.WriteString(fragment.Contents)
					if fragment.Contents != "" && fragment.Contents[len(fragment.Contents)-1] != '\n' {
						builder.WriteString("\n")
					}
					builder.WriteString("```")
					builder.WriteString("\n")
				case TaggedFragment:
					if index > 0 {
						builder.WriteString("\n")
					}
					var tags []string
					err := json.Unmarshal([]byte(fragment.FileNameTags), &tags)
					if err != nil {
						return result, err
					}
					if len(tags) < 2 {
						return result, fmt.Errorf("invalid tags count in metadata for tagged fragment with index: %d", index)
					}
					builder.WriteString(tags[0])
					builder.WriteString(fragment.Contents)
					builder.WriteString(tags[1])
					builder.WriteString("\n")
				case MultilineTaggedFragment:
					if index > 0 {
						builder.WriteString("\n")
					}
					var tags []string
					err := json.Unmarshal([]byte(fragment.FileNameTags), &tags)
					if err != nil {
						return result, err
					}
					if len(tags) < 2 {
						return result, fmt.Errorf("invalid tags count in metadata for tagged fragment with index: %d", index)
					}
					builder.WriteString(tags[0])
					builder.WriteString("\n")
					builder.WriteString(fragment.Contents)
					if fragment.Contents != "" && fragment.Contents[len(fragment.Contents)-1] != '\n' {
						builder.WriteString("\n")
					}
					builder.WriteString(tags[1])
					builder.WriteString("\n")
				default:
					return result, fmt.Errorf("invalid fragment type: %d, index: %d", fragment.Type, index)
				}
			}
			if lastMessage {
				builder.WriteString(msgSuffix)
			}
			llmMessage.Parts = []llms.ContentPart{llms.TextContent{Text: builder.String()}}
		}
		result = append(result, llmMessage)
	}
	if len(result) < 1 {
		return result, errors.New("no messages was generated")
	}
	return result, nil
}

func RenderMessagesToAIStrings(filesToMdLangMappings [][]string, messages []Message) ([]string, error) {
	messageContents, err := renderMessagesToGenericAILangChainFormat(filesToMdLangMappings, messages, "", "")
	if err != nil {
		return nil, err
	}
	var result []string
	for _, messageContent := range messageContents {
		var sb strings.Builder
		for _, part := range messageContent.Parts {
			text := ""
			switch pp := part.(type) {
			case llms.TextContent:
				text = pp.Text
			case llms.ImageURLContent:
				text = pp.URL
			case llms.BinaryContent:
				text = "<binary data>"
			case llms.ToolCall:
				text = "<tool call>"
			case llms.ToolCallResponse:
				text = "<tool call response>"
			default:
				text = "<unknown>"
			}
			sb.WriteString(text)
		}
		result = append(result, sb.String())
	}
	return result, nil
}
