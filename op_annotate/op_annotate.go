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
	"github.com/DarkCaster/Perpetual/shared"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const OpName = "annotate"
const OpDesc = "Generate annotations for project files"

func annotateFlags() *flag.FlagSet {
	return flag.NewFlagSet(OpName, flag.ExitOnError)
}

func Run(args []string, innerCall bool, logger logging.ILogger) {
	// Setup
	var help, verbose, trace bool
	var descFile, inputFile, userFilterFile, contextSaving, mode string

	flags := annotateFlags()
	flags.StringVar(&contextSaving, "c", "auto", "Context saving mode, reduce LLM context use for large projects (valid values: auto|off|medium|high)")
	flags.StringVar(&descFile, "df", "", "Optional path to project description file for adding into LLM context (valid values: file-path|disabled)")
	flags.BoolVar(&help, "h", false, "This help message")
	flags.StringVar(&inputFile, "i", "", "Forcefully (re)annotate a single file (after whitelist/blacklist and user-filter processing)")
	flags.StringVar(&mode, "m", "", "Select operation mode (valid values: normal|dryrun|full).\n"+
		"normal: reannotate only changed files.\n"+
		"dryrun: do not generate annotations, just list files that would be annotated.\n"+
		"full:   reannotate all files, even those with up-to-date annotations.")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	mode = strings.ToUpper(mode)
	if mode == "" {
		usage.PrintOperationUsage("You must provide a valid operation mode with the '-m' flag (valid values: normal|dryrun|full)", flags)
	}

	if mode != "NORMAL" && mode != "DRYRUN" && mode != "FULL" {
		logger.Errorln("Invalid mode:", mode)
		usage.PrintOperationUsage("You must provide a valid operation mode with the '-m' flag (valid values: normal|dryrun|full)", flags)
	}

	dryRun := mode == "DRYRUN"
	force := mode == "FULL"

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

	contextSaving = shared.ValidateContextSavingValue(contextSaving, logger)

	outerCallLogger := logger.Clone()
	if innerCall {
		outerCallLogger.DisableLevel(logging.ErrorLevel)
		outerCallLogger.DisableLevel(logging.WarnLevel)
		outerCallLogger.DisableLevel(logging.InfoLevel)
	}

	projectRootDir, perpetualDir, err := utils.FindProjectRoot(outerCallLogger, innerCall)
	if err != nil {
		logger.Panicln("Error finding project root directory:", err)
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
		utils.LoadEnvFiles(logger, perpetualDir, globalConfigDir)
	}

	projectConfig := config.LoadProjectConfig(perpetualDir, logger)
	annotateConfig := config.LoadOpAnnotateConfig(perpetualDir, logger)

	// Load project description
	projectDesc := ""
	wrn := ""
	if descFile == "" {
		projectDesc, wrn, err = utils.LoadTextFile(filepath.Join(perpetualDir, config.ProjectDescriptionFile))
		if err != nil {
			if os.IsNotExist(err) {
				logger.Infoln("Not loading missing project description file (description.md)")
			} else {
				logger.Panicln("Failed to load project description file:", err)
			}
		}
		if wrn != "" {
			logger.Warnf("%s: %s", config.ProjectDescriptionFile, wrn)
		}
	} else if strings.ToLower(descFile) != "disabled" {
		projectDesc, wrn, err = utils.LoadTextFile(descFile)
		if err != nil {
			logger.Panicln("Failed to load project description file:", err)
		}
		if wrn != "" {
			logger.Warnf("%s: %s", descFile, wrn)
		}
	} else {
		logger.Infoln("Loading of project description file (description.md) is disabled")
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
	if (contextSaving == "AUTO" && len(fileNames) >= projectConfig.Integer(config.K_ProjectMediumContextSavingFileCount)) ||
		contextSaving == "MEDIUM" ||
		contextSaving == "HIGH" {
		logger.Infoln("Context saving enabled: generating short annotations")
		contextSavingMode = 1
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
	if inputFile != "" {
		// Check if requested file is within fileNames array
		requestedFile, err := utils.MakePathRelative(projectRootDir, inputFile, false)
		if err != nil {
			logger.Panicln("Requested file is not inside project root", requestedFile)
		}
		requestedFile, found := utils.CaseInsensitiveFileSearch(requestedFile, fileNames)
		if !found {
			logger.Panicln("Requested file not found in project")
		}
		filesToAnnotate = utils.NewSlice(requestedFile)
	} else if force {
		filesToAnnotate = utils.NewSlice(fileNames...)
		sort.Strings(filesToAnnotate)
	} else {
		filesToAnnotate, err = utils.GetChangedAnnotations(annotationsFilePath, fileChecksums)
		if err != nil {
			logger.Panicln("error getting changed files:", err)
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
		return
	}

	// Create LLM connector for annotation generation
	connector, err := llm.NewLLMConnector(OpName,
		annotateConfig.String(config.K_SystemPrompt),
		annotateConfig.String(config.K_SystemPromptAck),
		projectConfig.TextMatcherString(config.K_ProjectMdCodeMappings),
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create LLM connector:", err)
	}

	// Generate file annotations
	logger.Infoln("Annotating files, count:", len(filesToAnnotate))

	if !innerCall && len(filesToAnnotate) > 0 {
		logger.Debugln("Rotating log file")
		if err := llm.RotateLLMRawLogFile(perpetualDir); err != nil {
			logger.Panicln("Failed to rotate log file:", err)
		}
	}

	if len(filesToAnnotate) > 0 {
		debugString := connector.GetDebugString()
		logger.Notifyln(debugString)
		llm.GetSimpleRawMessageLogger(perpetualDir)(fmt.Sprintf("=== Annotate: %s\n\n\n", debugString))
	}

	sortBySize := func(fileList []string) {
		logger.Traceln("Sorting files according to sizes:", len(fileList))
		fileSizes := utils.GetFileSizes(projectRootDir, fileList)
		// Sort files according to sizes
		sort.Slice(fileList, func(i, j int) bool {
			return fileSizes[fileList[i]] < fileSizes[fileList[j]]
		})
	}

	fileGroups := [][]string{}
	if connector.GetCachingEnabled() {
		// Compose files for annotation into the groups that use same prompts for better caching
		// Also sort files by size and sort file-groups by count of files inside
		logger.Traceln("Splitting files by prompt-groups")
		fileGroupMap := map[int][]string{}
		//split filesToAnnotate to groups according to its prompts
		for _, filePath := range filesToAnnotate {
			if matched, _, index := annotateConfig.TextMatcherString(config.K_AnnotateFilePrompts).TryMatch(filePath); matched {
				if fileGroup, exist := fileGroupMap[index]; exist {
					fileGroup = append(fileGroup, filePath)
					fileGroupMap[index] = fileGroup
				} else {
					fileGroup = utils.NewSlice(filePath)
					fileGroupMap[index] = fileGroup
				}
			} else {
				logger.Errorln("Failed to detect annotation prompt for file:", filePath)
				continue
			}
		}
		//compose 2d array for groups
		logger.Traceln("Composing file-groups:", len(fileGroupMap))
		for _, fileGroup := range fileGroupMap {
			fileGroups = append(fileGroups, fileGroup)
		}
		//sort file groups according to file count inside it
		sort.Slice(fileGroups, func(i, j int) bool {
			return len(fileGroups[i]) > len(fileGroups[j])
		})
		//sort each group according to size
		for i, fileGroup := range fileGroups {
			logger.Tracef("Sorting file group %d/%d", i+1, len(fileGroups))
			sortBySize(fileGroup)
		}
	} else {
		// When caching is disabled, just use single file group and sort files inside it according to sizes
		sortBySize(filesToAnnotate)
		fileGroups = utils.NewSlice(filesToAnnotate)
	}

	errorFlag := false
	newAnnotations := make(map[string]string)
	for gi, fileGroup := range fileGroups {
		if len(fileGroups) > 1 {
			logger.Infof("Annotating file group %d/%d", gi+1, len(fileGroups))
		}

		for i, filePath := range fileGroup {
			// Detect actual prompt for annotating this particular file
			annotatePrompt := ""
			if matched, values, _ := annotateConfig.TextMatcherString(config.K_AnnotateFilePrompts).TryMatch(filePath); matched {
				annotatePrompt = values[contextSavingMode]
			} else {
				logger.Errorln("Failed to detect annotation prompt for file:", filePath)
				continue
			}

			// Read file contents and generate annotation
			fileBytes, wrn, err := utils.LoadTextFile(filepath.Join(projectRootDir, filePath))
			if err != nil {
				logger.Errorf("Failed to read file %s: %s", filePath, err)
				continue
			}
			if wrn != "" {
				logger.Warnf("%s: %s", filePath, wrn)
			}
			fileContents := string(fileBytes)

			// Build message chain with project description if available
			var messages []llm.Message

			// Add project description if available
			if projectDesc != "" {
				projectDescPrompt := llm.AddPlainTextFragment(
					llm.NewMessage(llm.UserRequest),
					projectConfig.String(config.K_ProjectDescriptionPrompt))
				projectDescPrompt = llm.AddPlainTextFragment(projectDescPrompt, projectDesc)
				projectDescResponse := llm.AddPlainTextFragment(
					llm.NewMessage(llm.SimulatedAIResponse),
					projectConfig.String(config.K_ProjectDescriptionResponse))
				messages = append(messages, projectDescPrompt, projectDescResponse)
			}

			// Add annotation prompt and simulated response
			annotateRequest := llm.AddPlainTextFragment(
				llm.NewMessage(llm.UserRequest),
				annotatePrompt)
			annotateSimulatedResponse := llm.AddPlainTextFragment(
				llm.NewMessage(llm.SimulatedAIResponse),
				annotateConfig.String(config.K_AnnotateFileResponse))
			// we can benefit from caching here, all messages up to this should be the same
			annotateSimulatedResponse.CacheBreakpoint = true

			// Add file contents
			fileContentsRequest := llm.AddFileFragment(
				llm.NewMessage(llm.UserRequest),
				filePath,
				fileContents,
				projectConfig.Tags(config.K_ProjectFilenameTags))

			// Combine all messages
			messages = append(messages, annotateRequest, annotateSimulatedResponse, fileContentsRequest)

			llm.GetSimpleRawMessageLogger(perpetualDir)(fmt.Sprintf("=== Annotate: %s\n\n\n", filePath))

			onFailRetriesLeft := max(connector.GetOnFailureRetryLimit(), 1)
			allowCaching := len(fileGroup) >= connector.GetMinPrefixRepsForCaching()
			for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
				logger.Infof("%d: %s", i+1, filePath)
				// Perform actual query, caching may be beneficial if annotating more than 1 file
				annotationResponse, status, err := connector.Query(allowCaching, messages...)
				if perfString := connector.GetPerfString(); perfString != "" {
					logger.Traceln(perfString)
				}
				// Check for general error on query
				if err != nil {
					logger.Errorf("LLM query failed with status %d, error: %s", status, err)
					if onFailRetriesLeft < 1 {
						fileChecksums[filePath] = oldChecksums[filePath]
						errorFlag = true
					}
					continue
				}
				// Check for hitting token limit - there are no responses below token limit, we will try to regenerate from scratch if possible
				if status == llm.QueryMaxTokens {
					logger.Errorln("LLM response reached max tokens, consider increasing the limit")
					//TODO: find out do we have seed parameter set, because regenerating with same seed will fail again, so if true -> make onFailRetriesLeft = 0
					if onFailRetriesLeft < 1 {
						fileChecksums[filePath] = oldChecksums[filePath]
						errorFlag = true
					}
					continue
				}
				// Some final filtering and preparations of produced annotation response
				finalResponse := utils.FilterAndTrimResponse(annotationResponse, projectConfig.RegexpArray(config.K_ProjectCodeTagsRx), logger)
				// Stop there if no responses available for further processing
				if len(finalResponse) < 1 {
					logger.Errorln("No LLM response available")
					if onFailRetriesLeft < 1 {
						fileChecksums[filePath] = oldChecksums[filePath]
						errorFlag = true
					}
					continue
				}

				newAnnotations[filePath] = finalResponse
				break
			}
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
