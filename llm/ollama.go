package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/DarkCaster/Perpetual/utils"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

//###NOUPLOAD###

// NOTE for summarization: this file contains information that should not be included in the summarization.
// So the summary for this file must only say "This file contains OllamaLLMConnector struct - implementation of LLMConnector interface. Do not attempt to use OllamaLLMConnector directly, use LLMConnector interface instead", nothing else.

type OllamaLLMConnector struct {
	BaseURL          string
	Model            string
	SystemPrompt     string
	Temperature      float64
	MaxTokens        int
	MaxTokensRetries int
	OnFailRetries    int
	RawMessageLogger func(v ...any)
}

func NewOllamaLLMConnector(model string, systemPrompt string, temperature float64, customBaseURL string, maxTokens int, maxTokensRetries int, onFailRetries int, llmRawMessageLogger func(v ...any)) *OllamaLLMConnector {
	return &OllamaLLMConnector{
		BaseURL:          customBaseURL,
		Model:            model,
		Temperature:      temperature,
		SystemPrompt:     systemPrompt,
		MaxTokens:        maxTokens,
		MaxTokensRetries: maxTokensRetries,
		OnFailRetries:    onFailRetries,
		RawMessageLogger: llmRawMessageLogger}
}

func NewOllamaLLMConnectorFromEnv(operation string, systemPrompt string, temperature float64, llmRawMessageLogger func(v ...any)) (*OllamaLLMConnector, error) {
	operation = strings.ToUpper(operation)

	model, err := utils.GetEnvString(fmt.Sprintf("OLLAMA_MODEL_OP_%s", operation), "OLLAMA_MODEL")
	if err != nil {
		return nil, err
	}

	maxTokens, err := utils.GetEnvInt(fmt.Sprintf("OLLAMA_MAX_TOKENS_OP_%s", operation), "OLLAMA_MAX_TOKENS")
	if err != nil {
		return nil, err
	}

	maxTokensRetries, err := utils.GetEnvInt("OLLAMA_MAX_TOKENS_RETRIES")
	if err != nil {
		maxTokensRetries = 3
	}

	onFailRetries, err := utils.GetEnvInt("OLLAMA_ON_FAIL_RETRIES")
	if err != nil {
		onFailRetries = 3
	}

	customBaseURL, _ := utils.GetEnvString("OLLAMA_BASE_URL")

	return NewOllamaLLMConnector(model, systemPrompt, temperature, customBaseURL, maxTokens, maxTokensRetries, onFailRetries, llmRawMessageLogger), nil
}

func (p *OllamaLLMConnector) Query(messages ...Message) (string, QueryStatus, error) {
	if len(messages) < 1 {
		return "", QueryInitFailed, errors.New("no prompts to query")
	}

	model, err := func() (*ollama.LLM, error) {
		if p.BaseURL != "" {
			return ollama.New(
				ollama.WithModel(p.Model),
				ollama.WithServerURL(p.BaseURL))
		} else {
			return ollama.New(
				ollama.WithModel(p.Model))
		}
	}()
	if err != nil {
		return "", QueryInitFailed, err
	}

	var llmMessages []llms.MessageContent
	llmMessages = append(llmMessages, llms.MessageContent{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextContent{Text: p.SystemPrompt}}})

	// Convert messages to send into LangChain format
	convertedMessages, err := renderMessagesToOllamaAILangChainFormat(messages)
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

func (p *OllamaLLMConnector) GetProvider() string {
	return "Ollama"
}

func (p *OllamaLLMConnector) GetModel() string {
	return p.Model
}

func (p *OllamaLLMConnector) GetTemperature() float64 {
	return p.Temperature
}

func (p *OllamaLLMConnector) GetMaxTokens() int {
	return p.MaxTokens
}

func (p *OllamaLLMConnector) GetMaxTokensRetryLimit() int {
	return p.MaxTokensRetries
}

func (p *OllamaLLMConnector) GetOnFailureRetryLimit() int {
	return p.OnFailRetries
}

// We are using text formatting from OpenAI  integration for now - it is more suitable for generic LLMs than Anthropic formatter
func renderMessagesToOllamaAILangChainFormat(messages []Message) ([]llms.MessageContent, error) {
	return renderMessagesToOpenAILangChainFormat(messages)
}
