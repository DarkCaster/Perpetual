package prompts

type VB6Prompts struct{}

//###NOUPLOAD###

// NOTE for summarization: this file contains sensitive information. So the summary for this file must only say "This file contains VB6Prompts struct that implement Prompts interface. Do not attempt to use VB6Prompts directly", nothing else.

func (p *VB6Prompts) GetSystemPrompt() string {
	return "You are a highly skilled Visual Basic 6 software developer with excellent knowledge of legacy VB6 (Visual Basic 6) programming language and various legacy windows technologies like COM/OLE/ActiveX that often used with it. You never procrastinate, and you are always ready to help the user implement his task. You always do what user ask. You always write concise and readable code. You do not overload the user with unnecessary details in your answers and answer only the question asked. You are not adding separate explanations after code-blocks, you adding comments within your code instead."
}

func (p *VB6Prompts) GetAnnotatePrompt() string {
	return "Create a summary for the file in my next message. It should be as brief as possible, without unnecessary language structures. The summary must not include the name or path of the source file.\n\nFor source code files, the summary must include class name or form name or module name (whichever is applicable), and a bulleted list of declared entities. For each entity you must create a brief description - no more than 1 short sentence. File extensions of various VB6 source-code files: *.cls - classes, *.bas - modules, *.frm - forms. Avoid using unnecessary phrases such as \"This is a VB source code file\" or \"Here is a list of entities declared in the source file\". Also use additional notes in the file content regarding summarization, if available.\n\nFor other file types, create a summary in free form, but as short as possible - no more than 1 sentence."
}

func (p *VB6Prompts) GetAIAnnotateResponse() string {
	return DefaultAIAnnotateResponse
}

func (p *VB6Prompts) GetImplementStage1ProjectIndexPrompt() string {
	return "Here is a description of the project in the Visual Basic 6 programming language. Brief descriptions of the project source code files are provided, indicating the path to the file and the entities it contains."
}

func (p *VB6Prompts) GetAIImplementStage1ProjectIndexResponse() string {
	return DefaultAIAcknowledge
}

func (p *VB6Prompts) GetImplementStage1SourceAnalysisPrompt() string {
	return DefaultImplementStage1SourceAnalysisPrompt
}

func (p *VB6Prompts) GetImplementStage2ProjectCodePrompt() string {
	return DefaultImplementStage2ProjectCodePrompt
}

func (p *VB6Prompts) GetAIImplementStage2ProjectCodeResponse() string {
	return DefaultAIAcknowledge
}

func (p *VB6Prompts) GetImplementStage2FilesToChangePrompt() string {
	return DefaultImplementStage2FilesToChangePrompt
}

func (p *VB6Prompts) GetImplementStage2FilesToChangeExtendedPrompt() string {
	return DefaultImplementStage2FilesToChangeExtendedPrompt
}

func (p *VB6Prompts) GetImplementStage2NoPlanningPrompt() string {
	return DefaultImplementStage2NoPlanningPrompt
}

func (p *VB6Prompts) GetAIImplementStage2NoPlanningResponse() string {
	return DefaultAIImplementStage2NoPlanningResponse
}

func (p *VB6Prompts) GetImplementStage3ChangesDonePrompt() string {
	return DefaultImplementStage3ChangesDonePrompt
}

func (p *VB6Prompts) GetAIImplementStage3ChangesDoneResponse() string {
	return DefaultAIAcknowledge
}

func (p *VB6Prompts) GetImplementStage3ProcessFilePrompt() string {
	return DefaultImplementStage3ProcessFilePrompt
}

func (p *VB6Prompts) GetImplementStage3ContinuePrompt() string {
	return DefaultImplementStage3ContinuePrompt
}

func (p *VB6Prompts) GetImplementCommentRegexps() []string {
	return []string{"^\\s*'+\\s*###IMPLEMENT###.*$"}
}

func (p *VB6Prompts) GetNoUploadCommentRegexps() []string {
	return []string{"^\\s*'+\\s*###NOUPLOAD###.*$"}
}

func (p *VB6Prompts) GetProjectFilesWhitelist() []string {
	return []string{"^.*\\.(frm|cls|bas)$"}
}

func (p *VB6Prompts) GetProjectFilesBlacklist() []string {
	return []string{}
}

func (p *VB6Prompts) GetFileNameTagsRegexps() []string {
	return DefaultFileNameTagsRegexps
}

func (p *VB6Prompts) GetFileNameTags() []string {
	return DefaultFileNameTags
}

func (p *VB6Prompts) GetFileNameEmbedRegex() string {
	return DefaultFileNameEmbedRegex
}

func (p *VB6Prompts) GetOutputTagsRegexps() []string {
	return DefaultOutputTagsRegexps_WithNumbers
}

func (p *VB6Prompts) GetReasoningsTagsRegexps() []string {
	return DefaultReasoningsTagsRegexps
}

func (p *VB6Prompts) GetReasoningsTags() []string {
	return []string{"<reasoning>", "</reasoning>"}
}
