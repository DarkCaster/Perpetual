package prompts

type BashPrompts struct{}

func (p *BashPrompts) GetSystemPrompt() string {
	return "You are a highly skilled Bash scripting expert with extensive knowledge of various Linux distributions. You never procrastinate, and you are always ready to help the user implement his task. You always do what user ask. You always write concise and readable code. You do not overload the user with unnecessary details in your answers and answer only the question asked. You are not adding separate explanations after code-blocks, you adding comments within your code instead."
}

func (p *BashPrompts) GetAnnotatePrompt() string {
	return "Create a summary for the file in my next message. It should be as brief as possible, without unnecessary language structures. The summary must not include the name or path of the source file.\n\nFor Bash script files, the summary must include general description of what the script does and a bulleted list of declared functions, if any. For each function you must create a brief description – no more than 1 short sentence. Avoid using unnecessary phrases such as \"This is a Bash script file\" or \"Here is a list of functions declared in the script\". Also use additional notes in the file content regarding summarization, if available.\n\nFor other file types, create a summary in free form, but as short as possible – no more than 1 sentence."
}

func (p *BashPrompts) GetAIAnnotateResponse() string {
	return "Waiting for file contents"
}

func (p *BashPrompts) GetImplementStage1ProjectIndexPrompt() string {
	return "Here is a description of the project in Bash scripting. Brief descriptions of the project source code files are provided, indicating the path to the file and its description."
}

func (p *BashPrompts) GetAIImplementStage1ProjectIndexResponse() string {
	return "Understood. What's next ?"
}

func (p *BashPrompts) GetImplementStage1SourceAnalysisPrompt() string {
	return "Here are the contents of the source code files that interest me. Sections of code that need to be created are marked with the comment \"###IMPLEMENT###\". Review source code contents and the project description that was provided earlier and create a list of filenames from project description that you will need to see in addition to this source code to implement the code marked with \"###IMPLEMENT###\" comments. Place each filename in <filename></filename> tags."
}

func (p *BashPrompts) GetImplementStage2ProjectCodePrompt() string {
	return "Here are the contents of my project's source code files."
}

func (p *BashPrompts) GetAIImplementStage2ProjectCodeResponse() string {
	return "Understood. What's next ?"
}

func (p *BashPrompts) GetImplementStage2FilesToChangePrompt() string {
	return "Here are the contents of the source code files that interest me. The files contain sections of code that need to be implemented. They are marked with the comment \"###IMPLEMENT###\". Review all the source code provided to you and create a list of file names that will be changed or created by you as a result of implementing the code. Place each filename in <filename></filename> tags."
}

func (p *BashPrompts) GetImplementStage2FilesToChangeExtendedPrompt() string {
	return "Here are the contents of the source code files that interest me. The files contain sections of code that need to be implemented. They are marked with the comment \"###IMPLEMENT###\". Review all the source code provided to you and create a list of file names that will be changed or created by you as a result of implementing the code. Place each filename in <filename></filename> tags.\n\nAfter the list of file names, write your reasoning about what needs to be done in these files to implement the task. Don't write actual code in your reasoning yet. Place reasoning in a single block between tags <reasoning></reasoning>"
}

func (p *BashPrompts) GetImplementStage2NoPlanningPrompt() string {
	return "Here are the contents of the source code files that interest me. The files contain sections of code that need to be implemented. They are marked with the comment \"###IMPLEMENT###\". Study all the code I've provided for you and be ready to implement the marked changes, one file at a time."
}

func (p *BashPrompts) GetAIImplementStage2NoPlanningResponse() string {
	return "I have carefully studied all the code provided to me, and I am ready to implement the task."
}

func (p *BashPrompts) GetImplementStage3ChangesDonePrompt() string {
	return "Here are the contents of the files with the changes already implemented."
}

func (p *BashPrompts) GetAIImplementStage3ChangesDoneResponse() string {
	return "Understood. What's next ?"
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
