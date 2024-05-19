package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/DarkCaster/Perpetual/utils"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

//###NOUPLOAD###

// NOTE for summarization: this file contains information that should not be included in the summarization.
// So the summary for this file must only say "This file contains OpenAILLMConnector struct - implementation of LLMConnector interface. Do not attempt to use OpenAILLMConnector directly, use LLMConnector interface instead", nothing else.

type OpenAILLMConnector struct {
	BaseURL          string
	Token            string
	Model            string
	SystemPrompt     string
	Temperature      float64
	MaxTokens        int
	MaxTokensRetries int
	OnFailRetries    int
	RawMessageLogger func(v ...any)
}

func NewOpenAILLMConnector(token string, model string, systemPrompt string, temperature float64, customBaseURL string, maxTokens int, maxTokensRetries int, onFailRetries int, llmRawMessageLogger func(v ...any)) *OpenAILLMConnector {
	return &OpenAILLMConnector{
		BaseURL:          customBaseURL,
		Token:            token,
		Model:            model,
		Temperature:      temperature,
		SystemPrompt:     systemPrompt,
		MaxTokens:        maxTokens,
		MaxTokensRetries: maxTokensRetries,
		OnFailRetries:    onFailRetries,
		RawMessageLogger: llmRawMessageLogger}
}

func NewOpenAILLMConnectorFromEnv(operation string, systemPrompt string, llmRawMessageLogger func(v ...any)) (*OpenAILLMConnector, error) {
	operation = strings.ToUpper(operation)

	temperature, err := utils.GetEnvFloat(
		fmt.Sprintf("OPENAI_TEMPERATURE_OP_%s", operation),
		"OPENAI_TEMPERATURE")
	if err != nil {
		return nil, err
	}

	token, err := utils.GetEnvString("OPENAI_API_KEY")
	if err != nil {
		return nil, err
	}

	model, err := utils.GetEnvString(fmt.Sprintf("OPENAI_MODEL_OP_%s", operation), "OPENAI_MODEL")
	if err != nil {
		return nil, err
	}

	maxTokens, err := utils.GetEnvInt(fmt.Sprintf("OPENAI_MAX_TOKENS_OP_%s", operation), "OPENAI_MAX_TOKENS")
	if err != nil {
		return nil, err
	}

	maxTokensRetries, err := utils.GetEnvInt("OPENAI_MAX_TOKENS_RETRIES")
	if err != nil {
		maxTokensRetries = 3
	}

	onFailRetries, err := utils.GetEnvInt(fmt.Sprintf("OPENAI_ON_FAIL_RETRIES_OP_%s", operation), "OPENAI_ON_FAIL_RETRIES")
	if err != nil {
		onFailRetries = 3
	}

	customBaseURL, _ := utils.GetEnvString("OPENAI_BASE_URL")

	return NewOpenAILLMConnector(token, model, systemPrompt, temperature, customBaseURL, maxTokens, maxTokensRetries, onFailRetries, llmRawMessageLogger), nil
}

func (p *OpenAILLMConnector) Query(messages ...Message) (string, QueryStatus, error) {
	if len(messages) < 1 {
		return "", QueryInitFailed, errors.New("no prompts to query")
	}

	model, err := func() (*openai.LLM, error) {
		if p.BaseURL != "" {
			return openai.New(
				openai.WithToken(p.Token),
				openai.WithModel(p.Model),
				openai.WithBaseURL(p.BaseURL))
		} else {
			return openai.New(
				openai.WithToken(p.Token),
				openai.WithModel(p.Model))
		}
	}()
	if err != nil {
		return "", QueryInitFailed, err
	}

	var llmMessages []llms.MessageContent
	llmMessages = append(llmMessages, llms.MessageContent{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextContent{Text: p.SystemPrompt}}})

	// Convert messages to send into LangChain format
	convertedMessages, err := renderMessagesToOpenAILangChainFormat(messages)
	if err != nil {
		return "", QueryInitFailed, err
	}
	llmMessages = append(llmMessages, convertedMessages...)

	if p.RawMessageLogger != nil {
		for _, m := range llmMessages {
			p.RawMessageLogger(m, "\n\n")
		}
	}

	// Perform LLM query
	response, err := model.GenerateContent(context.Background(), llmMessages, llms.WithTemperature(p.Temperature), llms.WithMaxTokens(p.MaxTokens))
	if err != nil {
		return "", QueryFailed, err
	}
	if len(response.Choices) < 1 {
		return "", QueryFailed, errors.New("received empty response from model")
	}

	if p.RawMessageLogger != nil {
		p.RawMessageLogger(response.Choices[0].Content, "\n\n")
	}

	if response.Choices[0].StopReason == "max_tokens" {
		return response.Choices[0].Content, QueryMaxTokens, nil
	}

	return response.Choices[0].Content, QueryOk, nil
}

