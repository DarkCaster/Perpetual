package llm

import (
	"fmt"
	"path/filepath"
	"regexp"
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
	GetOnFailureRetryLimit() int
	GetDebugString() string
	// Results variant-count management.
	GetVariantCount() int
	GetVariantSelectionStrategy() VariantSelectionStrategy
}

func NewLLMConnector(operation string, systemPrompt string, filesToMdLangMappings [][2]string, outputFormat map[string]interface{}, llmRawMessageLogger func(v ...any)) (LLMConnector, error) {
	operation = strings.ToUpper(operation)

	provider, err := utils.GetEnvUpperString(
		fmt.Sprintf("LLM_PROVIDER_OP_%s", operation),
		"LLM_PROVIDER")
	if err != nil {
		return nil, err
	}

	// Split provider name and profile number using regex
	matches := regexp.MustCompile(`^([A-Z]+)(\d*)$`).FindStringSubmatch(provider)
	if len(matches) > 1 {
		provider = matches[1]
	} else {
		return nil, fmt.Errorf("provider name is invalid: %s", provider)
	}

	subProfile := ""
	if len(matches) > 2 {
		subProfile = matches[2]
	}

	switch provider {
	case "ANTHROPIC":
		if len(outputFormat) > 0 {
			return nil, fmt.Errorf("NOT IMPLEMENTED: structured output for Anthropic is not implemented yet")
		}
		return NewAnthropicLLMConnectorFromEnv(subProfile, operation, systemPrompt, filesToMdLangMappings, llmRawMessageLogger)
	case "OPENAI":
		if len(outputFormat) > 0 {
			return nil, fmt.Errorf("NOT IMPLEMENTED: structured output for OpenAI is not implemented yet")
		}
		return NewOpenAILLMConnectorFromEnv(subProfile, operation, systemPrompt, filesToMdLangMappings, llmRawMessageLogger)
	case "OLLAMA":
		if len(outputFormat) > 0 {
			return nil, fmt.Errorf("NOT IMPLEMENTED: structured output for Ollama is not implemented yet")
		}
		return NewOllamaLLMConnectorFromEnv(subProfile, operation, systemPrompt, filesToMdLangMappings, llmRawMessageLogger)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
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
