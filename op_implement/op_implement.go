package op_implement

import (
	"flag"
	"path/filepath"
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_annotate"
	"github.com/DarkCaster/Perpetual/op_embed"
	"github.com/DarkCaster/Perpetual/op_stash"
	"github.com/DarkCaster/Perpetual/shared"
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
	var forceUpload, help, noAnnotate, planning, reasonings, taskMode, verbose, trace, includeTests, notEnforceTargetFiles bool
	var taskFile, userFilterFile, contextSaving string
	var searchLimit, selectionPasses int

	// Parse flags for the "implement" operation
	flags := implementFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.StringVar(&contextSaving, "c", "auto", "Context saving mode, reduce LLM context use for large projects (valid values: auto|off|medium|high)")
	flags.StringVar(&taskFile, "i", "", "When using task mode (-t flag), read task instructions from file, plain text or markdown format")
	flags.BoolVar(&noAnnotate, "n", false, "No annotate mode: skip re-annotating of changed files and use current annotations if any")
	flags.BoolVar(&planning, "p", false, "Enable planning, needed for bigger modifications that may create new files, not needed on single file modifications. Disabled by default to save tokens.")
	flags.BoolVar(&reasonings, "pr", false, "Enables extended planning with additional reasoning. May produce improved results for complex or abstractly described tasks, but can also lead to flawed reasoning and worsen the final outcome. This flag includes the -p flag.")
	flags.BoolVar(&forceUpload, "f", false, "Disable 'no-upload' file-filter and upload such files for review and processing if reqested")
	flags.IntVar(&searchLimit, "s", 5, "Limit number of files related to the task returned by local search (0 = disable local search, only use LLM-requested files)")
	flags.IntVar(&selectionPasses, "sp", 1, "Set number of passes for related files selection at stage 1")
	flags.BoolVar(&taskMode, "t", false, "Implement the task directly from instructions read from stdin (or file if -i flag is specified). This flag includes the -p flag.")
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

	contextSaving = shared.ValidateContextSavingValue(contextSaving, logger)

	if searchLimit < 0 {
		logger.Panicln("Search limit (-s flag) value is invalid", searchLimit)
	}
	if selectionPasses < 1 {
		logger.Panicln("Selection passes count (-sp flag) value is invalid", selectionPasses)
	}

	// Set planning mode
	planningMode := 0
	if planning {
		planningMode = 1
	}
	if reasonings {
		planningMode = 2
	}
	if taskMode && planningMode < 1 {
		planningMode = 1
	}

	// task file checks
	if taskFile != "" && !taskMode {
		logger.Panicln("Cannot read task from file without enabling task mode (-t flag)")
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

	utils.LoadEnvFiles(logger, perpetualDir, globalConfigDir)

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
	logger.Infoln("Fetching project files")
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

	// Read input from file or stdin
	var task string
	if taskMode {
		if taskFile != "" {
			data, err := utils.LoadTextFile(taskFile)
			if err != nil {
				logger.Panicln("Error reading task from input file:", err)
			}
			task = data
		} else {
			logger.Infoln("Reading task from stdin")
			data, err := utils.LoadTextStdin()
			if err != nil {
				logger.Panicln("Error reading from stdin:", err)
			}
			task = string(data)
		}
		// Trim excess line breaks at both sides of task, and stop on empty input
		task = strings.Trim(task, "\n")
		if len(task) < 1 {
			logger.Panicln("Task is empty, cannot continue")
		}
	}

	var targetFiles []string
	if task != "" {
		logger.Debugln("Skipping search of source files with implement comment")
	} else {
		// Find files for operation. Select files that contains implement-mark
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
		// Log files to process
		if len(targetFiles) < 1 {
			logger.Panicln("No files found for processing")
		}
		logger.Infoln("Files for processing:")
		for _, targetFile := range targetFiles {
			logger.Infoln(targetFile)
		}
	}

	logger.Debugln("Rotating log file")
	if err := llm.RotateLLMRawLogFile(perpetualDir); err != nil {
		logger.Panicln("Failed to rotate log file:", err)
	}

	// Check if target files includes all project files, and run annotate if needed
	skipStage1 := false
	if len(targetFiles) == len(fileNames) {
		logger.Warnln("All project files selected for processing, no need to run annotate and stage1")
		skipStage1 = true
	} else if !noAnnotate {
		logger.Debugln("Running 'annotate' operation to update file annotations")
		op_annotate_params, op_embed_params := shared.GetAnnotateAndEmbedCmdLineFlags(userFilterFile, contextSaving)
		op_annotate.Run(op_annotate_params, true, logger, logger)
		op_embed.Run(op_embed_params, true, logger, logger)
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
		var promptPlain string
		var promptJson string
		if task != "" {
			promptPlain = implementConfig.String(config.K_ImplementTaskStage1AnalysisPrompt)
			promptJson = implementConfig.String(config.K_ImplementTaskStage1AnalysisJsonModePrompt)
		} else if len(targetFiles) > 0 {
			promptPlain = implementConfig.String(config.K_ImplementStage1AnalysisPrompt)
			promptJson = implementConfig.String(config.K_ImplementStage1AnalysisJsonModePrompt)
		} else {
			logger.Panicln("No task or files with implement-comments provided for processing, cannot continue!")
		}

		if nonTargetFilesAnnotationsCount > 0 {
			// Perform context saving measures - use local search to pre-select only some percentage of the most relevant project files
			filesPercent, randomizePercent := shared.GetLocalSearchLimitsForContextSaving(contextSaving, len(fileNames), projectConfig)
			preselectedFileNames := shared.Stage1Preselect(
				perpetualDir,
				projectRootDir,
				filesPercent,
				randomizePercent,
				fileNames,
				task,
				targetFiles,
				annotations,
				selectionPasses,
				logger)
			// Prepare for multi-pass stage 1
			selectionPasses = len(preselectedFileNames)
			stage1Logger := logger.Clone()
			if selectionPasses > 1 {
				stage1Logger.DisableLevel(logging.InfoLevel)
			}
			fileLists := make([][]string, selectionPasses)
			for pass := range selectionPasses {
				// Run stage 1
				fileLists[pass] = shared.Stage1(
					OpName,
					projectRootDir,
					perpetualDir,
					implementConfig,
					projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
					preselectedFileNames[pass],
					fileNames,
					annotations,
					[]string{}, []string{}, []string{},
					promptPlain,
					promptJson,
					task,
					targetFiles,
					pass+1,
					stage1Logger)
				// Prepare for local similarity search
				searchQueries, searchTags := op_embed.GetQueriesForSimilaritySearch(task, targetFiles, annotations)
				// Compose list of already requested files
				requestedFiles := append(utils.NewSlice(fileLists[pass]...), targetFiles...)
				// Select search mode
				searchMode := shared.GetLocalSearchModeFromContextSavingValue(contextSaving, len(requestedFiles), searchLimit)
				// Local similarity search stage
				similarFiles := op_embed.SimilaritySearchStage(
					searchMode,
					min(searchLimit, len(fileLists[pass])),
					perpetualDir,
					searchQueries,
					searchTags,
					fileNames,
					requestedFiles,
					stage1Logger)
				fileLists[pass] = append(fileLists[pass], similarFiles...)
			}
			// Merge fileLists together
			if selectionPasses > 1 {
				stage1Logger.EnableLevel(logging.InfoLevel)
			} else {
				stage1Logger.DisableLevel(logging.InfoLevel)
			}
			filesToReview = shared.MergeFileLists(fileLists, stage1Logger)
		} else {
			logger.Warnln("All source code files already selected for review, no need to run stage1")
		}
	}

	// Filter filesToReview files for presence of "no-upload" mark
	if !forceUpload {
		filesToReview = utils.FilterNoUploadProjectFiles(
			projectRootDir,
			filesToReview,
			implementConfig.RegexpArray(config.K_NoUploadCommentsRx),
			false,
			logger)
	}

	if planningMode == 0 {
		logger.Infoln("Running stage2: planning disabled, not generating work plan")
		shared.Stage2(OpName,
			projectRootDir,
			perpetualDir,
			implementConfig,
			projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
			[]string{},
			filesToReview,
			map[string]string{},
			[]string{implementConfig.String(config.K_ImplementStage2NoPlanningPrompt)},
			[]interface{}{targetFiles},
			[]string{implementConfig.String(config.K_ImplementStage2NoPlanningResponse)},
			"",
			"",
			false,
			logger,
		)
	}

	// Run stage 2 - create file review, create reasonings
	_, messages, msgIndexToAddExtraFiles := Stage2(projectRootDir,
		perpetualDir,
		implementConfig,
		projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
		planningMode,
		filesToReview,
		targetFiles,
		task,
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
		task,
		logger)

	if !forceUpload {
		otherFilesToModify = utils.FilterNoUploadProjectFiles(
			projectRootDir,
			otherFilesToModify,
			implementConfig.RegexpArray(config.K_NoUploadCommentsRx),
			true,
			logger)
	}

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
