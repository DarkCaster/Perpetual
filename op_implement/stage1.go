package op_implement

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
	fileNames []string,
	annotations map[string]string,
	targetFiles []string,
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

	// Create project-index request message
	projectIndexRequest := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), cfg.String(config.K_ImplementStage1IndexPrompt))
	for _, filename := range fileNames {
		projectIndexRequest = llm.AddIndexFragment(projectIndexRequest, filename, cfg.StringArray(config.K_FilenameTags))
		annotation := annotations[filename]
		if annotation == "" {
			annotation = "No annotation available"
		}
		projectIndexRequest = llm.AddPlainTextFragment(projectIndexRequest, annotation)
	}
	logger.Debugln("Created project-index request message")

	// Create project-index simulated response
	projectIndexResponse := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_ImplementStage1IndexResponse))
	logger.Debugln("Created project-index simulated response message")

	// Create target files analisys request message
	analisysPrompt := cfg.String(config.K_ImplementStage1AnalisysPrompt)
	if connector.GetOutputFormat() == llm.OutputJson {
		analisysPrompt = cfg.String(config.K_ImplementStage1AnalisysJsonModePrompt)
	}

	analysisRequest := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), analisysPrompt)
	for _, filename := range targetFiles {
		contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, filename))
		if err != nil {
			logger.Panicln("failed to add file contents to stage1 prompt", err)
		}
		analysisRequest = llm.AddFileFragment(analysisRequest, filename, contents, cfg.StringArray(config.K_FilenameTags))
	}
	logger.Debugln("Created target files analysis request message")

	var filesForReviewRaw []string
	onFailRetriesLeft := connector.GetOnFailureRetryLimit()
	if onFailRetriesLeft < 1 {
		onFailRetriesLeft = 1
	}
	for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
		logger.Infoln("Running stage1: find project files for review")
		var status llm.QueryStatus
		aiResponses, status, err := connector.Query(1, projectIndexRequest, projectIndexResponse, analysisRequest)
		// Handle LLM query errors
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
		// Handle empty response
		if len(aiResponses) < 1 || aiResponses[0] == "" {
			if onFailRetriesLeft < 1 {
				logger.Panicln("Got empty response from AI")
			} else {
				logger.Warnln("Got empty response from AI, retrying")
			}
			continue
		}
		//TODO: add JSON-mode response parsing here
		if connector.GetOutputFormat() == llm.OutputJson {
			logger.Panicln("JSON")
		}
		// Use regular parsing to extract file-list
		filesForReviewRaw, err = utils.ParseTaggedTextRx(aiResponses[0],
			cfg.RegexpArray(config.K_FilenameTagsRx)[0],
			cfg.RegexpArray(config.K_FilenameTagsRx)[1],
			false)
		if err != nil {
			logger.Panicln("Failed to parse list of files for review", err)
		}
		logger.Debugln("Parsed list of files for review from LLM response")
		break
	}
	// Filter all requested files through project file-list, return only files found in project file-list
	return utils.FilterRequestedProjectFiles(projectRootDir, filesForReviewRaw, targetFiles, fileNames, logger)
}
