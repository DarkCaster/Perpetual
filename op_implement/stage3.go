package op_implement

import (
	"path/filepath"
	"strings"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/prompts"
	"github.com/DarkCaster/Perpetual/utils"
)

// Perform the actual code implementation process based on the Stage2 answer, which includes the contents of other files related to the files for which we need to implement the code.
func Stage3(projectRootDir string, perpetualDir string, promptsDir string, systemPrompt string, outputTagsRxStrings []string, fileNameEmbedTag string,
	stage2Messages []llm.Message, otherFiles []string, targetFiles []string, logger logging.ILogger) map[string]string {

	logger.Traceln("Stage3: Starting")       // Add trace logging
	defer logger.Traceln("Stage3: Finished") // Add trace logging

	// Create stage3 llm connector
	stage3Connector, err := llm.NewLLMConnector(OpName+"_stage3", systemPrompt, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage3 LLM connector:", err)
	}
	logger.Debugln(llm.GetDebugString(stage3Connector))

	loadPrompt := func(filePath string) string {
		text, err := utils.LoadTextFile(filepath.Join(promptsDir, filePath))
		if err != nil {
			logger.Panicln("Failed to load prompt:", err)
		}
		return text
	}

	stage3ChangesDonePromptTemplate := loadPrompt(prompts.ImplementStage3ChangesDonePromptFile)
	stage3ChangesDoneResponse := loadPrompt(prompts.AIImplementStage3ChangesDoneResponseFile)
	stage3ProcessFilePromptTemplate := loadPrompt(prompts.ImplementStage3ProcessFilePromptFile)
	stage3ContinuePromptTemplate := loadPrompt(prompts.ImplementStage3ContinuePromptFile)

	processedFileContents := make(map[string]string)
	var processedFiles []string

	// Main processing loop
	for workPending := true; workPending; workPending = len(otherFiles) > 0 || len(targetFiles) > 0 {
		logger.Debugln("Work pending:", workPending) // Add debug logging

		// Generate change-done message from already processed files
		stage3ChangesDoneMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), stage3ChangesDonePromptTemplate)
		for _, item := range processedFiles {
			contents, ok := processedFileContents[item]
			if !ok {
				logger.Errorln("Failed to add file contents to stage3:", err)
			} else {
				stage3ChangesDoneMessage = llm.AddFileFragment(stage3ChangesDoneMessage, item, contents)
			}
		}

		stage3ChangesDoneResponseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), stage3ChangesDoneResponse)

		stage3Messages := append([]llm.Message(nil), stage2Messages...)
		if len(processedFiles) > 0 {
			stage3Messages = append(stage3Messages, stage3ChangesDoneMessage)
			stage3Messages = append(stage3Messages, stage3ChangesDoneResponseMessage)
		}

		pendingFile := ""
		if len(otherFiles) > 0 {
			pendingFile, otherFiles = otherFiles[0], otherFiles[1:]
		} else if len(targetFiles) > 0 {
			pendingFile, targetFiles = targetFiles[0], targetFiles[1:]
		}

		if pendingFile == "" {
			break
		}

		logger.Debugln("Processing file:", pendingFile) // Add debug logging
		// Create prompt from stage3ProcessFilePromptTemplate
		stage3ProcessFilePrompt, err := utils.ReplaceTag(stage3ProcessFilePromptTemplate, fileNameEmbedTag, pendingFile)
		if err != nil {
			logger.Errorln("Failed to replace filename tag", err)
			stage3ProcessFilePrompt = stage3ProcessFilePromptTemplate
		}

		// Create prompt for to implement one of the files
		stage3Messages = append(stage3Messages, llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), stage3ProcessFilePrompt))

		var responses []string
		continueGeneration := true
		generateTry := 1
		for continueGeneration {
			// Log messages we are going to send
			llm.LogMessages(logger, perpetualDir, stage3Connector, stage3Messages)

			continueGeneration = false
			logger.Infoln("Running stage3: implementing code for:", pendingFile)
			aiResponse, status, err := stage3Connector.Query(stage3Messages...)
			if err != nil {
				logger.Panicln("LLM query failed: ", err)
			} else if status == llm.QueryMaxTokens {
				if generateTry >= stage3Connector.GetMaxTokensRetryLimit() {
					logger.Errorln("LLM query reached token limit, and we are reached retry limit, not attempting to continue")
				} else {
					logger.Warnln("LLM query reached token limit, attempting to continue")
					continueGeneration = true
					generateTry++
				}
				// Add partial response to stage3 messages, with request to continue
				stage3Messages = append(stage3Messages, llm.SetRawResponse(llm.NewMessage(llm.SimulatedAIResponse), aiResponse))
				stage3Messages = append(stage3Messages, llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), stage3ContinuePromptTemplate))
			}

			// Log LLM response
			responseMessage := llm.SetRawResponse(llm.NewMessage(llm.RealAIResponse), aiResponse)
			llm.LogMessage(logger, perpetualDir, stage3Connector, &responseMessage)

			// Append response fragment
			responses = append(responses, aiResponse)
		}

		// Remove extra output tag from the start from non first response-fragments
		for i := range responses {
			if i > 0 {
				responses[i], err = utils.GetTextAfterFirstMatch(responses[i], outputTagsRxStrings[0])
				if err != nil {
					logger.Panicln("Error while parsing output response fragment:", err)
				}
			}
		}

		// Parse LLM output, detect file body in response
		combinedResponse := strings.Join(responses, "")
		fileBodies, err := utils.ParseTaggedText(combinedResponse, outputTagsRxStrings[0], outputTagsRxStrings[1])
		if err != nil {
			logger.Errorln("Error while parsing LLM response with output file:", err)
			// Try to remove only first match then
			fileBody, err := utils.GetTextAfterFirstMatch(combinedResponse, outputTagsRxStrings[0])
			if err != nil {
				logger.Panicln("Error while parsing body from combined fragments:", err)
			}
			fileBodies = []string{fileBody}
		}

		if len(fileBodies) > 1 {
			logger.Errorln("Multiple file bodies detected in LLM response:", len(fileBodies))
		}

		// Save body to processedFileContents and add record to processedFiles
		if len(fileBodies) > 0 {
			logger.Debugln("Found output for:", pendingFile)
			processedFileContents[pendingFile] = fileBodies[0]
			processedFiles = append(processedFiles, pendingFile)
		} else {
			logger.Errorln("No output found for file:", pendingFile)
		}
	}

	return processedFileContents
}
