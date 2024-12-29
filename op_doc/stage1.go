package op_doc

import (
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
	connector, err := llm.NewLLMConnector(OpName+"_stage1", cfg.String(config.K_SystemPrompt), filesToMdLangMappings, cfg.Object(config.K_Stage1OutputScheme), llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage1 LLM connector:", err)
	}
	logger.Debugln(connector.GetDebugString())

	var messages []llm.Message
	// Create project-index request message
	indexRequest := llm.ComposeMessageWithAnnotations(
		cfg.String(config.K_DocProjectIndexPrompt),
		projectFiles,
		cfg.StringArray(config.K_FilenameTags),
		annotations,
		logger)
	messages = append(messages, indexRequest)
	logger.Debugln("Created project-index request message")

	// Create project-index simulated response
	indexResponse := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_DocProjectIndexResponse))
	messages = append(messages, indexResponse)
	logger.Debugln("Created project-index simulated response message")

	if exampleDocuemnt != "" {
		// Create document-example request message
		exampleDocRequest := llm.ComposeMessageFromPromptAndTextFile(projectRootDir, cfg.String(config.K_DocExamplePrompt), exampleDocuemnt, logger)
		messages = append(messages, exampleDocRequest)
		logger.Debugln("Created document-example request message")
		// Create document-example simulated response
		exampleDocResponse := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_DocExampleResponse))
		messages = append(messages, exampleDocResponse)
		logger.Debugln("Created document-example simulated response message")
	}

	// Create project-files analisys request message
	var prompt string
	if action == "WRITE" {
		prompt = cfg.String(config.K_DocStage1WritePrompt)
		if connector.GetOutputFormat() == llm.OutputJson {
			prompt = cfg.String(config.K_DocStage1WriteJsonModePrompt)
		}
	} else if action == "REFINE" {
		prompt = cfg.String(config.K_DocStage1RefinePrompt)
		if connector.GetOutputFormat() == llm.OutputJson {
			prompt = cfg.String(config.K_DocStage1RefineJsonModePrompt)
		}
	} else {
		logger.Panicln("Invalid action:", action)
	}

	analysisRequest := llm.ComposeMessageFromPromptAndTextFile(projectRootDir, prompt, targetDocument, logger)
	messages = append(messages, analysisRequest)
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
