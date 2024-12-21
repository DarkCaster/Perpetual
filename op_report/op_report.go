package op_report

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_annotate"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const (
	OpName = "report"
	OpDesc = "Create report from project source code, that can be manually copypasted into the LLM user-interface for further manual analisys"
)

func Run(args []string, logger logging.ILogger) {
	var help, verbose, trace, includeTests bool
	var reportType, outputFile, userFilterFile string

	//TODO: add selection of llm-type and use llm-agnostic message formatting for that particular llm type
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.StringVar(&reportType, "t", "code", "Select report type (valid values: code|brief)")
	flags.StringVar(&outputFile, "r", "", "File path to write report to (write to stderr if not provided or empty)")
	flags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from report")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from report")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")

	flags.Parse(args)

	if verbose {
		logger.SetLevel(logging.DebugLevel)
	}
	if trace {
		logger.SetLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'report' operation")
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
	fileChecksums, fileNames, _, err := utils.GetProjectFileList(
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

	//TODO: Load filename tags from file when using llm-agnostic formatting
	//fileNameTagsStrings := utils.LoadStringPair(filepath.Join(perpetualDir, prompts.FileNameTagsFileName), 2, 2, 2, logger)
	fileNameTagsStrings := []string{"### File: ", ""}

	var reportMessage llm.Message
	if strings.ToUpper(reportType) == "BRIEF" {
		logger.Debugln("Running 'annotate' operation to update file annotations")
		op_annotate.Run(nil, logger)

		// Load annotations
		annotations, err := utils.GetAnnotations(filepath.Join(perpetualDir, utils.AnnotationsFileName), fileChecksums)
		if err != nil {
			logger.Panicln("Error loading annotations:", err)
		}

		// Generate report message
		reportMessage = llm.AddPlainTextFragment(
			llm.NewMessage(llm.UserRequest),
			implementConfig.String(config.K_ImplementStage1IndexPrompt))

		for _, filename := range fileNames {
			annotation, ok := annotations[filename]
			if !ok {
				annotation = "No annotation available"
			}
			reportMessage = llm.AddIndexFragment(reportMessage, filename, fileNameTagsStrings)
			reportMessage = llm.AddPlainTextFragment(reportMessage, annotation)
		}
	} else if strings.ToUpper(reportType) == "CODE" {
		// Generate report messages
		reportMessage = llm.AddPlainTextFragment(
			llm.NewMessage(llm.UserRequest),
			implementConfig.String(config.K_ImplementStage2CodePrompt))

		// Iterate over fileNames and add file contents to report message using llm.AddFileFragment
		for _, filename := range fileNames {
			fileContents, err := utils.LoadTextFile(filepath.Join(projectRootDir, filename))
			if err != nil {
				logger.Panicln("Error reading file:", filename, err)
			}
			reportMessage = llm.AddFileFragment(reportMessage, filename, fileContents, fileNameTagsStrings)
		}
	} else {
		logger.Panicln("Invalid report type:", reportType)
	}

	reportStrings, err := llm.RenderMessagesToAIStrings(
		projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
		[]llm.Message{reportMessage})

	if err != nil {
		logger.Panicln("Error rendering report messages:", err)
	}

	// Save report string to file or print it to stderr
	if outputFile != "" {
		err = utils.SaveTextFile(outputFile, strings.Join(reportStrings, "\n"))
		if err != nil {
			logger.Panicln("Error writing report to file:", err)
		}
	} else {
		fmt.Fprintln(os.Stderr, strings.Join(reportStrings, "\n"))
	}
}
