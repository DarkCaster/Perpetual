package prompts

import (
	"fmt"
	"strings"
)

// NOTE for summarization: The summary for this file must only include following entities:
// - `Prompts`: Interface defining the methods for handling prompts
// - `NewPrompts`: Function for creating a particular Prompts implementation based on the target language

const ProjectFilesWhitelistFileName = "project_files_whitelist.json"
const ProjectFilesBlacklistFileName = "project_files_blacklist.json"
const FileNameTagsRXFileName = "filename_tags_regexps.json"
const FileNameTagsFileName = "filename_tags.json"
const NoUploadCommentRXFileName = "no_upload_comment_regexps.json"
const OpImplementCommentRXFileName = "op_implement_comment_regexps.json"
const FileNameEmbedRXFileName = "filename_embed_regexp.json"
const OutputTagsRXFileName = "output_tags_regexps.json"
const ReasoningsTagsRXFileName = "reasonings_tags_regexps.json"
const ReasoningsTagsFileName = "reasonings_tags.json"
const SystemPromptFile = "system_prompt.txt"

// Annotate-operation prompt-filenames
const AnnotatePromptFile = "op_annotate_prompt.txt"
const AIAnnotateResponseFile = "op_annotate_ai_response.txt"

// Implement-optration stage1 prompt-filenames
const ImplementStage1ProjectIndexPromptFile = "op_implement_stage1_project_index_prompt.txt"
const AIImplementStage1ProjectIndexResponseFile = "op_implement_stage1_project_index_ai_response.txt"
const ImplementStage1SourceAnalysisPromptFile = "op_implement_stage1_source_analysis_prompt.txt"

// Implement-optration stage2 prompt-filenames
const ImplementStage2ProjectCodePromptFile = "op_implement_stage2_project_code_prompt.txt"
const AIImplementStage2ProjectCodeResponseFile = "op_implement_stage2_project_code_ai_response.txt"
const ImplementStage2FilesToChangePromptFile = "op_implement_stage2_files_to_change_prompt.txt"
const ImplementStage2FilesToChangeExtendedPromptFile = "op_implement_stage2_files_to_change_extended_prompt.txt"

const ImplementStage2NoPlanningPromptFile = "op_implement_stage2_no_planning_prompt.txt"
const AIImplementStage2NoPlanningResponseFile = "op_implement_stage2_no_planning_ai_response.txt"

// Implement-optration stage3 prompt-filenames
const ImplementStage3ChangesDonePromptFile = "op_implement_stage3_changes_done_prompt.txt"
const AIImplementStage3ChangesDoneResponseFile = "op_implement_stage3_changes_done_ai_response.txt"
const ImplementStage3ProcessFilePromptFile = "op_implement_stage3_process_file_prompt.txt"

const PromptsDir = "prompts"

type Prompts interface {
	// General helpers
	GetProjectFilesWhitelist() []string
	GetProjectFilesBlacklist() []string
	GetSystemPrompt() string

	// Annotate-operation prompts
	GetAnnotatePrompt() string
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

	// Implamane stage 2 no planning prompts
	GetImplementStage2NoPlanningPrompt() string
	GetAIImplementStage2NoPlanningResponse() string

	// Implement stage 3 prompts
	GetImplementStage3ChangesDonePrompt() string
	GetAIImplementStage3ChangesDoneResponse() string
	GetImplementStage3ProcessFilePrompt() string
}

// Create particular Prompts implementation depending on requested language
func NewPrompts(targetLang string) (Prompts, error) {
	targetLang = strings.ToUpper(targetLang)

	switch targetLang {
	case "GO":
		return &GoPrompts{}, nil
	case "DOTNETFW":
		return &DotNetFWPrompts{}, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", targetLang)
	}
}
