package op_implement

import (
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

// Perform the actual code implementation process based on the Stage2 answer, which includes the contents of other files related to the files for which we need to implement the code.
func Stage3(projectRootDir string,
	perpetualDir string,
	cfg map[string]interface{},
	filesToMdLangMappings [][2]string,
	stage2Messages []llm.Message,
	otherFiles []string,
	targetFiles []string,
	logger logging.ILogger) map[string]string {

	logger.Traceln("Stage3: Starting")       // Add trace logging
	defer logger.Traceln("Stage3: Finished") // Add trace logging

	// Create stage3 llm connector
	stage3Connector, err := llm.NewLLMConnector(OpName+"_stage3", cfg[config.K_SystemPrompt].(string), filesToMdLangMappings, map[string]interface{}{}, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage3 LLM connector:", err)
	}
	logger.Debugln(stage3Connector.GetDebugString())

	processedFileContents := make(map[string]string)
	var processedFiles []string

	// Main processing loop
	for workPending := true; workPending; workPending = len(otherFiles) > 0 || len(targetFiles) > 0 {
		logger.Debugln("Work pending:", workPending) // Add debug logging

		// Generate change-done message from already processed files
		stage3ChangesDoneMessage := llm.AddPlainTextFragment(
			llm.NewMessage(llm.UserRequest),
			cfg[config.K_ImplementStage3ChangesDonePrompt].(string))

		for _, item := range processedFiles {
			contents, ok := processedFileContents[item]
			if !ok {
				logger.Errorln("Failed to add file contents to stage3:", err)
			} else {
				stage3ChangesDoneMessage = llm.AddFileFragment(
					stage3ChangesDoneMessage,
					item,
					contents,
					utils.InterfaceToStringArray(cfg[config.K_FilenameTags]))
			}
		}

		stage3ChangesDoneResponseMessage := llm.AddPlainTextFragment(
			llm.NewMessage(llm.SimulatedAIResponse),
			cfg[config.K_ImplementStage3ChangesDoneResponse].(string))

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
		stage3ProcessFilePrompt, err := utils.ReplaceTag(
			cfg[config.K_ImplementStage3ProcessPrompt].(string),
			cfg[config.K_FilenameEmbedRx].(string),
			pendingFile)

		if err != nil {
			logger.Errorln("Failed to replace filename tag", err)
			stage3ProcessFilePrompt = cfg[config.K_ImplementStage3ProcessPrompt].(string)
		}

		// Create prompt for to implement one of the files
		stage3Messages = append(stage3Messages, llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), stage3ProcessFilePrompt))

		var fileBodies []string
		onFailRetriesLeft := stage3Connector.GetOnFailureRetryLimit()
		if onFailRetriesLeft < 1 {
			onFailRetriesLeft = 1
		}
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			// Create a copy, so it will be automatically discarded on file retry
			stage3MessagesTry := append([]llm.Message(nil), stage3Messages...)
			// Initialize temporary variables for handling partial answers
			var responses []string
			continueGeneration := true
			ignoreUnclosedTagErrors := false
			generateTry := 1
			fileRetry := false
			for continueGeneration && !fileRetry {
				// Run query
				continueGeneration = false
				logger.Infoln("Running stage3: implementing code for:", pendingFile)
				aiResponses, status, err := stage3Connector.Query(1, stage3MessagesTry...)
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
					if generateTry >= stage3Connector.GetMaxTokensSegments() {
						logger.Errorln("LLM query reached token limit, and we are reached segment limit, not attempting to continue")
					} else {
						logger.Warnln("LLM query reached token limit, attempting to continue and file recover")
						continueGeneration = true
						generateTry++
						// Disable some possible parsing errors in future
						ignoreUnclosedTagErrors = true
					}
					// Add partial response to stage3 messages, with request to continue
					stage3MessagesTry = append(stage3MessagesTry, llm.SetRawResponse(llm.NewMessage(llm.SimulatedAIResponse), aiResponses[0]))
					stage3MessagesTry = append(stage3MessagesTry, llm.AddPlainTextFragment(
						llm.NewMessage(llm.UserRequest),
						cfg[config.K_ImplementStage3ContinuePrompt].(string)))
				}

				// Append response fragment
				responses = append(responses, aiResponses[0])
			}
			if fileRetry {
				continue
			}

			// Remove extra output tag from the start from non first response-fragments
			for i := range responses {
				if i > 0 {
					responses[i], err = utils.GetTextAfterFirstMatches(
						responses[i],
						getEvenIndexElements(utils.InterfaceToStringArray(cfg[config.K_CodeTagsRx])))
					if err != nil {
						logger.Panicln("Error while parsing output response fragment:", err)
					}
				}
			}

			// Parse LLM output, detect file body in response
			combinedResponse := strings.Join(responses, "")
			fileBodies, err = utils.ParseMultiTaggedText(
				combinedResponse,
				getEvenIndexElements(utils.InterfaceToStringArray(cfg[config.K_CodeTagsRx])),
				getOddIndexElements(utils.InterfaceToStringArray(cfg[config.K_CodeTagsRx])),
				ignoreUnclosedTagErrors)
			if err != nil {
				if onFailRetriesLeft < 1 {
					logger.Errorln("Error while parsing LLM response with output file:", err)
				} else {
					logger.Warnln("Error while parsing LLM response with output file, retrying:", err)
					continue
				}
				// Try to remove only first match then, last resort
				fileBody, err := utils.GetTextAfterFirstMatches(
					combinedResponse,
					getEvenIndexElements(utils.InterfaceToStringArray(cfg[config.K_CodeTagsRx])))
				if err != nil {
					logger.Panicln("Error while parsing body from combined fragments:", err)
				}
				fileBodies = []string{fileBody}
			}

			if len(fileBodies) > 1 {
				if onFailRetriesLeft < 1 {
					logger.Errorln("Multiple file bodies detected in LLM response")
				} else {
					logger.Warnln("Multiple file bodies detected in LLM response")
					continue
				}
			}

			if len(fileBodies) < 1 {
				if onFailRetriesLeft < 1 {
					logger.Errorln("No file bodies detected in LLM response")
				} else {
					logger.Warnln("No file bodies detected in LLM response")
					continue
				}
			}

			// We are done here, exiting loop
			break
		}

		// Save body to processedFileContents and add record to processedFiles
		if len(fileBodies) > 0 {
			logger.Debugln("Found output for:", pendingFile)
			processedFileContents[pendingFile] = fileBodies[0]
			processedFiles = append(processedFiles, pendingFile)
		}
	}

	return processedFileContents
}

func getEvenIndexElements(arr []string) []string {
	var evenIndexElements []string
	for i := 0; i < len(arr); i += 2 {
		evenIndexElements = append(evenIndexElements, arr[i])
	}
	return evenIndexElements
}

func getOddIndexElements(arr []string) []string {
	var evenIndexElements []string
	for i := 1; i < len(arr); i += 2 {
		evenIndexElements = append(evenIndexElements, arr[i])
	}
	return evenIndexElements
}
