package op_implement

import (
	"fmt"
	"slices"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage3(projectRootDir string,
	perpetualDir string,
	prCfg config.Config,
	opCfg config.Config,
	filesToMdLangMappings config.TextMatcher[string],
	planningMode int,
	allFileNames []string,
	filesForReview []string,
	targetFiles []string,
	notEnforceTargetFiles bool,
	messages []llm.Message,
	task string,
	logger logging.ILogger) ([]llm.Message, []string, []string) {

	logger.Traceln("Stage3: Starting")
	defer logger.Traceln("Stage3: Finished")

	// Create stage3 llm connector
	connector, err := llm.NewLLMConnector(
		OpName+"_stage3",
		opCfg.String(config.K_SystemPrompt),
		opCfg.String(config.K_SystemPromptAck),
		filesToMdLangMappings,
		opCfg.Object(config.K_ImplementStage3OutputSchema),
		opCfg.String(config.K_ImplementStage3OutputSchemaName),
		opCfg.String(config.K_ImplementStage3OutputSchemaDesc),
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage3 LLM connector:", err)
	}

	// Initial position in message history to append content of extra target-files found out at this stage
	msgIndexToAddExtraFiles := max(len(messages)-2, 0)

	// Resulted filenames
	var targetFilesToModify []string
	var otherFilesToModify []string

	if !notEnforceTargetFiles || planningMode == 0 {
		targetFilesToModify = append(targetFilesToModify, targetFiles...)
		logger.Debugln("Target files added to modify list")
	}

	// When planning disabled, just copy target files into results without real LLM interaction in order to save tokens
	if planningMode == 0 {
		logger.Infoln("Running stage3: planning disabled, not generating list of files for processing")
	}

	// Declare jsonModeMessages, it will be used as messages history sent to llm when using json mode
	jsonModeMessages := utils.NewSlice(messages...)

	// When using planning without reasoning, create request that will include target files content
	if planningMode == 1 {
		// Create normal mode request and add it to history
		var request llm.Message
		if task == "" {
			request = llm.ComposeMessageWithFiles(
				projectRootDir,
				opCfg.String(config.K_ImplementStage3PlanningPrompt),
				targetFiles,
				prCfg.StringArray(config.K_ProjectFilenameTags),
				logger)
		} else {
			request = llm.AddPlainTextFragment(
				llm.AddPlainTextFragment(
					llm.NewMessage(llm.UserRequest),
					opCfg.String(config.K_ImplementTaskStage3PlanningPrompt)),
				task)
		}
		messages = append(messages, request)
		msgIndexToAddExtraFiles = len(messages) - 1

		// Create json mode request and add it to json mode history
		var jsonModeRequest llm.Message
		if task == "" {
			jsonModeRequest = llm.ComposeMessageWithFiles(
				projectRootDir,
				opCfg.String(config.K_ImplementStage3PlanningJsonModePrompt),
				targetFiles,
				prCfg.StringArray(config.K_ProjectFilenameTags),
				logger)
		} else {
			jsonModeRequest = llm.AddPlainTextFragment(
				llm.AddPlainTextFragment(
					llm.NewMessage(llm.UserRequest),
					opCfg.String(config.K_ImplementTaskStage3PlanningJsonModePrompt)),
				task)
		}
		jsonModeMessages = append(jsonModeMessages, jsonModeRequest)
		logger.Debugln("Files-to-change request message created (full)")
	}

	// When using planning WITH reasoning, create request that will only ask to create list of files to be changed
	if planningMode == 2 {
		// Create normal mode request and add it to history
		request := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), opCfg.String(config.K_ImplementStage3PlanningLitePrompt))
		messages = append(messages, request)
		// Create json mode request and add it to json mode history
		jsonModeRequest := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), opCfg.String(config.K_ImplementStage3PlanningLiteJsonModePrompt))
		jsonModeMessages = append(jsonModeMessages, jsonModeRequest)
		logger.Debugln("Files-to-change request message created (simplified)")
	}

	// Send request
	if planningMode > 0 {
		logger.Infoln("Running stage3: generating list of files for processing")
		debugString := connector.GetDebugString()
		logger.Notifyln(debugString)
		llm.GetSimpleRawMessageLogger(perpetualDir)(fmt.Sprintf("=== Implement (stage 3): %s\n\n\n", debugString))

		var filesToProcessRaw []string
		onFailRetriesLeft := connector.GetOnFailureRetryLimit()
		if onFailRetriesLeft < 1 {
			onFailRetriesLeft = 1
		}
		// Make request and retry on errors
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			// Request LLM to provide file list that will be modified (or created) while implementing code
			var status llm.QueryStatus
			// Select messages to send, depending on mode
			targetMessages := messages
			if connector.GetOutputFormat() == llm.OutputJson {
				targetMessages = jsonModeMessages
			}
			aiResponses, status, err := connector.Query(1, targetMessages...)
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
			if len(aiResponses) < 1 || aiResponses[0] == "" {
				if onFailRetriesLeft < 1 {
					logger.Panicln("Got empty response from AI")
				} else {
					logger.Warnln("Got empty response from AI, retrying")
				}
				continue
			}
			// Process response, parse files that will be created
			if connector.GetOutputFormat() == llm.OutputJson {
				// Use json-mode parsing
				filesToProcessRaw, err = utils.ParseListFromJSON(aiResponses[0], opCfg.String(config.K_ImplementStage3OutputKey))
			} else {
				// Use regular parsing to extract file-list
				filesToProcessRaw, err = utils.ParseTaggedTextRx(
					aiResponses[0],
					prCfg.RegexpArray(config.K_ProjectFilenameTagsRx)[0],
					prCfg.RegexpArray(config.K_ProjectFilenameTagsRx)[1],
					false)
			}
			if err != nil {
				if onFailRetriesLeft < 1 {
					logger.Panicln("Failed to parse list of files for review", err)
				} else {
					logger.Warnln("Failed to parse list of files for review, retrying", err)
				}
				continue
			}
			break
		}

		extra_task_prompt_added := false
		newFilesIndex := 0
		// Sort and filter file list provided by LLM
		logger.Debugln("Raw file-list to modify by LLM:", filesToProcessRaw)
		logger.Infoln("Files for processing selected by LLM:")
		for _, check := range filesToProcessRaw {
			// remove new line from the end of filename, if present
			if check != "" && check[len(check)-1] == '\n' {
				check = check[:len(check)-1]
			}
			// remove \r from the end of filename, if present
			if check != "" && check[len(check)-1] == '\r' {
				check = check[:len(check)-1]
			}
			//replace possibly-invalid path separators
			check = utils.ConvertFilePathToOSFormat(check)
			//make file path relative to project root
			file, err := utils.MakePathRelative(projectRootDir, check, true)
			if err != nil {
				logger.Errorln("Not using file, because it is outside project root directory", check)
				continue
			}
			// Sort files selected by LLM
			file, found := utils.CaseInsensitiveFileSearch(file, targetFiles)
			if found {
				file, found := utils.CaseInsensitiveFileSearch(file, targetFilesToModify)
				if found {
					logger.Debugln("Skipping file that already among target files:", file)
				} else {
					// This file among files to modify
					targetFilesToModify = append(targetFilesToModify, file)
					logger.Infoln(file, "(among initial target files)")
				}
			} else {
				file, found := utils.CaseInsensitiveFileSearch(file, otherFilesToModify)
				if found {
					logger.Warnln("Skipping already requested file:", file)
				} else {
					// Check if this file among files for review or not
					file, found := utils.CaseInsensitiveFileSearch(file, filesForReview)
					if found {
						otherFilesToModify = append(otherFilesToModify, file)
						logger.Infoln(file)
					} else {
						// Check if this file conflicts with any other file inside project directory
						file, found = utils.CaseInsensitiveFileSearch(file, allFileNames)
						if found {
							// Add extra prompt indicating addition of file content if using task mode
							if task != "" && !extra_task_prompt_added {
								extra_task_prompt_added = true
								messages[msgIndexToAddExtraFiles] = llm.AddPlainTextFragment(
									messages[msgIndexToAddExtraFiles],
									opCfg.String(config.K_ImplementTaskStage3ExtraFilesPrompt))
							}
							// Add the file contents so that LLM doesn't overwrite it from scratch, thus destroying it.
							messages[msgIndexToAddExtraFiles] = llm.AppendFileToMessage(
								messages[msgIndexToAddExtraFiles],
								projectRootDir,
								file,
								prCfg.StringArray(config.K_ProjectFilenameTags),
								logger)
							otherFilesToModify = append(otherFilesToModify, file)
							logger.Warnln("File exist in the project but was not requested previously, adding it to avoid corruption", file)
						} else {
							//insert new files to the beginning of file-list
							otherFilesToModify = slices.Insert(otherFilesToModify, newFilesIndex, file)
							newFilesIndex++
							logger.Infoln(file, "(new file)")
						}
					}
				}
			}
		}
		logger.Debugln("Files to modify parsed")

		// Generate simulated AI message, with list of files
		response := llm.NewMessage(llm.SimulatedAIResponse)
		for _, item := range otherFilesToModify {
			response = llm.AddTaggedFragment(response, item, prCfg.StringArray(config.K_ProjectFilenameTags))
		}
		for _, item := range targetFilesToModify {
			response = llm.AddTaggedFragment(response, item, prCfg.StringArray(config.K_ProjectFilenameTags))
		}
		// Add response to the normal-mode message history, because it better aligns with tags used to denote the filenames
		messages = append(messages, response)
		logger.Debugln("File-list response message created")
	}

	// Always return normal-mode message history
	return messages, otherFilesToModify, targetFilesToModify
}
