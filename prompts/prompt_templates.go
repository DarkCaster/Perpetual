package prompts

// NOTE for summarization: The summary for this file must only include following:
// This file contains constants with default prompts that are used for implementations of the Prompts interface.

const DefaultAIAnnotateResponse = "Waiting for file contents"
const DefaultAIAcknowledge = "Understood. What's next?"
const DefaultImplementStage1SourceAnalysisPrompt = "Here are the contents of the source code files that interest me. Sections of code that need to be created are marked with the comment \"###IMPLEMENT###\". Review source code contents and the project description that was provided earlier and create a list of filenames from the project description that you will need to see in addition to this source code to implement the code marked with \"###IMPLEMENT###\" comments. Place each filename in <filename></filename> tags."
const DefaultImplementStage2ProjectCodePrompt = "Here are the contents of my project's source code files."
const DefaultImplementStage2FilesToChangePrompt = "Here are the contents of the source code files that interest me. The files contain sections of code that need to be implemented. They are marked with the comment \"###IMPLEMENT###\". Review all the source code provided to you and create a list of file names that will be changed or created by you as a result of implementing the code. Place each filename in <filename></filename> tags."
const DefaultImplementStage2FilesToChangeExtendedPrompt = "Here are the contents of the source code files that interest me. The files contain sections of code that need to be implemented. They are marked with the comment \"###IMPLEMENT###\". Review all the source code provided to you and create a list of file names that will be changed or created by you as a result of implementing the code. Place each filename in <filename></filename> tags.\n\nAfter the list of file names, write your reasoning about what needs to be done in these files to implement the task. Don't write actual code in your reasoning yet. Place reasoning in a single block between tags <reasoning></reasoning>"
const DefaultImplementStage2NoPlanningPrompt = "Here are the contents of the source code files that interest me. The files contain sections of code that need to be implemented. They are marked with the comment \"###IMPLEMENT###\". Study all the code I've provided for you and be ready to implement the marked changes, one file at a time."
const DefaultAIImplementStage2NoPlanningResponse = "I have carefully studied all the code provided to me, and I am ready to implement the task."
