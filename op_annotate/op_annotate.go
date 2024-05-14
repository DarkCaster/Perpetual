package op_annotate

import (
	"flag"
	"os"
	"path/filepath"
	"sort"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/prompts"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const OpName = "annotate"
const OpDesc = "Annotate project files with LLM-generated comments"

func annotateFlags() *flag.FlagSet {
	return flag.NewFlagSet(OpName, flag.ExitOnError)
}

func Run(args []string, logger logging.ILogger) {
	flags := annotateFlags()

	var help, force, dryRun, verbose, trace bool
	var requestedFile string
	flags.BoolVar(&force, "f", false, "Force annotation of all files, even for files which annotations are up to date")
	flags.BoolVar(&dryRun, "d", false, "Perform a dry run without actually generating annotations, list of files that will be annotated and annotations that will be removed")
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

	logger.Infoln("Project root directory:", projectRootDir)
	logger.Debugln("Perpetual directory:", perpetualDir)

	err = utils.LoadEnvFile(filepath.Join(perpetualDir, utils.DotEnvFileName))

	if err != nil {
		logger.Panicln("error loading environment variables:", err)
	}

	var projectFilesWhitelist []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesWhitelistFileName), &projectFilesWhitelist)
	if err != nil {
		logger.Panicln("error reading project-files whitelist regexps:", err)
	}

	var projectFilesBlacklist []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesBlacklistFileName), &projectFilesBlacklist)
	if err != nil {
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

	promptsDir := filepath.Join(perpetualDir, prompts.PromptsDir)

	loadPrompt := func(filePath string) string {
		text, err := utils.LoadTextFile(filepath.Join(promptsDir, filePath))
		if err != nil {
			logger.Panicln("Failed to load prompt:", err)
		}
		return text
	}

	annotatePrompt := loadPrompt(prompts.AnnotatePromptFile)
	annotateResponse := loadPrompt(prompts.AIAnnotateResponseFile)
	systemPrompt := loadPrompt(prompts.SystemPromptFile)

	// Announce start of new LLM session
	llm.LogStartSession(logger, perpetualDir, "annotate", args...)

	// Create llm connector
	connector, err := llm.NewLLMConnector(OpName, systemPrompt, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create LLM connector:", err)
	}
	logger.Debugln(llm.GetDebugString(connector))

	// Generate file annotations
	logger.Infoln("Annotating files, count:", len(filesToAnnotate))
	errorFlag := false
	newAnnotations := make(map[string]string)
	for _, filePath := range filesToAnnotate {
		logger.Infoln(filePath)
		// Read file contents and generate annotation
		fileBytes, err := utils.LoadTextFile(filepath.Join(projectRootDir, filePath))
		if err != nil {
			logger.Panicln("failed to read file:", err)
		}
		fileContents := string(fileBytes)

		annotateRequest := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), annotatePrompt)
		annotateSimulatedResponse := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), annotateResponse)
		fileContentsRequest := llm.AddFileFragment(llm.NewMessage(llm.UserRequest), filePath, fileContents)

		// Log messages
		llm.LogMessage(logger, perpetualDir, connector, &annotateRequest)
		llm.LogMessage(logger, perpetualDir, connector, &annotateSimulatedResponse)
		llm.LogMessage(logger, perpetualDir, connector, &fileContentsRequest)

		// Perform actual query
		annotation, status, err := connector.Query(annotateRequest, annotateSimulatedResponse, fileContentsRequest)

		if err != nil {
			logger.Errorf("LLM query failed with status %d, error: %s", status, err)
			errorFlag = true
			fileChecksums[filePath] = "error"
		} else if status == llm.QueryMaxTokens {
			logger.Errorf("LLM response reached max tokens, try to increase the limit and run again")
			errorFlag = true
			fileChecksums[filePath] = "error"
		} else {
			newAnnotations[filePath] = annotation
		}

		// Log LLM response
		responseMessage := llm.SetRawResponse(llm.NewMessage(llm.RealAIResponse), annotation)
		llm.LogMessage(logger, perpetualDir, connector, &responseMessage)
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
