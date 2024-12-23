package op_implement

import (
	"path/filepath"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage2(projectRootDir string,
	perpetualDir string,
	cfg config.Config,
	filesToMdLangMappings [][]string,
	planningMode int,
	filesForReview []string,
	targetFiles []string,
	logger logging.ILogger) []llm.Message {

	logger.Traceln("Stage2: Starting")
	defer logger.Traceln("Stage2: Finished")

	// Create stage2 llm connector
	stage2Connector, err := llm.NewLLMConnector(OpName+"_stage2", cfg.String(config.K_SystemPrompt), filesToMdLangMappings, map[string]interface{}{}, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage2 LLM connector:", err)
	}
	logger.Debugln(stage2Connector.GetDebugString())

	// This will store message history to re-use on this and next stages
	var messages []llm.Message

	// Generate messages with listing of source code files requested at stage 1 (if any)
	if len(filesForReview) > 0 {
		// Create target files analisys request message
		stage2ProjectSourceCodeMessage := llm.AddPlainTextFragment(
			llm.NewMessage(llm.UserRequest),
			cfg.String(config.K_ImplementStage2CodePrompt))
		// Add actual files to it
		for _, item := range filesForReview {
			contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, item))
			if err != nil {
				logger.Panicln("Failed to add file contents to stage2 prompt", err)
			}
			stage2ProjectSourceCodeMessage = llm.AddFileFragment(
				stage2ProjectSourceCodeMessage,
				item,
				contents,
				cfg.StringArray(config.K_FilenameTags))
		}
		// Add message to history
		messages = append(messages, stage2ProjectSourceCodeMessage)
		logger.Debugln("Project source code message created")
		// Create simulated response
		stage2ProjectSourceCodeResponseMessage := llm.AddPlainTextFragment(
			llm.NewMessage(llm.SimulatedAIResponse),
			cfg.String(config.K_ImplementStage2CodeResponse))
		// Add message to history
		messages = append(messages, stage2ProjectSourceCodeResponseMessage)
		logger.Debugln("Project source code simulated response added")
	} else {
		logger.Infoln("Not creating extra source-code review")
	}

	// When planning is disabled, just create messages with listing of files marked to implement and request for step-by-step implementation
	if planningMode == 0 {
		logger.Infoln("Running stage2: planning disabled")
		// Create files to request for non-planning mode
		stage2FilesNoPlanningMessage := llm.AddPlainTextFragment(
			llm.NewMessage(llm.UserRequest),
			cfg.String(config.K_ImplementStage2NoPlanningPrompt))
		// Attach target files
		for _, item := range targetFiles {
			contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, item))
			if err != nil {
				logger.Panicln("Failed to add file contents to stage1 prompt", err)
			}
			stage2FilesNoPlanningMessage = llm.AddFileFragment(
				stage2FilesNoPlanningMessage,
				item,
				contents,
				cfg.StringArray(config.K_FilenameTags))
		}
		// Add message to history
		messages = append(messages, stage2FilesNoPlanningMessage)
		logger.Debugln("Files for no planning message created")
		// Create simulated response
		stage2FilesNoPlanningResponseMessage := llm.AddPlainTextFragment(
			llm.NewMessage(llm.SimulatedAIResponse),
			cfg.String(config.K_ImplementStage2NoPlanningResponse))
		// Add message to history
		messages = append(messages, stage2FilesNoPlanningResponseMessage)
		logger.Debugln("Files for no planning simulated response added")
	}

	// When planning mode set to extended mode, create list of files with request
	// to generate a reasonings/work plan of what needs to be done in order to implement the task
	if planningMode == 2 {
		logger.Infoln("Running stage2: generating reasonings")
		stage2ReasoningsRequestMessage := llm.AddPlainTextFragment(
			llm.NewMessage(llm.UserRequest),
			cfg.String(config.K_ImplementStage2ReasoningsPrompt))
		// Attach target files
		for _, item := range targetFiles {
			contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, item))
			if err != nil {
				logger.Panicln("Failed to add file contents to stage2 prompt", err)
			}
			stage2ReasoningsRequestMessage = llm.AddFileFragment(
				stage2ReasoningsRequestMessage,
				item,
				contents,
				cfg.StringArray(config.K_FilenameTags))
		}
		// Add message to history
		messages = append(messages, stage2ReasoningsRequestMessage)
		logger.Debugln("Planning request message created")
		//query LLM to generate reasonings
		onFailRetriesLeft := stage2Connector.GetOnFailureRetryLimit()
		if onFailRetriesLeft < 1 {
			onFailRetriesLeft = 1
		}
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			aiResponses, status, err := stage2Connector.Query(1, messages...)
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
					logger.Warnln("LLM query failed, retrying: ", err)
				}
				continue
			}
			// Reasonings should not include code blocks
			aiResponses = utils.FilterAndTrimResponses(aiResponses, cfg.RegexpArray(config.K_CodeTagsRx), logger)
			//TODO: for multi-response mode, add code here that combine responses together or select the most appropriate response
			reasonings := ""
			if len(aiResponses) > 0 {
				reasonings = aiResponses[0]
			}
			// Final check
			if reasonings == "" {
				if onFailRetriesLeft < 1 {
					logger.Panicln("Filtered reasonings response from AI is empty")
				} else {
					logger.Warnln("Filtered reasonings response from AI is empty, retrying")
				}
				continue
			}
			// Save reasonings to message-history
			stage2ReasoningsResponseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), reasonings)
			messages = append(messages, stage2ReasoningsResponseMessage)
			break
		}
	}
	return messages
}
