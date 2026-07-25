package op_init

import (
	"flag"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/usage"
)

const OpName = "init"
const OpDesc = "Initialize new .perpetual directory, will store project configuration"

func initFlags() *flag.FlagSet {
	return flag.NewFlagSet(OpName, flag.ExitOnError)
}

func Run(version string, args []string, logger logging.ILogger) {
	var lang, mode string
	var help, verbose, trace, envExamples bool

	// Parse flags for the "init" operation
	initFlags := initFlags()
	initFlags.StringVar(&mode, "m", "", "Select operation mode: init, ...")
	initFlags.BoolVar(&help, "h", false, "Show usage")
	initFlags.StringVar(&lang, "l", "", "Select programming language for '-m init' to setup project configuration with default LLM prompts (valid values: go|dotnet|bash|python3|vb6|c|cpp|arduino|flutter)")
	initFlags.BoolVar(&envExamples, "ex", false, "Create env-file examples inside .perpetual dir, for use with '-m init'")
	initFlags.BoolVar(&verbose, "v", false, "Enable debug logging")
	initFlags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
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
