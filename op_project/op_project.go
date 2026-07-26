package op_project

import (
	"flag"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/usage"
)

const (
	OpName = "project"
	OpDesc = "Check project configuration or initialize new at the .perpetual directory"
)

func Run(version string, args []string, logger logging.ILogger) {
	var lang, mode, descFile, userFilterFile string
	var help, verbose, trace, envExamples, includeTests bool

	projectFlags := flag.NewFlagSet(OpName, flag.ExitOnError)
	projectFlags.StringVar(&mode, "m", "", "Select operation mode: init, test, list, check-read, check-ascii, save-utf.\n"+
		"init:        Initialize .perpetual dir, write default configuration files for selected programming language or project type.\n"+
		"test:        Search for .perpetual dir, starting from curdir and check json configs. On success show absolute path of .perpetual dir.\n"+
		"list:        List project files accessible by perpetual, relative to project root.\n"+
		"check-read:  Try reading project files as text, on error will print paths of failed files to stdout (relative to project root).\n"+
		"check-ascii: Read project files and ensure it contains only ASCII characters (0-127), on error will print paths of failed files to stdout (relative to project root).\n"+
		"save-utf:    Read project files and convert non-UTF8/16/32 files to UTF8, print paths of affected files to stdout (relative to project root).")
	projectFlags.BoolVar(&help, "h", false, "Show usage")
	projectFlags.StringVar(&lang, "l", "", "Select programming language for '-m init' to setup project configuration with default LLM prompts (valid values: go|dotnet|bash|python3|vb6|c|cpp|arduino|flutter)")
	projectFlags.BoolVar(&envExamples, "ex", false, "Create env-file examples inside .perpetual dir, for use with '-m init'")
	projectFlags.BoolVar(&verbose, "v", false, "Enable debug logging")
	projectFlags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	// for non "init" modes
	projectFlags.StringVar(&descFile, "df", "", "Optional path to project description file (valid values: file-path|disabled), when used with '-m init' it will copy description as default to .perpetual dir")
	projectFlags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from processing when running project-files tests")
	projectFlags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing, when used with '-m init' it will append filter definitions to the project-wide blacklist")
	projectFlags.Parse(args)

	if verbose {
		logger.EnableLevel(logging.DebugLevel)
	}
	if trace {
		logger.EnableLevel(logging.DebugLevel)
		logger.EnableLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'init' operation")
	logger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", projectFlags)
	}

	var err error = nil
	switch mode {
	case "init":
		err = Init(version, lang, envExamples, logger)
	case "test":
		CheckProjectConfig(descFile, logger)
	case "list":
		FetchProjectFiles(descFile, userFilterFile, includeTests, true, logger)
	case "check-read":
		CheckFilesRead(descFile, userFilterFile, includeTests, logger)
	case "check-ascii":
		CheckFilesReadAsAscii(descFile, userFilterFile, includeTests, logger)
	case "save-utf":
		CheckFilesAndSaveAsUTF(descFile, userFilterFile, includeTests, logger)
	case "":
		usage.PrintOperationUsage("You must provide a valid operation mode with the '-m' flag", projectFlags)
	default:
		logger.Errorln("Invalid operation mode:", mode)
		usage.PrintOperationUsage("You must provide a valid operation mode with the '-m' flag", projectFlags)
	}

	if err != nil {
		usage.PrintOperationUsage(err.Error(), projectFlags)
	}
}
