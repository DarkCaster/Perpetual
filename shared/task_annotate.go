package shared

import (
	"fmt"
	"path/filepath"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

// This func used to extract `IMPLEMENT` comments from source code file,
// and create short task-annotation from it that will be used later to pre-select files using local similarity search.
func TaskAnnotate(targetFiles []string, logger logging.ILogger) []string {
	// Add trace and debug logging
	logger.Traceln("TaskAnnotate: Starting")
	defer logger.Traceln("TaskAnnotate: Finished")

	silentLogger := logger.Clone()
	silentLogger.DisableLevel(logging.ErrorLevel)
	silentLogger.DisableLevel(logging.WarnLevel)
	silentLogger.DisableLevel(logging.InfoLevel)

	projectRootDir, perpetualDir, err := utils.FindProjectRoot(silentLogger)
	if err != nil {
		logger.Panicln("Error finding project root directory:", err)
	}

	projectConfig := config.LoadProjectConfig(perpetualDir, logger)
	annotateConfig := config.LoadOpAnnotateConfig(perpetualDir, logger)

	// Create llm connector for annotate stage1
	connector, err := llm.NewLLMConnector("annotate",
		annotateConfig.String(config.K_SystemPrompt),
		annotateConfig.String(config.K_SystemPromptAck),
		projectConfig.TextMatcherString(config.K_ProjectMdCodeMappings),
		map[string]interface{}{},
		"", "",
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create LLM connector:", err)
	}

	debugString := connector.GetDebugString()
	logger.Notifyln(debugString)
	llm.GetSimpleRawMessageLogger(perpetualDir)(fmt.Sprintf("=== Annotate task: %s\n\n\n", debugString))

	results := []string{}
	for _, filePath := range targetFiles {
		// Read file contents and generate task annotation
		fileContents, wrn, err := utils.LoadTextFile(filepath.Join(projectRootDir, filePath))
		if err != nil {
			logger.Panicf("Failed to read file %s: %s", filePath, err)
		}
		if wrn != "" {
			logger.Warnf("%s: %s", filePath, wrn)
		}

		annotateRequest := llm.AddPlainTextFragment(
			llm.NewMessage(llm.UserRequest),
			annotateConfig.String(config.K_AnnotateTaskPrompt))

		annotateSimulatedResponse := llm.AddPlainTextFragment(
			llm.NewMessage(llm.SimulatedAIResponse),
			annotateConfig.String(config.K_AnnotateTaskResponse))

		fileContentsRequest := llm.AddFileFragment(
			llm.NewMessage(llm.UserRequest),
			filePath,
			fileContents,
			projectConfig.StringArray(config.K_ProjectFilenameTags))

		llm.GetSimpleRawMessageLogger(perpetualDir)(fmt.Sprintf("=== Annotate task: %s\n\n\n", filePath))
		logger.Infoln("Creating task summary for:", filePath)

		onFailRetriesLeft := max(connector.GetOnFailureRetryLimit(), 1)
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			// Perform actual query
			annotationVariants, status, err := connector.Query(1, annotateRequest, annotateSimulatedResponse, fileContentsRequest)
			if perfString := connector.GetPerfString(); perfString != "" {
				logger.Traceln(perfString)
			}
			// Check for general error on query
			if err != nil {
				if onFailRetriesLeft < 1 {
					logger.Panicf("LLM query failed with status %d, error: %s", status, err)
				}
				logger.Errorf("LLM query failed with status %d, error: %s", status, err)
				continue
			}
			// Check for hitting token limit - there are no response variants below token limit, we will try to regenerate from scratch if possible
			if status == llm.QueryMaxTokens {
				if onFailRetriesLeft < 1 {
					logger.Panicln("LLM response(s) reached max tokens, consider increasing the limit")
				}
				logger.Errorln("LLM response(s) reached max tokens, consider increasing the limit")
				continue
			}
			// Some final filtering and preparations of produced annotation variants
			finalVariants := utils.FilterAndTrimResponses(annotationVariants, projectConfig.RegexpArray(config.K_ProjectCodeTagsRx), logger)
			// Stop there if no responses available for further processing
			if len(finalVariants) < 1 {
				if onFailRetriesLeft < 1 {
					logger.Panicln("No LLM responses available")
				}
				logger.Errorln("No LLM responses available")
				continue
			}
			// Exit here if only one variant is available after filtering
			if len(finalVariants) != 1 {
				if onFailRetriesLeft < 1 {
					logger.Panicln("Invalid count of LLM responses detected:", len(finalVariants))
				}
				logger.Errorln("Invalid count of LLM responses detected:", len(finalVariants))
				continue
			}

			results = append(results, finalVariants[0])
			break
		}
	}

	return results
}
