package op_project

import (
	"flag"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/usage"
)

const OpName = "project"
const OpDesc = "Check project configuration or initialize new at the .perpetual directory"

func initFlags() *flag.FlagSet {
	return flag.NewFlagSet(OpName, flag.ExitOnError)
}

func Run(version string, args []string, logger logging.ILogger) {
	var lang, mode, descFile, userFilterFile string
	var help, verbose, trace, envExamples, includeTests bool

	// Parse flags for the "init" operation
	initFlags := initFlags()
	initFlags.StringVar(&mode, "m", "", "Select operation mode: init, test, list, check-read, check-ascii, save-utf.\n"+
		"init:        Initialize .perpetual dir, write default configuration files for selected programming language or project type.\n"+
		"test:        Search for .perpetual dir, starting from curdir and check json configs. On success show absolute path of .perpetual dir.\n"+
		"list:        List project files accessible by perpetual, relative to project root.\n"+
		"check-read:  Try reading project files as text, on error will print paths of failed files to stdout (relative to project root).\n"+
		"check-ascii: Read project files and ensure it contains only ASCII characters (0-127), on error will print paths of failed files to stdout (relative to project root).\n"+
		"save-utf:    Read project files and convert non-UTF8/16/32 files to UTF8, print paths of affected files to stdout (relative to project root).")
	initFlags.BoolVar(&help, "h", false, "Show usage")
	initFlags.StringVar(&lang, "l", "", "Select programming language for '-m init' to setup project configuration with default LLM prompts (valid values: go|dotnet|bash|python3|vb6|c|cpp|arduino|flutter)")
	initFlags.BoolVar(&envExamples, "ex", false, "Create env-file examples inside .perpetual dir, for use with '-m init'")
	initFlags.BoolVar(&verbose, "v", false, "Enable debug logging")
	initFlags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	// for non "init" modes
	initFlags.StringVar(&descFile, "df", "", "Optional path to project description file (valid values: file-path|disabled)")
	initFlags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from processing with '-m' flag")
	initFlags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing with '-m' flag")
	initFlags.Parse(args)

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
		usage.PrintOperationUsage("", initFlags)
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
		usage.PrintOperationUsage("You must provide a valid operation mode with the '-m' flag", initFlags)
	default:
		logger.Errorln("Invalid operation mode:", mode)
		usage.PrintOperationUsage("You must provide a valid operation mode with the '-m' flag", initFlags)
	}

	if err != nil {
		usage.PrintOperationUsage(err.Error(), initFlags)
	}
}
