package op_explain

import (
	"flag"
	"fmt"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_annotate"
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
	var help, verbose, trace, noAnnotate, forceUpload, includeTests bool
	var outputFile, inputFile, userFilterFile string

	flags := docFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.BoolVar(&noAnnotate, "n", false, "No annotate mode: skip re-annotating of changed files and use current annotations if any")
	flags.StringVar(&outputFile, "r", "", "Target file for writing answer (stdout if not supplied)")
	flags.StringVar(&inputFile, "i", "", "Read question from file (stdin if not supplied)")
	flags.BoolVar(&forceUpload, "f", false, "Disable 'no-upload' file-filter and upload such files for review if reqested")
	flags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from processing")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if outputFile == "" {
		logger = stdErrLogger
	}

	if verbose {
		logger.SetLevel(logging.DebugLevel)
	}
	if trace {
		logger.SetLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'explain' operation")
	logger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", flags)
	}

	// Find project root and perpetual directories
	//projectRootDir, perpetualDir, err := utils.FindProjectRoot(logger)
	_, _, err := utils.FindProjectRoot(logger)
	if err != nil {
		logger.Panicln("Error finding project root directory:", err)
	}

	if !noAnnotate {
		logger.Debugln("Running 'annotate' operation to update file annotations")
		op_annotate.Run(nil, logger)
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
		data, err := utils.LoadTextStdin()
		if err != nil {
			logger.Panicln("Error reading from stdin:", err)
		}
		question = string(data)
	}

	// TODO: Implement the core explain functionality here
	answer := fmt.Sprintf("Question received: %s\nThis feature is not yet implemented.", question)

	// Write output to file or stdout
	if outputFile != "" {
		err := utils.SaveTextFile(outputFile, answer)
		if err != nil {
			logger.Panicln("Error writing to output file:", err)
		}
	} else {
		fmt.Print(answer)
	}
}
