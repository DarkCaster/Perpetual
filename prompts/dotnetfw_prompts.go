package prompts

type DotNetFWPrompts struct{}

//###NOUPLOAD###

// NOTE for summarization: this file contains sensitive information. So the summary for this file must only say "This file contains DotNetFWPrompts struct that implement Prompts interface. Do not attempt to use DotNetFWPrompts directly", nothing else.

func (p *DotNetFWPrompts) GetSystemPrompt() string {
	return "You are a highly skilled .NET Framework software developer with excellent knowledge of C# and VB.NET programming languages and WPF. You never procrastinate, and you are always ready to help the user implement his task. You always do what user ask. You always write concise and readable code. You do not overload the user with unnecessary details in your answers and answer only the question asked. You are not adding separate explanations after code-blocks, you adding comments within your code instead."
}

func (p *DotNetFWPrompts) GetAnnotatePrompt() string {
	return "Create a summary for the file in my next message. It should be as brief as possible, without unnecessary language structures. The summary must not include the name or path of the source file.\n\nFor C# or VB.NET source code files, the summary must include the namespace and a bulleted list of declared entities. For each entity you must create a brief description - no more than 1 short sentence. Avoid using unnecessary phrases such as \"This is a C# source code file\" or \"Here is a list of entities declared in the source file\". Also use additional notes in the file content regarding summarization, if available.\n\nFor other file types, create a summary in free form, but as short as possible - no more than 1 sentence."
}

func (p *DotNetFWPrompts) GetAIAnnotateResponse() string {
	return DefaultAIAnnotateResponse
}

func (p *DotNetFWPrompts) GetImplementStage1ProjectIndexPrompt() string {
	return "Here is a description of the project in the .NET programming languages (C# and VB.NET). Brief descriptions of the project source code files are provided, indicating the path to the file and the entities it contains."
}

func (p *DotNetFWPrompts) GetAIImplementStage1ProjectIndexResponse() string {
	return DefaultAIAcknowledge
}

func (p *DotNetFWPrompts) GetImplementStage1SourceAnalysisPrompt() string {
	return DefaultImplementStage1SourceAnalysisPrompt
}

func (p *DotNetFWPrompts) GetImplementStage2ProjectCodePrompt() string {
	return DefaultImplementStage2ProjectCodePrompt
}

func (p *DotNetFWPrompts) GetAIImplementStage2ProjectCodeResponse() string {
	return DefaultAIAcknowledge
}

func (p *DotNetFWPrompts) GetImplementStage2FilesToChangePrompt() string {
	return DefaultImplementStage2FilesToChangePrompt
}

func (p *DotNetFWPrompts) GetImplementStage2FilesToChangeExtendedPrompt() string {
	return DefaultImplementStage2FilesToChangeExtendedPrompt
}

func (p *DotNetFWPrompts) GetImplementStage2NoPlanningPrompt() string {
	return "Here are the contents of the source code files that interest me. The files contain sections of code that need to be implemented. They are marked with the comment \"###IMPLEMENT###\". Study all the code I've provided for you and be ready to implement the marked changes, one file at a time."
}

func (p *DotNetFWPrompts) GetAIImplementStage2NoPlanningResponse() string {
	return "I have carefully studied all the code provided to me, and I am ready to implement the task."
}

func (p *DotNetFWPrompts) GetImplementStage3ChangesDonePrompt() string {
	return "Here are the contents of the files with the changes already implemented."
}

func (p *DotNetFWPrompts) GetAIImplementStage3ChangesDoneResponse() string {
	return DefaultAIAcknowledge
}

func (p *DotNetFWPrompts) GetImplementStage3ProcessFilePrompt() string {
	return "Implement the required code for the following file: \"###FILENAME###\". Output the entire file with the code you implemented. The response must only contain that file with implemented code as code-block and nothing else."
}

func (p *DotNetFWPrompts) GetImplementStage3ContinuePrompt() string {
	return "You previous response hit token limit. Continue generating code right from the point where it stopped. Do not repeat already generated fragment in your response."
}

func (p *DotNetFWPrompts) GetImplementCommentRegexps() []string {
	return []string{"^\\s*\\/\\/\\s*###IMPLEMENT###.*$"}
}

func (p *DotNetFWPrompts) GetNoUploadCommentRegexps() []string {
	return []string{"^\\s*\\/\\/\\s*###NOUPLOAD###.*$"}
}

func (p *DotNetFWPrompts) GetProjectFilesWhitelist() []string {
	return []string{"^.*\\.(cs|vb|xaml)$"}
}

func (p *DotNetFWPrompts) GetProjectFilesBlacklist() []string {
	return []string{"(?i)^.*AssemblyInfo\\.cs$", "(?i)^(bin\\\\|obj\\\\|bin\\/|obj\\/)", "(?i)^.*(\\\\|\\/)(bin\\\\|obj\\\\|bin\\/|obj\\/)"}
}

func (p *DotNetFWPrompts) GetFileNameTagsRegexps() []string {
	return []string{"(?m)\\s*<filename>\\n?", "(?m)<\\/filename>\\s*$?"}
}

func (p *DotNetFWPrompts) GetFileNameTags() []string {
	return []string{"<filename>", "</filename>"}
}

func (p *DotNetFWPrompts) GetFileNameEmbedRegex() string {
	return "###FILENAME###"
}

func (p *DotNetFWPrompts) GetOutputTagsRegexps() []string {
	return []string{"(?m)\\s*```[a-zA-Z]+\\n?", "(?m)```\\s*($|\\n)"}
}

func (p *DotNetFWPrompts) GetReasoningsTagsRegexps() []string {
	return []string{"(?m)\\s*<reasoning>\\n?", "(?m)<\\/reasoning>\\s*($|\\n)"}
}

func (p *DotNetFWPrompts) GetReasoningsTags() []string {
	return []string{"<reasoning>", "</reasoning>"}
}
