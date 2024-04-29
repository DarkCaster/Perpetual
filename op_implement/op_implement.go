package op_implement

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/op_annotate"
	"github.com/DarkCaster/Perpetual/prompts"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
	"github.com/sirupsen/logrus"
)

const OpName = "implement"
const OpDesc = "Implement code accodring instructions marked with ###IMPLEMENT### comments"

func implementFlags() *flag.FlagSet {
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)

	return flags
}

func Run(args []string, logger *logrus.Logger) {
	logger.Debugln("Starting 'implement' operation")

	var help, noAnnotate, allFiles, planning, verbose, trace bool
	var manualFilePath string

	// Parse flags for the "implement" operation
	flags := implementFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.BoolVar(&noAnnotate, "n", false, "No annotate mode: skip re-annotating of changed files and use current annotations if any")
	flags.BoolVar(&allFiles, "a", false, "Send all project files to context (warning: may incur high costs, may overflow LLM context)")
	flags.BoolVar(&planning, "p", false, "Enable extended planning stage, needed for bigger modifications that may create new files, not needed on single file modifications. Disabled by default to save tokens.")
	flags.StringVar(&manualFilePath, "r", "", "Manually request a file for the operation, otherwise select files automatically")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if verbose {
		logger.SetLevel(logrus.DebugLevel)
	}
	if trace {
		logger.SetLevel(logrus.TraceLevel)
	}
	logger.Traceln("Parsed flags:", "help:", help, "noAnnotate:", noAnnotate, "allFiles:", allFiles, "manualFilePath:", manualFilePath, "verbose:", verbose, "trace:", trace)

	if help {
		usage.PrintOperationUsage("", flags)
	}

	// We do not need file annotations when sending all the files content to the context
	if allFiles {
		noAnnotate = true
	}

	if !noAnnotate {
		logger.Debugln("Running 'annotate' operation to update file annotations")
		op_annotate.Run(nil, logger)
	} else if !allFiles {
		logger.Warnln("File-annotations update disabled, this may worsen the final result")
	}

	projectRootDir, perpetualDir, err := utils.FindProjectRoot(logger)
	if err != nil {
		logger.Fatalln("error finding project root directory:", err)
	}

	logger.Println("Project root directory:", projectRootDir)
	logger.Debugln("Perpetual directory:", perpetualDir)

	err = utils.LoadEnvFile(filepath.Join(perpetualDir, utils.DotEnvFileName))
	if err != nil {
		logger.Fatalln("error loading environment variables:", err)
	}

	var projectFilesWhitelist []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesWhitelistFileName), &projectFilesWhitelist)
	if err != nil {
		logger.Fatalln("error reading project-files whitelist regexps:", err)
	}

	var projectFilesBlacklist []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesBlacklistFileName), &projectFilesBlacklist)
	if err != nil {
		logger.Fatalln("error reading project-files blacklist regexps:", err)
	}

	var implementRxStrings []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.OpImplementCommentRXFileName), &implementRxStrings)
	if err != nil {
		logger.Fatalln("error reading implement-operation regexps:", err)
	}

	var noUploadRxStrings []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.NoUploadCommentRXFileName), &noUploadRxStrings)
	if err != nil {
		logger.Fatalln("error reading no-upload regexps:", err)
	}

	loadStringPair := func(file string) []string {
		var result []string
		err = utils.LoadJsonFile(filepath.Join(perpetualDir, file), &result)
		if err != nil {
			logger.Fatalln("error loading json:", err)
		}
		if len(result) != 2 {
			logger.Fatalln("file may only contain 2 tags and nothing more:", file)
		}
		return result
	}

	fileNameTagsRxStrings := loadStringPair(prompts.FileNameTagsRXFileName)
	fileNameTagsStrings := loadStringPair(prompts.FileNameTagsFileName)
	outputTagsRxStrings := loadStringPair(prompts.OutputTagsRXFileName)
	var fileNameEmbedRXString string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.FileNameEmbedRXFileName), &fileNameEmbedRXString)
	if err != nil {
		logger.Fatalln("error loading filename-embed regexp json:", err)
	}

	fileChecksums, fileNames, allFileNames, err := utils.GetProjectFileList(projectRootDir, perpetualDir, projectFilesWhitelist, projectFilesBlacklist)
	if err != nil {
		logger.Fatalln("error getting project file-list:", err)
	}

	// Check fileNames array for case collisions
	if !utils.CheckFilenameCaseCollisions(fileNames) {
		logger.Fatalln("Filename case collisions detected in project files")
	}
	// File names and dir-names must not contain path separators characters
	if !utils.CheckForPathSeparatorsInFilenames(fileNames) {
		logger.Fatalln("Invalid characters detected in project filenames or directories: / and \\ characters are not allowed!")
	}

	var implementRegexps []*regexp.Regexp
	for _, rx := range implementRxStrings {
		crx, err := regexp.Compile(rx)
		if err != nil {
			logger.Fatalln("failed to compile 'implement' comment search regexp: ", err)
		}
		implementRegexps = append(implementRegexps, crx)
	}

	var noUploadRegexps []*regexp.Regexp
	for _, rx := range noUploadRxStrings {
		crx, err := regexp.Compile(rx)
		if err != nil {
			logger.Fatalln("failed to compile 'no-upload' comment search regexp: ", err)
		}
		noUploadRegexps = append(noUploadRegexps, crx)
	}

	// Find files for operation
	var targetFiles []string

	if manualFilePath != "" {
		//check path relative to project root directory, and make it path relative to it
		targetFile, err := utils.MakePathRelative(projectRootDir, manualFilePath, false)
		if err != nil {
			logger.Fatalln("failed to process file path: ", err)
		}
		targetFile, found := utils.CaseInsensitiveFileSearch(targetFile, fileNames)
		if !found {
			logger.Fatalln("Requested file not found in project:", targetFile)
		}
		found, err = utils.FindInFile(filepath.Join(projectRootDir, targetFile), implementRegexps)
		if err != nil {
			logger.Fatalln("Failed to search 'implement' comment in file: ", err)
		}
		if !found {
			logger.Errorln("Cannot find 'implement' comment in manually provided file, expect LLM to provide wrong results for the file", targetFile)
		}
		targetFiles = append(targetFiles, targetFile)
	} else {
		for _, filePath := range fileNames {
			logger.Debugln("File:", filePath)
			found, err := utils.FindInFile(filepath.Join(projectRootDir, filePath), implementRegexps)
			if err != nil {
				logger.Fatalln("failed to search 'implement' comment in file: ", err)
			}
			if found {
				targetFiles = append(targetFiles, filePath)
			}
		}
	}

	checkNoUpload := func(filePath string) bool {
		targetFile, err := utils.MakePathRelative(projectRootDir, filePath, true)
		if err != nil {
			logger.Fatalln("failed to process file path: ", err)
		}
		found, err := utils.FindInFile(filepath.Join(projectRootDir, targetFile), noUploadRegexps)
		if os.IsNotExist(err) {
			return true
		}
		if err != nil {
			logger.Fatalln("failed to search 'no-upload' comment in file: ", err)
		}
		return !found
	}

	// Log files to process
	if len(targetFiles) < 1 {
		logger.Fatalln("No files found for processing")
	}
	logger.Println("Files for processing:")
	for _, targetFile := range targetFiles {
		logger.Println(targetFile)
	}

	// Create prompts
	promptsDir := filepath.Join(perpetualDir, prompts.PromptsDir)
	loadPrompt := func(filePath string, errorMsg string) string {
		bytes, err := utils.LoadTextFile(filepath.Join(promptsDir, filePath))
		if err != nil {
			logger.Fatalln(errorMsg, err)
		}
		return string(bytes)
	}

	annotations, err := utils.GetAnnotations(filepath.Join(perpetualDir, utils.AnnotationsFileName), fileChecksums)
	if err != nil {
		logger.Fatalln("error reading annotations:", err)
	}

	systemPrompt := loadPrompt(prompts.SystemPromptFile, "failed to read system prompt:")

	// Announce start of new LLM session
	llm.LogStartSession(logger, perpetualDir, "implement (stage1)", args...)

	// Run stage 1
	filesToReview := Stage1(projectRootDir, perpetualDir, promptsDir, systemPrompt, fileNameTagsRxStrings, fileNames, annotations, targetFiles, logger)

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
