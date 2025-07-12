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

// Shared stage 2, used by `doc` and `explain` operations, `implement` operation using it's own stage 2 implementation
func Stage2(
	opName string,
	projectRootDir string,
	perpetualDir string,
	cfg config.Config,
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
	logger logging.ILogger) (string, []llm.Message, int) {

	logger.Traceln("Stage2: Starting")
	defer logger.Traceln("Stage2: Finished")

	// Create stage2 llm connector
	connector, err := llm.NewLLMConnector(
		opName+"_stage2",
		cfg.String(config.K_SystemPrompt),
		cfg.String(config.K_SystemPromptAck),
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
	} else {
		logger.Infoln("Not adding project-annotations")
	}

	// Add files requested by LLM
	if len(filesForReview) > 0 {
		// Create request with file-contents
		requestMessage := llm.ComposeMessageWithFiles(
			projectRootDir,
			cfg.String(config.K_ProjectCodePrompt),
			filesForReview,
			cfg.StringArray(config.K_FilenameTags),
			logger)
		messages = append(messages, requestMessage)
		logger.Debugln("Project source code message created")
		// Create simulated response
		messages = append(messages, llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), cfg.String(config.K_ProjectCodeResponse)))
		logger.Debugln("Project source code simulated response added")
	} else {
		logger.Infoln("Not creating extra source-code review")
	}

	// Contain index of the last pre-query message, can be used as quick way to insert or modify stage 2 message-history before LLM request on next stages:
	// lastPreQueryMessageIndex is the last stage 2 pre-request message index,
	// lastPreQueryMessageIndex + 1 is response message index for the last stage 2 pre-request message
	// lastPreQueryMessageIndex + 2 is main stage 2 request message index
	// lastPreQueryMessageIndex + 3 is main stage 2 response message index
	lastPreQueryMessageIndex := 0

	// Create extra history of queries with LLM responses that will be inserted before main query
	for i := range preQueriesPrompts {
		var request llm.Message
		//check body type, it can be either text content (string) or list of filenames ([]string)
		if text, isText := preQueriesBodies[i].(string); isText {
			if text == "" {
				continue
			}
			// Create prompt with query + text content
			request = llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), preQueriesPrompts[i])
			request = llm.AddPlainTextFragment(request, text)
			logger.Debugf("Created pre-request message #%d (with text)", i)
		} else if fileNames, isFnList := preQueriesBodies[i].([]string); isFnList {
			if len(fileNames) < 1 {
				continue
			}
			// Create prompt with query + files' contents
			request = llm.ComposeMessageWithFiles(projectRootDir, preQueriesPrompts[i], fileNames, cfg.StringArray(config.K_FilenameTags), logger)
			logger.Debugf("Created pre-request message #%d (with files)", i)
		} else {
			logger.Panicln("Unsupported pre-query body type, index:", i)
		}
		messages = append(messages, request)
		lastPreQueryMessageIndex = len(messages) - 1
		// Create simulated response and add it to history
		messages = append(messages, llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), preQueriesResponses[i]))
		logger.Debugf("Created simulated response for pre-request message #%d", i)
	}

	// Exit here if no main LLM request present
	if mainPrompt == "" {
		// Just return generated message-history from annotations, files for review list and pre-request messages if present for later use
		return "", messages, lastPreQueryMessageIndex
	}

	// Create query-processing request message
	var requestMessage llm.Message
	if text, isText := mainPromptBody.(string); isText {
		requestMessage = llm.AddPlainTextFragment(llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), mainPrompt), text)
	} else if fileNames, isFnList := mainPromptBody.([]string); isFnList {
		requestMessage = llm.ComposeMessageWithFiles(projectRootDir, mainPrompt, fileNames, cfg.StringArray(config.K_FilenameTags), logger)
	} else {
		logger.Panicln("Unsupported main query body type")
	}

	// realMessages message-history will be used for actual LLM prompt, it will be replaced with simplified prompts at the end
	realMessages := append(utils.NewSlice(messages...), requestMessage)
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
				messagesTry = append(messagesTry, llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), cfg.String(config.K_Stage2ContinuePrompt)))
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
			filteredResponses := utils.FilterAndTrimResponses([]string{response}, cfg.RegexpArray(config.K_CodeTagsRx), logger)
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
			messages = append(messages, requestMessage)
		} else {
			var finalRequestMessage llm.Message
			// Add final request message to history - it use simplier instructions that will less likely affect next stages
			if text, isText := mainPromptBody.(string); isText {
				finalRequestMessage = llm.AddPlainTextFragment(llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), mainPromptFinal), text)
			} else if fileNames, isFnList := mainPromptBody.([]string); isFnList {
				finalRequestMessage = llm.ComposeMessageWithFiles(projectRootDir, mainPromptFinal, fileNames, cfg.StringArray(config.K_FilenameTags), logger)
			} else {
				logger.Panicln("Unsupported main query body type")
			}
			messages = append(messages, finalRequestMessage)
		}
		lastPreQueryMessageIndex = len(messages) - 1
		logger.Debugln("Created final request message")
		// Add response to message-history
		messages = append(messages, llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), response))
		logger.Debugln("Created final response message")
	}
	return response, messages, lastPreQueryMessageIndex
}
