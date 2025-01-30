package op_doc

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
	targetDocument string,
	exampleDocuemnt string,
	action string,
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

	// Create project-index request message
	indexRequest := llm.ComposeMessageWithAnnotations(
		cfg.String(config.K_DocProjectIndexPrompt),
		projectFiles,
		cfg.StringArray(config.K_FilenameTags),
		annotations,
		logger)
	messages = append(messages, indexRequest)
	logger.Debugln("Created project-index request message")

	// Create project-index simulated response
	indexResponse := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_DocProjectIndexResponse))
	messages = append(messages, indexResponse)
	logger.Debugln("Created project-index simulated response message")

	// Add files requested by LLM
	if len(filesForReview) > 0 {
		// Create request with file-contents
		requestMessage := llm.ComposeMessageWithFiles(
			projectRootDir,
			cfg.String(config.K_DocProjectCodePrompt),
			filesForReview,
			cfg.StringArray(config.K_FilenameTags),
			logger)
		messages = append(messages, requestMessage)
		logger.Debugln("Project source code message created")
		// Create simulated response
		responseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_DocProjectCodeResponse))
		messages = append(messages, responseMessage)
		logger.Debugln("Project source code simulated response added")
	} else {
		logger.Infoln("Not creating extra source-code review")
	}

	if exampleDocuemnt != "" {
		// Create document-example request message
		requestMessage := llm.ComposeMessageFromPromptAndTextFile(projectRootDir, cfg.String(config.K_DocExamplePrompt), exampleDocuemnt, logger)
		messages = append(messages, requestMessage)
		logger.Debugln("Created document-example request message")
		// Create document-example simulated response
		responseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_DocExampleResponse))
		messages = append(messages, responseMessage)
		logger.Debugln("Created document-example simulated response message")
	}

	// Create document-processing request message
	prompt := cfg.String(config.K_DocStage2WritePrompt)
	if action == "REFINE" {
		prompt = cfg.String(config.K_DocStage2RefinePrompt)
	}
	documentRequest := llm.ComposeMessageFromPromptAndTextFile(projectRootDir, prompt, targetDocument, logger)
	messages = append(messages, documentRequest)
	logger.Debugln("Created document processing request message")

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
			logger.Infoln("Running stage2: processing document")
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
				messagesTry = append(messagesTry, llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), cfg.String(config.K_DocStage2ContinuePrompt)))
			}
			// Append response fragment
			responses = append(responses, aiResponses[0])
		}
		if fileRetry {
			continue
		}
		// Join responses together to form the final document contents
		return strings.Join(responses, "")
	}
	return ""
}
