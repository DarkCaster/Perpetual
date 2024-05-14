package op_implement

import (
	"path/filepath"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/prompts"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage1(projectRootDir string, perpetualDir string, promptsDir string, systemPrompt string, fileNameTagsRxStrings []string,
	fileNames []string, annotations map[string]string, targetFiles []string, logger logging.ILogger) []string {

	// Add trace and debug logging
	logger.Traceln("Stage1: Starting")
	defer logger.Traceln("Stage1: Finished")

	// Create stage1 llm connector
	stage1Connector, err := llm.NewLLMConnector(OpName+"_stage1", systemPrompt, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("failed to create stage1 LLM connector:", err)
	}
	logger.Debugln("Stage1: Created LLM connector")

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
		stage1ProjectIndexRequestMessage = llm.AddIndexFragment(stage1ProjectIndexRequestMessage, item)
		annotation := annotations[item]
		if annotation == "" {
			annotation = "No annotation available"
		}
		stage1ProjectIndexRequestMessage = llm.AddPlainTextFragment(stage1ProjectIndexRequestMessage, annotation)
	}
	logger.Debugln("Stage1: Created project-index request message")

	// Create project-index simulated response
	stage1ProjectIndexResponseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), loadPrompt(prompts.AIImplementStage1ProjectIndexResponseFile))
	logger.Debugln("Stage1: Created project-index simulated response message")

	// Create target files analisys request message
	stage1SourceAnalysisRequestMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(prompts.ImplementStage1SourceAnalysisPromptFile))
	for _, item := range targetFiles {
		contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, item))
		if err != nil {
			logger.Panicln("failed to add file contents to stage1 prompt", err)
		}
		stage1SourceAnalysisRequestMessage = llm.AddFileFragment(stage1SourceAnalysisRequestMessage, item, contents)
	}
	logger.Debugln("Stage1: Created target files analysis request message")

	// Log messages
	llm.LogMessage(logger, perpetualDir, stage1Connector, &stage1ProjectIndexRequestMessage)
	llm.LogMessage(logger, perpetualDir, stage1Connector, &stage1ProjectIndexResponseMessage)
	llm.LogMessage(logger, perpetualDir, stage1Connector, &stage1SourceAnalysisRequestMessage)

	logger.Infoln("Running stage1: find project files for review")
	aiResponse, status, err := stage1Connector.Query(
		stage1ProjectIndexRequestMessage,
		stage1ProjectIndexResponseMessage,
		stage1SourceAnalysisRequestMessage)
	if err != nil {
		logger.Panicln("LLM query failed: ", err)
	} else if status == llm.QueryMaxTokens {
		logger.Panicln("LLM query reached token limit")
	}
	logger.Debugln("Stage1: LLM query completed")

	// Log LLM response
	responseMessage := llm.SetRawResponse(llm.NewMessage(llm.RealAIResponse), aiResponse)
	llm.LogMessage(logger, perpetualDir, stage1Connector, &responseMessage)

	// Process response, parse files of interest from ai response
	filesForReviewRaw, err := utils.ParseTaggedText(aiResponse, fileNameTagsRxStrings[0], fileNameTagsRxStrings[1])
	if err != nil {
		logger.Panicln("Failed to parse list of files for review", err)
	}
	logger.Debugln("Stage1: Parsed list of files for review from LLM response")

	// Check all requested files are among initial project file-list
	var filesForReview []string
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
