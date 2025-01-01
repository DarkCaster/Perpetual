package op_implement

import (
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
	logger logging.ILogger) ([]llm.Message, llm.Message) {

	logger.Traceln("Stage2: Starting")
	defer logger.Traceln("Stage2: Finished")

	// Create stage2 llm connector
	stage2Connector, err := llm.NewLLMConnector(
		OpName+"_stage2",
		cfg.String(config.K_SystemPrompt),
		filesToMdLangMappings,
		map[string]interface{}{},
		"", "",
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage2 LLM connector:", err)
	}
	logger.Debugln(stage2Connector.GetDebugString())

	// This will store message history to re-use on this and next stages
	var messages []llm.Message

	// Generate messages with listing of source files requested for review at stage 1, if any
	if len(filesForReview) > 0 {
		// Create target files analysis request message
		realRequestMessage := llm.ComposeMessageWithFiles(
			projectRootDir,
			cfg.String(config.K_ImplementStage2CodePrompt),
			filesForReview,
			cfg.StringArray(config.K_FilenameTags),
			logger)
		// Add message to history
		messages = append(messages, realRequestMessage)
		logger.Debugln("Project source code message created")
		// Create simulated response and add message to history
		responseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_ImplementStage2CodeResponse))
		messages = append(messages, responseMessage)
		logger.Debugln("Project source code simulated response added")
	} else {
		logger.Infoln("Not adding any source code files for review")
	}

	var messageWithTargetFiles llm.Message
	// When planning is disabled, just create messages with listing of files marked to implement and request for step-by-step implementation
	if planningMode == 0 {
		logger.Infoln("Running stage2: planning disabled, not generating work plan")
		// Create files to request for non-planning mode
		requestMessage := llm.ComposeMessageWithFiles(
			projectRootDir,
			cfg.String(config.K_ImplementStage2NoPlanningPrompt),
			targetFiles,
			cfg.StringArray(config.K_FilenameTags),
			logger)
		// Add message to history
		messageWithTargetFiles = requestMessage
		messages = append(messages, requestMessage)
		logger.Debugln("Files for no planning message created")
		// Create simulated response and add it to history
		responseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_ImplementStage2NoPlanningResponse))
		messages = append(messages, responseMessage)
		logger.Debugln("Files for no planning simulated response added")
	}

	// When planning mode set to extended mode, create list of files with request
	// to generate a reasonings/work plan of what needs to be done in order to implement the task
	if planningMode == 2 {
		logger.Infoln("Running stage2: generating work plan")
		// Generate actual request message that will me used with LLM
		requestMessage := llm.ComposeMessageWithFiles(
			projectRootDir,
			cfg.String(config.K_ImplementStage2ReasoningsPrompt),
			targetFiles,
			cfg.StringArray(config.K_FilenameTags),
			logger)
		// realMessages message-history will be used for actual LLM prompt
		realMessages := make([]llm.Message, len(messages), len(messages)+1)
		copy(realMessages, messages)
		realMessages = append(realMessages, requestMessage)
		// Query LLM to generate reasonings
		onFailRetriesLeft := stage2Connector.GetOnFailureRetryLimit()
		if onFailRetriesLeft < 1 {
			onFailRetriesLeft = 1
		}
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			aiResponses, status, err := stage2Connector.Query(1, realMessages...)
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
			// Add final request message to history - it use simplier instructions that will less likely affect next stages
			finalRequestMessage := llm.ComposeMessageWithFiles(
				projectRootDir,
				cfg.String(config.K_ImplementStage2ReasoningsPromptFinal),
				targetFiles,
				cfg.StringArray(config.K_FilenameTags),
				logger)
			messageWithTargetFiles = finalRequestMessage
			messages = append(messages, finalRequestMessage)
			logger.Debugln("Planning request message created")
			// Save reasonings to message-history
			responseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), reasonings)
			messages = append(messages, responseMessage)
			break
		}
	}
	return messages, messageWithTargetFiles
}
