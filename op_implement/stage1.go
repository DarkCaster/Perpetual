package op_implement

import (
	"path/filepath"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/prompts"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage1(projectRootDir string, perpetualDir string, promptsDir string, systemPrompt string, filesToMdLangMappings [][2]string, fileNameTagsRxStrings []string, fileNameTags []string,
	fileNames []string, annotations map[string]string, targetFiles []string, logger logging.ILogger) []string {

	// Add trace and debug logging
	logger.Traceln("Stage1: Starting")
	defer logger.Traceln("Stage1: Finished")

	// Create stage1 llm connector
	stage1Connector, err := llm.NewLLMConnector(OpName+"_stage1", systemPrompt, filesToMdLangMappings, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage1 LLM connector:", err)
	}
	logger.Debugln(llm.GetDebugString(stage1Connector))

	loadPrompt := func(filePath string) string {
		text, err := utils.LoadTextFile(filepath.Join(promptsDir, filePath))
		if err != nil {
			logger.Panicln("Failed to load prompt:", err)
		}
		return text
	}

	// Create project-index request message
	stage1ProjectIndexRequestMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(prompts.ImplementStage1ProjectIndexPromptFile))
	for _, item := range fileNames {
		stage1ProjectIndexRequestMessage = llm.AddIndexFragment(stage1ProjectIndexRequestMessage, item, fileNameTags)
		annotation := annotations[item]
		if annotation == "" {
			annotation = "No annotation available"
		}
		stage1ProjectIndexRequestMessage = llm.AddPlainTextFragment(stage1ProjectIndexRequestMessage, annotation)
	}
	logger.Debugln("Created project-index request message")

	// Create project-index simulated response
	stage1ProjectIndexResponseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), loadPrompt(prompts.AIImplementStage1ProjectIndexResponseFile))
	logger.Debugln("Created project-index simulated response message")

	// Create target files analisys request message
	stage1SourceAnalysisRequestMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(prompts.ImplementStage1SourceAnalysisPromptFile))
	for _, item := range targetFiles {
		contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, item))
		if err != nil {
			logger.Panicln("failed to add file contents to stage1 prompt", err)
		}
		stage1SourceAnalysisRequestMessage = llm.AddFileFragment(stage1SourceAnalysisRequestMessage, item, contents, fileNameTags)
	}
	logger.Debugln("Created target files analysis request message")

	aiResponse := ""
	onFailRetriesLeft := stage1Connector.GetOnFailureRetryLimit()
	if onFailRetriesLeft < 1 {
		onFailRetriesLeft = 1
	}
	for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
		logger.Infoln("Running stage1: find project files for review")
		var status llm.QueryStatus
		//NOTE: do not use := here, looks like it will make copy of aiResponse, and effectively result in empty file-list (tested on golang 1.22.3)
		aiResponse, status, err = stage1Connector.Query(
			stage1ProjectIndexRequestMessage,
			stage1ProjectIndexResponseMessage,
			stage1SourceAnalysisRequestMessage)
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
		if aiResponse == "" {
			logger.Warnln("Got empty response from AI, retrying")
			continue
		}
		logger.Debugln("LLM query completed")
		break
	}

	// Process response, parse files of interest from ai response
	if aiResponse == "" {
		logger.Errorln("Got empty response from AI")
	}
	filesForReviewRaw, err := utils.ParseTaggedText(aiResponse, fileNameTagsRxStrings[0], fileNameTagsRxStrings[1])
	if err != nil {
		logger.Panicln("Failed to parse list of files for review", err)
	}
	logger.Debugln("Parsed list of files for review from LLM response")

	// Check all requested files are among initial project file-list
	var filesForReview []string
	logger.Debugln("Raw file-list requested by LLM:", filesForReviewRaw)
	logger.Infoln("Files requested by LLM:")
	for _, check := range filesForReviewRaw {
		//remove new line from the end of filename, if present
		if check != "" && check[len(check)-1] == '\n' {
			check = check[:len(check)-1]
		}
		//remove \r from the end of filename, if present
		if check != "" && check[len(check)-1] == '\r' {
			check = check[:len(check)-1]
		}
		//replace possibly-invalid path separators
		check = utils.ConvertFilePathToOSFormat(check)
		//make file path relative to project root
		file, err := utils.MakePathRelative(projectRootDir, check, true)
		if err != nil {
			logger.Errorln("Failed to validate filename requested by LLM for review:", check)
			continue
		}
		// Do not add file for review if it among files for implement, also fix case if so
		file, found := utils.CaseInsensitiveFileSearch(file, targetFiles)
		if found {
			logger.Warnln("Not adding file for review, this file already marked for implementation:", file)
		} else {
			file, found := utils.CaseInsensitiveFileSearch(file, filesForReview)
			if found {
				logger.Warnln("Not adding file for review, it is already added or having filename case conflict:", file)
			} else {
				file, found := utils.CaseInsensitiveFileSearch(file, fileNames)
				if found {
					filesForReview = append(filesForReview, file)
					logger.Infoln(file)
				} else {
					logger.Warnln("Not adding file for review, it is not found in filtered project file-list:", file)
				}
			}
		}
	}

	return filesForReview
}
