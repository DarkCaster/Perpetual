package op_annotate

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const OpName = "annotate"
const OpDesc = "Generate annotations for project files"

func annotateFlags() *flag.FlagSet {
	return flag.NewFlagSet(OpName, flag.ExitOnError)
}

func Run(args []string, innerCall bool, logger, stdErrLogger logging.ILogger) {

	// Setup
	var help, force, dryRun, verbose, trace bool
	var requestedFile, userFilterFile, contextSaving string

	flags := annotateFlags()
	flags.StringVar(&contextSaving, "c", "auto", "Context saving mode, reduce LLM context use for large projects (valid values: auto|off|medium|high)")
	flags.BoolVar(&force, "f", false, "Force annotation of all files, even for files which annotations are up to date")
	flags.BoolVar(&dryRun, "d", false, "Perform a dry run without actually generating annotations, list of files that will be annotated")
	flags.BoolVar(&help, "h", false, "This help message")
	flags.StringVar(&requestedFile, "r", "", "Only annotate single file provided with this flag, even if its annotation is already up to date (implies -f flag)")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if dryRun {
		logger = stdErrLogger
	}
	if verbose {
		logger.EnableLevel(logging.DebugLevel)
	}
	if trace {
		logger.EnableLevel(logging.DebugLevel)
		logger.EnableLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'annotate' operation")
	logger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", flags)
	}

	contextSaving = strings.ToUpper(contextSaving)
	if contextSaving != "AUTO" && contextSaving != "OFF" && contextSaving != "MEDIUM" && contextSaving != "HIGH" {
		logger.Panicln("Invalid context saving mode value provided")
	}

	if requestedFile != "" {
		force = true
	}

	outerCallLogger := logger.Clone()
	if innerCall {
		outerCallLogger.DisableLevel(logging.ErrorLevel)
		outerCallLogger.DisableLevel(logging.WarnLevel)
		outerCallLogger.DisableLevel(logging.InfoLevel)
	}

	projectRootDir, perpetualDir, err := utils.FindProjectRoot(outerCallLogger)
	if err != nil {
		logger.Panicln("error finding project root directory:", err)
	}

	globalConfigDir, err := utils.FindConfigDir()
	if err != nil {
		logger.Panicln("Error finding perpetual config directory:", err)
	}

	outerCallLogger.Infoln("Project root directory:", projectRootDir)
	outerCallLogger.Debugln("Perpetual directory:", perpetualDir)

	if innerCall {
		logger.Debugln("Not re-loading env files for inner call of annotate operation")
	} else {
		utils.LoadEnvFiles(logger, filepath.Join(perpetualDir, utils.DotEnvFileName), filepath.Join(globalConfigDir, utils.DotEnvFileName))
	}

	projectConfig, err := config.LoadProjectConfig(perpetualDir)
	if err != nil {
		logger.Panicf("Error loading project config: %s", err)
	}

	annotateConfig, err := config.LoadOpAnnotateConfig(perpetualDir)
	if err != nil {
		logger.Panicf("Error loading op_annotate config: %s", err)
	}

	var userBlacklist []*regexp.Regexp
	if userFilterFile != "" {
		userBlacklist, err = utils.AppendUserFilterFromFile(userFilterFile, userBlacklist)
		if err != nil {
			logger.Panicln("Error processing user blacklist-filter:", err)
		}
	}

	// Preparation of project files
	outerCallLogger.Infoln("Fetching project files")
	fileNames, _, err := utils.GetProjectFileList(
		projectRootDir,
		perpetualDir,
		projectConfig.RegexpArray(config.K_ProjectFilesWhitelist),
		projectConfig.RegexpArray(config.K_ProjectFilesBlacklist))

	if err != nil {
		logger.Panicln("Error getting project file-list:", err)
	}

	contextSavingMode := 0
	if (contextSaving == "AUTO" &&
		len(fileNames) >= annotateConfig.Integer(config.K_AnnotateContextSavingFilesCount1) &&
		len(fileNames) < annotateConfig.Integer(config.K_AnnotateContextSavingFilesCount2)) ||
		contextSaving == "MEDIUM" {
		logger.Infoln("Medium context saving selected: generating short annotations")
		contextSavingMode = 1
	}

	if (contextSaving == "AUTO" &&
		len(fileNames) >= annotateConfig.Integer(config.K_AnnotateContextSavingFilesCount2)) ||
		contextSaving == "HIGH" {
		logger.Infoln("Maximum context saving selected: generating tiny annotations")
		contextSavingMode = 2
	}

	// Check fileNames array for case collisions
	if !utils.CheckFilenameCaseCollisions(fileNames) {
		logger.Panicln("Filename case collisions detected in project files")
	}
	// File names and dir-names must not contain path separators characters
	if !utils.CheckForPathSeparatorsInFilenames(fileNames) {
		logger.Panicln("Invalid characters detected in project filenames or directories: / and \\ characters are not allowed!")
	}

	logger.Infoln("Calculating checksums for project files")
	fileChecksums, err := utils.CalculateFilesChecksums(projectRootDir, fileNames)
	if err != nil {
		logger.Panicln("Error getting project-files checksums:", err)
	}

	annotationsFilePath := filepath.Join(perpetualDir, utils.AnnotationsFileName)
	var filesToAnnotate []string
	if !force {
		filesToAnnotate, err = utils.GetChangedAnnotations(annotationsFilePath, fileChecksums)
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
			filesToAnnotate = utils.NewSlice(requestedFile)
		} else {
			filesToAnnotate = utils.NewSlice(fileNames...)
			sort.Strings(filesToAnnotate)
		}
	}

	oldChecksums := utils.GetChecksumsFromAnnotations(annotationsFilePath, fileNames)

	//filter filesToAnnotate with user-blacklist, revert checksum for dropped files, so they can be reevaluated next time
	filesToAnnotate, droppedFiles := utils.FilterFilesWithBlacklist(filesToAnnotate, userBlacklist)
	if len(droppedFiles) > 0 {
		logger.Infoln("Number of files to annotate, filtered by user-provided blacklist:", len(droppedFiles))
	}
	for _, file := range droppedFiles {
		fileChecksums[file] = oldChecksums[file]
		logger.Debugln("Filtered-out:", file)
	}

	if dryRun {
		logger.Infoln("Files to annotate:")
		for _, file := range filesToAnnotate {
			fmt.Println(file)
		}
		os.Exit(0)
	}

	// Create llm connector for annotate stage1
	connector, err := llm.NewLLMConnector(OpName,
		annotateConfig.String(config.K_SystemPrompt),
		annotateConfig.String(config.K_SystemPromptAck),
		projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
		map[string]interface{}{},
		"", "",
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create LLM connector:", err)
	}

	// Create new connector for "annotate_post" operation (stage2)
	connectorPost, err := llm.NewLLMConnector(OpName+"_post",
		annotateConfig.String(config.K_SystemPrompt),
		annotateConfig.String(config.K_SystemPromptAck),
		projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
		map[string]interface{}{},
		"", "",
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create LLM connector:", err)
	}

	// Generate file annotations
	logger.Infoln("Annotating files, count:", len(filesToAnnotate))

	if len(filesToAnnotate) > 0 && connector.GetVariantCount() <= 1 {
		logger.Infoln(connector.GetDebugString())
	}

	if len(filesToAnnotate) > 0 && connector.GetVariantCount() > 1 {
		logger.Infoln("Annotate LLM config for generating summary variants:")
		logger.Infoln(connector.GetDebugString())
		logger.Infoln("Annotate LLM config for post-processing:")
		logger.Infoln(connectorPost.GetDebugString())
	}

	if !innerCall && len(filesToAnnotate) > 0 {
		logger.Debugln("Rotating log file")
		if err := llm.RotateLLMRawLogFile(perpetualDir); err != nil {
			logger.Panicln("Failed to rotate log file:", err)
		}
	}

	logger.Debugln("Sorting files according to sizes")
	fileSizes := utils.GetFileSizes(projectRootDir, filesToAnnotate)

	// Sort files according to sizes
	sort.Slice(filesToAnnotate, func(i, j int) bool {
		return fileSizes[filesToAnnotate[i]] < fileSizes[filesToAnnotate[j]]
	})

	errorFlag := false
	newAnnotations := make(map[string]string)
	for i, filePath := range filesToAnnotate {
		annotatePrompt := ""
		//detect actual prompt for annotating this particular file
		for _, mapping := range annotateConfig.StringArray2D(config.K_AnnotateStage1Prompts) {
			matched, err := regexp.MatchString(mapping[0], filePath)
			if err == nil && matched {
				annotatePrompt = mapping[1+contextSavingMode]
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
			logger.Errorf("Failed to read file %s: %s", filePath, err)
			continue
		}
		fileContents := string(fileBytes)

		annotateRequest := llm.AddPlainTextFragment(
			llm.NewMessage(llm.UserRequest),
			annotatePrompt)

		annotateSimulatedResponse := llm.AddPlainTextFragment(
			llm.NewMessage(llm.SimulatedAIResponse),
			annotateConfig.String(config.K_AnnotateStage1Response))

		fileContentsRequest := llm.AddFileFragment(
			llm.NewMessage(llm.UserRequest),
			filePath,
			fileContents,
			annotateConfig.StringArray(config.K_FilenameTags))

		onFailRetriesLeft := max(connector.GetOnFailureRetryLimit(), 1)
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			logger.Infof("%d: %s", i+1, filePath)
			// Get max number of variants to generate on query
			variantCount := connector.GetVariantCount()
			// Perform actual query
			annotationVariants, status, err := connector.Query(variantCount, annotateRequest, annotateSimulatedResponse, fileContentsRequest)
			// Check for general error on query
			if err != nil {
				logger.Errorf("LLM query failed with status %d, error: %s", status, err)
				if onFailRetriesLeft < 1 {
					fileChecksums[filePath] = oldChecksums[filePath]
					errorFlag = true
				}
				continue
			}
			// Check for hitting token limit - there are no response variants below token limit, we will try to regenerate from scratch if possible
			if status == llm.QueryMaxTokens {
				logger.Errorln("LLM response(s) reached max tokens, consider increasing the limit")
				//TODO: find out do we have seed parameter set, because regenerating with same seed will fail again, so if true -> make onFailRetriesLeft = 0
				if onFailRetriesLeft < 1 {
					fileChecksums[filePath] = oldChecksums[filePath]
					errorFlag = true
				}
				continue
			}
			// Some final filtering and preparations of produced annotation variants
			finalVariants := utils.FilterAndTrimResponses(annotationVariants, annotateConfig.RegexpArray(config.K_CodeTagsRx), logger)
			// Stop there if no responses available for further processing
			if len(finalVariants) < 1 {
				logger.Errorln("No LLM responses available")
				if onFailRetriesLeft < 1 {
					fileChecksums[filePath] = oldChecksums[filePath]
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
			if variantSelectionStrategy == llm.Combine || variantSelectionStrategy == llm.Best {
				// Create message-chain for request
				combinedMessages := utils.NewSlice(annotateRequest, annotateSimulatedResponse, fileContentsRequest)
				for i, variant := range finalVariants {
					combinedMessages = append(combinedMessages, llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), variant))
					if i < len(finalVariants)-1 {
						combinedMessages = append(combinedMessages, llm.AddPlainTextFragment(
							llm.NewMessage(llm.UserRequest),
							annotateConfig.String(config.K_AnnotateStage2PromptVariant)))
					} else if variantSelectionStrategy == llm.Combine {
						combinedMessages = append(combinedMessages, llm.AddPlainTextFragment(
							llm.NewMessage(llm.UserRequest),
							annotateConfig.String(config.K_AnnotateStage2PromptCombine)))
					} else if variantSelectionStrategy == llm.Best {
						combinedMessages = append(combinedMessages, llm.AddPlainTextFragment(
							llm.NewMessage(llm.UserRequest),
							annotateConfig.String(config.K_AnnotateStage2PromptBest)))
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
				} else if blocks, err := utils.ParseMultiTaggedTextRx(
					combinedAnnotation[0],
					utils.GetEvenRegexps(annotateConfig.RegexpArray(config.K_CodeTagsRx)),
					utils.GetOddRegexps(annotateConfig.RegexpArray(config.K_CodeTagsRx)),
					true); err != nil || len(blocks) > 0 {
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
	annotations, err := utils.GetAnnotations(annotationsFilePath, fileNames)
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
