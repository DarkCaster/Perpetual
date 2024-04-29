package op_annotate

import (
	"flag"
	"os"
	"path/filepath"
	"sort"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/prompts"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
	"github.com/sirupsen/logrus"
)

const OpName = "annotate"
const OpDesc = "Annotate project files with LLM-generated comments"

func annotateFlags() *flag.FlagSet {
	return flag.NewFlagSet(OpName, flag.ExitOnError)
}

func Run(args []string, logger *logrus.Logger) {
	logger.Debugln("Starting 'annotate' operation")
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
		logger.SetLevel(logrus.DebugLevel)
	}
	if trace {
		logger.SetLevel(logrus.TraceLevel)
	}
	logger.Traceln("Parsed flags:", "help:", help, "force:", force, "dryRun:", dryRun, "requestedFile:", requestedFile, "verbose:", verbose, "trace:", trace)

	if help {
		usage.PrintOperationUsage("", flags)
	}

	if requestedFile != "" {
		force = true
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

	fileChecksums, fileNames, _, err := utils.GetProjectFileList(projectRootDir, perpetualDir, projectFilesWhitelist, projectFilesBlacklist)
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

	annotationsFilePath := filepath.Join(perpetualDir, utils.AnnotationsFileName)
	var filesToAnnotate []string
	if !force {
		filesToAnnotate, err = utils.GetChangedFiles(annotationsFilePath, fileChecksums)
		if err != nil {
			logger.Fatalln("error getting changed files:", err)
		}
	} else {
		if requestedFile != "" {
			// Check if requested file is within fileNames array
			requestedFile, err := utils.MakePathRelative(projectRootDir, requestedFile, false)
			if err != nil {
				logger.Fatalln("Requested file is not inside project root", requestedFile)
			}
			requestedFile, found := utils.CaseInsensitiveFileSearch(requestedFile, fileNames)
			if !found {
				logger.Fatalln("Requested file not found in project")
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
		logger.Println("Files to annotate:")
		for _, file := range filesToAnnotate {
			logger.Println(file)
		}
		os.Exit(0)
	}

	promptsDir := filepath.Join(perpetualDir, prompts.PromptsDir)

	loadPrompt := func(filePath string, errorMsg string) string {
		bytes, err := utils.LoadTextFile(filepath.Join(promptsDir, filePath))
		if err != nil {
			logger.Fatalln(errorMsg, err)
		}
		return string(bytes)
	}

	annotatePrompt := loadPrompt(prompts.AnnotatePromptFile, "failed to read annotation prompt:")
	annotateResponse := loadPrompt(prompts.AIAnnotateResponseFile, "failed to read annotate response prompt:")
	systemPrompt := loadPrompt(prompts.SystemPromptFile, "failed to read system prompt:")

	// Announce start of new LLM session
	llm.LogStartSession(logger, perpetualDir, "annotate", args...)

	// Create llm connector
	connector, err := llm.NewLLMConnector(OpName, systemPrompt, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Fatalln("failed to create LLM connector:", err)
	}

	// Generate file annotations
	logger.Println("Annotating files, count:", len(filesToAnnotate))
	errorFlag := false
	newAnnotations := make(map[string]string)
	for _, filePath := range filesToAnnotate {
		logger.Println(filePath)
		// Read file contents and generate annotation
		fileBytes, err := utils.LoadTextFile(filepath.Join(projectRootDir, filePath))
		if err != nil {
			logger.Fatalln("failed to read file:", err)
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
		annotation, err := connector.Query(annotateRequest, annotateSimulatedResponse, fileContentsRequest)

		if err != nil {
			logger.Error("LLM query failed: ", err)
			errorFlag = true
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
		logger.Fatalln("Failed to read old annotations:", err)
	}

	// Copy new annotations back to old annotations
	for element := range newAnnotations {
		annotations[element] = newAnnotations[element]
	}

	// Save updated annotations
	logger.Println("Saving annotations")
	if err := utils.SaveAnnotations(annotationsFilePath, fileChecksums, annotations); err != nil {
		logger.Fatalln("Failed to save annotations:", err)
	}

	if errorFlag {
		logger.Fatalln("Not all files were successfully annotated. Run annotate again to try to index the failed files.")
	}
}
