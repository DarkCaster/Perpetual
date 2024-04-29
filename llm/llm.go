package llm

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type LLMConnector interface {
	// Main interaction point with LLM
	Query(messages ...Message) (string, error)
	// Following functions needed for logging, consider not to use it anywhere else
	GetProvider() string
	GetModel() string
	GetTemperature() float64
	GetMaxTokens() int
}

func NewLLMConnector(operation string, systemPrompt string, rawMessageLogger func(v ...any)) (LLMConnector, error) {
	operation = strings.ToUpper(operation)
	provider := os.Getenv(fmt.Sprintf("LLM_PROVIDER_OP_%s", operation))
	if provider == "" {
		provider = os.Getenv("LLM_PROVIDER")
	}
	if provider == "" {
		return nil, fmt.Errorf("LLM_PROVIDER_OP_%s or LLM_PROVIDER env var not set", operation)
	}
	provider = strings.ToUpper(provider)
	tempStr := os.Getenv(fmt.Sprintf("TEMPERATURE_%s_OP_%s", provider, operation))
	if tempStr == "" {
		tempStr = os.Getenv(fmt.Sprintf("TEMPERATURE_%s", provider))
	}
	if tempStr == "" {
		tempStr = os.Getenv("TEMPERATURE")
	}
	if tempStr == "" {
		return nil, fmt.Errorf("TEMPERATURE_%s_OP_%s or TEMPERATURE_%s or TEMPERATURE env var not set", provider, operation, provider)
	}
	temperature, err := strconv.ParseFloat(tempStr, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert temperature env variable to float64: %s", err)
	}
	switch provider {
	case "ANTHROPIC":
		return NewAnthropicLLMConnectorFromEnv(operation, systemPrompt, temperature, rawMessageLogger)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}
