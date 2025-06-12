package op_implement

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
	preQueriesPrompts []string,
	preQueriesBodies []string,
	preQueriesResponses []string,
	query string,
	targetFiles []string,
	logger logging.ILogger) []string {

	// Add trace and debug logging
	logger.Traceln("Stage1: Starting")
	defer logger.Traceln("Stage1: Finished")

	// Some safety checks
	if len(preQueriesPrompts) != len(preQueriesBodies) || len(preQueriesPrompts) != len(preQueriesResponses) {
		logger.Panicf("Pre-queries arrays are different length!")
	}

	// Create stage1 llm connector
	connector, err := llm.NewLLMConnector(
		OpName+"_stage1",
		cfg.String(config.K_SystemPrompt),
		cfg.String(config.K_SystemPromptAck),
		filesToMdLangMappings,
		cfg.Object(config.K_Stage1OutputSchema),
		cfg.String(config.K_Stage1OutputSchemaName),
		cfg.String(config.K_Stage1OutputSchemaDesc),
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage1 LLM connector:", err)
	}

	var messages []llm.Message
	// Create project-index request message
	indexRequest := llm.ComposeMessageWithAnnotations(
		cfg.String(config.K_ProjectIndexPrompt),
		projectFiles,
		cfg.StringArray(config.K_FilenameTags),
		annotations,
		logger)
	messages = append(messages, indexRequest)
	logger.Debugln("Created project-index request message")

	// Create project-index simulated response
	indexResponse := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_ProjectIndexResponse))
	messages = append(messages, indexResponse)
	logger.Debugln("Created project-index simulated response message")

	// Create target files analysis request message
	var analysisRequest llm.Message
	if query == "" {
		analysisPrompt := cfg.String(config.K_ImplementStage1AnalysisPrompt)
		if connector.GetOutputFormat() == llm.OutputJson {
			analysisPrompt = cfg.String(config.K_ImplementStage1AnalysisJsonModePrompt)
		}
		analysisRequest = llm.ComposeMessageWithFiles(projectRootDir, analysisPrompt, targetFiles, cfg.StringArray(config.K_FilenameTags), logger)
	} else {
		analysisPrompt := cfg.String(config.K_ImplementTaskStage1AnalysisPrompt)
		if connector.GetOutputFormat() == llm.OutputJson {
			analysisPrompt = cfg.String(config.K_ImplementTaskStage1AnalysisJsonModePrompt)
		}
		analysisRequest = llm.AddPlainTextFragment(llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), analysisPrompt), query)
	}

	messages = append(messages, analysisRequest)
	logger.Debugln("Created code-analysis request message")

	logger.Infoln("Running stage1: find project files for review")
	logger.Infoln(connector.GetDebugString())

	// Perform LLM query
	var filesForReviewRaw []string
	onFailRetriesLeft := connector.GetOnFailureRetryLimit()
	if onFailRetriesLeft < 1 {
		onFailRetriesLeft = 1
	}
	for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
		aiResponses, status, err := connector.Query(1, messages...)
		if perfString := connector.GetPerfString(); perfString != "" {
			logger.Traceln(perfString)
		}
		// Handle LLM query errors
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
		// Handle empty response
		if len(aiResponses) < 1 || aiResponses[0] == "" {
			if onFailRetriesLeft < 1 {
				logger.Panicln("Got empty response from AI")
			} else {
				logger.Warnln("Got empty response from AI, retrying")
			}
			continue
		}
		if connector.GetOutputFormat() == llm.OutputJson {
			// Use json-mode parsing
			filesForReviewRaw, err = utils.ParseListFromJSON(aiResponses[0], cfg.String(config.K_Stage1OutputKey))
		} else {
			// Use regular parsing to extract file-list
			filesForReviewRaw, err = utils.ParseTaggedTextRx(aiResponses[0],
				cfg.RegexpArray(config.K_FilenameTagsRx)[0],
				cfg.RegexpArray(config.K_FilenameTagsRx)[1],
				false)
		}
		if err != nil {
			if onFailRetriesLeft < 1 {
				logger.Panicln("Failed to parse list of files for review", err)
			} else {
				logger.Warnln("Failed to parse list of files for review", err)
			}
			continue
		}
		logger.Debugln("Parsed list of files for review from LLM response")
		break
	}
	// Filter all requested files through project file-list, return only files found in project file-list
	return utils.FilterRequestedProjectFiles(projectRootDir, filesForReviewRaw, targetFiles, projectFiles, logger)
}
