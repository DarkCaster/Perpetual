package llm

import (
	"errors"
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

type OutputFormat int

const (
	OutputPlain OutputFormat = iota
	OutputJson
)

type VariantSelectionStrategy int

const (
	Short VariantSelectionStrategy = iota
	Long
	Combine
	Best
)

type LLMConnector interface {
	// Main interaction point with LLM
	Query(maxCandidates int, messages ...Message) ([]string, QueryStatus, error)
	// When response bumps max token limit, try to continue generating next segment, until reaching this limit
	GetMaxTokensSegments() int
	GetOnFailureRetryLimit() int
	GetDebugString() string
	GetOutputFormat() OutputFormat
	// Results variant-count management.
	GetVariantCount() int
	GetVariantSelectionStrategy() VariantSelectionStrategy
}

func NewLLMConnector(operation string,
	systemPrompt string,
	filesToMdLangMappings [][]string,
	outputSchema map[string]interface{},
	outputSchemaName string,
	outputSchemaDesc string,
	llmRawMessageLogger func(v ...any)) (LLMConnector, error) {
	// Input parameters check
	if operation == "" {
		return nil, errors.New("operation name is empty")
	}
	if systemPrompt == "" {
		return nil, errors.New("system prompt name is empty")
	}

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

	prefix := fmt.Sprintf("%s%s", strings.ToUpper(provider), strings.ToUpper(subProfile))
	outputFormat := OutputPlain

	// try to setup output format from config and outputschema if applicable
	// particular llm provider can still switch to plain format if schema is invalid or structured JSON output format is not supported
	if format, err := utils.GetEnvString(fmt.Sprintf("%s_FORMAT_OP_%s", prefix, operation)); err == nil {
		if strings.ToUpper(format) == "JSON" {
			if len(outputSchema) < 1 {
				return nil, fmt.Errorf("output schema is empty, cannot use JSON mode")
			}
			if len(outputSchemaName) < 1 {
				return nil, fmt.Errorf("output schema name is empty, cannot use JSON mode")
			}
			if len(outputSchemaDesc) < 1 {
				return nil, fmt.Errorf("output schema description is empty, cannot use JSON mode")
			}
			outputFormat = OutputJson
		} else {
			outputFormat = OutputPlain
			outputSchema = map[string]interface{}{}
		}
	} else {
		outputFormat = OutputPlain
		outputSchema = map[string]interface{}{}
	}

	switch provider {
	case "ANTHROPIC":
		return NewAnthropicLLMConnectorFromEnv(
			subProfile,
			operation,
			systemPrompt,
			filesToMdLangMappings,
			outputSchema,
			outputSchemaName,
			outputSchemaDesc,
			outputFormat,
			llmRawMessageLogger)
	case "OPENAI":
		return NewOpenAILLMConnectorFromEnv(
			subProfile,
			operation,
			systemPrompt,
			filesToMdLangMappings,
			outputSchema,
			outputSchemaName,
			outputSchemaDesc,
			outputFormat,
			llmRawMessageLogger)
	case "OLLAMA":
		return NewOllamaLLMConnectorFromEnv(
			subProfile,
			operation,
			systemPrompt,
			filesToMdLangMappings,
			outputSchema,
			outputFormat,
			llmRawMessageLogger)
	case "GENERIC":
		return NewGenericLLMConnectorFromEnv(
			subProfile,
			operation,
			systemPrompt,
			filesToMdLangMappings,
			llmRawMessageLogger)
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
