package op_implement

import (
	"flag"
	"path/filepath"
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_annotate"
	"github.com/DarkCaster/Perpetual/op_stash"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const OpName = "implement"
const OpDesc = "Implement code accodring instructions marked with ###IMPLEMENT### comments"

func implementFlags() *flag.FlagSet {
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)

	return flags
}

func Run(args []string, logger logging.ILogger) {
	var help, trySalvageFiles, noAnnotate, planning, reasonings, verbose, trace, includeTests, notEnforceTargetFiles bool
	var manualFilePath, userFilterFile, contextSaving string

	// Parse flags for the "implement" operation
	flags := implementFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.StringVar(&contextSaving, "c", "auto", "Context saving measures, reduce LLM context use for large projects (valid values: auto|on|off)")
	flags.BoolVar(&noAnnotate, "n", false, "No annotate mode: skip re-annotating of changed files and use current annotations if any")
	flags.BoolVar(&planning, "p", false, "Enable extended planning stage, needed for bigger modifications that may create new files, not needed on single file modifications. Disabled by default to save tokens.")
	flags.BoolVar(&reasonings, "pr", false, "Enables planning with additional reasoning. May produce improved results for complex or abstractly described tasks, but can also lead to flawed reasoning and worsen the final outcome. This flag includes the -p flag.")
	flags.StringVar(&manualFilePath, "r", "", "Manually request a file for the operation, otherwise select files automatically")
	flags.BoolVar(&trySalvageFiles, "s", false, "Try to salvage incorrect filenames on stage 1. Experimental, use in projects with a large number of files where LLM tends to make more mistakes when generating list of files to analyze")
	flags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from processing")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing")
	flags.BoolVar(&notEnforceTargetFiles, "z", false, "When using -p or -pr flags, do not enforce initial sources to file-lists produced by planning")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if verbose {
		logger.EnableLevel(logging.DebugLevel)
	}
	if trace {
		logger.EnableLevel(logging.DebugLevel)
		logger.EnableLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'implement' operation")
	logger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", flags)
	}

	contextSaving = strings.ToUpper(contextSaving)
	if contextSaving != "AUTO" && contextSaving != "ON" && contextSaving != "OFF" {
		logger.Panicln("Invalid context saving measures mode value provided")
	}

	// Set planning mode
	planningMode := 0
	if planning {
		planningMode = 1
	}
	if reasonings {
		planningMode = 2
	}

	// Initialize: detect work directories, load .env file with LLM settings, load file filtering regexps
	projectRootDir, perpetualDir, err := utils.FindProjectRoot(logger)
	if err != nil {
		logger.Panicln("Error finding project root directory:", err)
	}

	globalConfigDir, err := utils.FindConfigDir()
	if err != nil {
		logger.Panicln("Error finding perpetual config directory:", err)
	}

	logger.Infoln("Project root directory:", projectRootDir)
	logger.Debugln("Perpetual directory:", perpetualDir)

	utils.LoadEnvFiles(logger, filepath.Join(perpetualDir, utils.DotEnvFileName), filepath.Join(globalConfigDir, utils.DotEnvFileName))

	implementConfig, err := config.LoadOpImplementConfig(perpetualDir)
	if err != nil {
		logger.Panicf("Error loading op_implement config: %s", err)
	}

	projectConfig, err := config.LoadProjectConfig(perpetualDir)
	if err != nil {
		logger.Panicf("Error loading project config: %s", err)
	}

	projectFilesBlacklist := projectConfig.RegexpArray(config.K_ProjectFilesBlacklist)

	if userFilterFile != "" {
		projectFilesBlacklist, err = utils.AppendUserFilterFromFile(userFilterFile, projectFilesBlacklist)
		if err != nil {
			logger.Panicln("Error appending user blacklist-filter:", err)
		}
	}

	if !includeTests {
		projectFilesBlacklist = append(projectFilesBlacklist, projectConfig.RegexpArray(config.K_ProjectTestFilesBlacklist)...)
	}

	// Get project files, which names selected with whitelist regexps and filtered with blacklist regexps
	fileNames, allFileNames, err := utils.GetProjectFileList(
		projectRootDir,
		perpetualDir,
		projectConfig.RegexpArray(config.K_ProjectFilesWhitelist),
		projectFilesBlacklist)

	if err != nil {
		logger.Panicln("Error getting project file-list:", err)
	}

	// Check fileNames array for case collisions
	if !utils.CheckFilenameCaseCollisions(fileNames) {
		logger.Panicln("Filename case collisions detected in project files")
	}
	// File names and dir-names must not contain path separators characters
	if !utils.CheckForPathSeparatorsInFilenames(fileNames) {
		logger.Panicln("Invalid characters detected in project filenames or directories: / and \\ characters are not allowed!")
	}

	// Find files for operation. Select files that contains implement-mark
	var targetFiles []string
	if manualFilePath != "" {
		//check path relative to project root directory, and make it path relative to it
		targetFile, err := utils.MakePathRelative(projectRootDir, manualFilePath, false)
		if err != nil {
			logger.Panicln("Failed to process file path: ", err)
		}
		targetFile, found := utils.CaseInsensitiveFileSearch(targetFile, fileNames)
		if !found {
			logger.Panicln("Requested file not found in project (make sure it is not excluded from processing by filters):", targetFile)
		}
		found, err = utils.FindInFile(
			filepath.Join(projectRootDir, targetFile),
			implementConfig.RegexpArray(config.K_ImplementCommentsRx))
		if err != nil {
			logger.Panicln("Failed to search 'implement' comment in file: ", err)
		}
		if !found {
			logger.Errorln("Cannot find 'implement' comment in manually provided file, expect LLM to provide wrong results for the file", targetFile)
		}
		targetFiles = append(targetFiles, targetFile)
	} else {
		logger.Debugln("Searching project files for implement comment")
		for _, filePath := range fileNames {
			logger.Traceln(filePath)
			found, err := utils.FindInFile(
				filepath.Join(projectRootDir, filePath),
				implementConfig.RegexpArray(config.K_ImplementCommentsRx))
			if err != nil {
				logger.Panicln("Failed to search 'implement' comment in file: ", err)
			}
			if found {
				targetFiles = append(targetFiles, filePath)
			}
		}
	}

	// Log files to process
	if len(targetFiles) < 1 {
		logger.Panicln("No files found for processing")
	}
	logger.Infoln("Files for processing:")
	for _, targetFile := range targetFiles {
		logger.Infoln(targetFile)
	}

	// Check if target files includes all project files, and run annotate if needed
	skipStage1 := false
	if len(targetFiles) == len(fileNames) {
		logger.Warnln("All project files selected for processing, no need to run annotate and stage1")
		skipStage1 = true
	} else if !noAnnotate {
		logger.Debugln("Running 'annotate' operation to update file annotations")
		op_annotate_params := []string{}
		if userFilterFile != "" {
			op_annotate_params = append(op_annotate_params, "-x", userFilterFile)
		}
		if contextSaving != "AUTO" {
			op_annotate_params = append(op_annotate_params, "-c", contextSaving)
		}
		op_annotate.Run(op_annotate_params, true, logger, logger)
	} else {
		logger.Warnln("File-annotations update disabled, this may worsen the final result")
	}

	var filesToReview []string
	if !skipStage1 {
		// Load annotations needed for stage1
		annotations, err := utils.GetAnnotations(filepath.Join(perpetualDir, utils.AnnotationsFileName), fileNames)
		if err != nil {
			logger.Panicln("Error reading annotations:", err)
		}
		// Find out do we have annotations for files not in targetFiles
		nonTargetFilesAnnotationsCount := 0
		for filename := range annotations {
			found := false
			for _, targetFile := range targetFiles {
				if filename == targetFile {
					found = true
					break
				}
			}
			if !found {
				nonTargetFilesAnnotationsCount++
			}
		}
		if nonTargetFilesAnnotationsCount > 0 {
			// Run stage 1
			filesToReview = Stage1(
				projectRootDir,
				perpetualDir,
				implementConfig,
				projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
				fileNames,
				annotations,
				targetFiles,
				trySalvageFiles,
				logger)
		} else {
			logger.Warnln("No annotaions found for files not in to-implement list, no need to run stage1")
		}
	}

	// Filter filesToReview files for presence of "no-upload" mark
	filesToReview = utils.FilterNoUploadProjectFiles(
		projectRootDir,
		filesToReview,
		implementConfig.RegexpArray(config.K_NoUploadCommentsRx),
		false,
		logger)

	// Run stage 2 - create file review, create reasonings
	messages, msgIndexToAddExtraFiles := Stage2(projectRootDir,
		perpetualDir,
		implementConfig,
		projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
		planningMode,
		filesToReview,
		targetFiles,
		logger)

	// Run stage 3 - get list of files to modify
	messages, otherFilesToModify, targetFilesToModify := Stage3(
		projectRootDir,
		perpetualDir,
		implementConfig,
		projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
		planningMode,
		allFileNames,
		filesToReview,
		targetFiles,
		notEnforceTargetFiles,
		messages,
		msgIndexToAddExtraFiles,
		logger)

	otherFilesToModify = utils.FilterNoUploadProjectFiles(
		projectRootDir,
		otherFilesToModify,
		implementConfig.RegexpArray(config.K_NoUploadCommentsRx),
		true,
		logger)

	otherFilesToModify, droppedFiles := utils.FilterFilesWithWhitelist(otherFilesToModify, projectConfig.RegexpArray(config.K_ProjectFilesWhitelist))
	for _, file := range droppedFiles {
		logger.Warnln("File was filtered-out with project whitelist:", file)
	}
	otherFilesToModify, droppedFiles = utils.FilterFilesWithBlacklist(otherFilesToModify, projectFilesBlacklist)
	for _, file := range droppedFiles {
		logger.Warnln("File was filtered-out with project or user blacklist:", file)
	}

	// Run stage 4 - implement code in selected files
	results := Stage4(
		projectRootDir,
		perpetualDir,
		implementConfig,
		projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
		messages,
		otherFilesToModify,
		targetFilesToModify,
		logger)

	// Extra failsafe: filter-out files from results that not among initial files to modify
	var filteredResults = make(map[string]string)
	finalFilesToModify := append(utils.NewSlice(targetFilesToModify...), otherFilesToModify...)
	for file, content := range results {
		skip := true
		for _, targetFile := range finalFilesToModify {
			if file == targetFile {
				skip = false
				break
			}
		}
		if skip {
			logger.Warnln("Skipping file from results that not among files to modify:", file)
			continue
		}
		filteredResults[file] = content
	}

	// Create and apply stash from generated results
	newStashFileName := op_stash.CreateStash(filteredResults, fileNames, logger)
	op_stash.Run([]string{"-a", "-n", newStashFileName}, true, logger)
}
