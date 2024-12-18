package op_implement

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_annotate"
	"github.com/DarkCaster/Perpetual/op_stash"
	"github.com/DarkCaster/Perpetual/prompts"
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
	var help, noAnnotate, planning, reasonings, verbose, trace, includeTests bool
	var manualFilePath, userFilterFile string

	// Parse flags for the "implement" operation
	flags := implementFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.BoolVar(&noAnnotate, "n", false, "No annotate mode: skip re-annotating of changed files and use current annotations if any")
	flags.BoolVar(&planning, "p", false, "Enable extended planning stage, needed for bigger modifications that may create new files, not needed on single file modifications. Disabled by default to save tokens.")
	flags.BoolVar(&reasonings, "pr", false, "Enables planning with additional reasoning. May produce improved results for complex or abstractly described tasks, but can also lead to flawed reasoning and worsen the final outcome. This flag includes the -p flag.")
	flags.StringVar(&manualFilePath, "r", "", "Manually request a file for the operation, otherwise select files automatically")
	flags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from processing")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if verbose {
		logger.SetLevel(logging.DebugLevel)
	}
	if trace {
		logger.SetLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'implement' operation")
	logger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", flags)
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

	systemPrompts := map[string]string{}
	if err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.SystemPromptsConfigFile), &systemPrompts); err != nil {
		logger.Panicf("Error loading %s config :%s", prompts.SystemPromptsConfigFile, err)
	}

	implementConfig := map[string]interface{}{}
	if err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.OpImplementConfigFile), &implementConfig); err != nil {
		logger.Panicf("Error loading %s config :%s", prompts.OpImplementConfigFile, err)
	}

	var projectFilesWhitelist []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesWhitelistFileName), &projectFilesWhitelist)
	if err != nil {
		logger.Panicln("Error reading project-files whitelist regexps:", err)
	}

	var filesToMdLangMappings [][2]string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesToMarkdownLangMappingFileName), &filesToMdLangMappings)
	if err != nil {
		logger.Warnln("Error reading optional filename to markdown-lang mappings:", err)
	}

	var projectFilesBlacklist []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesBlacklistFileName), &projectFilesBlacklist)
	if err != nil {
		logger.Panicln("Error reading project-files blacklist regexps:", err)
	}

	if userFilterFile != "" {
		var userFilesBlacklist []string
		err := utils.LoadJsonFile(userFilterFile, &userFilesBlacklist)
		if err != nil {
			logger.Panicln("Error reading user-supplied blacklist regexps:", err)
		}
		projectFilesBlacklist = append(projectFilesBlacklist, userFilesBlacklist...)
	}

	if !includeTests {
		var testFilesBlacklist []string
		err := utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectTestFilesBlacklistFileName), &testFilesBlacklist)
		if err != nil {
			logger.Panicln("Error reading project-files blacklist regexps for unit-tests, you may have to rerun init:", err)
		}
		projectFilesBlacklist = append(projectFilesBlacklist, testFilesBlacklist...)
	}

	// Get project files, which names selected with whitelist regexps and filtered with blacklist regexps
	fileChecksums, fileNames, allFileNames, err := utils.GetProjectFileList(projectRootDir, perpetualDir, projectFilesWhitelist, projectFilesBlacklist)
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

	var implementRegexps []*regexp.Regexp
	for _, rx := range utils.InterfaceToStringArray(implementConfig[prompts.ImplementCommentsRxName]) {
		crx, err := regexp.Compile(rx)
		if err != nil {
			logger.Panicln("Failed to compile 'implement' comment search regexp: ", err)
		}
		implementRegexps = append(implementRegexps, crx)
	}

	var noUploadRegexps []*regexp.Regexp
	for _, rx := range utils.InterfaceToStringArray(implementConfig[prompts.NoUploadCommentsRxName]) {
		crx, err := regexp.Compile(rx)
		if err != nil {
			logger.Panicln("Failed to compile 'no-upload' comment search regexp: ", err)
		}
		noUploadRegexps = append(noUploadRegexps, crx)
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
		found, err = utils.FindInFile(filepath.Join(projectRootDir, targetFile), implementRegexps)
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
			found, err := utils.FindInFile(filepath.Join(projectRootDir, filePath), implementRegexps)
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
		op_annotate.Run(nil, logger)
	} else {
		logger.Warnln("File-annotations update disabled, this may worsen the final result")
	}

	var filesToReview []string
	if !skipStage1 {
		// Load annotations needed for stage1
		annotations, err := utils.GetAnnotations(filepath.Join(perpetualDir, utils.AnnotationsFileName), fileChecksums)
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
				systemPrompts[prompts.DefaultSystemPromptName],
				implementConfig,
				filesToMdLangMappings,
				fileNames,
				annotations,
				targetFiles,
				logger)
		} else {
			logger.Warnln("No annotaions found for files not in to-implement list, no need to run stage1")
		}
	}

	// Check filesToReview files for presence of "no-upload" mark
	var filteredFilesToReview []string
	for _, file := range filesToReview {
		if found, err := utils.FindInRelativeFile(projectRootDir, file, noUploadRegexps); err == nil && !found {
			filteredFilesToReview = append(filteredFilesToReview, file)
		} else if found {
			logger.Warnln("Skipping file marked with 'no-upload' comment:", file)
		} else {
			logger.Errorln("Error searching for 'no-upload' comment in file:", file, err)
		}
	}
	filesToReview = filteredFilesToReview

	// Run stage 2
	stage2Messages, otherFilesToModify, targetFilesToModify := Stage2(
		projectRootDir,
		perpetualDir,
		systemPrompts[prompts.DefaultSystemPromptName],
		implementConfig,
		filesToMdLangMappings,
		planningMode,
		allFileNames,
		filesToReview,
		targetFiles,
		logger)

	var filteredOtherFilesToModify []string
	for _, file := range otherFilesToModify {
		if found, err := utils.FindInRelativeFile(projectRootDir, file, noUploadRegexps); (err == nil || os.IsNotExist(err)) && !found {
			filteredOtherFilesToModify = append(filteredOtherFilesToModify, file)
		} else if found {
			logger.Warnln("Skipping file marked with 'no-upload' comment:", file)
		} else {
			logger.Errorln("Error searching for 'no-upload' comment in file:", file, err)
		}
	}
	otherFilesToModify = filteredOtherFilesToModify

	// Run stage 3
	results := Stage3(
		projectRootDir,
		perpetualDir,
		systemPrompts[prompts.DefaultSystemPromptName],
		implementConfig,
		filesToMdLangMappings,
		stage2Messages,
		otherFilesToModify,
		targetFilesToModify,
		logger)

	// Extra failsafe: filter-out files from results that not among initial files to modify
	var filteredResults = make(map[string]string)
	finalFilesToModify := append(append([]string{}, targetFilesToModify...), otherFilesToModify...)
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
	op_stash.Run([]string{"-a", "-n", newStashFileName}, logger)
}
