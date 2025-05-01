package op_report

import (
	"flag"
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
	OpDesc = "Create report from project source code, that can be manually copypasted into the LLM user-interface for further manual analysis"
)

func Run(args []string, logger, stdErrLogger logging.ILogger) {
	var help, verbose, trace, includeTests bool
	var reportType, outputFile, userFilterFile, contextSaving string

	//TODO: add selection of llm-type and use llm-agnostic message formatting for that particular llm type
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.StringVar(&contextSaving, "c", "auto", "Context saving mode, reduce LLM context use for large projects (valid values: auto|off|medium|high)")
	flags.StringVar(&reportType, "t", "code", "Select report type (valid values: code|brief)")
	flags.StringVar(&outputFile, "r", "", "File path to write report to (write to stdout if not provided or empty)")
	flags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from report")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from report")
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

	logger.Debugln("Starting 'report' operation")
	logger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", flags)
	}

	contextSaving = strings.ToUpper(contextSaving)
	if contextSaving != "AUTO" && contextSaving != "OFF" && contextSaving != "MEDIUM" && contextSaving != "HIGH" {
		logger.Panicln("Invalid context saving mode value provided")
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

	reportConfig, err := config.LoadOpReportConfig(perpetualDir)
	if err != nil {
		logger.Panicf("Error loading op_report config: %s", err)
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

	var reportMessage llm.Message
	if strings.ToUpper(reportType) == "BRIEF" {
		logger.Debugln("Running 'annotate' operation to update file annotations")
		op_annotate_params := []string{}
		if userFilterFile != "" {
			op_annotate_params = append(op_annotate_params, "-x", userFilterFile)
		}
		if contextSaving != "AUTO" {
			op_annotate_params = append(op_annotate_params, "-c", contextSaving)
		}
		logger.Debugln("Rotating log file")
		if err := llm.RotateLLMRawLogFile(perpetualDir); err != nil {
			logger.Panicln("Failed to rotate log file:", err)
		}
		op_annotate.Run(op_annotate_params, true, logger, stdErrLogger)
		// Load annotations
		annotations, err := utils.GetAnnotations(filepath.Join(perpetualDir, utils.AnnotationsFileName), fileNames)
		if err != nil {
			logger.Panicln("Error loading annotations:", err)
		}
		// Generate report message
		reportMessage = llm.ComposeMessageWithAnnotations(
			reportConfig.String(config.K_ReportBriefPrompt),
			fileNames,
			reportConfig.StringArray(config.K_FilenameTags),
			annotations,
			logger)
	} else if strings.ToUpper(reportType) == "CODE" {
		// Generate report messages
		reportMessage = llm.ComposeMessageWithFiles(
			projectRootDir,
			reportConfig.String(config.K_ReportCodePrompt),
			fileNames,
			reportConfig.StringArray(config.K_FilenameTags),
			logger)
	} else {
		logger.Panicln("Invalid report type:", reportType)
	}
	reportStrings, err := llm.RenderMessagesToAIStrings(projectConfig.StringArray2D(config.K_ProjectMdCodeMappings), []llm.Message{reportMessage})

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
		err = utils.WriteTextStdout(strings.Join(reportStrings, "\n"))
		if err != nil {
			logger.Panicln("Error writing report to stdout:", err)
		}
	}
}
