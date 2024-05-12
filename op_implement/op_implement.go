package op_implement

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_annotate"
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
	var help, noAnnotate, planning, verbose, trace bool
	var manualFilePath string

	// Parse flags for the "implement" operation
	flags := implementFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.BoolVar(&noAnnotate, "n", false, "No annotate mode: skip re-annotating of changed files and use current annotations if any")
	flags.BoolVar(&planning, "p", false, "Enable extended planning stage, needed for bigger modifications that may create new files, not needed on single file modifications. Disabled by default to save tokens.")
	flags.StringVar(&manualFilePath, "r", "", "Manually request a file for the operation, otherwise select files automatically")
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

	// Initialize: detect work directories, load .env file with LLM settings, load file filtering regexps
	projectRootDir, perpetualDir, err := utils.FindProjectRoot(logger)
	if err != nil {
		logger.Panicln("Error finding project root directory:", err)
	}

	logger.Infoln("Project root directory:", projectRootDir)
	logger.Debugln("Perpetual directory:", perpetualDir)

	promptsDir := filepath.Join(perpetualDir, prompts.PromptsDir)
	logger.Debugln("Prompts directory:", promptsDir)

	err = utils.LoadEnvFile(filepath.Join(perpetualDir, utils.DotEnvFileName))
	if err != nil {
		logger.Panicln("Error loading environment variables:", err)
	}

	var projectFilesWhitelist []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesWhitelistFileName), &projectFilesWhitelist)
	if err != nil {
		logger.Panicln("Error reading project-files whitelist regexps:", err)
	}

	var projectFilesBlacklist []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesBlacklistFileName), &projectFilesBlacklist)
	if err != nil {
		logger.Panicln("Error reading project-files blacklist regexps:", err)
	}

	var implementRxStrings []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.OpImplementCommentRXFileName), &implementRxStrings)
	if err != nil {
		logger.Panicln("Error reading implement-operation regexps:", err)
	}

	var noUploadRxStrings []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.NoUploadCommentRXFileName), &noUploadRxStrings)
	if err != nil {
		logger.Panicln("Error reading no-upload regexps:", err)
	}

	loadStringPair := func(file string) []string {
		var result []string
		err = utils.LoadJsonFile(filepath.Join(perpetualDir, file), &result)
		if err != nil {
			logger.Panicln("Error loading json:", err)
		}
		if len(result) != 2 {
			logger.Panicln("File may only contain 2 tags and nothing more:", file)
		}
		return result
	}

	fileNameTagsRxStrings := loadStringPair(prompts.FileNameTagsRXFileName)
	fileNameTagsStrings := loadStringPair(prompts.FileNameTagsFileName)
	outputTagsRxStrings := loadStringPair(prompts.OutputTagsRXFileName)
	var fileNameEmbedRXString string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.FileNameEmbedRXFileName), &fileNameEmbedRXString)
	if err != nil {
		logger.Panicln("Error loading filename-embed regexp json:", err)
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
	for _, rx := range implementRxStrings {
		crx, err := regexp.Compile(rx)
		if err != nil {
			logger.Panicln("Failed to compile 'implement' comment search regexp: ", err)
		}
		implementRegexps = append(implementRegexps, crx)
	}

	var noUploadRegexps []*regexp.Regexp
	for _, rx := range noUploadRxStrings {
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
			logger.Panicln("Requested file not found in project:", targetFile)
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

	systemPrompt, err := utils.LoadTextFile(filepath.Join(promptsDir, prompts.SystemPromptFile))
	if err != nil {
		logger.Warnln("Failed to read system prompt:", err)
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
			// Announce start of new LLM session
			llm.LogStartSession(logger, perpetualDir, "implement (stage1)", args...)
			// Run stage 1
			filesToReview = Stage1(projectRootDir, perpetualDir, promptsDir, systemPrompt, fileNameTagsRxStrings, fileNames, annotations, targetFiles, logger)
		} else {
			logger.Warnln("No annotaions found for files not in to-implement list, no need to run stage1")
		}
	}

	checkNoUpload := func(filePath string) bool {
		targetFile, err := utils.MakePathRelative(projectRootDir, filePath, true)
		if err != nil {
			logger.Panicln("Failed to process file path: ", err)
		}
		found, err := utils.FindInFile(filepath.Join(projectRootDir, targetFile), noUploadRegexps)
		if os.IsNotExist(err) {
			return true
		}
		if err != nil {
			logger.Panicln("Failed to search 'no-upload' comment in file: ", err)
		}
		return !found
	}

	// Check filesToReview with checkNoUpload function and remove files from the list for which checkNoUpload returns false
	var filteredFilesToReview []string
	for _, file := range filesToReview {
		if checkNoUpload(file) {
			filteredFilesToReview = append(filteredFilesToReview, file)
		} else {
			logger.Warnln("Skipping file marked with 'no-upload' comment:", file)
		}
	}
	filesToReview = filteredFilesToReview

	// Announce start of new LLM session
	llm.LogStartSession(logger, perpetualDir, "implement (stage2,stage3,stage4)", args...)

	// Run stage 2
	stage2Messages, otherFilesToModify, targetFilesToModify := Stage2(projectRootDir, perpetualDir, promptsDir, systemPrompt, planning, fileNameTagsRxStrings, fileNameTagsStrings, allFileNames, filesToReview, targetFiles, logger)

	var filteredOtherFilesToModify []string
	for _, file := range otherFilesToModify {
		if checkNoUpload(file) {
			filteredOtherFilesToModify = append(filteredOtherFilesToModify, file)
		} else {
			logger.Warnln("Skipping file marked with 'no-upload' comment:", file)
		}
	}
	otherFilesToModify = filteredOtherFilesToModify

	// Run stage 3
	results := Stage3(projectRootDir, perpetualDir, promptsDir, systemPrompt, outputTagsRxStrings, fileNameEmbedRXString, stage2Messages, otherFilesToModify, targetFilesToModify, logger)

	// Filter results similar to filteredOtherFilesToModify: remove files from map that marked with no-upload comment
	var filteredResults = make(map[string]string)
	for file, content := range results {
		shouldSkip := false
		for _, targetFile := range targetFilesToModify {
			if file == targetFile {
				shouldSkip = false
				break
			}
			shouldSkip = true
		}
		if shouldSkip && !checkNoUpload(file) {
			logger.Warnln("Skipping file marked with 'no-upload' comment:", file)
		} else {
			filteredResults[file] = content
		}
	}

	// Run stage 4
	Stage4(projectRootDir, filteredResults, fileNames, logger)
}
