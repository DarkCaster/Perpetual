package op_implement

import (
	"fmt"
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

// Perform the actual code implementation process based on the Stage 2 and 3 answers, which includes the contents of other files related to the files for which we need to implement the code and extra reasonings (if enabled).
func Stage4(projectRootDir string,
	perpetualDir string,
	prCfg config.Config,
	cfg config.Config,
	filesToMdLangMappings utils.TextMatcher[string],
	stage2Messages []llm.Message,
	otherFiles []string,
	targetFiles []string,
	logger logging.ILogger) map[string]string {

	logger.Traceln("Stage4: Starting")       // Add trace logging
	defer logger.Traceln("Stage4: Finished") // Add trace logging

	// Create stage4 llm connector
	connector, err := llm.NewLLMConnector(
		OpName+"_stage4",
		cfg.String(config.K_SystemPrompt),
		cfg.String(config.K_SystemPromptAck),
		filesToMdLangMappings,
		map[string]interface{}{},
		"", "",
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage4 LLM connector:", err)
	}

	processedFileContents := make(map[string]string)
	var processedFiles []string

	logger.Infoln("Running stage4: implementing code")
	debugString := connector.GetDebugString()
	logger.Notifyln(debugString)
	llm.GetSimpleRawMessageLogger(perpetualDir)(fmt.Sprintf("=== Implement (stage 4): %s\n\n\n", debugString))

	// Main processing loop
	for workPending := true; workPending; workPending = len(otherFiles) > 0 || len(targetFiles) > 0 {
		logger.Debugln("Work pending:", workPending) // Add debug logging

		// Generate change-done message from already processed files
		stage4ChangesDoneMessage := llm.AddPlainTextFragment(
			llm.NewMessage(llm.UserRequest),
			cfg.String(config.K_ImplementStage4ChangesDonePrompt))

		for _, item := range processedFiles {
			contents, ok := processedFileContents[item]
			if !ok {
				logger.Errorln("Failed to add file contents to stage4:", err)
			} else {
				stage4ChangesDoneMessage = llm.AddFileFragment(
					stage4ChangesDoneMessage,
					item,
					contents,
					prCfg.StringArray(config.K_ProjectFilenameTags))
			}
		}

		stage4ChangesDoneResponseMessage := llm.AddPlainTextFragment(
			llm.NewMessage(llm.SimulatedAIResponse),
			cfg.String(config.K_ImplementStage4ChangesDoneResponse))

		stage4Messages := utils.NewSlice(stage2Messages...)
		if len(processedFiles) > 0 {
			stage4Messages = append(stage4Messages, stage4ChangesDoneMessage)
			stage4Messages = append(stage4Messages, stage4ChangesDoneResponseMessage)
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

		logger.Infoln(pendingFile)
		llm.GetSimpleRawMessageLogger(perpetualDir)(fmt.Sprintf("=== Implement (stage 4): %s\n\n\n", pendingFile))

		// Create prompt from stage4ProcessFilePromptTemplate
		stage4ProcessFilePrompt, err := utils.ReplaceTagRx(
			cfg.String(config.K_ImplementStage4ProcessPrompt),
			cfg.Regexp(config.K_ImplementFilenameEmbedRx),
			pendingFile)

		if err != nil {
			logger.Errorln("Failed to replace filename tag", err)
			stage4ProcessFilePrompt = cfg.String(config.K_ImplementStage4ProcessPrompt)
		}

		// Create prompt for to implement one of the files
		stage4Messages = append(stage4Messages, llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), stage4ProcessFilePrompt))

		var fileBodies []string
		onFailRetriesLeft := connector.GetOnFailureRetryLimit()
		if onFailRetriesLeft < 1 {
			onFailRetriesLeft = 1
		}
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			// Create a copy, so it will be discarded on file retry
			stage4MessagesTry := utils.NewSlice(stage4Messages...)
			// Initialize temporary variables for handling partial answers
			var responses []string
			continueGeneration := true
			ignoreUnclosedTagErrors := false
			generateTry := 1
			fileRetry := false
			for continueGeneration && !fileRetry {
				// Run query
				continueGeneration = false
				aiResponses, status, err := connector.Query(1, stage4MessagesTry...)
				if perfString := connector.GetPerfString(); perfString != "" {
					logger.Traceln(perfString)
				}
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
						// Disable some possible parsing errors in future
						ignoreUnclosedTagErrors = true
					}
					// Add partial response to stage4 messages, with request to continue
					stage4MessagesTry = append(stage4MessagesTry, llm.SetRawResponse(llm.NewMessage(llm.SimulatedAIResponse), aiResponses[0]))
					stage4MessagesTry = append(stage4MessagesTry, llm.AddPlainTextFragment(
						llm.NewMessage(llm.UserRequest),
						cfg.String(config.K_ImplementStage4ContinuePrompt)))
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
					responses[i] = utils.GetTextAfterFirstMatchesRx(responses[i], utils.GetEvenRegexps(prCfg.RegexpArray(config.K_ProjectCodeTagsRx)))
				}
			}

			// Parse LLM output, detect file body in response
			combinedResponse := strings.Join(responses, "")
			fileBodies, err = utils.ParseMultiTaggedTextRx(
				combinedResponse,
				utils.GetEvenRegexps(prCfg.RegexpArray(config.K_ProjectCodeTagsRx)),
				utils.GetOddRegexps(prCfg.RegexpArray(config.K_ProjectCodeTagsRx)),
				ignoreUnclosedTagErrors)
			if err != nil {
				if onFailRetriesLeft < 1 {
					logger.Errorln("Error while parsing LLM response with output file:", err)
				} else {
					logger.Warnln("Error while parsing LLM response with output file, retrying:", err)
					continue
				}
				// Try to remove only first match then, last resort
				fileBody := utils.GetTextAfterFirstMatchesRx(combinedResponse, utils.GetEvenRegexps(prCfg.RegexpArray(config.K_ProjectCodeTagsRx)))
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
