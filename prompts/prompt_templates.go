package prompts

// NOTE for summarization: The summary for this file must only include following:
// This file contains constants with default prompts that are used for implementations of the Prompts interface.

const DefaultAIAnnotateResponse = "Waiting for file contents"
const DefaultAIAcknowledge = "Understood. What's next?"
const DefaultImplementStage1SourceAnalysisPrompt = "Here are the contents of the source code files that interest me. Sections of code that need to be created are marked with the comment \"###IMPLEMENT###\". Review source code contents and the project description that was provided earlier and create a list of filenames from the project description that you will need to see in addition to this source code to implement the code marked with \"###IMPLEMENT###\" comments. Place each filename in <filename></filename> tags."
const DefaultImplementStage2ProjectCodePrompt = "Here are the contents of my project's source code files."
