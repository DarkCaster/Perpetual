package op_implement

import (
	"path/filepath"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage1(projectRootDir string,
	perpetualDir string,
	cfg config.Config,
	filesToMdLangMappings [][2]string,
	fileNames []string,
	annotations map[string]string,
	targetFiles []string,
	logger logging.ILogger) []string {

	// Add trace and debug logging
	logger.Traceln("Stage1: Starting")
	defer logger.Traceln("Stage1: Finished")

	// Create stage1 llm connector
	stage1Connector, err := llm.NewLLMConnector(OpName+"_stage1", cfg.String(config.K_SystemPrompt), filesToMdLangMappings, map[string]interface{}{}, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage1 LLM connector:", err)
	}
	logger.Debugln(stage1Connector.GetDebugString())

	// Create project-index request message
	stage1ProjectIndexRequestMessage := llm.AddPlainTextFragment(
		llm.NewMessage(llm.UserRequest),
		cfg.String(config.K_ImplementStage1IndexPrompt))

	for _, item := range fileNames {
		stage1ProjectIndexRequestMessage = llm.AddIndexFragment(
			stage1ProjectIndexRequestMessage,
			item,
			cfg.StringArray(config.K_FilenameTags))

		annotation := annotations[item]
		if annotation == "" {
			annotation = "No annotation available"
		}

		stage1ProjectIndexRequestMessage = llm.AddPlainTextFragment(
			stage1ProjectIndexRequestMessage,
			annotation)
	}
	logger.Debugln("Created project-index request message")

	// Create project-index simulated response
	stage1ProjectIndexResponseMessage := llm.AddPlainTextFragment(
		llm.NewMessage(llm.SimulatedAIResponse),
		cfg.String(config.K_ImplementStage1IndexResponse))

	logger.Debugln("Created project-index simulated response message")

	// Create target files analisys request message
	stage1SourceAnalysisRequestMessage := llm.AddPlainTextFragment(
		llm.NewMessage(llm.UserRequest),
		cfg.String(config.K_ImplementStage1AnalisysPrompt))

	for _, item := range targetFiles {
		contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, item))
		if err != nil {
			logger.Panicln("failed to add file contents to stage1 prompt", err)
		}
		stage1SourceAnalysisRequestMessage = llm.AddFileFragment(
			stage1SourceAnalysisRequestMessage,
			item,
			contents,
			cfg.StringArray(config.K_FilenameTags))
	}
	logger.Debugln("Created target files analysis request message")

	var aiResponses []string
	onFailRetriesLeft := stage1Connector.GetOnFailureRetryLimit()
	if onFailRetriesLeft < 1 {
		onFailRetriesLeft = 1
	}
	for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
		logger.Infoln("Running stage1: find project files for review")
		var status llm.QueryStatus
		//NOTE: do not use := here, looks like it will make copy of aiResponse, and effectively result in empty file-list (tested on golang 1.22.3)
		aiResponses, status, err = stage1Connector.Query(1,
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

	filesForReviewRaw, err := utils.ParseTaggedTextRx(aiResponses[0],
		cfg.RegexpArray(config.K_FilenameTagsRx)[0],
		cfg.RegexpArray(config.K_FilenameTagsRx)[1],
		false)

	if err != nil {
		logger.Panicln("Failed to parse list of files for review", err)
	}
	logger.Debugln("Parsed list of files for review from LLM response")

	// Filter all requested files through project file-list, return only files found in project file-list
	return utils.FilterRequestedProjectFiles(projectRootDir, filesForReviewRaw, targetFiles, fileNames, logger)
}
