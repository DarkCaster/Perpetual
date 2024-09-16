package op_doc

import (
	"path/filepath"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/prompts"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage2(projectRootDir string, perpetualDir string, promptsDir string, systemPrompt string, filesToMdLangMappings [][2]string, fileNameTagsRxStrings []string, fileNameTags []string, projectFiles []string, filesForReview []string, annotations map[string]string, targetDocument string, action string, logger logging.ILogger) {

	logger.Traceln("Stage2: Starting")
	defer logger.Traceln("Stage2: Finished")

	// Create stage2 llm connector
	stage2Connector, err := llm.NewLLMConnector(OpName+"_stage2", systemPrompt, filesToMdLangMappings, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage2 LLM connector:", err)
	}
	logger.Debugln(llm.GetDebugString(stage2Connector))

	loadPrompt := func(filePath string) string {
		text, err := utils.LoadTextFile(filepath.Join(promptsDir, filePath))
		if err != nil {
			logger.Panicln("Failed to load prompt:", err)
		}
		return text
	}

	var messages []llm.Message

	// Create project-index request message
	stage2ProjectIndexRequestMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(prompts.DocProjectIndexPromptFile))
	for _, item := range projectFiles {
		stage2ProjectIndexRequestMessage = llm.AddIndexFragment(stage2ProjectIndexRequestMessage, item, fileNameTags)
		annotation := annotations[item]
		if annotation == "" {
			annotation = "No annotation available"
		}
		stage2ProjectIndexRequestMessage = llm.AddPlainTextFragment(stage2ProjectIndexRequestMessage, annotation)
	}
	messages = append(messages, stage2ProjectIndexRequestMessage)
	logger.Debugln("Created project-index request message")

	// Create project-index simulated response
	stage2ProjectIndexResponseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), loadPrompt(prompts.AIDocProjectIndexResponseFile))
	messages = append(messages, stage2ProjectIndexResponseMessage)
	logger.Debugln("Created project-index simulated response message")

	// Add files requested by LLM
	if len(filesForReview) > 0 {
		// Create request with file-contents
		stage2ProjectSourceCodeMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(prompts.DocProjectCodePromptFile))
		for _, item := range filesForReview {
			contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, item))
			if err != nil {
				logger.Panicln("Failed to add file contents to stage2 prompt", err)
			}
			stage2ProjectSourceCodeMessage = llm.AddFileFragment(stage2ProjectSourceCodeMessage, item, contents, fileNameTags)
		}
		messages = append(messages, stage2ProjectSourceCodeMessage)
		logger.Debugln("Project source code message created")

		// Create simulated response
		stage2ProjectSourceCodeResponseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), loadPrompt(prompts.AIDocProjectCodeResponseFile))
		messages = append(messages, stage2ProjectSourceCodeResponseMessage)
		logger.Debugln("Project source code simulated response added")
	} else {
		logger.Infoln("Not creating extra source-code review")
	}

	// Create document-processing request message
	stage2PromptFile := prompts.DocStage2WritePromptFile
	if action == "REFINE" {
		stage2PromptFile = prompts.DocStage2RefinePromptFile
	}

	stage2DocProcessRequestMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), loadPrompt(stage2PromptFile))
	contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, targetDocument))
	if err != nil {
		logger.Panicln("failed to add document contents to stage1 prompt", err)
	}
	stage2DocProcessRequestMessage = llm.AddPlainTextFragment(stage2DocProcessRequestMessage, contents)
	messages = append(messages, stage2DocProcessRequestMessage)
	logger.Debugln("Created target files analysis request message")

	//TODO: make LLM request with retries and response merging
}
