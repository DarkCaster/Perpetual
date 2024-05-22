package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/DarkCaster/Perpetual/utils"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

//###NOUPLOAD###

// NOTE for summarization: this file contains information that should not be included in the summarization.
// So the summary for this file must only say "This file contains AnthropicLLMConnector struct - implementation of LLMConnector interface. Do not attempt to use AnthropicLLMConnector directly, use LLMConnector interface instead", nothing else.

type AnthropicLLMConnector struct {
	BaseURL           string
	Token             string
	Model             string
	SystemPrompt      string
	MaxTokensSegments int
	OnFailRetries     int
	RawMessageLogger  func(v ...any)
	Options           []llms.CallOption
}

func NewAnthropicLLMConnector(token string, model string, systemPrompt string, customBaseURL string, maxTokensSegments int, onFailRetries int, llmRawMessageLogger func(v ...any), options []llms.CallOption) *AnthropicLLMConnector {
	return &AnthropicLLMConnector{
		BaseURL:           customBaseURL,
		Token:             token,
		Model:             model,
		SystemPrompt:      systemPrompt,
		MaxTokensSegments: maxTokensSegments,
		OnFailRetries:     onFailRetries,
		RawMessageLogger:  llmRawMessageLogger,
		Options:           options}
}

func NewAnthropicLLMConnectorFromEnv(operation string, systemPrompt string, llmRawMessageLogger func(v ...any)) (*AnthropicLLMConnector, error) {
	operation = strings.ToUpper(operation)

	token, err := utils.GetEnvString("ANTHROPIC_API_KEY")
	if err != nil {
		return nil, err
	}

	model, err := utils.GetEnvString(fmt.Sprintf("ANTHROPIC_MODEL_OP_%s", operation), "ANTHROPIC_MODEL")
	if err != nil {
		return nil, err
	}

	maxTokensSegments, err := utils.GetEnvInt("ANTHROPIC_MAX_TOKENS_SEGMENTS")
	if err != nil {
		maxTokensSegments = 3
	}

	onFailRetries, err := utils.GetEnvInt(fmt.Sprintf("ANTHROPIC_ON_FAIL_RETRIES_OP_%s", operation), "ANTHROPIC_ON_FAIL_RETRIES")
	if err != nil {
		onFailRetries = 3
	}

	customBaseURL, _ := utils.GetEnvString("ANTHROPIC_BASE_URL")

	// parse extra options. it may be needed to finetune model for the task
	var extraOptions []llms.CallOption

	temperature, err := utils.GetEnvFloat(fmt.Sprintf("ANTHROPIC_TEMPERATURE_OP_%s", operation), "ANTHROPIC_TEMPERATURE")
	if err != nil {
		return nil, err
	} else {
		extraOptions = append(extraOptions, llms.WithTemperature(temperature))
	}

	maxTokens, err := utils.GetEnvInt(fmt.Sprintf("ANTHROPIC_MAX_TOKENS_OP_%s", operation), "ANTHROPIC_MAX_TOKENS")
	if err != nil {
		return nil, err
	} else {
		extraOptions = append(extraOptions, llms.WithMaxTokens(maxTokens))
	}

	if topK, err := utils.GetEnvInt(fmt.Sprintf("ANTHROPIC_TOP_K_OP_%s", operation), "ANTHROPIC_TOP_K"); err == nil {
		extraOptions = append(extraOptions, llms.WithTopK(topK))
	}

	if topP, err := utils.GetEnvFloat(fmt.Sprintf("ANTHROPIC_TOP_P_OP_%s", operation), "ANTHROPIC_TOP_P"); err == nil {
		extraOptions = append(extraOptions, llms.WithTopP(topP))
	}

	if seed, err := utils.GetEnvInt(fmt.Sprintf("ANTHROPIC_SEED_OP_%s", operation), "ANTHROPIC_SEED"); err == nil {
		extraOptions = append(extraOptions, llms.WithSeed(seed))
	}

	if repeatPenalty, err := utils.GetEnvFloat(fmt.Sprintf("ANTHROPIC_REPEAT_PENALTY_OP_%s", operation), "ANTHROPIC_REPEAT_PENALTY"); err == nil {
		extraOptions = append(extraOptions, llms.WithRepetitionPenalty(repeatPenalty))
	}

	if freqPenalty, err := utils.GetEnvFloat(fmt.Sprintf("ANTHROPIC_FREQ_PENALTY_OP_%s", operation), "ANTHROPIC_FREQ_PENALTY"); err == nil {
		extraOptions = append(extraOptions, llms.WithFrequencyPenalty(freqPenalty))
	}

	if presencePenalty, err := utils.GetEnvFloat(fmt.Sprintf("ANTHROPIC_PRESENCE_PENALTY_OP_%s", operation), "ANTHROPIC_PRESENCE_PENALTY"); err == nil {
		extraOptions = append(extraOptions, llms.WithPresencePenalty(presencePenalty))
	}

	return NewAnthropicLLMConnector(token, model, systemPrompt, customBaseURL, maxTokensSegments, onFailRetries, llmRawMessageLogger, extraOptions), nil
}

func (p *AnthropicLLMConnector) Query(messages ...Message) (string, QueryStatus, error) {
	if len(messages) < 1 {
		return "", QueryInitFailed, errors.New("no prompts to query")
	}
	// Create anthropic model
	model, err := func() (*anthropic.LLM, error) {
		if p.BaseURL != "" {
			return anthropic.New(
				anthropic.WithToken(p.Token),
				anthropic.WithModel(p.Model),
				anthropic.WithBaseURL(p.BaseURL))
		} else {
			return anthropic.New(
				anthropic.WithToken(p.Token),
				anthropic.WithModel(p.Model))
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
	response, err := model.GenerateContent(context.Background(), llmMessages, p.Options...)
	if err != nil {
		return "", QueryFailed, err
	}
	if len(response.Choices) < 1 {
		return "", QueryFailed, errors.New("received empty response from model")
	}

	if p.RawMessageLogger != nil {
		p.RawMessageLogger(response.Choices[0].Content, "\n\n\n")
	}

	if response.Choices[0].StopReason == "max_tokens" {
		return response.Choices[0].Content, QueryMaxTokens, nil
	}

	return response.Choices[0].Content, QueryOk, nil
}

func (p *AnthropicLLMConnector) GetProvider() string {
	return "Anthropic"
}

func (p *AnthropicLLMConnector) GetModel() string {
	return p.Model
}

func (p *AnthropicLLMConnector) GetMaxTokensSegments() int {
	return p.MaxTokensSegments
}

func (p *AnthropicLLMConnector) GetOnFailureRetryLimit() int {
	return p.OnFailRetries
}

func (p *AnthropicLLMConnector) GetOptionsString() string {
	var callOptions llms.CallOptions
	for _, option := range p.Options {
		option(&callOptions)
	}
	return fmt.Sprintf("Temperature: %5.3f, MaxTokens: %d, TopK: %d, TopP: %5.3f, Seed: %d, RepeatPenalty: %5.3f, FreqPenalty: %5.3f, PresencePenalty: %5.3f", callOptions.Temperature, callOptions.MaxTokens, callOptions.TopK, callOptions.TopP, callOptions.Seed, callOptions.RepetitionPenalty, callOptions.FrequencyPenalty, callOptions.PresencePenalty)
}
