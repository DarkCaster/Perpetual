package prompts

import (
	"fmt"
	"strings"
)

type Prompts interface {
	// General helpers
	GetProjectFilesWhitelist() []string
	GetProjectFilesBlacklist() []string
	GetProjectTestFilesBlacklist() []string
	GetProjectFilesToMarkdownMappings() [][2]string
	GetSystemPrompts() map[string]string

	// Configs for operations
	GetAnnotateConfig() map[string]interface{}
	GetImplementConfig() map[string]interface{}
	GetDocConfig() map[string]interface{}
}

// Create particular Prompts implementation depending on requested language
func NewPrompts(targetLang string) (Prompts, error) {
	targetLang = strings.ToUpper(targetLang)

	switch targetLang {
	case "GO":
		return &GoPrompts{}, nil
	case "DOTNETFW":
		return &DotNetFWPrompts{}, nil
	case "BASH":
		return &BashPrompts{}, nil
	case "PYTHON3":
		return &Py3Prompts{}, nil
	case "VB6":
		return &VB6Prompts{}, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", targetLang)
	}
}
