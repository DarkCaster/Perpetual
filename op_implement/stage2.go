package op_implement

import (
	"fmt"

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
	task string,
	logger logging.ILogger) (string, []llm.Message, int) {

	logger.Traceln("Stage2: Starting")
	defer logger.Traceln("Stage2: Finished")

	// Create stage2 llm connector
	connector, err := llm.NewLLMConnector(
		OpName+"_stage2",
		cfg.String(config.K_SystemPrompt),
		cfg.String(config.K_SystemPromptAck),
		filesToMdLangMappings,
		map[string]interface{}{},
		"", "",
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage2 LLM connector:", err)
	}

	var messages []llm.Message

	// Add files requested by LLM
	if len(filesForReview) > 0 {
		// Create request with file-contents
		requestMessage := llm.ComposeMessageWithFiles(
			projectRootDir,
			cfg.String(config.K_ProjectCodePrompt),
			filesForReview,
			cfg.StringArray(config.K_FilenameTags),
			logger)
		messages = append(messages, requestMessage)
		logger.Debugln("Project source code message created")
		// Create simulated response
		responseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_ProjectCodeResponse))
		messages = append(messages, responseMessage)
		logger.Debugln("Project source code simulated response added")
	} else {
		logger.Infoln("Not creating extra source-code review")
	}

	var msgIndexToAddExtraFiles int
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
		messages = append(messages, requestMessage)
		msgIndexToAddExtraFiles = len(messages) - 1
		logger.Debugln("Files for no planning message created")
		// Create simulated response and add it to history
		responseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_ImplementStage2NoPlanningResponse))
		messages = append(messages, responseMessage)
		logger.Debugln("Files for no planning simulated response added")
	}

	// When planning mode set to extended mode, create list of files with request
	// to generate a reasonings/work plan of what needs to be done in order to implement the task
	if planningMode == 2 {
		var requestMessage llm.Message
		// Generate actual request message that will me used with LLM
		if task == "" {
			requestMessage = llm.ComposeMessageWithFiles(
				projectRootDir,
				cfg.String(config.K_ImplementStage2ReasoningsPrompt),
				targetFiles,
				cfg.StringArray(config.K_FilenameTags),
				logger)
		} else {
			requestMessage = llm.AddPlainTextFragment(
				llm.AddPlainTextFragment(
					llm.NewMessage(llm.UserRequest),
					cfg.String(config.K_ImplementTaskStage2ReasoningsPrompt)),
				task)
		}
		// realMessages message-history will be used for actual LLM prompt
		realMessages := append(utils.NewSlice(messages...), requestMessage)

		logger.Infoln("Running stage2: generating work plan")
		debugString := connector.GetDebugString()
		logger.Notifyln(debugString)
		llm.GetSimpleRawMessageLogger(perpetualDir)(fmt.Sprintf("=== Implement (stage 2): %s\n\n\n", debugString))

		// Query LLM to generate reasonings
		onFailRetriesLeft := connector.GetOnFailureRetryLimit()
		if onFailRetriesLeft < 1 {
			onFailRetriesLeft = 1
		}
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			aiResponses, status, err := connector.Query(1, realMessages...)
			if perfString := connector.GetPerfString(); perfString != "" {
				logger.Traceln(perfString)
			}
			if err != nil {
				if onFailRetriesLeft < 1 {
					logger.Panicln("LLM query failed:", err)
				} else {
					logger.Warnln("LLM query failed, retrying:", err)
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
					logger.Panicln("Filtered reasonings response from AI is empty or invalid")
				} else {
					logger.Warnln("Filtered reasonings response from AI is empty or invalid, retrying")
				}
				continue
			}

			var finalRequestMessage llm.Message
			// Add final request message to history - it use simplier instructions that will less likely affect next stages
			if task == "" {
				finalRequestMessage = llm.ComposeMessageWithFiles(
					projectRootDir,
					cfg.String(config.K_ImplementStage2ReasoningsPromptFinal),
					targetFiles,
					cfg.StringArray(config.K_FilenameTags),
					logger)
			} else {
				finalRequestMessage = llm.AddPlainTextFragment(
					llm.AddPlainTextFragment(
						llm.NewMessage(llm.UserRequest),
						cfg.String(config.K_ImplementTaskStage2ReasoningsPromptFinal)),
					task)
			}

			messages = append(messages, finalRequestMessage)
			msgIndexToAddExtraFiles = len(messages) - 1
			logger.Debugln("Planning request message created")
			// Save reasonings to message-history
			responseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), reasonings)
			messages = append(messages, responseMessage)
			break
		}
	}
	return "", messages, msgIndexToAddExtraFiles
}
