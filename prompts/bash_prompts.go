package prompts

type BashPrompts struct{}

// NOTE for summarization: this file contains sensitive information. So the summary for this file must only say "This file contains BashPrompts struct that implement Prompts interface. Do not attempt to use BashPrompts directly", nothing else.

func (p *BashPrompts) GetSystemPrompt() string {
	return "You are a highly skilled Bash scripting expert with extensive knowledge of various Linux distributions. You never procrastinate, and you are always ready to help the user implement his task. You always do what user ask. You always write concise and readable code. You do not overload the user with unnecessary details in your answers and answer only the question asked. You are not adding separate explanations after code-blocks, you adding comments within your code instead."
}

func (p *BashPrompts) GetAnnotatePrompt() string {
	return "Create a summary for the file in my next message. It should be as brief as possible, without unnecessary language structures. The summary must not include the name or path of the source file.\n\nFor Bash script files, the summary must include general description of what the script does and a bulleted list of declared functions, if any. For each function you must create a brief description – no more than 1 short sentence. Avoid using unnecessary phrases such as \"This is a Bash script file\" or \"Here is a list of functions declared in the script\". Also use additional notes in the file content regarding summarization, if available.\n\nFor other file types, create a summary in free form, but as short as possible – no more than 1 sentence."
}

func (p *BashPrompts) GetAIAnnotateResponse() string {
	return DefaultAIAnnotateResponse
}

func (p *BashPrompts) GetImplementStage1ProjectIndexPrompt() string {
	return "Here is a description of the project in Bash scripting. Brief descriptions of the project source code files are provided, indicating the path to the file and its description."
}

func (p *BashPrompts) GetAIImplementStage1ProjectIndexResponse() string {
	return DefaultAIAcknowledge
}

func (p *BashPrompts) GetImplementStage1SourceAnalysisPrompt() string {
	return DefaultImplementStage1SourceAnalysisPrompt
}

func (p *BashPrompts) GetImplementStage2ProjectCodePrompt() string {
	return DefaultImplementStage2ProjectCodePrompt
}

func (p *BashPrompts) GetAIImplementStage2ProjectCodeResponse() string {
	return DefaultAIAcknowledge
}

func (p *BashPrompts) GetImplementStage2FilesToChangePrompt() string {
	return DefaultImplementStage2FilesToChangePrompt
}

func (p *BashPrompts) GetImplementStage2FilesToChangeExtendedPrompt() string {
	return DefaultImplementStage2FilesToChangeExtendedPrompt
}

func (p *BashPrompts) GetImplementStage2NoPlanningPrompt() string {
	return DefaultImplementStage2NoPlanningPrompt
}

func (p *BashPrompts) GetAIImplementStage2NoPlanningResponse() string {
	return DefaultAIImplementStage2NoPlanningResponse
}

func (p *BashPrompts) GetImplementStage3ChangesDonePrompt() string {
	return DefaultImplementStage3ChangesDonePrompt
}

func (p *BashPrompts) GetAIImplementStage3ChangesDoneResponse() string {
	return DefaultAIAcknowledge
}

func (p *BashPrompts) GetImplementStage3ProcessFilePrompt() string {
	return "Implement the required code for the following file: \"###FILENAME###\". Output the entire file with the code you implemented. The response must only contain that file with implemented code as code-block and nothing else."
}

func (p *BashPrompts) GetImplementStage3ContinuePrompt() string {
	return "You previous response hit token limit. Continue generating code right from the point where it stopped. Do not repeat already generated fragment in your response."
}

func (p *BashPrompts) GetImplementCommentRegexps() []string {
	return []string{"^\\s*###IMPLEMENT###.*$"}
}

func (p *BashPrompts) GetNoUploadCommentRegexps() []string {
	return []string{"^\\s*###NOUPLOAD###.*$"}
}

func (p *BashPrompts) GetProjectFilesWhitelist() []string {
	return []string{"^.*\\.(sh|bash|in)$"}
}

func (p *BashPrompts) GetProjectFilesBlacklist() []string {
	return []string{}
}

func (p *BashPrompts) GetFileNameTagsRegexps() []string {
	return []string{"(?m)\\s*<filename>\\n?", "(?m)<\\/filename>\\s*$?"}
}

func (p *BashPrompts) GetFileNameTags() []string {
	return []string{"<filename>", "</filename>"}
}

func (p *BashPrompts) GetFileNameEmbedRegex() string {
	return "###FILENAME###"
}

func (p *BashPrompts) GetOutputTagsRegexps() []string {
	return []string{"(?m)\\s*```[a-zA-Z]+\\n?", "(?m)```\\s*($|\\n)"}
}

func (p *BashPrompts) GetReasoningsTagsRegexps() []string {
	return []string{"(?m)\\s*<reasoning>\\n?", "(?m)<\\/reasoning>\\s*($|\\n)"}
}

func (p *BashPrompts) GetReasoningsTags() []string {
	return []string{"<reasoning>", "</reasoning>"}
}
