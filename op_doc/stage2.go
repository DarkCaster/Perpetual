package op_doc

import (
	"path/filepath"
	"strings"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/prompts"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage2(projectRootDir string, perpetualDir string, promptsDir string, systemPrompt string, filesToMdLangMappings [][2]string, fileNameTags []string, projectFiles []string, filesForReview []string, annotations map[string]string, targetDocument string, exampleDocuemnt string, action string, logger logging.ILogger) string {

	logger.Traceln("Stage2: Starting")
	defer logger.Traceln("Stage2: Finished")

	// Create stage2 llm connector
	connector, err := llm.NewLLMConnector(OpName+"_stage2", systemPrompt, filesToMdLangMappings, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage2 LLM connector:", err)
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

	// Add files requested by LLM
	if len(filesForReview) > 0 {
		// Create request with file-contents
		sourceCodeMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(prompts.DocProjectCodePromptFile))
		for _, item := range filesForReview {
			contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, item))
			if err != nil {
				logger.Panicln("Failed to add file contents to stage2 prompt", err)
			}
			sourceCodeMessage = llm.AddFileFragment(sourceCodeMessage, item, contents, fileNameTags)
		}
		messages = append(messages, sourceCodeMessage)
		logger.Debugln("Project source code message created")

		// Create simulated response
		sourceCodeResponseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), loadPrompt(prompts.AIDocProjectCodeResponseFile))
		messages = append(messages, sourceCodeResponseMessage)
		logger.Debugln("Project source code simulated response added")
	} else {
		logger.Infoln("Not creating extra source-code review")
	}

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

	// Create document-processing request message
	docPromptFile := prompts.DocStage2WritePromptFile
	if action == "REFINE" {
		docPromptFile = prompts.DocStage2RefinePromptFile
	}

	docRequestMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(docPromptFile))
	contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, targetDocument))
	if err != nil {
		logger.Panicln("Failed to add document contents to stage2 prompt", err)
	}
	docRequestMessage = llm.AddPlainTextFragment(docRequestMessage, contents)
	messages = append(messages, docRequestMessage)
	logger.Debugln("Created document processing request message")

	continuePrompt := loadPrompt(prompts.DocStage2ContinuePromptFile)

	//Make LLM request, process response
	onFailRetriesLeft := connector.GetOnFailureRetryLimit()
	if onFailRetriesLeft < 1 {
		onFailRetriesLeft = 1
	}

	for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
		// Work with a copy of message-chain, to discard it on retry
		messagesTry := append([]llm.Message(nil), messages...)
		// Initialize temporary variables for handling partial answers
		var responses []string
		continueGeneration := true
		generateTry := 1
		fileRetry := false
		for continueGeneration && !fileRetry {
			// Run query
			continueGeneration = false
			logger.Infoln("Running stage2: processing document")
			aiResponse, status, err := connector.Query(messagesTry...)
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
				messagesTry = append(messagesTry, llm.SetRawResponse(llm.NewMessage(llm.SimulatedAIResponse), aiResponse))
				messagesTry = append(messagesTry, llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), continuePrompt))
			}
			// Append response fragment
			responses = append(responses, aiResponse)
		}
		if fileRetry {
			continue
		}
		// Join responses together to form the final document contents
		return strings.Join(responses, "")
	}
	return ""
}
