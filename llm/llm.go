package llm

import (
	"fmt"
	"os"
	"strconv"
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
	// Limit maximum re-tries to get extra fragments of code when generation hits LLM token-limit
	GetMaxTokensRetryLimit() int
	// Following functions needed for LLM messages logging, consider not to use it anywhere else
	GetProvider() string
	GetModel() string
	GetTemperature() float64
	GetMaxTokens() int
}

func NewLLMConnector(operation string, systemPrompt string, llmRawMessageLogger func(v ...any)) (LLMConnector, error) {
	operation = strings.ToUpper(operation)
	provider := os.Getenv(fmt.Sprintf("LLM_PROVIDER_OP_%s", operation))
	if provider == "" {
		provider = os.Getenv("LLM_PROVIDER")
	}
	if provider == "" {
		return nil, fmt.Errorf("LLM_PROVIDER_OP_%s or LLM_PROVIDER env var not set", operation)
	}
	provider = strings.ToUpper(provider)
	tempStr := os.Getenv(fmt.Sprintf("%s_TEMPERATURE_OP_%s", provider, operation))
	if tempStr == "" {
		tempStr = os.Getenv(fmt.Sprintf("%s_TEMPERATURE", provider))
	}
	if tempStr == "" {
		tempStr = os.Getenv("TEMPERATURE")
	}
	if tempStr == "" {
		return nil, fmt.Errorf("%s_TEMPERATURE_OP_%s or %s_TEMPERATURE or TEMPERATURE env var not set", provider, operation, provider)
	}
	temperature, err := strconv.ParseFloat(tempStr, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert temperature env variable to float64: %s", err)
	}
	switch provider {
	case "ANTHROPIC":
		return NewAnthropicLLMConnectorFromEnv(operation, systemPrompt, temperature, llmRawMessageLogger)
	case "OPENAI":
		return NewOpenAILLMConnectorFromEnv(operation, systemPrompt, temperature, llmRawMessageLogger)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}

func GetDebugString(llm LLMConnector) string {
	return fmt.Sprintf("Provider: %s, Model: %s, Temperature: %5.3f, MaxTokens: %d, MaxTokensRetries: %d", llm.GetProvider(), llm.GetModel(), llm.GetTemperature(), llm.GetMaxTokens(), llm.GetMaxTokensRetryLimit())
}
