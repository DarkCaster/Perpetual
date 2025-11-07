package op_misc

import (
	"flag"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/usage"
)

const OpName = "misc"
const OpDesc = "Helper functions not covered by other operations. All human-readable logging goes to stderr and machine parsable output to stdout."

func miscFlags() *flag.FlagSet {
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)

	return flags
}

func Run(args []string, stdErrLogger logging.ILogger) {
	var help, verbose, trace, includeTests bool
	var descFile, userFilterFile string

	// Parse flags for the "misc" operation
	flags := miscFlags()
	flags.BoolVar(&help, "h", false, "Show usage")

	flags.StringVar(&descFile, "df", "", "Optional path to project description file (valid values: file-path|disabled)")
	flags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from processing")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if verbose {
		stdErrLogger.EnableLevel(logging.DebugLevel)
	}
	if trace {
		stdErrLogger.EnableLevel(logging.DebugLevel)
		stdErrLogger.EnableLevel(logging.TraceLevel)
	}

	stdErrLogger.Debugln("Starting 'implement' operation")
	stdErrLogger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", flags)
	}
}
