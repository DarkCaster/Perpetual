package op_implement

import (
	"path/filepath"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/prompts"
	"github.com/DarkCaster/Perpetual/utils"
	"github.com/sirupsen/logrus"
)

func Stage2(projectRootDir string, perpetualDir string, promptsDir string, systemPrompt string, planning bool, fileNameTagsRxStrings []string, fileNameTags []string,
	allFileNames []string, filesForReview []string, targetFiles []string, logger *logrus.Logger) ([]llm.Message, []string, []string) {

	logger.Traceln("Stage2: Starting")
	defer logger.Traceln("Stage2: Finished")

	// Create stage2 llm connector
	stage2Connector, err := llm.NewLLMConnector(OpName+"_stage2", systemPrompt, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Fatalln("failed to create stage2 LLM connector:", err)
	}

	loadPrompt := func(filePath string, errorMsg string) string {
		text, err := utils.LoadTextFile(filepath.Join(promptsDir, filePath))
		if err != nil {
			logger.Fatalln(errorMsg, err)
		}
		return text
	}

	var messages []llm.Message

	// Create target files analisys request message
	stage2ProjectSourceCodeMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(prompts.ImplementStage2ProjectCodePromptFile, "failed to load implement stage2-project-code prompt file:"))
	for _, item := range filesForReview {
		contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, item))
		if err != nil {
			logger.Fatalln("failed to add file contents to stage2 prompt", err)
		}
		stage2ProjectSourceCodeMessage = llm.AddFileFragment(stage2ProjectSourceCodeMessage, item, contents)
	}
	messages = append(messages, stage2ProjectSourceCodeMessage)
	logger.Debugln("Stage2: Project source code message created")

	// Create simulated response
	stage2ProjectSourceCodeResponseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), loadPrompt(prompts.AIImplementStage2ProjectCodeResponseFile, "failed to load ai-stage2-project-code response file:"))
	messages = append(messages, stage2ProjectSourceCodeResponseMessage)
	logger.Debugln("Stage2: Project source code simulated response added")

	if planning {
		// Create files to change request message
		stage2FilesToChangeMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(prompts.ImplementStage2FilesToChangePromptFile, "failed to load implement stage2-files-to-change prompt file:"))
		for _, item := range targetFiles {
			contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, item))
			if err != nil {
				logger.Fatalln("failed to add file contents to stage1 prompt", err)
			}
			stage2FilesToChangeMessage = llm.AddFileFragment(stage2FilesToChangeMessage, item, contents)
		}
		messages = append(messages, stage2FilesToChangeMessage)
		logger.Debugln("Stage2: Files to change message created")
	} else {
		// Create files to request for non-planning mode
		stage2FilesNoPlanningMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(prompts.ImplementStage2NoPlanningPromptFile, "failed to load implement stage2-no-planning prompt file:"))
		for _, item := range targetFiles {
			contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, item))
			if err != nil {
				logger.Fatalln("failed to add file contents to stage1 prompt", err)
			}
			stage2FilesNoPlanningMessage = llm.AddFileFragment(stage2FilesNoPlanningMessage, item, contents)
		}
		messages = append(messages, stage2FilesNoPlanningMessage)
		logger.Debugln("Stage2: Files for no planning message created")

		// Create simulated response
		stage2FilesNoPlanningResponseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), loadPrompt(prompts.AIImplementStage2NoPlanningResponseFile, "failed to load ai-stage2-no-planning response file:"))
		messages = append(messages, stage2FilesNoPlanningResponseMessage)
		logger.Debugln("Stage2: Files for no planning simulated response added")
	}

	// Log messages
	llm.LogMessages(logger, perpetualDir, stage2Connector, messages)
	logger.Debugln("Stage2: Messages logged")

	// Resulted filenames
	var targetFilesToModify []string
	var otherFilesToModify []string

	// Only perform real planning step if enabled
	if planning {
		// Request LLM to provide file list that will be modified (or created) while implementing code
		logger.Println("Running stage2: planning changes")
		aiResponse, err := stage2Connector.Query(messages...)
		if err != nil {
			logger.Fatalln("LLM query failed: ", err)
		}
		logger.Traceln("Stage2: LLM query completed")

		// Log LLM response
		responseMessage := llm.SetRawResponse(llm.NewMessage(llm.RealAIResponse), aiResponse)
		llm.LogMessage(logger, perpetualDir, stage2Connector, &responseMessage)
		logger.Debugln("Stage2: LLM response logged")

		// Process response, parse files that will be created
		filesToProcessRaw, err := utils.ParseTaggedText(aiResponse, fileNameTagsRxStrings[0], fileNameTagsRxStrings[1])
		if err != nil {
			logger.Fatalln("Failed to parse list of files for review", err)
		}
		logger.Traceln("Stage2: Files to process parsed")

		// Check all selected files
		logger.Println("Files to modify selected by LLM:")
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
					logger.Println(file, "(requested by User)")
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
						logger.Println(file, "(requested by LLM)")
					} else {
						// Check if this file conflicts with any other file inside project directory
						file, found = utils.CaseInsensitiveFileSearch(file, allFileNames)
						if found {
							logger.Fatalln("File requested by LLM is among project files not provided for review, this will cause file corruption:", file)
						}
						otherFilesToModify = append(otherFilesToModify, file)
						logger.Println(file, "(requested by LLM, new file)")
					}
				}
			}
		}
		logger.Debugln("Stage2: Files to modify parsed")

		// Generate simplified ai message, with just a list of files
		simplifiedResponseMessage := llm.NewMessage(llm.SimulatedAIResponse)
		for _, item := range otherFilesToModify {
			simplifiedResponseMessage = llm.AddTaggedFragment(simplifiedResponseMessage, item, fileNameTags)
		}
		for _, item := range targetFilesToModify {
			simplifiedResponseMessage = llm.AddTaggedFragment(simplifiedResponseMessage, item, fileNameTags)
		}

		// Log message before response, to mark it as logged here, because stage3 actively copying and reusing old messages
		llm.LogMessage(logger, perpetualDir, stage2Connector, &simplifiedResponseMessage)
		messages = append(messages, simplifiedResponseMessage)
		logger.Debugln("Stage2: Simplified response message created")
	} else {
		// Just copy target files into results without real LLM interaction in order to save tokens
		logger.Println("Running stage2: planning disabled")
		targetFilesToModify = append(targetFilesToModify, targetFiles...)
		logger.Debugln("Stage2: Target files added to modify list")
	}

	return messages, otherFilesToModify, targetFilesToModify
}
