package op_doc

import (
	"path/filepath"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/prompts"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage1(projectRootDir string, perpetualDir string, promptsDir string, systemPrompt string, filesToMdLangMappings [][2]string, fileNameTagsRxStrings []string, fileNameTags []string, projectFiles []string, annotations map[string]string, targetDocument string, exampleDocuemnt string, action string, logger logging.ILogger) []string {

	// Add trace and debug logging
	logger.Traceln("Stage1: Starting")
	defer logger.Traceln("Stage1: Finished")

	// Create stage1 llm connector
	connector, err := llm.NewLLMConnector(OpName+"_stage1", systemPrompt, filesToMdLangMappings, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage1 LLM connector:", err)
	}
	logger.Debugln(llm.GetDebugString(connector))

	loadPrompt := func(filePath string) string {
		text, err := utils.LoadTextFile(filepath.Join(promptsDir, filePath))
		if err != nil {
			logger.Panicln("Failed to load prompt:", err)
		}
		return text
	}

	var messages []llm.Message
	// Create project-index request message
	projectIndexRequestMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(prompts.DocProjectIndexPromptFile))
	for _, item := range projectFiles {
		projectIndexRequestMessage = llm.AddIndexFragment(projectIndexRequestMessage, item, fileNameTags)
		annotation := annotations[item]
		if annotation == "" {
			annotation = "No annotation available"
		}
		projectIndexRequestMessage = llm.AddPlainTextFragment(projectIndexRequestMessage, annotation)
	}
	messages = append(messages, projectIndexRequestMessage)
	logger.Debugln("Created project-index request message")

	// Create project-index simulated response
	projectIndexResponseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), loadPrompt(prompts.AIDocProjectIndexResponseFile))
	messages = append(messages, projectIndexResponseMessage)
	logger.Debugln("Created project-index simulated response message")

	if exampleDocuemnt != "" {
		// Create document-example request message
		docExampleRequestMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(prompts.DocExamplePromptFile))
		exampleContents, err := utils.LoadTextFile(filepath.Join(projectRootDir, exampleDocuemnt))
		if err != nil {
			logger.Panicln("Failed to load example document:", err)
		}
		docExampleRequestMessage = llm.AddPlainTextFragment(docExampleRequestMessage, exampleContents)
		messages = append(messages, docExampleRequestMessage)
		logger.Debugln("Created document-example request message")

		// Create document-example simulated response
		docExampleResponseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), loadPrompt(prompts.AIDocExampleResponseFile))
		messages = append(messages, docExampleResponseMessage)
		logger.Debugln("Created document-example simulated response message")
	}

	// Create project-files analisys request message
	promptFile := prompts.DocStage1WritePromptFile
	if action == "REFINE" {
		promptFile = prompts.DocStage1RefinePromptFile
	}

	codeAnalysisRequestMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(promptFile))
	contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, targetDocument))
	if err != nil {
		logger.Panicln("failed to add file contents to stage1 prompt", err)
	}
	codeAnalysisRequestMessage = llm.AddPlainTextFragment(codeAnalysisRequestMessage, contents)
	messages = append(messages, codeAnalysisRequestMessage)
	logger.Debugln("Created code-analysis request message")

	// Perform LLM query
	var aiResponses []string
	onFailRetriesLeft := connector.GetOnFailureRetryLimit()
	if onFailRetriesLeft < 1 {
		onFailRetriesLeft = 1
	}
	for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
		logger.Infoln("Running stage1: find project files for review")
		var status llm.QueryStatus
		//NOTE: do not use := here, looks like it will make copy of aiResponse, and effectively result in empty file-list (tested on golang 1.22.3)
		aiResponses, status, err = connector.Query(1, messages...)
		if err != nil {
			if onFailRetriesLeft < 1 {
				logger.Panicln("LLM query failed: ", err)
			} else {
				logger.Warnln("LLM query failed, retrying: ", err)
			}
			continue
		} else if status == llm.QueryMaxTokens {
			if onFailRetriesLeft < 1 {
				logger.Panicln("LLM query reached token limit")
			} else {
				logger.Warnln("LLM query reached token limit, retrying")
			}
			continue
		}
		if len(aiResponses) < 1 || aiResponses[0] == "" {
			logger.Warnln("Got empty response from AI, retrying")
			continue
		}
		logger.Debugln("LLM query completed")
		break
	}

	// Process response, parse files of interest from ai response
	if len(aiResponses) < 1 || aiResponses[0] == "" {
		logger.Errorln("Got empty response from AI")
	}
	llmRequestedFiles, err := utils.ParseTaggedText(aiResponses[0], fileNameTagsRxStrings[0], fileNameTagsRxStrings[1], false)
	if err != nil {
		logger.Panicln("Failed to parse list of files for review", err)
	}
	logger.Debugln("Parsed list of files for review from LLM response")

	// Filter all requested files through project file-list, return only files found in project file-list
	return utils.FilterRequestedProjectFiles(projectRootDir, llmRequestedFiles, []string{targetDocument}, projectFiles, logger)
}
