package shared

import (
	"fmt"
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Shared stage 2, used by `doc`, `explain` and `implement` operations.
// Main purpose of stage 2 is to process query and generate response based on project-files selected on stage 1.
// Depending on what operation using stage 2 - this query may be a document content generation task, project-related question,
// or query for generating workplan for further code implementation.
func Stage2(
	opName string,
	projectRootDir string,
	perpetualDir string,
	opCfg config.Config,
	filesToMdLangMappings [][]string,
	projectFiles []string,
	filesForReview []string,
	annotations map[string]string,
	addAnnotations bool,
	preQueriesPrompts []string,
	preQueriesBodies []interface{},
	preQueriesResponses []string,
	mainPrompt string,
	mainPromptFinal string,
	mainPromptBody interface{},
	continueOnMaxTokens bool,
	filterResponseWithCodeRx bool,
	logger logging.ILogger) (string, []llm.Message) {

	logger.Traceln("Stage2: Starting")
	defer logger.Traceln("Stage2: Finished")

	// Create stage2 llm connector
	connector, err := llm.NewLLMConnector(
		opName+"_stage2",
		opCfg.String(config.K_SystemPrompt),
		opCfg.String(config.K_SystemPromptAck),
		filesToMdLangMappings,
		map[string]interface{}{},
		"", "",
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create stage2 LLM connector:", err)
	}

	var messages []llm.Message

	if addAnnotations {
		// Create project-index request message
		indexRequest := llm.ComposeMessageWithAnnotations(
			opCfg.String(config.K_ProjectIndexPrompt),
			projectFiles,
			opCfg.StringArray(config.K_FilenameTags),
			annotations,
			logger)
		messages = append(messages, indexRequest)
		logger.Debugln("Created project-index request message")

		// Create project-index simulated response
		indexResponse := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), opCfg.String(config.K_ProjectIndexResponse))
		messages = append(messages, indexResponse)
		logger.Debugln("Created project-index simulated response message")
	} else {
		logger.Infoln("Not adding project-annotations")
	}

	// Add files requested by LLM
	if len(filesForReview) > 0 {
		// Create request with file-contents
		reviewRequest := llm.ComposeMessageWithFiles(
			projectRootDir,
			opCfg.String(config.K_ProjectCodePrompt),
			filesForReview,
			opCfg.StringArray(config.K_FilenameTags),
			logger)
		messages = append(messages, reviewRequest)
		logger.Debugln("Created source code review request message")
		// Create simulated response
		messages = append(messages, llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), opCfg.String(config.K_ProjectCodeResponse)))
		logger.Debugln("Created source code review simulated response message")
	} else {
		logger.Infoln("Not creating extra source-code review")
	}

	// Create extra history of queries with LLM responses that will be inserted before main query
	for i := range preQueriesPrompts {
		if preRequest, ok := llm.ComposeMessageWithFilesOrText(projectRootDir,
			preQueriesPrompts[i],
			preQueriesBodies[i],
			opCfg.StringArray(config.K_FilenameTags),
			logger,
		); ok {
			messages = append(messages, preRequest)
			logger.Debugf("Created pre-request message #%d", i)
			// Create simulated response and add it to history
			messages = append(messages, llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), preQueriesResponses[i]))
			logger.Debugf("Created simulated response for pre-request message #%d", i)
		} else {
			continue
		}
	}

	// Exit here if no main LLM request present
	if mainPrompt == "" {
		// Just return generated message-history from annotations, files for review list and pre-request messages if present for later use
		return "", messages
	}

	// Create query-processing mainRequest message
	mainRequest, ok := llm.ComposeMessageWithFilesOrText(projectRootDir, mainPrompt, mainPromptBody, opCfg.StringArray(config.K_FilenameTags), logger)
	if !ok {
		logger.Panicln("Failed to create main prompt message")
	}

	// realMessages message-history will be used for actual LLM prompt, it will be replaced with simplified prompts at the end
	realMessages := append(utils.NewSlice(messages...), mainRequest)
	response := ""

	logger.Infoln("Running stage2: processing query")
	debugString := connector.GetDebugString()
	logger.Notifyln(debugString)
	llm.GetSimpleRawMessageLogger(perpetualDir)(fmt.Sprintf("=== %s (stage 2): %s\n\n\n", cases.Title(language.English, cases.Compact).String(opName), debugString))

	//Make LLM request, process response
	onFailRetriesLeft := connector.GetOnFailureRetryLimit()
	if onFailRetriesLeft < 1 {
		onFailRetriesLeft = 1
	}
	for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
		// Work with a copy of message-chain, to discard it on full retry, append it with temporary response on partial answer
		messagesTry := utils.NewSlice(realMessages...)
		// Initialize temporary variables for handling partial answers
		var responses []string
		continueGeneration := true
		generateTry := 1
		fileRetry := false
		for continueGeneration && !fileRetry {
			// Run query
			continueGeneration = false
			aiResponses, status, err := connector.Query(1, messagesTry...)
			if perfString := connector.GetPerfString(); perfString != "" {
				logger.Traceln(perfString)
			}
			if err != nil {
				// Retry file on LLM error
				if onFailRetriesLeft < 1 {
					logger.Panicln("LLM query failed:", err)
				} else {
					logger.Warnln("LLM query failed, retrying:", err)
				}
				fileRetry = true
				break
			} else if status == llm.QueryMaxTokens && continueOnMaxTokens {
				// Try to recover other parts of the file if reached max tokens
				if generateTry >= connector.GetMaxTokensSegments() {
					logger.Errorln("LLM query reached token limit, and we are reached segment limit, not attempting to continue")
				} else {
					logger.Warnln("LLM query reached token limit, attempting to continue and file recover")
					continueGeneration = true
					generateTry++
				}
				// Add partial response to stage2 messages, with request to continue
				messagesTry = append(messagesTry, llm.SetRawResponse(llm.NewMessage(llm.SimulatedAIResponse), aiResponses[0]))
				messagesTry = append(messagesTry, llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), opCfg.String(config.K_Stage2ContinuePrompt)))
			} else if status == llm.QueryMaxTokens {
				if onFailRetriesLeft < 1 {
					logger.Panicln("LLM query reached token limit")
				} else {
					logger.Warnln("LLM query reached token limit, retrying")
				}
				fileRetry = true
				break
			}
			// Append response fragment
			responses = append(responses, aiResponses[0])
		}
		if fileRetry {
			continue
		}
		response = strings.Join(responses, "")
		if filterResponseWithCodeRx {
			// Filter-out code blocks from response
			filteredResponses := utils.FilterAndTrimResponses([]string{response}, opCfg.RegexpArray(config.K_CodeTagsRx), logger)
			if len(filteredResponses) < 1 || filteredResponses[0] == "" {
				if onFailRetriesLeft < 1 {
					logger.Panicln("Filtered reasonings response from AI is empty or invalid")
				} else {
					logger.Warnln("Filtered reasonings response from AI is empty or invalid, retrying")
				}
				continue
			}
			response = filteredResponses[0]
		}
		if mainPromptFinal == "" {
			// Add request message to history
			messages = append(messages, mainRequest)
		} else {
			// Add final request message to history - it use simplier instructions that will less likely affect next stages
			finalRequest, ok := llm.ComposeMessageWithFilesOrText(projectRootDir, mainPromptFinal, mainPromptBody, opCfg.StringArray(config.K_FilenameTags), logger)
			if !ok {
				logger.Panicln("Failed to create final main prompt message")
			}
			messages = append(messages, finalRequest)
		}
		logger.Debugln("Created final request message")
		// Add response to message-history
		messages = append(messages, llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), response))
		logger.Debugln("Created final response message")
		break
	}
	return response, messages
}
