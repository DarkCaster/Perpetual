package llm

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/DarkCaster/Perpetual/utils"
)

const LLMRawLogFile = ".message_log.txt"

type QueryStatus int

const (
	QueryOk QueryStatus = iota
	QueryInitFailed
	QueryMaxTokens
	QueryFailed
)

type VariantSelectionStrategy int

const (
	Short VariantSelectionStrategy = iota
	Long
	Combine
)

type LLMConnector interface {
	// Main interaction point with LLM
	Query(maxCandidates int, messages ...Message) ([]string, QueryStatus, error)
	// When response bumps max token limit, try to continue generating next segment, until reaching this limit
	GetMaxTokensSegments() int
	// Get retry limit on general query fail
	GetOnFailureRetryLimit() int
	// Following functions needed for LLM messages logging, consider not to use it anywhere else
	GetProvider() string
	GetModel() string
	GetOptionsString() string
	// Results variant-count management.
	GetVariantCount() int
	GetVariantSelectionStrategy() VariantSelectionStrategy
}

func NewLLMConnector(operation string, systemPrompt string, filesToMdLangMappings [][2]string, llmRawMessageLogger func(v ...any)) (LLMConnector, error) {
	operation = strings.ToUpper(operation)

	provider, err := utils.GetEnvUpperString(
		fmt.Sprintf("LLM_PROVIDER_OP_%s", operation),
		"LLM_PROVIDER")
	if err != nil {
		return nil, err
	}

	switch provider {
	case "ANTHROPIC":
		return NewAnthropicLLMConnectorFromEnv(operation, systemPrompt, filesToMdLangMappings, llmRawMessageLogger)
	case "OPENAI":
		return NewOpenAILLMConnectorFromEnv(operation, systemPrompt, filesToMdLangMappings, llmRawMessageLogger)
	case "OLLAMA":
		return NewOllamaLLMConnectorFromEnv(operation, systemPrompt, filesToMdLangMappings, llmRawMessageLogger)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}

func GetDebugString(llm LLMConnector) string {
	return fmt.Sprintf("Provider: %s, Model: %s, OnFailureRetries: %d, %s", llm.GetProvider(), llm.GetModel(), llm.GetOnFailureRetryLimit(), llm.GetOptionsString())
}

func GetSimpleRawMessageLogger(perpetualDir string) func(v ...any) {
	logFunc := func(v ...any) {
		if len(v) < 1 {
			return
		}
		str := fmt.Sprintf(v[0].(string), v[1:]...)
		utils.AppendToTextFile(filepath.Join(perpetualDir, LLMRawLogFile), str)
	}
	return logFunc
}
