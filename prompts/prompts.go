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

	// Annotate-operation prompts
	GetAnnotateConfig() map[string]interface{}

	// Annotate-operation prompts
	GetImplementConfig() map[string]interface{}

	// Doc project index and code prompts
	GetDocProjectIndexPrompt() string
	GetAIDocProjectIndexResponse() string
	GetDocProjectCodePrompt() string
	GetAIDocProjectCodeResponse() string
	GetDocExamplePrompt() string
	GetAIDocExampleResponse() string

	// Doc stage1 prompts
	GetDocStage1WritePrompt() string
	GetDocStage1RefinePrompt() string

	// Doc stage2 prompts
	GetDocStage2WritePrompt() string
	GetDocStage2RefinePrompt() string
	GetDocStage2ContinuePrompt() string
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
