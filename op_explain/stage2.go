package op_explain

import (
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage2(projectRootDir string,
	perpetualDir string,
	cfg config.Config,
	filesToMdLangMappings [][]string,
	projectFiles []string,
	filesForReview []string,
	annotations map[string]string,
	question string,
	addAnnotations bool,
	logger logging.ILogger) string {

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
	logger.Debugln(connector.GetDebugString())

	var messages []llm.Message

	if addAnnotations {
		// Create project-index request message
		indexRequest := llm.ComposeMessageWithAnnotations(
			cfg.String(config.K_ExplainProjectIndexPrompt),
			projectFiles,
			cfg.StringArray(config.K_FilenameTags),
			annotations,
			logger)
		messages = append(messages, indexRequest)
		logger.Debugln("Created project-index request message")
		// Create project-index simulated response
		indexResponse := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_ExplainProjectIndexResponse))
		messages = append(messages, indexResponse)
		logger.Debugln("Created project-index simulated response message")
	} else {
		logger.Infoln("Not adding project-annotations")
	}

	// Add files requested by LLM
	if len(filesForReview) > 0 {
		// Create request with file-contents
		requestMessage := llm.ComposeMessageWithFiles(
			projectRootDir,
			cfg.String(config.K_ExplainStage2FilesPrompt),
			filesForReview,
			cfg.StringArray(config.K_FilenameTags),
			logger)
		messages = append(messages, requestMessage)
		logger.Debugln("Project source code message created")
		// Create simulated response
		responseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_ExplainStage2FilesResponse))
		messages = append(messages, responseMessage)
		logger.Debugln("Project source code simulated response added")
	} else {
		logger.Infoln("Not creating extra source-code review")
	}

	// Create question-processing request message
	explainRequest := llm.AddPlainTextFragment(
		llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), cfg.String(config.K_ExplainStage2QuestionPrompt)),
		question)
	messages = append(messages, explainRequest)
	logger.Debugln("Created question processing request message")

	//Make LLM request, process response
	onFailRetriesLeft := connector.GetOnFailureRetryLimit()
	if onFailRetriesLeft < 1 {
		onFailRetriesLeft = 1
	}
	for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
		// Work with a copy of message-chain, to discard it on retry
		messagesTry := utils.NewSlice(messages...)
		// Initialize temporary variables for handling partial answers
		var responses []string
		continueGeneration := true
		generateTry := 1
		fileRetry := false
		for continueGeneration && !fileRetry {
			// Run query
			continueGeneration = false
			logger.Infoln("Running stage2: processing question")
			aiResponses, status, err := connector.Query(1, messagesTry...)
			if err != nil {
				// Retry file on LLM error
				if onFailRetriesLeft < 1 {
					logger.Panicln("LLM query failed:", err)
				} else {
					logger.Warnln("LLM query failed, retrying:", err)
					fileRetry = true
					break
				}
			} else if status == llm.QueryMaxTokens {
				// Try to recover other parts of the file if reached max tokens
				if generateTry >= connector.GetMaxTokensSegments() {
					logger.Errorln("LLM query reached token limit, and we are reached segment limit, not attempting to continue")
				} else {
					logger.Warnln("LLM query reached token limit, attempting to continue and file recover")
					continueGeneration = true
					generateTry++
				}
				// Add partial response to stage2 messages, with request to continue
				messagesTry = append(messagesTry, llm.SetRawResponse(llm.NewMessage(llm.SimulatedAIResponse), aiResponses[0]))
				messagesTry = append(messagesTry, llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), cfg.String(config.K_ExplainStage2ContinuePrompt)))
			}
			// Append response fragment
			responses = append(responses, aiResponses[0])
		}
		if fileRetry {
			continue
		}
		// Join responses together to form the final answer
		return strings.Join(responses, "")
	}
	return ""
}
