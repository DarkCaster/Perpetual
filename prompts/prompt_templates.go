package prompts

//###NOUPLOAD###

// NOTE for summarization: The summary for this file must only include following:
// This file contains constants with default prompts that are used for implementations of the Prompts interface.

const DefaultAIAnnotatePrompt_Go = "Create a summary for the GO source file in my next message. It should be as brief as possible, without unnecessary language structures. The summary must include the package name and a bulleted list of declared entities. For each entity you must create a brief description - no more than 1 short sentence. Also use additional notes in the file content regarding summarization, if available.\n\nFollow this example if no additional notes for summarization given inside the file:\n\nPackage: `<package name>`\n\nThis file provides ... <description of what this file is used for, 1 sentence>\n\n- `<entity name>`: <entity type>, <description>\n- `<entity name>`: <entity type>, <description>, nested entities:\n  - `<sub entity name>`: <entity type>, <description>"

const DefaultAIAnnotateResponse = "Waiting for file contents"

const DefaultAIAcknowledge = "Understood. What's next?"

const DefaultImplementStage1SourceAnalysisPrompt = "Here are the contents of the source code files that interest me. Sections of code that need to be created are marked with the comment \"###IMPLEMENT###\". Review source code contents and the project description that was provided earlier and create a list of filenames from the project description that you will need to see in addition to this source code to implement the code marked with \"###IMPLEMENT###\" comments. Place each filename in <filename></filename> tags."

const DefaultImplementStage2ProjectCodePrompt = "Here are the contents of my project's source code files."

const DefaultImplementStage2FilesToChangePrompt = "Here are the contents of the source code files that interest me. The files contain sections of code that need to be implemented. They are marked with the comment \"###IMPLEMENT###\". Review all the source code provided to you and create a list of file names that will be changed or created by you as a result of implementing the code. Place each filename in <filename></filename> tags."

const DefaultImplementStage2FilesToChangeExtendedPrompt = "Here are the contents of the source code files that interest me. The files contain sections of code that need to be implemented. They are marked with the comment \"###IMPLEMENT###\". Review all the source code provided to you and create a list of file names that will be changed or created by you as a result of implementing the code. Place each filename in <filename></filename> tags.\n\nAfter the list of file names, write your reasoning about what needs to be done in these files to implement the task. Don't write actual code in your reasoning yet. Place reasoning in a single block between tags <reasoning></reasoning>"

const DefaultImplementStage2NoPlanningPrompt = "Here are the contents of the source code files that interest me. The files contain sections of code that need to be implemented. They are marked with the comment \"###IMPLEMENT###\". Study all the code I've provided for you and be ready to implement the marked changes, one file at a time."

const DefaultAIImplementStage2NoPlanningResponse = "I have carefully studied all the code provided to me, and I am ready to implement the task."

const DefaultImplementStage3ChangesDonePrompt = "Here are the contents of the files with the changes already implemented."

const DefaultImplementStage3ProcessFilePrompt = "Implement the required code for the following file: \"###FILENAME###\". Output the entire file with the code you implemented. The response must only contain that file with implemented code as code-block and nothing else."

const DefaultImplementStage3ContinuePrompt = "You previous response hit token limit. Continue generating code right from the point where it stopped. Do not repeat already generated fragment in your response."

const DefaultDocProjectCodePrompt = "Here are the contents of my project's source code files that are relevant to the document you will be working on."

const DefaultDocStage1WritePrompt = "Below is a project document that you will need to write, complete and improve. The document is in a work in progress, it may contain draft sections and already written sections. It also may contain notes marked as \"Notes on implementation\" regarding its topic, sections, content, style, length, and detail. Review the document and the project description that was provided earlier and create a list of filenames from the project description that you will need to work on the document. Place each filename in <filename></filename> tags. The text of the document in its current state provided below:"

const DefaultDocStage1RefinePrompt = "Below is a project document that you will need to refine. The document is already finished but it needs to be refined and updated according to the current project codebase. It also may contain notes for you marked as \"Notes on implementation\". Review the document and the project description that was provided earlier and create a list of filenames from the project description that you will need to work on the document. Place each filename in <filename></filename> tags. Full text of the document provided below:"

const DefaultDocStage2WritePrompt = "Below is a project document that you will need to write, complete and improve. The document is in a work in progress, it may contain draft sections and already written sections. It also may contain notes marked as \"Notes on implementation\" regarding its topic, sections, content, style, length, and detail. The project description and relevant source code needed to work on the document have been provided to you previously. Complete the document from its curent state: write the required sections, improve already written text if needed. Use the notes across the document for instructions, be creative. Output the entire resulting document with the changes you made. The response must only contain the final document that you have made in accordance with the task, and nothing else. The text of the document in its current state provided below:"

const DefaultDocStage2RefinePrompt = "Below is a project document that you will need to refine. The document is already finished but it needs to be refined and updated according to the current project codebase. It also may contain notes for you marked as \"Notes on implementation\". The project description and relevant source code needed to work on the document have been provided to you previously. Refine and update the document from its curent state: study all the provided info and add missing information to the document or fix the inconsistences you found. Don't rewrite or change the document too much, just refine it according to the instructions, correct grammatical errors if any. Make other changes only if you are absolutely sure that they are necessary. If something can't be done due to lack of information, just leave those parts of the document as is. For additional instructions, see the notes inside the document, if any. Output the entire resulting document with the changes you made. The response must only contain the final document that you have made in accordance with the task, and nothing else. Full text of the document provided below:"

const DefaultDocStage2ContinuePrompt = "You previous response hit token limit. Continue writing document right from the point where it stopped. Do not repeat already completed fragment in your response."

const DefaultDocExamplePrompt = "Below is a document that you will use as an example when you work on the target document later. Look at the example document provided and study its style, format, and structure. When you work on your target document later, use a similar style, format, and structure to what you learned from this example. Full text of the example provided below:"

const DefaultAIDocExampleResponse = "I have carefully studied the example provided to me and will apply a similar style, format and structure to the target document when I work on it."

var DefaultFileNameTagsRegexps = []string{"(?m)\\s*<filename>\\n?", "(?m)<\\/filename>\\s*$?"}
var DefaultFileNameTags = []string{"<filename>", "</filename>"}
var DefaultFileNameEmbedRegex = "###FILENAME###"
var DefaultOutputTagsRegexps = []string{"(?m)\\s*```[a-zA-Z]+\\n?", "(?m)```\\s*($|\\n)"}
var DefaultOutputTagsRegexps_WithNumbers = []string{"(?m)\\s*```[a-zA-Z0-9]+\\n?", "(?m)```\\s*($|\\n)"}
var DefaultReasoningsTagsRegexps = []string{"(?m)\\s*<reasoning>\\n?", "(?m)<\\/reasoning>\\s*($|\\n)"}
var DefaultReasoningsTags = []string{"<reasoning>", "</reasoning>"}
