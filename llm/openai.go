package llm

import (
	"context"
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
	BaseURL           string
	Token             string
	Model             string
	SystemPrompt      string
	Temperature       float64
	MaxTokens         int
	MaxTokensSegments int
	OnFailRetries     int
	RawMessageLogger  func(v ...any)
}

func NewOpenAILLMConnector(token string, model string, systemPrompt string, temperature float64, customBaseURL string, maxTokens int, maxTokensSegments int, onFailRetries int, llmRawMessageLogger func(v ...any)) *OpenAILLMConnector {
	return &OpenAILLMConnector{
		BaseURL:           customBaseURL,
		Token:             token,
		Model:             model,
		Temperature:       temperature,
		SystemPrompt:      systemPrompt,
		MaxTokens:         maxTokens,
		MaxTokensSegments: maxTokensSegments,
		OnFailRetries:     onFailRetries,
		RawMessageLogger:  llmRawMessageLogger}
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

	maxTokensSegments, err := utils.GetEnvInt("OPENAI_MAX_TOKENS_SEGMENTS")
	if err != nil {
		maxTokensSegments = 3
	}

	onFailRetries, err := utils.GetEnvInt(fmt.Sprintf("OPENAI_ON_FAIL_RETRIES_OP_%s", operation), "OPENAI_ON_FAIL_RETRIES")
	if err != nil {
		onFailRetries = 3
	}

	customBaseURL, _ := utils.GetEnvString("OPENAI_BASE_URL")

	return NewOpenAILLMConnector(token, model, systemPrompt, temperature, customBaseURL, maxTokens, maxTokensSegments, onFailRetries, llmRawMessageLogger), nil
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
	convertedMessages, err := renderMessagesToGenericAILangChainFormat(messages)
	if err != nil {
		return "", QueryInitFailed, err
	}
	llmMessages = append(llmMessages, convertedMessages...)

	if p.RawMessageLogger != nil {
		for _, m := range llmMessages {
			p.RawMessageLogger(m, "\n\n\n")
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
		p.RawMessageLogger(response.Choices[0].Content, "\n\n\n")
	}

	if response.Choices[0].StopReason == "length" {
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

func (p *OpenAILLMConnector) GetMaxTokensSegments() int {
	return p.MaxTokensSegments
}

func (p *OpenAILLMConnector) GetOnFailureRetryLimit() int {
	return p.OnFailRetries
}

func (p *OpenAILLMConnector) GetOptionsString() string {
	return fmt.Sprintf("MaxTokens: %d, Temperature: %5.3f", p.MaxTokens, p.Temperature)
}
