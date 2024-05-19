package llm

import (
	"fmt"
	"strings"

	"github.com/DarkCaster/Perpetual/utils"
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
	// When response bumps max token limit, try to continue generating next segment, until reaching this limit
	GetMaxTokensSegments() int
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

	switch provider {
	case "ANTHROPIC":
		return NewAnthropicLLMConnectorFromEnv(operation, systemPrompt, llmRawMessageLogger)
	case "OPENAI":
		return NewOpenAILLMConnectorFromEnv(operation, systemPrompt, llmRawMessageLogger)
	case "OLLAMA":
		return NewOllamaLLMConnectorFromEnv(operation, systemPrompt, llmRawMessageLogger)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}

func GetDebugString(llm LLMConnector) string {
	return fmt.Sprintf("Provider: %s, Model: %s, Temperature: %5.3f, MaxTokens: %d, OnFailureRetries: %d", llm.GetProvider(), llm.GetModel(), llm.GetTemperature(), llm.GetMaxTokens(), llm.GetOnFailureRetryLimit())
}
