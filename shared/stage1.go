package shared

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Stage1(
	opName string,
	projectRootDir string,
	perpetualDir string,
	cfg config.Config,
	filesToMdLangMappings [][]string,
	preselectedProjectFiles []string,
	allProjectFiles []string,
	annotations map[string]string,
	preQueriesPrompts []string,
	preQueriesBodies []string,
	preQueriesResponses []string,
	mainPromptPlain string,
	mainPromptJson string,
	mainPromptBody string,
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
		opName+"_stage1",
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
		preselectedProjectFiles,
		cfg.StringArray(config.K_FilenameTags),
		annotations,
		logger)
	messages = append(messages, indexRequest)
	logger.Debugln("Created project-index request message")

	// Create project-index simulated response
	indexResponse := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_ProjectIndexResponse))
	messages = append(messages, indexResponse)
	logger.Debugln("Created project-index simulated response message")

	// Create extra history of queries with LLM responses that will be inserted before main query
	for i := range preQueriesPrompts {
		if preQueriesBodies[i] == "" {
			continue
		}
		// Create prompt
		request := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), preQueriesPrompts[i])
		request = llm.AddPlainTextFragment(request, preQueriesBodies[i])
		messages = append(messages, request)
		logger.Debugf("Created pre-request message #%d", i)
		// Create response
		response := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), preQueriesResponses[i])
		messages = append(messages, response)
		logger.Debugf("Created simulated response for pre-request message #%d", i)
	}

	prompt := mainPromptPlain
	if connector.GetOutputFormat() == llm.OutputJson {
		prompt = mainPromptJson
	}
	analysisRequest := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), prompt)
	// Add main body
	if mainPromptBody != "" {
		analysisRequest = llm.AddPlainTextFragment(analysisRequest, mainPromptBody)
	}
	// Add file contents
	for _, item := range targetFiles {
		analysisRequest = llm.AppendFileToMessage(analysisRequest, projectRootDir, item, cfg.StringArray(config.K_FilenameTags), logger)
	}

	messages = append(messages, analysisRequest)
	logger.Debugln("Created main request message")

	logger.Notifyln("Running stage1: find project files for review")
	debugString := connector.GetDebugString()
	logger.Notifyln(debugString)
	llm.GetSimpleRawMessageLogger(perpetualDir)(fmt.Sprintf("=== %s (stage 1): %s\n\n\n", cases.Title(language.English, cases.Compact).String(opName), debugString))

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
	return filterRequestedProjectFiles(projectRootDir, filesForReviewRaw, targetFiles, allProjectFiles, logger)
}

func filterRequestedProjectFiles(projectRootDir string, llmRequestedFiles []string, userRequestedFiles []string, projectFiles []string, logger logging.ILogger) []string {
	var filteredResult []string
	logger.Debugln("Unfiltered file-list requested by LLM:", llmRequestedFiles)
	logger.Infoln("Files requested by LLM:")
	for _, check := range llmRequestedFiles {
		//Remove new line from the end of filename, if present
		if check != "" && check[len(check)-1] == '\n' {
			check = check[:len(check)-1]
		}
		//Remove \r from the end of filename, if present
		if check != "" && check[len(check)-1] == '\r' {
			check = check[:len(check)-1]
		}
		//Replace possibly-invalid path separators
		check = utils.ConvertFilePathToOSFormat(check)
		//Make file path relative to project root
		file, err := utils.MakePathRelative(projectRootDir, check, true)
		if err != nil {
			logger.Errorln("Failed to validate filename requested by LLM:", check)
			continue
		}
		//Filter-out file if it is among files reqested by user, also fix case if so
		file, found := utils.CaseInsensitiveFileSearch(file, userRequestedFiles)
		if found {
			logger.Warnln("Filtering-out file, it is already requested by user:", file)
		} else {
			file, found := utils.CaseInsensitiveFileSearch(file, filteredResult)
			if found {
				logger.Warnln("Filtering-out file, it is already processed or having name-case conflict:", file)
			} else {
				file, found := utils.CaseInsensitiveFileSearch(file, projectFiles)
				if found {
					filteredResult = append(filteredResult, file)
					logger.Infoln(file)
				} else if file, found = tryToSalvageFilename(projectFiles, file, logger); found {
					_, found1 := utils.CaseInsensitiveFileSearch(file, userRequestedFiles)
					_, found2 := utils.CaseInsensitiveFileSearch(file, filteredResult)
					if found1 || found2 {
						logger.Warnln("Filtering-out salvaged file, because it is already in filtered or user files", file)
					} else {
						filteredResult = append(filteredResult, file)
					}
				}
			}
		}
	}

	return filteredResult
}

func tryToSalvageFilename(projectFiles []string, fileToRecover string, logger logging.ILogger) (string, bool) {
	filename := strings.ToLower(filepath.Base(fileToRecover))
	var matches []string

	// Find all files that end with the same filename
	for _, projectFile := range projectFiles {
		if strings.ToLower(filepath.Base(projectFile)) == filename {
			matches = append(matches, projectFile)
		}
	}

	if len(matches) == 1 {
		logger.Infoln("Salvaged filename:", matches[0], "from:", fileToRecover)
		return matches[0], true
	} else if len(matches) > 1 {
		logger.Warnln("Multiple matches found while salvaging filename:", fileToRecover)
	} else {
		logger.Warnln("No matches found while salvaging filename:", fileToRecover)
	}
	return "", false
}
