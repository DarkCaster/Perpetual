package prompts

import (
	"fmt"
	"strings"
)

type Prompts interface {
	// General helpers
	GetProjectFilesWhitelist() []string
	GetProjectFilesBlacklist() []string
	GetProjectFilesToMarkdownMappings() [][2]string
	GetSystemPrompt() string

	// Annotate-operation prompts
	GetAnnotatePrompt() [][2]string
	GetAIAnnotateResponse() string

	// Implement-operation helpers
	GetFileNameTags() []string
	GetFileNameTagsRegexps() []string
	GetImplementCommentRegexps() []string
	GetNoUploadCommentRegexps() []string
	GetFileNameEmbedRegex() string
	GetOutputTagsRegexps() []string
	GetReasoningsTags() []string
	GetReasoningsTagsRegexps() []string

	// Implement stage 1 prompts
	GetImplementStage1ProjectIndexPrompt() string
	GetAIImplementStage1ProjectIndexResponse() string
	GetImplementStage1SourceAnalysisPrompt() string

	// Implement stage 2 prompts
	GetImplementStage2ProjectCodePrompt() string
	GetAIImplementStage2ProjectCodeResponse() string
	GetImplementStage2FilesToChangePrompt() string
	GetImplementStage2FilesToChangeExtendedPrompt() string

	// Implement stage 2 no planning prompts
	GetImplementStage2NoPlanningPrompt() string
	GetAIImplementStage2NoPlanningResponse() string

	// Implement stage 3 prompts
	GetImplementStage3ChangesDonePrompt() string
	GetAIImplementStage3ChangesDoneResponse() string
	GetImplementStage3ProcessFilePrompt() string
	GetImplementStage3ContinuePrompt() string

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
