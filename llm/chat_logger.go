package llm

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

//###NOUPLOAD###

const ChatLogFile = ".chatlog.md"
const LLMRawLogFile = ".raw_message_log.txt"

func LogStartSession(logger logging.ILogger, perpetualDir string, operation string, args ...string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	caption := fmt.Sprintf("# New Session Started at %s\n\n**Operation:** %s  \n**Args:** %s  ", now, operation, strings.Join(args, " "))
	logEntry := fmt.Sprintf("%s\n___\n", caption)
	err := utils.AppendToTextFile(filepath.Join(perpetualDir, ChatLogFile), logEntry)
	if err != nil {
		logger.Panicln("Failed to add start session log-record to chat-log", err)
	}
}

func formatMDHeadingLine(line string) string {
	if strings.HasPrefix(line, "#") {
		return "##" + line
	}
	return line
}

func formatTagsInLine(line string) string {
	// Find valid XML-style tags in the line
	tagRegex := `<[^>]+>`
	tags := regexp.MustCompile(tagRegex).FindAllString(line, -1)
	// Replace each tag with a code block
	for _, tag := range tags {
		line = strings.ReplaceAll(line, tag, fmt.Sprintf(" ```%s```", tag))
	}
	return line
}

func LogMessages(logger logging.ILogger, perpetualDir string, connector LLMConnector, messages []Message) {
	for index := range messages {
		LogMessage(logger, perpetualDir, connector, &messages[index])
	}
}

func LogMessage(logger logging.ILogger, perpetualDir string, connector LLMConnector, message *Message) {
	// Skip already logger messages
	if message.IsLogged {
		return
	}
	// Flag message as logged
	message.IsLogged = true

	var builder strings.Builder

	// Header
	switch message.Type {
	case UserRequest:
		builder.WriteString("\n## User Request\n\n")
	case SimulatedAIResponse:
		builder.WriteString("\n## Simulated AI Response\n\n")
	case RealAIResponse:
		builder.WriteString("\n## AI Response\n\n")
		builder.WriteString(fmt.Sprintf("**Provider:** %s  \n", connector.GetProvider()))
		builder.WriteString(fmt.Sprintf("**Model:** %s  \n", connector.GetModel()))
		builder.WriteString(fmt.Sprintf("**Temperature:** %f  \n", connector.GetTemperature()))
		builder.WriteString(fmt.Sprintf("**Max Tokens:** %d  \n", connector.GetMaxTokens()))
		builder.WriteString("___\n")
	}

	// Log message content
	if message.RawText != "" {
		builder.WriteString("\n````````````````````````````````````````````````````````````````````````````````text\n")
		builder.WriteString(message.RawText)
		// Ensure we have a new line at the end
		if message.RawText != "" && message.RawText[len(message.RawText)-1] != '\n' {
			builder.WriteString("\n")
		}
		builder.WriteString("````````````````````````````````````````````````````````````````````````````````\n")
	} else {
		for index, fragment := range message.Fragments {
			switch fragment.Type {
			// We are assuming that text in plain text fragments is already Markdown compatible
			case PlainTextFragment:
				// Each additional plain text fragment should have a blank line between
				if index > 0 {
					builder.WriteString("\n")
				}
				lines := strings.Split(fragment.Payload, "\n")
				for index, line := range lines {
					builder.WriteString(formatMDHeadingLine(formatTagsInLine(line)))
					if index < len(lines)-1 {
						builder.WriteString("\n")
					}
				}
				// Add new line after end of plain text block, if missing
				if fragment.Payload != "" && fragment.Payload[len(fragment.Payload)-1] != '\n' {
					builder.WriteString("\n")
				}
			case IndexFragment:
				// Each index fragment should have a blank line between it and previous text
				if index > 0 {
					builder.WriteString("\n")
				}
				builder.WriteString(formatMDHeadingLine("# File: " + fragment.Payload))
				builder.WriteString("\n")
			case FileFragment:
				// Each file fragment must have a blank line between it and previous text
				if index > 0 {
					builder.WriteString("\n")
				}
				builder.WriteString(formatMDHeadingLine("# Content of the file: " + fragment.Metadata))
				builder.WriteString("\n")
				builder.WriteString("\n")
				builder.WriteString("````````````````````````````````````````````````````````````````````````````````" + getMarkdownCodeBlockType(fragment.Metadata) + "\n")
				builder.WriteString(fragment.Payload)
				if fragment.Payload != "" && fragment.Payload[len(fragment.Payload)-1] != '\n' {
					builder.WriteString("\n")
				}
				builder.WriteString("````````````````````````````````````````````````````````````````````````````````\n")
			case MultilineTaggedFragment:
				// Each file fragment must have a blank line between it and previous text
				if index > 0 {
					builder.WriteString("\n")
				}
				var tags []string
				err := json.Unmarshal([]byte(fragment.Metadata), &tags)
				if err != nil {
					logger.Errorln("Failed to unmarshal tags metadata for tagged fragment with index:", index, "error:", err)
					continue
				}
				if len(tags) < 2 {
					logger.Errorln("It must be 2 tags in metadata for tagged fragment with index:", index)
					continue
				}
				builder.WriteString("````````````````````````````````````````````````````````````````````````````````text\n")
				builder.WriteString(tags[0])
				builder.WriteString("\n")
				builder.WriteString(fragment.Payload)
				if fragment.Payload != "" && fragment.Payload[len(fragment.Payload)-1] != '\n' {
					builder.WriteString("\n")
				}
				builder.WriteString(tags[1])
				builder.WriteString("\n")
				builder.WriteString("````````````````````````````````````````````````````````````````````````````````\n")
			case TaggedFragment:
				if index > 0 {
					builder.WriteString("\n")
				}
				var tags []string
				err := json.Unmarshal([]byte(fragment.Metadata), &tags)
				if err != nil {
					logger.Errorln("Failed to unmarshal tags metadata for tagged fragment with index:", index, "error:", err)
					continue
				}
				if len(tags) < 2 {
					logger.Errorln("It must be 2 tags in metadata for tagged fragment with index:", index)
					continue
				}
				builder.WriteString("```")
				builder.WriteString(tags[0])
				builder.WriteString(fragment.Payload)
				builder.WriteString(tags[1])
				builder.WriteString("```")
				builder.WriteString("\n")
			default:
				logger.Errorln("Invalid fragment type:", fragment.Type, "index:", index)
			}
		}
	}

	builder.WriteString("\n___\n")

	logEntry := builder.String()
	err := utils.AppendToTextFile(filepath.Join(perpetualDir, ChatLogFile), logEntry)
	if err != nil {
		logger.Errorln("Failed to add log-record to chat-log", err)
	}
}

func getMarkdownCodeBlockType(fileName string) string {
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
	case ".c", ".cpp", ".h", ".hpp":
		return "c"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
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
	case ".sh", ".bash":
		return "bash"
	case ".sql":
		return "sql"
	default:
		return "text"
	}
}

func GetSimpleRawMessageLogger(perpetualDir string) func(v ...any) {
	logFunc := func(v ...any) {
		for _, msg := range v {
			str := fmt.Sprintln(msg)
			utils.AppendToTextFile(filepath.Join(perpetualDir, LLMRawLogFile), str)
		}
	}
	return logFunc
}
