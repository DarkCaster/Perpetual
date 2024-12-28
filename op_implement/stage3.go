package op_implement

import (
	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage3(projectRootDir string,
	perpetualDir string,
	cfg config.Config,
	filesToMdLangMappings [][]string,
	planningMode int,
	allFileNames []string,
	filesForReview []string,
	targetFiles []string,
	messages []llm.Message,
	logger logging.ILogger) ([]llm.Message, []string, []string) {

	logger.Traceln("Stage3: Starting")
	defer logger.Traceln("Stage3: Finished")

	// Create stage3 llm connector
	stage3Connector, err := llm.NewLLMConnector(OpName+"_stage3", cfg.String(config.K_SystemPrompt), filesToMdLangMappings, map[string]interface{}{}, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage3 LLM connector:", err)
	}
	logger.Debugln(stage3Connector.GetDebugString())

	// Resulted filenames
	var targetFilesToModify []string
	var otherFilesToModify []string

	// When planning disabled, just copy target files into results without real LLM interaction in order to save tokens
	if planningMode == 0 {
		logger.Infoln("Running stage3: planning disabled")
		targetFilesToModify = append(targetFilesToModify, targetFiles...)
		logger.Debugln("Target files added to modify list")
	}

	// When using planning without reasoning, create request that will include target files content
	if planningMode == 1 {
		request := llm.ComposeMessageWithFiles(
			projectRootDir,
			cfg.String(config.K_ImplementStage3PlanningPrompt),
			targetFiles,
			cfg.StringArray(config.K_FilenameTags),
			logger)
		// Add message to history
		messages = append(messages, request)
		logger.Debugln("Files-to-change request message created (full)")
	}

	// When using planning WITH reasoning, create request that will only ask to create list of files to be changed
	if planningMode == 2 {
		request := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), cfg.String(config.K_ImplementStage3PlanningLitePrompt))
		//TODO: place json-mode request here
		// Add message to history
		messages = append(messages, request)
		logger.Debugln("Files-to-change request message created (simplified)")
	}

	// Send request
	if planningMode > 0 {
		filesToProcessRaw := []string{}
		onFailRetriesLeft := stage3Connector.GetOnFailureRetryLimit()
		if onFailRetriesLeft < 1 {
			onFailRetriesLeft = 1
		}
		// Make request and retry on errors
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			// Request LLM to provide file list that will be modified (or created) while implementing code
			logger.Infoln("Running stage3: requesting list of files for changes")
			var status llm.QueryStatus
			aiResponses, status, err := stage3Connector.Query(1, messages...)
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
			if len(aiResponses) < 1 || aiResponses[0] == "" {
				filesToProcessRaw = []string{}
				if onFailRetriesLeft < 1 {
					logger.Errorln("Got empty response from AI")
				} else {
					logger.Warnln("Got empty response from AI, retrying")
				}
				continue
			}
			//TODO: add list-parsing from JSON here
			// Process response, parse files that will be created
			filesToProcessRaw, err = utils.ParseTaggedTextRx(
				aiResponses[0],
				cfg.RegexpArray(config.K_FilenameTagsRx)[0],
				cfg.RegexpArray(config.K_FilenameTagsRx)[1],
				false)
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

		// Sort and filter file list provided by LLM
		logger.Debugln("Raw file-list to modify by LLM:", filesToProcessRaw)
		logger.Infoln("Files to modify selected by LLM:")
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
				logger.Errorln("Not using file mentioned by LLM, because it is outside project root directory", check)
				continue
			}
			// Sort files selected by LLM
			file, found := utils.CaseInsensitiveFileSearch(file, targetFiles)
			if found {
				file, found := utils.CaseInsensitiveFileSearch(file, targetFilesToModify)
				if found {
					logger.Warnln("Skipping file that already in target files:", file)
				} else {
					// This file among files to modify
					targetFilesToModify = append(targetFilesToModify, file)
					logger.Infoln(file, "(requested by User)")
				}
			} else {
				file, found := utils.CaseInsensitiveFileSearch(file, otherFilesToModify)
				if found {
					logger.Warnln("Skipping file that already in requested files:", file)
				} else {
					// Check if this file among files for review or not
					file, found := utils.CaseInsensitiveFileSearch(file, filesForReview)
					if found {
						otherFilesToModify = append(otherFilesToModify, file)
						logger.Infoln(file, "(requested by LLM)")
					} else {
						// Check if this file conflicts with any other file inside project directory
						file, found = utils.CaseInsensitiveFileSearch(file, allFileNames)
						if found {
							logger.Panicln("File requested by LLM is among project files not provided for review, this will cause file corruption:", file)
						}
						otherFilesToModify = append(otherFilesToModify, file)
						logger.Infoln(file, "(requested by LLM, new file)")
					}
				}
			}
		}
		logger.Debugln("Files to modify parsed")

		// Generate simulated AI message, with list of files
		response := llm.NewMessage(llm.SimulatedAIResponse)
		for _, item := range otherFilesToModify {
			response = llm.AddTaggedFragment(response, item, cfg.StringArray(config.K_FilenameTags))
		}
		for _, item := range targetFilesToModify {
			response = llm.AddTaggedFragment(response, item, cfg.StringArray(config.K_FilenameTags))
		}
		messages = append(messages, response)
		logger.Debugln("File-list response message created")
	}

	return messages, otherFilesToModify, targetFilesToModify
}
