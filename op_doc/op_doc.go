package op_doc

import (
	"flag"
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

const OpName = "doc"
const OpDesc = "Create or rework documentation files (in markdown or plain-text format)"

func docFlags() *flag.FlagSet {
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)
	return flags
}

func Run(args []string, logger, stdErrLogger logging.ILogger) {
	var help, verbose, trace, noAnnotate, forceUpload, includeTests bool
	var docFile, docExample, action, userFilterFile, contextSaving string
	var searchLimit, selectionPasses int

	flags := docFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.StringVar(&contextSaving, "c", "auto", "Context saving mode, reduce LLM context use for large projects (valid values: auto|off|medium|high)")
	flags.BoolVar(&noAnnotate, "n", false, "No annotate mode: skip re-annotating of changed files and use current annotations if any")
	flags.StringVar(&docFile, "r", "", "Target documentation file for processing (if omited, read from stdin and write result to stdout)")
	flags.StringVar(&docExample, "e", "", "Optional documentation file to use as an example/reference for style, structure and format, but not for content")
	flags.StringVar(&action, "a", "write", "Select action to perform (valid values: draft|write|refine)")
	flags.BoolVar(&forceUpload, "f", false, "Disable 'no-upload' file-filter and upload such files for review if reqested")
	flags.IntVar(&searchLimit, "s", 7, "Limit number of files related to target document returned by local search (0 = disable local search, only use LLM-requested files)")
	flags.IntVar(&selectionPasses, "sp", 1, "Set number of passes for related files selection at stage 1")
	flags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from processing")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if docFile == "" {
		logger = stdErrLogger
	}

	if verbose {
		logger.EnableLevel(logging.DebugLevel)
	}
	if trace {
		logger.EnableLevel(logging.DebugLevel)
		logger.EnableLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'doc' operation")
	logger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", flags)
	}

	action = strings.ToUpper(action)
	if action != "DRAFT" && action != "WRITE" && action != "REFINE" {
		logger.Panicln("Invalid action provided")
	}

	contextSaving = shared.ValidateContextSavingValue(contextSaving, logger)

	if searchLimit < 0 {
		logger.Panicln("Search limit (-s flag) value is invalid", searchLimit)
	}
	if selectionPasses < 1 {
		logger.Panicln("Selection passes count (-sp flag) value is invalid", selectionPasses)
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

	var docExampleContent string
	var docContent string

	if action == "DRAFT" {
		docContent = MDDocDraft
	} else {
		docConfig, err := config.LoadOpDocConfig(perpetualDir)
		if err != nil {
			logger.Panicf("Error loading op_doc config: %s", err)
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

		//Load example if provided
		if docExample != "" {
			logger.Infoln("Loading example document:", docExample)
			docExampleContent, err = utils.LoadTextFile(docExample)
			if err != nil {
				logger.Panicln("Error reading example document:", err)
			}
		}

		//Load main document
		if docFile != "" {
			docContent, err = utils.LoadTextFile(docFile)
			if err != nil {
				logger.Panicln("Error reading target document:", err)
			}
		} else {
			logger.Infoln("Reading document from stdin")
			docContent, err = utils.LoadTextStdin()
			if err != nil {
				logger.Panicln("Error reading from stdin:", err)
			}
		}

		if docContent == "" {
			logger.Panicln("Document content is empty, please provide at least a minimal draft to proceed")
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

		// Load annotations needed for stage1
		annotations, err := utils.GetAnnotations(filepath.Join(perpetualDir, utils.AnnotationsFileName), fileNames)
		if err != nil {
			logger.Panicln("Error reading annotations:", err)
		}

		var docPromptPlain string
		var docPromptJson string
		if action == "WRITE" {
			docPromptPlain = docConfig.String(config.K_DocStage1WritePrompt)
			docPromptJson = docConfig.String(config.K_DocStage1WriteJsonModePrompt)
		} else if action == "REFINE" {
			docPromptPlain = docConfig.String(config.K_DocStage1RefinePrompt)
			docPromptJson = docConfig.String(config.K_DocStage1RefineJsonModePrompt)
		} else {
			logger.Panicln("Invalid action:", action)
		}

		// Perform context saving measures - use local search to select only selected percentage of the most relevant files
		filesPercent, randomizePercent := shared.GetLocalSearchLimitsForContextSaving(contextSaving, len(fileNames), projectConfig)
		preselectedFileNames := shared.Stage1Preselect(
			perpetualDir,
			projectRootDir,
			filesPercent,
			randomizePercent,
			fileNames,
			docContent,
			[]string{},
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
			// Run stage1 to find out what project-files contents we need to work on document
			fileLists[pass] = shared.Stage1(
				OpName,
				projectRootDir,
				perpetualDir,
				docConfig,
				projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
				preselectedFileNames[pass],
				fileNames,
				annotations,
				[]string{docConfig.String(config.K_DocExamplePrompt)},
				[]string{docExampleContent},
				[]string{docConfig.String(config.K_DocExampleResponse)},
				docPromptPlain,
				docPromptJson,
				docContent,
				[]string{},
				pass+1,
				stage1Logger)
			// Prepare for local similarity search
			searchMode := shared.GetLocalSearchModeFromContextSavingValue(contextSaving, len(fileLists[pass]), searchLimit)
			// Local similarity search stage
			searchQueries, searchTags := op_embed.GetQueriesForSimilaritySearch(docContent, []string{}, annotations)
			similarFiles := op_embed.SimilaritySearchStage(
				searchMode,
				min(searchLimit, len(fileLists[pass])),
				perpetualDir,
				searchQueries,
				searchTags,
				fileNames,
				fileLists[pass],
				logger)
			fileLists[pass] = append(fileLists[pass], similarFiles...)
		}
		// Merge fileLists together
		if selectionPasses > 1 {
			stage1Logger.EnableLevel(logging.InfoLevel)
		} else {
			stage1Logger.DisableLevel(logging.InfoLevel)
		}
		requestedFiles := shared.MergeFileLists(fileLists, stage1Logger)

		// Check requested files for no-upload mark and filter it out
		var filteredRequestedFiles []string
		if forceUpload {
			filteredRequestedFiles = requestedFiles
		} else {
			for _, file := range requestedFiles {
				if found, err := utils.FindInRelativeFile(
					projectRootDir,
					file,
					docConfig.RegexpArray(config.K_NoUploadCommentsRx)); err == nil && !found {
					filteredRequestedFiles = append(filteredRequestedFiles, file)
				} else if found {
					logger.Warnln("Skipping file marked with 'no-upload' comment:", file)
				} else {
					logger.Errorln("Error searching for 'no-upload' comment in file:", file, err)
				}
			}
		}

		docPrompt := docConfig.String(config.K_DocStage2WritePrompt)
		if action == "REFINE" {
			docPrompt = docConfig.String(config.K_DocStage2RefinePrompt)
		}

		// Run stage2 to make changes to the document and save it to docContent
		docContent = Stage2(
			OpName,
			projectRootDir,
			perpetualDir,
			docConfig,
			projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
			fileNames,
			filteredRequestedFiles,
			annotations,
			[]string{docConfig.String(config.K_DocExamplePrompt)},
			[]string{docExampleContent},
			[]string{docConfig.String(config.K_DocExampleResponse)},
			docPrompt,
			docContent,
			true,
			logger)
	}

	// Add extra newline if not present
	if !strings.HasSuffix(docContent, "\n") {
		docContent += "\n"
	}

	// Write output to file or stdout
	if docFile != "" {
		logger.Infoln("Writing document:", docFile)
		err := utils.SaveTextFile(docFile, docContent)
		if err != nil {
			logger.Panicln("Error writing to output file:", err)
		}
	} else {
		err := utils.WriteTextStdout(docContent)
		if err != nil {
			logger.Panicln("Error writing document to stdout:", err)
		}
	}
}
