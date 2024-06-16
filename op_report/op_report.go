package op_report

import (
	"flag"
	"path/filepath"
	"strings"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_annotate"
	"github.com/DarkCaster/Perpetual/prompts"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const (
	OpName = "report"
	OpDesc = "Create report from project source code, that can be manually copypasted into the LLM user-interface for further manual analisys"
)

func Run(args []string, logger logging.ILogger) {
	var help, verbose, trace bool
	var reportType, outputFile string

	flags := flag.NewFlagSet(OpName, flag.ExitOnError)
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.StringVar(&reportType, "t", "code", "Select report type (valid values: code|brief)")
	flags.StringVar(&outputFile, "r", "", "File path to write report to (write to stderr if not provided or empty)")
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

	promptsDir := filepath.Join(perpetualDir, prompts.PromptsDir)
	logger.Debugln("Prompts directory:", promptsDir)

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

	// Get project files, which names selected with whitelist regexps and filtered with blacklist regexps
	fileChecksums, fileNames, _, err := utils.GetProjectFileList(projectRootDir, perpetualDir, projectFilesWhitelist, projectFilesBlacklist)
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

	if strings.ToUpper(reportType) == "BRIEF" {
		logger.Debugln("Running 'annotate' operation to update file annotations")
		op_annotate.Run(nil, logger)

		// Load annotations
		annotations, err := utils.GetAnnotations(filepath.Join(perpetualDir, utils.AnnotationsFileName), fileChecksums)
		if err != nil {
			logger.Panicln("Error loading annotations:", err)
		}

		// Generate report messages
		reportMessage := llm.NewMessage(llm.UserRequest)
		for _, filename := range fileNames {
			annotation, ok := annotations[filename]
			if !ok {
				annotation = "No annotation available"
			}
			reportMessage = llm.AddIndexFragment(reportMessage, filename, nil)
			reportMessage = llm.AddPlainTextFragment(reportMessage, annotation)
		}

	} else if strings.ToUpper(reportType) == "CODE" {
	} else {
		logger.Panicln("Invalid report type:", reportType)
	}

}
