package op_onboard

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

type operationValidation struct {
	Operation string
	Provider  string
	Model     string
}

var validationOperations = []string{
	"ANNOTATE",
	"EMBED",
	"IMPLEMENT_STAGE1",
	"IMPLEMENT_STAGE2",
	"IMPLEMENT_STAGE3",
	"IMPLEMENT_STAGE4",
	"DOC_STAGE1",
	"DOC_STAGE2",
	"EXPLAIN_STAGE1",
	"EXPLAIN_STAGE2",
}

var providerNameRx = regexp.MustCompile(`^([A-Z]+)([0-9]*)$`)

func ValidateEnvironment(w io.Writer, env *ActiveEnvironment) {
	missing, configErrors, selected := validateEnvironment(env)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Validation:")

	if len(missing) > 0 {
		fmt.Fprintln(w, "Missing required env variables:")
		for _, item := range missing {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}

	if len(configErrors) > 0 {
		fmt.Fprintln(w, "Configuration errors:")
		for _, item := range configErrors {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}

	if len(missing) > 0 || len(configErrors) > 0 {
		return
	}

	fmt.Fprintln(w, "Selected providers and models:")
	for _, item := range selected {
		fmt.Fprintf(w, "%s: %s -> %s\n", item.Operation, strings.ToLower(item.Provider), item.Model)
	}
}

func validateEnvironment(env *ActiveEnvironment) ([]string, []string, []operationValidation) {
	missing := []string{}
	configErrors := []string{}
	selected := []operationValidation{}

	for _, operation := range validationOperations {
		provider, ok := getOperationProvider(env, operation)
		if !ok {
			missing = append(missing, fmt.Sprintf("%s: LLM_PROVIDER_OP_%s or LLM_PROVIDER", operation, operation))
			continue
		}

		baseProvider, prefix, ok := parseProvider(provider)
		if !ok {
			configErrors = append(configErrors, fmt.Sprintf("%s: invalid provider %q", operation, provider))
			continue
		}

		model := ""
		switch baseProvider {
		case "ANTHROPIC":
			if operation == "EMBED" {
				configErrors = append(configErrors, "EMBED: anthropic provider does not support embeddings")
				continue
			}
			missing = appendMissingRequired(env, missing, operation, prefix+"_AUTH", prefix+"_API_KEY")
			missing = appendMissingRequired(env, missing, operation, prefix+"_MODEL_OP_"+operation, prefix+"_MODEL")
			missing = appendMissingRequired(env, missing, operation, prefix+"_MAX_TOKENS_OP_"+operation, prefix+"_MAX_TOKENS")
			model = getFirstEnvValue(env, prefix+"_MODEL_OP_"+operation, prefix+"_MODEL")

		case "OPENAI":
			missing = appendMissingRequired(env, missing, operation, prefix+"_AUTH", prefix+"_API_KEY")
			if operation == "EMBED" {
				missing = appendMissingRequired(env, missing, operation, prefix+"_MODEL_OP_EMBED")
				model = getFirstEnvValue(env, prefix+"_MODEL_OP_EMBED")
			} else {
				missing = appendMissingRequired(env, missing, operation, prefix+"_MODEL_OP_"+operation, prefix+"_MODEL")
				model = getFirstEnvValue(env, prefix+"_MODEL_OP_"+operation, prefix+"_MODEL")
			}

		case "OLLAMA":
			if operation == "EMBED" {
				missing = appendMissingRequired(env, missing, operation, prefix+"_MODEL_OP_EMBED")
				model = getFirstEnvValue(env, prefix+"_MODEL_OP_EMBED")
			} else {
				missing = appendMissingRequired(env, missing, operation, prefix+"_MODEL_OP_"+operation, prefix+"_MODEL")
				missing = appendMissingRequired(env, missing, operation, prefix+"_MAX_TOKENS_OP_"+operation, prefix+"_MAX_TOKENS")
				model = getFirstEnvValue(env, prefix+"_MODEL_OP_"+operation, prefix+"_MODEL")
			}

		case "GENERIC":
			missing = appendMissingRequired(env, missing, operation, prefix+"_BASE_URL")
			if operation == "EMBED" {
				missing = appendMissingRequired(env, missing, operation, prefix+"_MODEL_OP_EMBED")
				model = getFirstEnvValue(env, prefix+"_MODEL_OP_EMBED")
			} else {
				missing = appendMissingRequired(env, missing, operation, prefix+"_MODEL_OP_"+operation, prefix+"_MODEL")
				model = getFirstEnvValue(env, prefix+"_MODEL_OP_"+operation, prefix+"_MODEL")
			}

		default:
			configErrors = append(configErrors, fmt.Sprintf("%s: unsupported provider %q", operation, provider))
			continue
		}

		selected = append(selected, operationValidation{
			Operation: operation,
			Provider:  provider,
			Model:     model,
		})
	}

	missing = uniqueSortedStrings(missing)
	sort.Strings(configErrors)

	return missing, configErrors, selected
}

func getOperationProvider(env *ActiveEnvironment, operation string) (string, bool) {
	if value, ok := getNonEmptyEnvValue(env, "LLM_PROVIDER_OP_"+operation); ok {
		return value, true
	}
	return getNonEmptyEnvValue(env, "LLM_PROVIDER")
}

func parseProvider(provider string) (string, string, bool) {
	provider = strings.ToUpper(strings.TrimSpace(provider))
	matches := providerNameRx.FindStringSubmatch(provider)
	if len(matches) < 2 {
		return "", "", false
	}

	switch matches[1] {
	case "ANTHROPIC", "OPENAI", "OLLAMA", "GENERIC":
		return matches[1], matches[1] + matches[2], true
	default:
		return "", "", false
	}
}

func appendMissingRequired(env *ActiveEnvironment, missing []string, operation string, names ...string) []string {
	for _, name := range names {
		if _, ok := getNonEmptyEnvValue(env, name); ok {
			return missing
		}
	}
	return append(missing, fmt.Sprintf("%s: %s", operation, strings.Join(names, " or ")))
}

func getFirstEnvValue(env *ActiveEnvironment, names ...string) string {
	for _, name := range names {
		if value, ok := getNonEmptyEnvValue(env, name); ok {
			return value
		}
	}
	return ""
}

func getNonEmptyEnvValue(env *ActiveEnvironment, name string) (string, bool) {
	if env == nil {
		return "", false
	}
	value, ok := env.Get(name)
	if !ok || strings.TrimSpace(value) == "" {
		return "", false
	}
	return value, true
}

func uniqueSortedStrings(items []string) []string {
	seen := map[string]bool{}
	result := []string{}

	for _, item := range items {
		if seen[item] {
			continue
		}
		seen[item] = true
		result = append(result, item)
	}

	sort.Strings(result)
	return result
}
