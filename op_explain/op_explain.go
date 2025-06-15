package op_explain

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_annotate"
	"github.com/DarkCaster/Perpetual/op_embed"
	"github.com/DarkCaster/Perpetual/shared"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const OpName = "explain"
const OpDesc = "Getting answers to questions and clarifications on the project (based on source code analysis)"

func docFlags() *flag.FlagSet {
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)
	return flags
}

func Run(args []string, logger, stdErrLogger logging.ILogger) {
	var help, addAnnotations, listFilesOnly, verbose, trace, noAnnotate, forceUpload, addQuestion, includeTests bool
	var outputFile, inputFile, userFilterFile, contextSaving string
	var searchLimit int

	flags := docFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.StringVar(&contextSaving, "c", "auto", "Context saving mode, reduce LLM context use for large projects (valid values: auto|off|medium|high)")
	flags.BoolVar(&addAnnotations, "a", false, "Add project annotation in addition to files requested by LLM to improve the quality of the answer")
	flags.BoolVar(&listFilesOnly, "l", false, "Only list files that LLM thinks are related to the question, do not generate the final answer. One filename per line, no formatting.")
	flags.BoolVar(&noAnnotate, "n", false, "No annotate mode: skip re-annotating of changed files and use current annotations if any")
	flags.StringVar(&outputFile, "r", "", "Target file for writing answer, markdown formatted (stdout if not supplied)")
	flags.StringVar(&inputFile, "i", "", "Read question from file, plain text or markdown format (stdin if not supplied)")
	flags.BoolVar(&forceUpload, "f", false, "Disable 'no-upload' file-filter and upload such files for review if reqested")
	flags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from processing")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing")
	flags.IntVar(&searchLimit, "s", 5, "Limit number of files related to question returned by local search (0 = disable local search, only use LLM-requested files)")
	flags.BoolVar(&addQuestion, "q", false, "Include the question text and the list of relevant files in the generated answer")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if outputFile == "" {
		logger = stdErrLogger
	}

	if verbose {
		logger.EnableLevel(logging.DebugLevel)
	}
	if trace {
		logger.EnableLevel(logging.DebugLevel)
		logger.EnableLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'explain' operation")
	logger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", flags)
	}

	contextSaving = shared.ValidateContextSavingValue(contextSaving, logger)

	if searchLimit < 0 {
		logger.Panicln("Similar files limit parameter cannot be less than 0")
	}

	// Find project root and perpetual directories
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

	explainConfig, err := config.LoadOpExplainConfig(perpetualDir)
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
	fileNames, _, err := utils.GetProjectFileList(
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
	var question string
	if inputFile != "" {
		data, err := utils.LoadTextFile(inputFile)
		if err != nil {
			logger.Panicln("Error reading input file:", err)
		}
		question = data
	} else {
		logger.Infoln("Reading question from stdin")
		data, err := utils.LoadTextStdin()
		if err != nil {
			logger.Panicln("Error reading from stdin:", err)
		}
		question = string(data)
	}

	// Trim excess line breaks at both sides of question, and stop on empty input
	question = strings.Trim(question, "\n")
	if len(question) < 1 {
		logger.Panicln("Question is empty, cannot continue")
	}

	logger.Debugln("Rotating log file")
	if err := llm.RotateLLMRawLogFile(perpetualDir); err != nil {
		logger.Panicln("Failed to rotate log file:", err)
	}

	if !noAnnotate {
		logger.Debugln("Running 'annotate' operation to update file annotations")
		op_annotate_params, op_embed_params := shared.GetAnnotateAndEmbedCmdLineFlags(userFilterFile, contextSaving)
		op_annotate.Run(op_annotate_params, true, logger, stdErrLogger)
		op_embed.Run(op_embed_params, true, logger, stdErrLogger)
	}

	// Load annotations
	annotations, err := utils.GetAnnotations(filepath.Join(perpetualDir, utils.AnnotationsFileName), fileNames)
	if err != nil {
		logger.Panicln("Error loading annotations:", err)
	}

	//TODO: preselect project files

	// Run stage 1
	requestedFiles := shared.Stage1(
		OpName,
		projectRootDir,
		perpetualDir,
		explainConfig,
		projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
		fileNames,
		fileNames,
		annotations,
		[]string{}, []string{}, []string{},
		explainConfig.String(config.K_ExplainStage1QuestionPrompt),
		explainConfig.String(config.K_ExplainStage1QuestionJsonModePrompt),
		question,
		[]string{},
		logger)

	searchMode := shared.GetLocalSearchModeFromContextSavingValue(contextSaving, len(requestedFiles), searchLimit)
	if searchLimit > len(requestedFiles) {
		searchLimit = len(requestedFiles)
	}

	// Local similarity search stage
	searchQueries, searchTags := op_embed.GetQueriesForSimilaritySearch(question, []string{}, annotations)
	similarFiles := op_embed.SimilaritySearchStage(searchMode, searchLimit, perpetualDir, searchQueries, searchTags, fileNames, requestedFiles, logger)
	requestedFiles = append(requestedFiles, similarFiles...)

	if listFilesOnly {
		// for this mode, just list files one file per line for easier parsing with 3-rd party tool
		if outputFile != "" {
			err := utils.SaveTextFile(outputFile, strings.Join(requestedFiles, "\n"))
			if err != nil {
				logger.Panicln("Error writing to output file:", err)
			}
		} else {
			for _, filename := range requestedFiles {
				fmt.Println(filename)
			}
		}
		return
	}

	var filteredRequestedFiles []string
	if forceUpload {
		filteredRequestedFiles = requestedFiles
	} else {
		for _, file := range requestedFiles {
			if found, err := utils.FindInRelativeFile(
				projectRootDir,
				file,
				explainConfig.RegexpArray(config.K_NoUploadCommentsRx)); err == nil && !found {
				filteredRequestedFiles = append(filteredRequestedFiles, file)
			} else if found {
				logger.Warnln("Skipping file marked with 'no-upload' comment:", file)
			} else {
				logger.Errorln("Error searching for 'no-upload' comment in file:", file, err)
			}
		}
	}

	// Run stage2 to generate answer to requested question
	answer := Stage2(projectRootDir,
		perpetualDir,
		explainConfig,
		projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
		fileNames,
		filteredRequestedFiles,
		annotations,
		question,
		addAnnotations,
		logger)

	// Here we proposing that LLM returned markdown-formatted content, so format file-list and the rest of the answer accordingly
	var outputMessage llm.Message
	outputMessage = llm.NewMessage(llm.UserRequest)

	if addQuestion {
		// Add question
		outputMessage = llm.AddPlainTextFragment(outputMessage, explainConfig.String(config.K_ExplainOutQuestionHeader))
		outputMessage = llm.AddPlainTextFragment(outputMessage, question)
		// Add header for file list
		if len(requestedFiles) > 0 {
			outputMessage = llm.AddPlainTextFragment(outputMessage, explainConfig.String(config.K_ExplainOutFilesHeader))
		}
		// Add files with status indicators
		for _, file := range requestedFiles {
			var isFiltered bool = true
			for _, filteredFile := range filteredRequestedFiles {
				if file == filteredFile {
					isFiltered = false
					break
				}
			}
			if isFiltered {
				outputMessage = llm.AddTaggedFragment(outputMessage, file, explainConfig.StringArray(config.K_ExplainOutFilteredFilenameTags))
			} else {
				outputMessage = llm.AddTaggedFragment(outputMessage, file, explainConfig.StringArray(config.K_ExplainOutFilenameTags))
			}
		}
		// Add header and answer text
		outputMessage = llm.AddPlainTextFragment(outputMessage, explainConfig.String(config.K_ExplainOutAnswerHeader))
	}

	outputMessage = llm.AddPlainTextFragment(outputMessage, answer)
	outputStrings, err := llm.RenderMessagesToAIStrings(projectConfig.StringArray2D(config.K_ProjectMdCodeMappings), []llm.Message{outputMessage})
	if err != nil {
		logger.Panicln("Error rendering report messages:", err)
	}

	// Write output to file or stdout
	if outputFile != "" {
		logger.Infoln("Writing answer to file:", outputFile)
		err := utils.SaveTextFile(outputFile, strings.Join(outputStrings, "\n"))
		if err != nil {
			logger.Panicln("Error writing to output file:", err)
		}
	} else {
		err := utils.WriteTextStdout(strings.Join(outputStrings, "\n"))
		if err != nil {
			logger.Panicln("Error writing answer to stdout:", err)
		}
	}
}
