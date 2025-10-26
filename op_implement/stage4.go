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
	noIncrMode bool,
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

	if noIncrMode {
		logger.Warnln("Incremental file-modification mode is manually disabled")
	}

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
					prCfg.Tags(config.K_ProjectFilenameTags))
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

		// Create prompt for incremental search-and-replace processing
		stage4ProcessIncrPrompt, err := utils.ReplaceTagRx(
			cfg.String(config.K_ImplementStage4ProcessIncrPrompt),
			cfg.Regexp(config.K_ImplementFilenameEmbedRx),
			pendingFile)
		if err != nil {
			logger.Errorln("Failed to replace filename tag", err)
			stage4ProcessIncrPrompt = cfg.String(config.K_ImplementStage4ProcessIncrPrompt)
		}

		useIncrMode := true
		if noIncrMode {
			useIncrMode = false
		} else {
			useIncrMode = connector.GetIncrModeSupport()
		}

		if useIncrMode {
			//detect if we can use incremental mode depending on file size and filename match
			_, fileSize, err := llm.GetSourceFileFromCache(pendingFile)
			if err != nil {
				useIncrMode = false
				logger.Infoln("Not using incremental mode, new file:", pendingFile)
			} else {
				matcher := prCfg.TextMatcherInteger(config.K_ProjectFilesIncrModeMinLen)
				if ok, v := matcher.TryMatch(pendingFile); ok {
					if fileSize < v[0] {
						useIncrMode = false
						logger.Infoln("Not using incremental mode, file too small:", pendingFile)
					}
				} else {
					logger.Infoln("Not using incremental mode for file:", pendingFile)
				}
			}
		}

		var fileBodies []string
		onFailRetriesLeft := max(connector.GetOnFailureRetryLimit(), 1)
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			// Create a copy, so it will be discarded on retry
			stage4MessagesTry := utils.NewSlice(stage4Messages...)
			// Create LLM message for incremental search-and-replace mode, or for regular mode
			if useIncrMode {
				stage4MessagesTry = append(stage4MessagesTry, llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), stage4ProcessIncrPrompt))
			} else {
				stage4MessagesTry = append(stage4MessagesTry, llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), stage4ProcessFilePrompt))
			}
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
					if useIncrMode {
						// Try to recover other parts of the file if reached max tokens
						if onFailRetriesLeft < 1 {
							logger.Errorln("LLM query reached token limit, no retries left!")
						} else {
							logger.Warnln("LLM query reached token limit, turning off incremental mode and retrying")
							useIncrMode = false
							fileRetry = true
							break
						}
					} else if generateTry >= connector.GetMaxTokensSegments() {
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

			if useIncrMode {
				//TODO:
				logger.Panicln("TODO: incremental mode parsing")
				// check we have inconsistent number of responses generated
				if len(responses) != 1 {
					//TODO
				}
				// filter search-and-replace blocks from response
				blocks, err := utils.ParseIncrBlocks(responses[0], prCfg.RegexpArray(config.K_ProjectFilesIncrModeRx))
				if err != nil {
					//TODO
				}
				if len(blocks) < 1 {
					useIncrMode = false
					logger.Warnln("No incremental search-and-replace blocks found in response, trying regular mode")
				} else {
					// apply additional code-blocks filter to the search-and-replace blocks
					blocks = utils.FilterIncrBlocks(blocks,
						utils.GetEvenRegexps(prCfg.RegexpArray(config.K_ProjectCodeTagsRx)),
						utils.GetOddRegexps(prCfg.RegexpArray(config.K_ProjectCodeTagsRx)))
					// get file from cache
					fileBody, _, err := llm.GetSourceFileFromCache(pendingFile)
					if err != nil {
						//should not occur, so this is an internal error
						logger.Panicf("Failed to get file from cache, trying regular mode: %v", err)
					}
					bool changedOk = true
					// iterate over each search-and-replace block
					for _, block := range blocks {
						// search requested text
						// replace requested text
					}
					if changedOk {
						// save modified file contents
						fileBodies = []string{fileBody}
					}
				}
			}

			if !useIncrMode {
				// Remove extra output tag from the start from non first response-fragments
				//TODO: fix: remove only in last response each time
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
