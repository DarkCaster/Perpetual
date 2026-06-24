package op_onboard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

type ActiveEnvironment struct {
	Vars                   map[string]string
	ProviderSelections     map[string]string
	ActiveProviderPrefixes map[string]bool
}

var providerSpecificEnvRx = regexp.MustCompile(`^(ANTHROPIC|OPENAI|OLLAMA|GENERIC)([0-9]*)_.+`)

func DetectActiveEnvironment(logger logging.ILogger) (*ActiveEnvironment, error) {
	globalConfigDir, err := utils.FindConfigDir()
	if err != nil {
		return nil, err
	}

	utils.LoadEnvFiles(logger, globalConfigDir)

	allVars := readEnvironmentVariables()
	providerSelections := detectProviderSelections(allVars)
	activeProviderPrefixes := detectActiveProviderPrefixes(providerSelections)

	result := &ActiveEnvironment{
		Vars:                   map[string]string{},
		ProviderSelections:     providerSelections,
		ActiveProviderPrefixes: activeProviderPrefixes,
	}

	for name, value := range allVars {
		switch {
		case isProviderSelectionVar(name):
			result.Vars[name] = value
		case name == "FALLBACK_TEXT_ENCODING":
			result.Vars[name] = value
		case isActiveProviderSpecificVar(name, activeProviderPrefixes):
			result.Vars[name] = value
		}
	}

	return result, nil
}

func PrintActiveEnvironment(w io.Writer, env *ActiveEnvironment) {
	fmt.Fprintln(w, "Active env variables:")

	providerVars := make([]string, 0)
	otherVars := make([]string, 0)

	for name := range env.Vars {
		if isProviderSelectionVar(name) {
			providerVars = append(providerVars, name)
		} else {
			otherVars = append(otherVars, name)
		}
	}

	sort.Strings(providerVars)
	sort.Strings(otherVars)

	for _, name := range providerVars {
		fmt.Fprintf(w, "%s=%s\n", name, env.Vars[name])
	}
	for _, name := range otherVars {
		fmt.Fprintf(w, "%s=%s\n", name, printableEnvValue(name, env.Vars[name]))
	}
}

func (env *ActiveEnvironment) Get(name string) (string, bool) {
	if env == nil {
		return "", false
	}
	value, ok := env.Vars[name]
	return value, ok
}

func (env *ActiveEnvironment) Has(name string) bool {
	if env == nil {
		return false
	}
	_, ok := env.Vars[name]
	return ok
}

func readEnvironmentVariables() map[string]string {
	result := map[string]string{}

	for _, item := range os.Environ() {
		name, value, ok := strings.Cut(item, "=")
		if !ok {
			continue
		}
		result[name] = value
	}

	return result
}

func detectProviderSelections(vars map[string]string) map[string]string {
	result := map[string]string{}

	for name, value := range vars {
		if isProviderSelectionVar(name) {
			result[name] = value
		}
	}

	return result
}

func detectActiveProviderPrefixes(providerSelections map[string]string) map[string]bool {
	result := map[string]bool{}

	for _, provider := range providerSelections {
		prefix := providerNameToEnvPrefix(provider)
		if prefix != "" {
			result[prefix] = true
		}
	}

	return result
}

func providerNameToEnvPrefix(provider string) string {
	provider = strings.ToUpper(strings.TrimSpace(provider))
	if provider == "" {
		return ""
	}

	matches := regexp.MustCompile(`^([A-Z]+)([0-9]*)$`).FindStringSubmatch(provider)
	if len(matches) < 2 {
		return ""
	}

	switch matches[1] {
	case "ANTHROPIC", "OPENAI", "OLLAMA", "GENERIC":
		return matches[1] + matches[2]
	default:
		return ""
	}
}

func isProviderSelectionVar(name string) bool {
	return name == "LLM_PROVIDER" || strings.HasPrefix(name, "LLM_PROVIDER_OP_")
}

func isActiveProviderSpecificVar(name string, activeProviderPrefixes map[string]bool) bool {
	matches := providerSpecificEnvRx.FindStringSubmatch(name)
	if len(matches) < 3 {
		return false
	}

	prefix := matches[1] + matches[2]
	return activeProviderPrefixes[prefix]
}

func printableEnvValue(name, value string) string {
	if isSensitiveEnvValue(name) {
		if value == "" {
			return ""
		}
		return "<hidden>"
	}
	return value
}

func isSensitiveEnvValue(name string) bool {
	return strings.HasSuffix(name, "_API_KEY") ||
		strings.HasSuffix(name, "_AUTH") ||
		strings.HasSuffix(name, "_BASE_URL")
}
