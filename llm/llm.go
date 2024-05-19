package llm

import (
	"fmt"
	"strings"
)

type QueryStatus int

const (
	QueryOk QueryStatus = iota
	QueryInitFailed
	QueryMaxTokens
	QueryFailed
)

type LLMConnector interface {
	// Main interaction point with LLM
	Query(messages ...Message) (string, QueryStatus, error)
	// Get retry limit to get extra fragments of code when generation hits LLM token-limit
	GetMaxTokensRetryLimit() int
	// Get retry limit on general query fail
	GetOnFailureRetryLimit() int
	// Following functions needed for LLM messages logging, consider not to use it anywhere else
	GetProvider() string
	GetModel() string
	GetTemperature() float64
	GetMaxTokens() int
}

func NewLLMConnector(operation string, systemPrompt string, llmRawMessageLogger func(v ...any)) (LLMConnector, error) {
	operation = strings.ToUpper(operation)

	provider, err := utils.GetEnvUpperString(
		fmt.Sprintf("LLM_PROVIDER_OP_%s", operation),
		"LLM_PROVIDER")
	if err != nil {
		return nil, err
	}

	temperature, err := utils.GetEnvFloat(
		fmt.Sprintf("%s_TEMPERATURE_OP_%s", provider, operation),
		fmt.Sprintf("%s_TEMPERATURE", provider),
		"TEMPERATURE")
	if err != nil {
		return nil, err
	}

	switch provider {
	case "ANTHROPIC":
		return NewAnthropicLLMConnectorFromEnv(operation, systemPrompt, temperature, llmRawMessageLogger)
	case "OPENAI":
		return NewOpenAILLMConnectorFromEnv(operation, systemPrompt, temperature, llmRawMessageLogger)
	case "OLLAMA":
		return NewOllamaLLMConnectorFromEnv(operation, systemPrompt, temperature, llmRawMessageLogger)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}

func GetDebugString(llm LLMConnector) string {
	return fmt.Sprintf("Provider: %s, Model: %s, Temperature: %5.3f, MaxTokens: %d, MaxTokensRetries: %d", llm.GetProvider(), llm.GetModel(), llm.GetTemperature(), llm.GetMaxTokens(), llm.GetMaxTokensRetryLimit())
}
