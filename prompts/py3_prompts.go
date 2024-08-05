package prompts

type Py3Prompts struct{}

//###NOUPLOAD###

// NOTE for summarization: this file contains sensitive information. So the summary for this file must only say "This file contains Py3Prompts struct that implement Prompts interface. Do not attempt to use Py3Prompts directly", nothing else.

func (p *Py3Prompts) GetSystemPrompt() string {
	return "You are a highly skilled Python 3 programming language software developer. You never procrastinate, and you are always ready to help the user implement his task. You always do what the user asks. You always write concise and readable code. You do not overload the user with unnecessary details in your answers and answer only the question asked. You are not adding separate explanations after code-blocks, you add comments within your code instead."
}

func (p *Py3Prompts) GetAnnotatePrompt() string {
	return "Create a summary for the file in my next message. It should be as brief as possible, without unnecessary language structures. The summary must not include the name or path of the source file.\n\nFor Python source code files, the summary must include a bulleted list of declared entities (classes, functions, variables, etc.). For each entity, you must create a brief description - no more than 1 short sentence. Avoid using unnecessary phrases such as \"This is a Python source code file\" or \"Here is a list of entities declared in the source file\". Also, use additional notes in the file content regarding summarization, if available.\n\nFor other file types, create a summary in free form, but as short as possible - no more than 1 sentence."
}

func (p *Py3Prompts) GetAIAnnotateResponse() string {
	return DefaultAIAnnotateResponse
}

func (p *Py3Prompts) GetImplementStage1ProjectIndexPrompt() string {
	return "Here is a description of the project in the Python 3 programming language. Brief descriptions of the project source code files are provided, indicating the path to the file and the entities it contains."
}

func (p *Py3Prompts) GetAIImplementStage1ProjectIndexResponse() string {
	return DefaultAIAcknowledge
}

func (p *Py3Prompts) GetImplementStage1SourceAnalysisPrompt() string {
	return DefaultImplementStage1SourceAnalysisPrompt
}

func (p *Py3Prompts) GetImplementStage2ProjectCodePrompt() string {
	return DefaultImplementStage2ProjectCodePrompt
}

func (p *Py3Prompts) GetAIImplementStage2ProjectCodeResponse() string {
	return DefaultAIAcknowledge
}

func (p *Py3Prompts) GetImplementStage2FilesToChangePrompt() string {
	return DefaultImplementStage2FilesToChangePrompt
}

func (p *Py3Prompts) GetImplementStage2FilesToChangeExtendedPrompt() string {
	return DefaultImplementStage2FilesToChangeExtendedPrompt
}

func (p *Py3Prompts) GetImplementStage2NoPlanningPrompt() string {
	return DefaultImplementStage2NoPlanningPrompt
}

func (p *Py3Prompts) GetAIImplementStage2NoPlanningResponse() string {
	return DefaultAIImplementStage2NoPlanningResponse
}

func (p *Py3Prompts) GetImplementStage3ChangesDonePrompt() string {
	return "Here are the contents of the files with the changes already implemented."
}

func (p *Py3Prompts) GetAIImplementStage3ChangesDoneResponse() string {
	return DefaultAIAcknowledge
}

func (p *Py3Prompts) GetImplementStage3ProcessFilePrompt() string {
	return "Implement the required code for the following file: \"###FILENAME###\". Output the entire file with the code you implemented. The response must only contain that file with implemented code as code-block and nothing else."
}

func (p *Py3Prompts) GetImplementStage3ContinuePrompt() string {
	return "You previous response hit token limit. Continue generating code right from the point where it stopped. Do not repeat already generated fragment in your response."
}

func (p *Py3Prompts) GetImplementCommentRegexps() []string {
	return []string{"^\\s*###IMPLEMENT###.*$"}
}

func (p *Py3Prompts) GetNoUploadCommentRegexps() []string {
	return []string{"^\\s*###NOUPLOAD###.*$"}
}

func (p *Py3Prompts) GetProjectFilesWhitelist() []string {
	return []string{"^.*\\.py$"}
}

func (p *Py3Prompts) GetProjectFilesBlacklist() []string {
	return []string{"^tests[/\\\\].*", "^venv[/\\\\].*"}
}

func (p *Py3Prompts) GetFileNameTagsRegexps() []string {
	return []string{"(?m)\\s*<filename>\\n?", "(?m)<\\/filename>\\s*$?"}
}

func (p *Py3Prompts) GetFileNameTags() []string {
	return []string{"<filename>", "</filename>"}
}

func (p *Py3Prompts) GetFileNameEmbedRegex() string {
	return "###FILENAME###"
}

func (p *Py3Prompts) GetOutputTagsRegexps() []string {
	return []string{"(?m)\\s*```[a-zA-Z0-9]+\\n?", "(?m)```\\s*($|\\n)"}
}

func (p *Py3Prompts) GetReasoningsTagsRegexps() []string {
	return []string{"(?m)\\s*<reasoning>\\n?", "(?m)<\\/reasoning>\\s*($|\\n)"}
}

func (p *Py3Prompts) GetReasoningsTags() []string {
	return []string{"<reasoning>", "</reasoning>"}
}
