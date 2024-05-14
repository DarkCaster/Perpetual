package llm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

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
	RawMessageLogger func(v ...any)
}

func NewOpenAILLMConnector(token string, model string, systemPrompt string, temperature float64, customBaseURL string, maxTokens int, maxTokensRetries int, llmRawMessageLogger func(v ...any)) *OpenAILLMConnector {
	return &OpenAILLMConnector{BaseURL: customBaseURL, Token: token, Model: model, Temperature: temperature, SystemPrompt: systemPrompt, MaxTokens: maxTokens, MaxTokensRetries: maxTokensRetries, RawMessageLogger: llmRawMessageLogger}
}

func NewOpenAILLMConnectorFromEnv(operation string, systemPrompt string, temperature float64, llmRawMessageLogger func(v ...any)) (*OpenAILLMConnector, error) {
	operation = strings.ToUpper(operation)

	token := os.Getenv("OPENAI_API_KEY")
	if token == "" {
		return nil, errors.New("OPENAI_API_KEY env var not set")
	}

	model := os.Getenv(fmt.Sprintf("OPENAI_MODEL_OP_%s", operation))
	if model == "" {
		model = os.Getenv("OPENAI_MODEL")
	}
	if model == "" {
		return nil, fmt.Errorf("OPENAI_MODEL_OP_%s or OPENAI_MODEL env var not set", operation)
	}

	maxTokensStr := os.Getenv(fmt.Sprintf("OPENAI_MAX_TOKENS_OP_%s", operation))
	if maxTokensStr == "" {
		maxTokensStr = os.Getenv("OPENAI_MAX_TOKENS")
	}
	if maxTokensStr == "" {
		return nil, fmt.Errorf("OPENAI_MAX_TOKENS_OP_%s or OPENAI_MAX_TOKENS env var not set", operation)
	}

	maxTokens, err := strconv.ParseInt(maxTokensStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert temperature env variable to int64: %s", err)
	}

	maxTokensRetriesStr := os.Getenv("OPENAI_MAX_TOKENS_RETRIES")
	if maxTokensRetriesStr == "" {
		maxTokensRetriesStr = "3"
	}

	maxTokensRetries, err := strconv.ParseInt(maxTokensRetriesStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert max tokens retries env variable to int64: %s", err)
	}

	customBaseURL := os.Getenv("OPENAI_BASE_URL")
	return NewOpenAILLMConnector(token, model, systemPrompt, temperature, customBaseURL, int(maxTokens), int(maxTokensRetries), llmRawMessageLogger), nil
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

func renderMessagesToOpenAILangChainFormat(messages []Message) ([]llms.MessageContent, error) {
	return renderMessagesToAnthropicLangChainFormat(messages)
}
