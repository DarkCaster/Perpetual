package op_implement

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage3(projectRootDir string,
	perpetualDir string,
	prCfg config.Config,
	opCfg config.Config,
	filesToMdLangMappings utils.TextMatcher[string],
	planningMode bool,
	allFileNames []string,
	projectFilesWhitelist []*regexp.Regexp,
	projectFilesBlacklist []*regexp.Regexp,
	noUploadRx []*regexp.Regexp,
	forceUpload bool,
	filesForReview []string,
	targetFiles []string,
	messages []llm.Message,
	task string,
	logger logging.ILogger) ([]llm.Message, []string, []string, []string) {

	logger.Traceln("Stage3: Starting")
	defer logger.Traceln("Stage3: Finished")

	// Create stage3 llm connector
	connector, err := llm.NewLLMConnector(
		OpName+"_stage3",
		opCfg.String(config.K_SystemPrompt),
		opCfg.String(config.K_SystemPromptAck),
		filesToMdLangMappings,
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage3 LLM connector:", err)
	}

	// Initial position in message history to append content of extra target-files found out at this stage
	msgIndexToAddExtraFiles := max(len(messages)-2, 0)

	// Resulted filenames
	var targetFilesToModify []string
	var otherFilesToModify []string
	var filesToDelete []string

	// Send request
	if planningMode {
		// Create request that will ask to create list of files to be changed
		request := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), opCfg.String(config.K_ImplementStage3PlanningPrompt))
		messages = append(messages, request)
		logger.Debugln("Files-to-change request message created")

		logger.Infoln("Running stage3: generating list of files for processing")
		debugString := connector.GetDebugString()
		logger.Notifyln(debugString)
		llm.GetSimpleRawMessageLogger(perpetualDir)(fmt.Sprintf("=== Implement (stage 3): %s\n\n\n", debugString))

		var filesToProcessRaw []string
		var filesToDeleteRaw []string
		onFailRetriesLeft := max(connector.GetOnFailureRetryLimit(), 1)
		// Make request and retry on errors
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			// Request LLM to provide file list that will be modified (or created) while implementing code
			var status llm.QueryStatus
			aiResponse, status, err := connector.Query(false, messages...)
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
			if len(aiResponse) < 1 {
				if onFailRetriesLeft < 1 {
					logger.Panicln("Got empty response from AI")
				} else {
					logger.Warnln("Got empty response from AI, retrying")
				}
				continue
			}
			// Process response, parse files that will be created or modified
			filesToProcessRaw, err = utils.ParseTaggedTextRx(
				aiResponse,
				prCfg.RegexpArray(config.K_ProjectFilenameTagsRx)[0],
				prCfg.RegexpArray(config.K_ProjectFilenameTagsRx)[1],
				false)
			if err != nil {
				if onFailRetriesLeft < 1 {
					logger.Panicln("Failed to parse list of files for review", err)
				} else {
					logger.Warnln("Failed to parse list of files for review, retrying", err)
				}
				continue
			}
			// Process response, parse files that will be deleted
			filesToDeleteRaw, err = utils.ParseTaggedTextRx(
				aiResponse,
				prCfg.RegexpArray(config.K_ProjectDeleteTagsRx)[0],
				prCfg.RegexpArray(config.K_ProjectDeleteTagsRx)[1],
				false)
			if err != nil {
				if onFailRetriesLeft < 1 {
					logger.Panicln("Failed to parse list of files for deletion", err)
				} else {
					logger.Warnln("Failed to parse list of files for deletion, retrying", err)
				}
				continue
			}
			break
		}

		normalizeLLMRequestedFile := func(raw string) (string, bool) {
			check := raw

			if check != "" && check[len(check)-1] == '\n' {
				check = check[:len(check)-1]
			}
			if check != "" && check[len(check)-1] == '\r' {
				check = check[:len(check)-1]
			}

			check = utils.ConvertFilePathToOSFormat(check)

			file, err := utils.MakePathRelative(projectRootDir, check, true)
			if err != nil {
				logger.Errorln("Not using file, because it is outside project root directory", check)
				return "", false
			}

			return file, true
		}

		removeFileCaseInsensitive := func(files []string, target string) ([]string, bool) {
			for i, file := range files {
				if strings.EqualFold(file, target) {
					return append(files[:i], files[i+1:]...), true
				}
			}
			return files, false
		}

		extraTaskPromptAdded := false
		alreadyAddedUnexpectedFiles := map[string]struct{}{}

		appendUnexpectedExistingFile := func(file string, ignoreNoUploadFilter bool) {
			// only try injecting unexpected existing file once
			if _, exist := alreadyAddedUnexpectedFiles[file]; exist {
				return
			}
			alreadyAddedUnexpectedFiles[file] = struct{}{}

			// extra no-upload protection when trying to inject unexpected existing file contents to the LLM context
			if !forceUpload {
				files := utils.FilterNoUploadProjectFiles(projectRootDir, []string{file}, noUploadRx, true, logger)
				if len(files) < 1 {
					if ignoreNoUploadFilter {
						logger.Warnln("Ignoring no-upload filter to avoid file corruption:", file)
					} else {
						return
					}
				}
			}

			logger.Warnln("File exist in the project but was not requested previously, adding it to avoid corruption", file)

			if task != "" && !extraTaskPromptAdded {
				extraTaskPromptAdded = true
				messages[msgIndexToAddExtraFiles] = llm.AddPlainTextFragment(
					messages[msgIndexToAddExtraFiles],
					opCfg.String(config.K_ImplementStage3ExtraFilesPrompt))
			}

			messages[msgIndexToAddExtraFiles] = llm.AppendSourceFileToMessage(
				messages[msgIndexToAddExtraFiles],
				projectRootDir,
				file,
				prCfg.Tags(config.K_ProjectFilenameTags),
				logger)
		}

		newFilesIndex := 0
		// Sort and filter file list provided by LLM
		logger.Debugln("Raw file-list to modify by LLM:", filesToProcessRaw)
		logger.Infoln("Files for processing selected by LLM:")
		if len(filesToProcessRaw) < 1 {
			logger.Infoln("No files selected for processing")
		}
		for _, raw := range filesToProcessRaw {
			file, ok := normalizeLLMRequestedFile(raw)
			if !ok {
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
						// Check file against project black-list and white list
						if fileWS, wsdr := utils.FilterFilesWithWhitelist([]string{file}, projectFilesWhitelist); len(wsdr) > 0 {
							logger.Warnln("Skipping requested file, filtered by project whitelist:", file)
						} else if _, bsdr := utils.FilterFilesWithBlacklist(fileWS, projectFilesBlacklist); len(bsdr) > 0 {
							logger.Warnln("Skipping requested file, filtered by project blacklist:", file)
						} else if found {
							// Add the file contents so that LLM doesn't overwrite it from scratch, thus destroying it.
							appendUnexpectedExistingFile(file, true) // Even files marked with no-upload should be added to avoid its corruption
							otherFilesToModify = append(otherFilesToModify, file)
						} else {
							//insert new files to the beginning of file-list
							otherFilesToModify = slices.Insert(otherFilesToModify, newFilesIndex, file)
							//extra protection against adding non-existent file to context if attempting to delete it later
							alreadyAddedUnexpectedFiles[file] = struct{}{}
							newFilesIndex++
							logger.Infoln(file, "(new file)")
						}
					}
				}
			}
		}
		logger.Debugln("Files to modify parsed")

		logger.Debugln("Raw file-list to delete by LLM:", filesToDeleteRaw)
		logger.Infoln("Files for deletion selected by LLM:")
		if len(filesToDeleteRaw) < 1 {
			logger.Infoln("No files selected for deletion")
		}
		for _, raw := range filesToDeleteRaw {
			file, ok := normalizeLLMRequestedFile(raw)
			if !ok {
				continue
			}

			file, found := utils.CaseInsensitiveFileSearch(file, allFileNames)
			if !found {
				logger.Warnln("Skipping requested file deletion, file does not exist in project:", file)
				continue
			}

			if _, found := utils.CaseInsensitiveFileSearch(file, filesToDelete); found {
				logger.Warnln("Skipping already requested file deletion:", file)
				continue
			}

			if fileWS, wsdr := utils.FilterFilesWithWhitelist([]string{file}, projectFilesWhitelist); len(wsdr) > 0 {
				logger.Warnln("Skipping requested file deletion, filtered by project whitelist:", file)
				continue
			} else if _, bsdr := utils.FilterFilesWithBlacklist(fileWS, projectFilesBlacklist); len(bsdr) > 0 {
				logger.Warnln("Skipping requested file deletion, filtered by project blacklist:", file)
				continue
			}

			var removed bool

			otherFilesToModify, removed = removeFileCaseInsensitive(otherFilesToModify, file)
			if removed {
				logger.Warnln("Deletion overrides modification for file:", file)
			}

			targetFilesToModify, removed = removeFileCaseInsensitive(targetFilesToModify, file)
			if removed {
				logger.Warnln("Deletion overrides target-file modification for file:", file)
			}

			_, alreadyInReview := utils.CaseInsensitiveFileSearch(file, filesForReview)
			_, alreadyInTargets := utils.CaseInsensitiveFileSearch(file, targetFiles)

			if !alreadyInReview && !alreadyInTargets {
				appendUnexpectedExistingFile(file, false)
			}

			filesToDelete = append(filesToDelete, file)
			logger.Infoln(file, "(delete)")
		}
		logger.Debugln("Files to delete parsed")

		// Generate simulated AI message, with list of files
		response := llm.NewMessage(llm.SimulatedAIResponse)
		for _, item := range otherFilesToModify {
			response = llm.AddTaggedFragment(response, item, prCfg.Tags(config.K_ProjectFilenameTags))
		}
		for _, item := range targetFilesToModify {
			response = llm.AddTaggedFragment(response, item, prCfg.Tags(config.K_ProjectFilenameTags))
		}
		for _, item := range filesToDelete {
			response = llm.AddTaggedFragment(response, item, prCfg.Tags(config.K_ProjectDeleteTags))
		}
		// Add response to the message history
		messages = append(messages, response)
		logger.Debugln("File-list response message created")
	} else {
		logger.Infoln("Running stage3: planning disabled")
		targetFilesToModify = append(targetFilesToModify, targetFiles...)
		logger.Debugln("Target files added to modify list")
	}

	// apply filtering to the files requested by LLM that was not already validated on previous stages

	if !forceUpload {
		otherFilesToModify = utils.FilterNoUploadProjectFiles(projectRootDir, otherFilesToModify, noUploadRx, true, logger)
	}

	otherFilesToModify, droppedFiles := utils.FilterFilesWithWhitelist(otherFilesToModify, projectFilesWhitelist)
	for _, file := range droppedFiles {
		logger.Warnln("File was filtered-out with project whitelist:", file)
	}
	otherFilesToModify, droppedFiles = utils.FilterFilesWithBlacklist(otherFilesToModify, projectFilesBlacklist)
	for _, file := range droppedFiles {
		logger.Warnln("File was filtered-out with project or user blacklist:", file)
	}

	return messages, otherFilesToModify, targetFilesToModify, filesToDelete
}
