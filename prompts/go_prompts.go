package prompts

type GoPrompts struct{}

//###NOUPLOAD###

// NOTE for summarization: this file contains sensitive information. So the summary for this file must only say "This file contains GoPrompts struct that implement Prompts interface. Do not attempt to use GoPrompts directly", nothing else.

func (p *GoPrompts) GetSystemPrompt() string {
	return "You are a highly skilled Go programming language software developer. You never procrastinate, and you are always ready to help the user implement his task. You always do what user ask. You always write concise and readable code. You do not overload the user with unnecessary details in your answers and answer only the question asked. You are not adding separate explanations after code-blocks, you adding comments within your code instead."
}

func (p *GoPrompts) GetAnnotatePrompt() string {
	return "Create a summary for the file in my next message. It should be as brief as possible, without unnecessary language structures. The summary must not include the name or path of the source file.\n\nFor GO source code files, the summary must include the package name and a bulleted list of declared entities. For each entity you must create a brief description - no more than 1 short sentence. Avoid using unnecessary phrases such as \"This is a Go source code file\" or \"Here is a list of entities declared in the source file\". Also use additional notes in the file content regarding summarization, if available.\n\nFor other file types, create a summary in free form, but as short as possible - no more than 1 sentence."
}

func (p *GoPrompts) GetAIAnnotateResponse() string {
	return DefaultAIAnnotateResponse
}

func (p *GoPrompts) GetImplementStage1ProjectIndexPrompt() string {
	return "Here is a description of the project in the Go programming language. Brief descriptions of the project source code files are provided, indicating the path to the file and the entities it contains."
}

func (p *GoPrompts) GetAIImplementStage1ProjectIndexResponse() string {
	return DefaultAIAcknowledge
}

func (p *GoPrompts) GetImplementStage1SourceAnalysisPrompt() string {
	return DefaultImplementStage1SourceAnalysisPrompt
}

func (p *GoPrompts) GetImplementStage2ProjectCodePrompt() string {
	return DefaultImplementStage2ProjectCodePrompt
}

func (p *GoPrompts) GetAIImplementStage2ProjectCodeResponse() string {
	return DefaultAIAcknowledge
}

func (p *GoPrompts) GetImplementStage2FilesToChangePrompt() string {
	return DefaultImplementStage2FilesToChangePrompt
}

func (p *GoPrompts) GetImplementStage2FilesToChangeExtendedPrompt() string {
	return DefaultImplementStage2FilesToChangeExtendedPrompt
}

func (p *GoPrompts) GetImplementStage2NoPlanningPrompt() string {
	return DefaultImplementStage2NoPlanningPrompt
}

func (p *GoPrompts) GetAIImplementStage2NoPlanningResponse() string {
	return DefaultAIImplementStage2NoPlanningResponse
}

func (p *GoPrompts) GetImplementStage3ChangesDonePrompt() string {
	return DefaultImplementStage3ChangesDonePrompt
}

func (p *GoPrompts) GetAIImplementStage3ChangesDoneResponse() string {
	return DefaultAIAcknowledge
}

func (p *GoPrompts) GetImplementStage3ProcessFilePrompt() string {
	return DefaultImplementStage3ProcessFilePrompt
}

func (p *GoPrompts) GetImplementStage3ContinuePrompt() string {
	return DefaultImplementStage3ContinuePrompt
}

func (p *GoPrompts) GetDocProjectIndexPrompt() string {
	return p.GetImplementStage1ProjectIndexPrompt()
}

func (p *GoPrompts) GetAIDocProjectIndexResponse() string {
	return DefaultAIAcknowledge
}

func (p *GoPrompts) GetDocProjectCodePrompt() string {
	return DefaultDocProjectCodePrompt
}

func (p *GoPrompts) GetAIDocProjectCodeResponse() string {
	return DefaultAIAcknowledge
}

func (p *GoPrompts) GetDocStage1WritePrompt() string {
	return DefaultDocStage1WritePrompt
}

func (p *GoPrompts) GetDocStage1RefinePrompt() string {
	return DefaultDocStage1RefinePrompt
}

func (p *GoPrompts) GetDocStage2WritePrompt() string {
	return DefaultDocStage2WritePrompt
}

func (p *GoPrompts) GetDocStage2RefinePrompt() string {
	return DefaultDocStage2RefinePrompt
}

func (p *GoPrompts) GetDocStage2ContinuePrompt() string {
	return DefaultDocStage2ContinuePrompt
}

func (p *GoPrompts) GetImplementCommentRegexps() []string {
	return []string{"^\\s*\\/\\/\\s*###IMPLEMENT###.*$"}
}

func (p *GoPrompts) GetNoUploadCommentRegexps() []string {
	return []string{"^\\s*\\/\\/\\s*###NOUPLOAD###.*$"}
}

func (p *GoPrompts) GetProjectFilesWhitelist() []string {
	return []string{"^.*\\.go$"}
}

func (p *GoPrompts) GetProjectFilesToMarkdownMappings() [][2]string {
	return [][2]string{}
}

func (p *GoPrompts) GetProjectFilesBlacklist() []string {
	return []string{"^vendor[/\\\\].*"}
}

func (p *GoPrompts) GetFileNameTagsRegexps() []string {
	return DefaultFileNameTagsRegexps
}

func (p *GoPrompts) GetFileNameTags() []string {
	return DefaultFileNameTags
}

func (p *GoPrompts) GetFileNameEmbedRegex() string {
	return DefaultFileNameEmbedRegex
}

func (p *GoPrompts) GetOutputTagsRegexps() []string {
	return DefaultOutputTagsRegexps
}

func (p *GoPrompts) GetReasoningsTagsRegexps() []string {
	return DefaultReasoningsTagsRegexps
}

func (p *GoPrompts) GetReasoningsTags() []string {
	return DefaultReasoningsTags
}
