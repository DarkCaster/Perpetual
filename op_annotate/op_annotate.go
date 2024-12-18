package op_annotate

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/prompts"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const OpName = "annotate"
const OpDesc = "Generate annotations for project files"

func annotateFlags() *flag.FlagSet {
	return flag.NewFlagSet(OpName, flag.ExitOnError)
}

func Run(args []string, logger logging.ILogger) {
	flags := annotateFlags()

	var help, force, dryRun, verbose, trace bool
	var requestedFile string
	flags.BoolVar(&force, "f", false, "Force annotation of all files, even for files which annotations are up to date")
	flags.BoolVar(&dryRun, "d", false, "Perform a dry run without actually generating annotations, list of files that will be annotated")
	flags.BoolVar(&help, "h", false, "This help message")
	flags.StringVar(&requestedFile, "r", "", "Only annotate single file provided with this flag, even if its annotation is already up to date (implies -f flag)")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if verbose {
		logger.SetLevel(logging.DebugLevel)
	}
	if trace {
		logger.SetLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'annotate' operation")
	logger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", flags)
	}

	if requestedFile != "" {
		force = true
	}

	projectRootDir, perpetualDir, err := utils.FindProjectRoot(logger)
	if err != nil {
		logger.Panicln("error finding project root directory:", err)
	}

	globalConfigDir, err := utils.FindConfigDir()
	if err != nil {
		logger.Panicln("Error finding perpetual config directory:", err)
	}

	logger.Infoln("Project root directory:", projectRootDir)
	logger.Debugln("Perpetual directory:", perpetualDir)

	utils.LoadEnvFiles(logger, filepath.Join(perpetualDir, utils.DotEnvFileName), filepath.Join(globalConfigDir, utils.DotEnvFileName))

	var projectFilesWhitelist []string
	if err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesWhitelistFileName), &projectFilesWhitelist); err != nil {
		logger.Panicln("error reading project-files whitelist regexps:", err)
	}

	var filesToMdLangMappings [][2]string
	if err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesToMarkdownLangMappingFileName), &filesToMdLangMappings); err != nil {
		logger.Warnln("Error reading optional filename to markdown-lang mappings:", err)
	}

	var projectFilesBlacklist []string
	if err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesBlacklistFileName), &projectFilesBlacklist); err != nil {
		logger.Panicln("error reading project-files blacklist regexps:", err)
	}

	fileChecksums, fileNames, _, err := utils.GetProjectFileList(projectRootDir, perpetualDir, projectFilesWhitelist, projectFilesBlacklist)
	if err != nil {
		logger.Panicln("error getting project file-list:", err)
	}

	// Check fileNames array for case collisions
	if !utils.CheckFilenameCaseCollisions(fileNames) {
		logger.Panicln("Filename case collisions detected in project files")
	}
	// File names and dir-names must not contain path separators characters
	if !utils.CheckForPathSeparatorsInFilenames(fileNames) {
		logger.Panicln("Invalid characters detected in project filenames or directories: / and \\ characters are not allowed!")
	}

	annotationsFilePath := filepath.Join(perpetualDir, utils.AnnotationsFileName)
	var filesToAnnotate []string
	if !force {
		filesToAnnotate, err = utils.GetChangedFiles(annotationsFilePath, fileChecksums)
		if err != nil {
			logger.Panicln("error getting changed files:", err)
		}
	} else {
		if requestedFile != "" {
			// Check if requested file is within fileNames array
			requestedFile, err := utils.MakePathRelative(projectRootDir, requestedFile, false)
			if err != nil {
				logger.Panicln("Requested file is not inside project root", requestedFile)
			}
			requestedFile, found := utils.CaseInsensitiveFileSearch(requestedFile, fileNames)
			if !found {
				logger.Panicln("Requested file not found in project")
			}
			filesToAnnotate = []string{requestedFile}
		} else {
			filesToAnnotate = make([]string, 0, len(fileChecksums))
			for file := range fileChecksums {
				filesToAnnotate = append(filesToAnnotate, file)
			}
			sort.Strings(filesToAnnotate)
		}
	}

	if dryRun {
		logger.Infoln("Files to annotate:")
		for _, file := range filesToAnnotate {
			logger.Infoln(file)
		}
		os.Exit(0)
	}

	systemPrompts := map[string]string{}
	if err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.SystemPromptsConfigFile), &systemPrompts); err != nil {
		logger.Panicf("Error loading %s config :%s", prompts.SystemPromptsConfigFile, err)
	}

	annotateConfig := map[string]interface{}{}
	if err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.OpAnnotateConfigFile), &annotateConfig); err != nil {
		logger.Panicf("Error loading %s config :%s", prompts.OpAnnotateConfigFile, err)
	}

	// Create llm connector for annotate stage1
	connector, err := llm.NewLLMConnector(OpName,
		systemPrompts[prompts.DefaultSystemPromptName],
		filesToMdLangMappings,
		map[string]interface{}{},
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create LLM connector:", err)
	}
	logger.Debugln(connector.GetDebugString())

	// Create new connector for "annotate_post" operation (stage2)
	connectorPost, err := llm.NewLLMConnector(OpName+"_post",
		systemPrompts[prompts.DefaultSystemPromptName],
		filesToMdLangMappings,
		map[string]interface{}{},
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create LLM connector:", err)
	}
	logger.Debugln(connectorPost.GetDebugString())

	// Load output tags regexps
	outputTagsRxStrings := utils.InterfaceToStringArray(annotateConfig[prompts.CodeTagsRxName])

	// Generate file annotations
	logger.Infoln("Annotating files, count:", len(filesToAnnotate))
	errorFlag := false
	newAnnotations := make(map[string]string)
	for _, filePath := range filesToAnnotate {
		annotatePrompt := ""
		//detect actual prompt for annotating this particular file
		for _, mapping := range utils.InterfaceTo2DStringArray(annotateConfig[prompts.AnnotateStage1PromptNames]) {
			matched, err := regexp.MatchString(mapping[0], filePath)
			if err == nil && matched {
				annotatePrompt = mapping[1]
				break
			}
		}
		if annotatePrompt == "" {
			logger.Errorln("Failed to detect annotation prompt for file:", filePath)
			continue
		}

		// Read file contents and generate annotation
		fileBytes, err := utils.LoadTextFile(filepath.Join(projectRootDir, filePath))
		if err != nil {
			logger.Panicln("failed to read file:", err)
		}
		fileContents := string(fileBytes)

		annotateRequest := llm.AddPlainTextFragment(
			llm.NewMessage(llm.UserRequest),
			annotatePrompt)

		annotateSimulatedResponse := llm.AddPlainTextFragment(
			llm.NewMessage(llm.SimulatedAIResponse),
			annotateConfig[prompts.AnnotateStage1ResponseName].(string))

		fileContentsRequest := llm.AddFileFragment(
			llm.NewMessage(llm.UserRequest),
			filePath,
			fileContents,
			utils.InterfaceToStringArray(annotateConfig[prompts.FilenameTagsName]))

		onFailRetriesLeft := connector.GetOnFailureRetryLimit()
		if onFailRetriesLeft < 1 {
			onFailRetriesLeft = 1
		}
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			logger.Infoln(filePath)
			// Get max number of variants to generate on query
			variantCount := connector.GetVariantCount()
			// Perform actual query
			annotationVariants, status, err := connector.Query(variantCount, annotateRequest, annotateSimulatedResponse, fileContentsRequest)
			// Check for general error on query
			if err != nil {
				logger.Errorf("LLM query failed with status %d, error: %s", status, err)
				if onFailRetriesLeft < 1 {
					fileChecksums[filePath] = "error"
					errorFlag = true
				}
				continue
			}
			// Check for hitting token limit - there are no response variants below token limit, we will try to regenerate from scratch if possible
			if status == llm.QueryMaxTokens {
				logger.Errorln("LLM response(s) reached max tokens, consider increasing the limit")
				//TODO: find out do we have seed parameter set, because regenerating with same seed will fail again, so if true -> make onFailRetriesLeft = 0
				if onFailRetriesLeft < 1 {
					fileChecksums[filePath] = "error"
					errorFlag = true
				}
				continue
			}
			// Some final filtering and preparations of produced annotation variants
			finalVariants := []string{}
			for i, variant := range annotationVariants {
				// Filter-out variants that contain code-blocks - this is not allowed
				if blocks, err := utils.ParseMultiTaggedText(variant, getEvenIndexElements(outputTagsRxStrings), getOddIndexElements(outputTagsRxStrings), true); err != nil || len(blocks) > 0 {
					logger.Warnf("LLM response #%d contains code blocks, which is not allowed", i)
					continue
				}
				// Trim unneded symbols from both ends of annotation
				variant = strings.Trim(variant, " \t\n") //note: there is a space character first, do not remove it
				if len(variant) < 1 {
					logger.Warnf("LLM response #%d is empty", i)
					continue
				}
				finalVariants = append(finalVariants, variant)
			}
			// Stop there if no responses available for further processing
			if len(finalVariants) < 1 {
				logger.Errorln("No LLM responses available")
				if onFailRetriesLeft < 1 {
					fileChecksums[filePath] = "error"
					errorFlag = true
				}
				continue
			}
			// Exit here if only one variant is available after filtering
			if len(finalVariants) == 1 {
				newAnnotations[filePath] = finalVariants[0]
				break
			}

			variantSelectionStrategy := connector.GetVariantSelectionStrategy()

			// Combine the annotation using LLM
			if variantSelectionStrategy == llm.Combine {
				// Create message-chain for request
				combinedMessages := []llm.Message{annotateRequest, annotateSimulatedResponse, fileContentsRequest}
				for i, variant := range finalVariants {
					combinedMessages = append(combinedMessages, llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), variant))
					if i < len(finalVariants)-1 {
						combinedMessages = append(combinedMessages, llm.AddPlainTextFragment(
							llm.NewMessage(llm.UserRequest),
							annotateConfig[prompts.AnnotateStage2PromptVariantName].(string)))
					} else {
						combinedMessages = append(combinedMessages, llm.AddPlainTextFragment(
							llm.NewMessage(llm.UserRequest),
							annotateConfig[prompts.AnnotateStage2PromptCombineName].(string)))
					}
				}
				// Perform the query
				combinedAnnotation, status, err := connectorPost.Query(1, combinedMessages...)
				// Check for general error on query, switch for using "short" variant selection strategy on error
				if err != nil {
					logger.Warnf("LLM query failed with status %d, error: %s", status, err)
					variantSelectionStrategy = llm.Short
				} else if status == llm.QueryMaxTokens {
					logger.Warnf("LLM combined annotation reached max tokens")
					variantSelectionStrategy = llm.Short
				} else if blocks, err := utils.ParseMultiTaggedText(combinedAnnotation[0], getEvenIndexElements(outputTagsRxStrings), getOddIndexElements(outputTagsRxStrings), true); err != nil || len(blocks) > 0 {
					logger.Warnln("LLM combined annotation contains code blocks, which is not allowed")
					variantSelectionStrategy = llm.Short
				} else {
					// Trim unneded symbols from both ends of annotation
					trimmedAnnotation := strings.Trim(combinedAnnotation[0], " \t\n") //note: there is a space character first, do not remove it
					if len(trimmedAnnotation) < 1 {
						logger.Warnln("LLM combined annotation is empty")
						variantSelectionStrategy = llm.Short
					} else {
						// Finally save our post-processed annotation
						newAnnotations[filePath] = combinedAnnotation[0]
						break
					}
				}
			}

			// Longest variant
			if variantSelectionStrategy == llm.Long {
				longestVariant := finalVariants[0]
				for _, variant := range finalVariants[1:] {
					if len(variant) > len(longestVariant) {
						longestVariant = variant
					}
				}
				newAnnotations[filePath] = longestVariant
				break
			}

			// Select shortest variant
			shortestVariant := finalVariants[0]
			for _, variant := range finalVariants[1:] {
				if len(variant) < len(shortestVariant) {
					shortestVariant = variant
				}
			}
			newAnnotations[filePath] = shortestVariant
			break
		}
	}

	// Get annotations for files listed in fileChecksums
	annotations, err := utils.GetAnnotations(annotationsFilePath, fileChecksums)
	if err != nil {
		logger.Panicln("Failed to read old annotations:", err)
	}

	// Copy new annotations back to old annotations
	for element := range newAnnotations {
		annotations[element] = newAnnotations[element]
	}

	// Save updated annotations
	logger.Infoln("Saving annotations")
	if err := utils.SaveAnnotations(annotationsFilePath, fileChecksums, annotations); err != nil {
		logger.Panicln("Failed to save annotations:", err)
	}

	if errorFlag {
		logger.Panicln("Not all files were successfully annotated. Run annotate again to try to index the failed files.")
	}
}

func getEvenIndexElements(arr []string) []string {
	var evenIndexElements []string
	for i := 0; i < len(arr); i += 2 {
		evenIndexElements = append(evenIndexElements, arr[i])
	}
	return evenIndexElements
}

func getOddIndexElements(arr []string) []string {
	var evenIndexElements []string
	for i := 1; i < len(arr); i += 2 {
		evenIndexElements = append(evenIndexElements, arr[i])
	}
	return evenIndexElements
}
