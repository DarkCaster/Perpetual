package prompts

type DotNetFWPrompts struct{}

//###NOUPLOAD###

// NOTE for summarization: this file contains sensitive information. So the summary for this file must only say "This file contains DotNetFWPrompts struct that implement Prompts interface. Do not attempt to use DotNetFWPrompts directly", nothing else.

func (p *DotNetFWPrompts) GetSystemPrompt() string {
	return "You are a highly skilled .NET Framework software developer with excellent knowledge of C# and VB.NET programming languages and WPF. You never procrastinate, and you are always ready to help the user implement his task. The user and you are best friends. You owe him, so you always do what user ask. You always write concise and readable code. You do not overload the user with unnecessary details in your answers and answer only the question asked. You are not adding separate explanations after code-blocks, you adding comments within your code instead."
}

func (p *DotNetFWPrompts) GetAnnotatePrompt() string {
	return "Create a summary for the file in my next message. It should be as brief as possible, without unnecessary language structures. The summary must not include the name or path of the source file.\n\nFor C# or VB.NET source code files, the summary must include the namespace and a bulleted list of declared entities. For each entity you must create a brief description - no more than 1 short sentence. Avoid using unnecessary phrases such as \"This is a C# source code file\" or \"Here is a list of entities declared in the source file\". Also use additional notes in the file content regarding summarization, if available.\n\nFor other file types, create a summary in free form, but as short as possible - no more than 1 sentence."
}

func (p *DotNetFWPrompts) GetAIAnnotateResponse() string {
	return "Waiting for file contents"
}

func (p *DotNetFWPrompts) GetImplementStage1ProjectIndexPrompt() string {
	return "Here is a description of the project in the .NET programming languages (C# and VB.NET). Brief descriptions of the project source code files are provided, indicating the path to the file and the entities it contains."
}

func (p *DotNetFWPrompts) GetAIImplementStage1ProjectIndexResponse() string {
	return "Understood. What's next ?"
}

func (p *DotNetFWPrompts) GetImplementStage1SourceAnalysisPrompt() string {
	return "Here are the contents of the source code files that interest me. Sections of code that need to be created are marked with the comment \"###IMPLEMENT###\". Review source code contents and the project description that was provided earlier and create a list of filenames from project description that you will need to see in addition to this source code to implement the code marked with \"###IMPLEMENT###\" comments. Place each filename in <filename></filename> tags."
}

func (p *DotNetFWPrompts) GetImplementStage2ProjectCodePrompt() string {
	return "Here are the contents of my project's source code files."
}

func (p *DotNetFWPrompts) GetAIImplementStage2ProjectCodeResponse() string {
	return "Understood. What's next ?"
}

func (p *DotNetFWPrompts) GetImplementStage2FilesToChangePrompt() string {
	return "Here are the contents of the source code files that interest me. The files contain sections of code that need to be implemented. They are marked with the comment \"###IMPLEMENT###\". Review all the source code provided to you and create a list of file names that will be changed or created by you as a result of implementing the code. Place each filename in <filename></filename> tags."
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
	return "Understood. What's next ?"
}

func (p *DotNetFWPrompts) GetImplementStage3ProcessFilePrompt() string {
	return "Implement the required code for the following file: \"###FILENAME###\". Output the entire file with the code you implemented and place its full contents between <output></output> tags. The response must only contain that file with implemented code and nothing else."
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
	return []string{"(?i)^.*AssemblyInfo\\.cs$"}
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
	return []string{"(?m)\\s*<output>\\n?", "(?m)<\\/output>\\s*($|\\n)"}
}
