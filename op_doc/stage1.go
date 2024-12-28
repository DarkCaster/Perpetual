package op_doc

import (
	"path/filepath"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage1(projectRootDir string,
	perpetualDir string,
	cfg config.Config,
	filesToMdLangMappings [][]string,
	projectFiles []string,
	annotations map[string]string,
	targetDocument string,
	exampleDocuemnt string,
	action string,
	logger logging.ILogger) []string {

	// Add trace and debug logging
	logger.Traceln("Stage1: Starting")
	defer logger.Traceln("Stage1: Finished")

	// Create stage1 llm connector
	connector, err := llm.NewLLMConnector(OpName+"_stage1", cfg.String(config.K_SystemPrompt), filesToMdLangMappings, map[string]interface{}{}, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage1 LLM connector:", err)
	}
	logger.Debugln(connector.GetDebugString())

	var messages []llm.Message
	// Create project-index request message
	projectIndexRequestMessage := llm.AddPlainTextFragment(
		llm.NewMessage(llm.UserRequest),
		cfg.String(config.K_DocProjectIndexPrompt))

	for _, item := range projectFiles {
		projectIndexRequestMessage = llm.AddIndexFragment(
			projectIndexRequestMessage,
			item,
			cfg.StringArray(config.K_FilenameTags))

		annotation := annotations[item]
		if annotation == "" {
			annotation = "No annotation available"
		}
		projectIndexRequestMessage = llm.AddPlainTextFragment(projectIndexRequestMessage, annotation)
	}
	messages = append(messages, projectIndexRequestMessage)
	logger.Debugln("Created project-index request message")

	// Create project-index simulated response
	projectIndexResponseMessage := llm.AddPlainTextFragment(
		llm.NewMessage(llm.SimulatedAIResponse),
		cfg.String(config.K_DocProjectIndexResponse))

	messages = append(messages, projectIndexResponseMessage)
	logger.Debugln("Created project-index simulated response message")

	if exampleDocuemnt != "" {
		// Create document-example request message
		requestMessage := llm.ComposeMessageFromPromptAndTextFile(projectRootDir, cfg.String(config.K_DocExamplePrompt), exampleDocuemnt, logger)
		messages = append(messages, requestMessage)
		logger.Debugln("Created document-example request message")
		// Create document-example simulated response
		responseMessage := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_DocExampleResponse))
		messages = append(messages, responseMessage)
		logger.Debugln("Created document-example simulated response message")
	}

	// Create project-files analisys request message
	prompt := cfg.String(config.K_DocStage1WritePrompt)
	if action == "REFINE" {
		prompt = cfg.String(config.K_DocStage1RefinePrompt)
	}

	codeAnalysisRequestMessage := llm.AddPlainTextFragment(
		llm.NewMessage(llm.UserRequest),
		prompt)

	contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, targetDocument))
	if err != nil {
		logger.Panicln("failed to add file contents to stage1 prompt", err)
	}
	codeAnalysisRequestMessage = llm.AddPlainTextFragment(codeAnalysisRequestMessage, contents)
	messages = append(messages, codeAnalysisRequestMessage)
	logger.Debugln("Created code-analysis request message")

	// Perform LLM query
	var aiResponses []string
	onFailRetriesLeft := connector.GetOnFailureRetryLimit()
	if onFailRetriesLeft < 1 {
		onFailRetriesLeft = 1
	}
	for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
		logger.Infoln("Running stage1: find project files for review")
		var status llm.QueryStatus
		//NOTE: do not use := here, looks like it will make copy of aiResponse, and effectively result in empty file-list (tested on golang 1.22.3)
		aiResponses, status, err = connector.Query(1, messages...)
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
				logger.Warnln("LLM query reached token limit, retrying")
			}
			continue
		}
		if len(aiResponses) < 1 || aiResponses[0] == "" {
			logger.Warnln("Got empty response from AI, retrying")
			continue
		}
		logger.Debugln("LLM query completed")
		break
	}

	// Process response, parse files of interest from ai response
	if len(aiResponses) < 1 || aiResponses[0] == "" {
		logger.Errorln("Got empty response from AI")
	}

	llmRequestedFiles, err := utils.ParseTaggedTextRx(aiResponses[0],
		cfg.RegexpArray(config.K_FilenameTagsRx)[0],
		cfg.RegexpArray(config.K_FilenameTagsRx)[1],
		false)

	if err != nil {
		logger.Panicln("Failed to parse list of files for review", err)
	}
	logger.Debugln("Parsed list of files for review from LLM response")

	// Filter all requested files through project file-list, return only files found in project file-list
	return utils.FilterRequestedProjectFiles(projectRootDir, llmRequestedFiles, []string{targetDocument}, projectFiles, logger)
}