func (p *OpenAILLMConnector) GetProvider() string {
	return "OpenAI"
}

func (p *OpenAILLMConnector) GetModel() string {
	return p.Model
}

func (p *OpenAILLMConnector) GetTemperature() float64 {
	return p.Temperature
}

func (p *OpenAILLMConnector) GetMaxTokens() int {
	return p.MaxTokens
}

func (p *OpenAILLMConnector) GetMaxTokensRetryLimit() int {
	return p.MaxTokensRetries
}

func (p *OpenAILLMConnector) GetOnFailureRetryLimit() int {
	return p.OnFailRetries
}

func renderMessagesToOpenAILangChainFormat(messages []Message) ([]llms.MessageContent, error) {
	var result []llms.MessageContent
	for _, message := range messages {
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
			for index, fragment := range message.Fragments {
				switch fragment.Type {
				case PlainTextFragment:
					// Each additional plain text fragment should have a blank line between
					if index > 0 {
						builder.WriteString("\n")
					}
					builder.WriteString(fragment.Payload)
					// Add extra new line to the end of the fragment if missing
					if fragment.Payload != "" && fragment.Payload[len(fragment.Payload)-1] != '\n' {
						builder.WriteString("\n")
					}
				case IndexFragment:
					// Each index fragment should have a blank line between it and previous text
					if index > 0 {
						builder.WriteString("\n")
					}
					// Placing filenames in a such way will serve as example how to deal with filenames in responses
					// Because GPT is not trained well to format results with XML tags
					builder.WriteString("<filename>" + fragment.Payload + "</filename>")
					builder.WriteString("\n")
				case FileFragment:
					// Each file fragment must have a blank line between it and previous text
					if index > 0 {
						builder.WriteString("\n")
					}
					// Following formatting will also show GPT how to deal with filenames and file contens in responses
					builder.WriteString("<filename>" + fragment.Metadata + "</filename>")
					builder.WriteString("\n")
					builder.WriteString("<output>")
					builder.WriteString("\n")
					builder.WriteString(fragment.Payload)
					if fragment.Payload != "" && fragment.Payload[len(fragment.Payload)-1] != '\n' {
						builder.WriteString("\n")
					}
					builder.WriteString("</output>")
					builder.WriteString("\n")
				case TaggedFragment:
					if index > 0 {
						builder.WriteString("\n")
					}
					var tags []string
					err := json.Unmarshal([]byte(fragment.Metadata), &tags)
					if err != nil {
						return result, err
					}
					if len(tags) < 2 {
						return result, fmt.Errorf("invalid tags count in metadata for tagged fragment with index: %d", index)
					}
					builder.WriteString(tags[0])
					builder.WriteString(fragment.Payload)
					builder.WriteString(tags[1])
					builder.WriteString("\n")
				case MultilineTaggedFragment:
					if index > 0 {
						builder.WriteString("\n")
					}
					var tags []string
					err := json.Unmarshal([]byte(fragment.Metadata), &tags)
					if err != nil {
						return result, err
					}
					if len(tags) < 2 {
						return result, fmt.Errorf("invalid tags count in metadata for tagged fragment with index: %d", index)
					}
					builder.WriteString(tags[0])
					builder.WriteString("\n")
					builder.WriteString(fragment.Payload)
					if fragment.Payload != "" && fragment.Payload[len(fragment.Payload)-1] != '\n' {
						builder.WriteString("\n")
					}
					builder.WriteString(tags[1])
					builder.WriteString("\n")
				default:
					return result, fmt.Errorf("invalid fragment type: %d, index: %d", fragment.Type, index)
				}
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
