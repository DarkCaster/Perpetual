package op_annotate

import (
	"flag"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"

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
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesWhitelistFileName), &projectFilesWhitelist)
	if err != nil {
		logger.Panicln("error reading project-files whitelist regexps:", err)
	}

	var filesToMdLangMappings [][2]string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesToMarkdownLangMappingFileName), &filesToMdLangMappings)
	if err != nil {
		logger.Warnln("Error reading optional filename to markdown-lang mappings:", err)
	}

	var projectFilesBlacklist []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesBlacklistFileName), &projectFilesBlacklist)
	if err != nil {
		logger.Panicln("error reading project-files blacklist regexps:", err)
	}

	var fileNameTags []string
	err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.FileNameTagsFileName), &fileNameTags)
	if err != nil {
		logger.Panicln("Error loading filename-tags json:", err)
	}
	if len(fileNameTags) != 2 {
		logger.Panicln("filename-tags json must contain exact 2 tags:", err)
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

	var annotatePromptMappings [][2]string
	err = utils.LoadJsonFile(filepath.Join(promptsDir, prompts.AnnotatePromptFile), &annotatePromptMappings)
	if err != nil {
		logger.Panicln("error getting annotate-prompts json file:", err)
	}

	loadPrompt := func(filePath string) string {
		text, err := utils.LoadTextFile(filepath.Join(promptsDir, filePath))
		if err != nil {
			logger.Panicln("Failed to load prompt:", err)
		}
		return text
	}

	annotateResponse := loadPrompt(prompts.AIAnnotateResponseFile)
	systemPrompt := loadPrompt(prompts.SystemPromptFile)

	// Create llm connector
	connector, err := llm.NewLLMConnector(OpName, systemPrompt, filesToMdLangMappings, llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create LLM connector:", err)
	}
	logger.Debugln(llm.GetDebugString(connector))

	// Load output tags regexps
	outputTagsRxStrings := utils.LoadStringPair(filepath.Join(perpetualDir, prompts.OutputTagsRXFileName), 2, math.MaxInt, 2, logger)

	// Generate file annotations
	logger.Infoln("Annotating files, count:", len(filesToAnnotate))
	errorFlag := false
	newAnnotations := make(map[string]string)
	for _, filePath := range filesToAnnotate {
		annotatePrompt := ""
		//detect actual prompt for annotating this particular file
		for _, mapping := range annotatePromptMappings {
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

		annotateRequest := llm.AddPlainTextFragment(llm.NewMessage(llm.UserRequest), annotatePrompt)
		annotateSimulatedResponse := llm.AddPlainTextFragment(llm.NewMessage(llm.SimulatedAIResponse), annotateResponse)
		fileContentsRequest := llm.AddFileFragment(llm.NewMessage(llm.UserRequest), filePath, fileContents, fileNameTags)

		onFailRetriesLeft := connector.GetOnFailureRetryLimit()
		if onFailRetriesLeft < 1 {
			onFailRetriesLeft = 1
		}
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			logger.Infoln(filePath)
			// Perform actual query
			annotation, status, err := connector.Query(annotateRequest, annotateSimulatedResponse, fileContentsRequest)
			if err != nil {
				logger.Errorf("LLM query failed with status %d, error: %s", status, err)
				if onFailRetriesLeft < 1 {
					fileChecksums[filePath] = "error"
					errorFlag = true
				}
			} else if status == llm.QueryMaxTokens {
				logger.Errorf("LLM response reached max tokens, consider increasing the limit")
				if onFailRetriesLeft < 1 {
					fileChecksums[filePath] = "error"
					errorFlag = true
				}
			} else if _, err := utils.GetTextAfterFirstMatches(annotation, getEvenIndexElements(outputTagsRxStrings)); err == nil {
				logger.Errorf("LLM response contains code blocks, which is not allowed", filePath)
				if onFailRetriesLeft < 1 {
					fileChecksums[filePath] = "error"
					errorFlag = true
				}
			} else {
				newAnnotations[filePath] = annotation
				break
			}
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
