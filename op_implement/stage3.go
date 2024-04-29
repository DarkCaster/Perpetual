package op_implement

import (
	"path/filepath"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/prompts"
	"github.com/DarkCaster/Perpetual/utils"
	"github.com/sirupsen/logrus"
)

// Perform the actual code implementation process based on the Stage2 answer, which includes the contents of other files related to the files for which we need to implement the code.
func Stage3(projectRootDir string, perpetualDir string, promptsDir string, systemPrompt string, outputTagsRxStrings []string, fileNameEmbedTag string,
	stage2Messages []llm.Message, otherFiles []string, targetFiles []string, logger *logrus.Logger) map[string]string {

	logger.Traceln("Stage3: Starting")       // Add trace logging
	defer logger.Traceln("Stage3: Finished") // Add trace logging

	// Create stage3 llm connector
	stage3Connector, err := llm.NewLLMConnector(OpName+"_stage3", systemPrompt, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Fatalln("failed to create stage3 LLM connector:", err)
	}

	loadPrompt := func(filePath string, errorMsg string) string {
		bytes, err := utils.LoadTextFile(filepath.Join(promptsDir, filePath))
		if err != nil {
			logger.Fatalln(errorMsg, err)
		}
		return string(bytes)
	}

	stage3ChangesDonePromptTemplate := loadPrompt(prompts.ImplementStage3ChangesDonePromptFile, "failed to load implement stage3-changes prompt file")
	stage3ChangesDoneResponse := loadPrompt(prompts.AIImplementStage3ChangesDoneResponseFile, "failed to load implement stage3-changes ai response file")
	stage3ProcessFilePromptTemplate := loadPrompt(prompts.ImplementStage3ProcessFilePromptFile, "failed to load implement stage3-process-file prompt file")

	processedFileContents := make(map[string]string)
	var processedFiles []string

	// Main processing loop
	for workPending := true; workPending; workPending = len(otherFiles) > 0 || len(targetFiles) > 0 {
		logger.Debugln("Stage3: Work pending:", workPending) // Add debug logging

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

		logger.Debugln("Stage3: Processing file:", pendingFile) // Add debug logging
		// Create prompt from stage3ProcessFilePromptTemplate
		stage3ProcessFilePrompt, err := utils.ReplaceTag(stage3ProcessFilePromptTemplate, fileNameEmbedTag, pendingFile)
		if err != nil {
			logger.Errorln("Failed to replace filename tag at stage3-process-file prompt:", err)
			stage3ProcessFilePrompt = stage3ProcessFilePromptTemplate
		}

		// Create prompt for to implement one of the files
		stage3Messages = append(stage3Messages, llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), stage3ProcessFilePrompt))

		// Log messages we are going to send
		llm.LogMessages(logger, perpetualDir, stage3Connector, stage3Messages)

		logger.Println("Running stage3: implementing code for:", pendingFile)
		aiResponse, err := stage3Connector.Query(stage3Messages...)
		if err != nil {
			logger.Fatalln("LLM query failed: ", err)
		}

		// Log LLM response
		responseMessage := llm.SetRawResponse(llm.NewMessage(llm.RealAIResponse), aiResponse)
		llm.LogMessage(logger, perpetualDir, stage3Connector, &responseMessage)

		// Parse LLM output, detect file body in response
		fileBodies, err := utils.ParseTaggedText(aiResponse, outputTagsRxStrings[0], outputTagsRxStrings[1])
		if err != nil {
			logger.Errorln("Error while parsing LLM response with output file:", err)
		}

		// Save body to processedFileContents and add record to processedFiles
		if len(fileBodies) > 0 {
			logger.Debugln("Stage3: Found output for:", pendingFile)
			processedFileContents[pendingFile] = fileBodies[0]
			processedFiles = append(processedFiles, pendingFile)
		} else {
			logger.Errorln("No output found for file:", pendingFile)
		}
	}

	return processedFileContents
}
