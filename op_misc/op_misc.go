package op_misc

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const OpName = "misc"
const OpDesc = "Helper functions not covered by other operations. All human-readable logging goes to stderr and machine parsable output to stdout."

func miscFlags() *flag.FlagSet {
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)

	return flags
}

func Run(args []string, stdErrLogger logging.ILogger) {
	var help, projTest, verbose, trace, includeTests bool
	var descFile, userFilterFile string

	// Parse flags for the "misc" operation
	flags := miscFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	// Main flags to perform particular function
	flags.BoolVar(&projTest, "p", false, "Search for .perpetual dir, starting from curdir and check json configs inside it. Output full path of .perpetual dir on success.")
	// Extra options, may be used with flags above to alter its behavior or test some more things
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

	//TODO: check for flag conflicts between: -p, ...
	stdErrLogger.Debugln("Starting 'misc' operation")
	stdErrLogger.Traceln("Args:", args)

	if !projTest || help {
		usage.PrintOperationUsage("One of the following flags must be provided: -p, ...", flags)
	}

	// Initialize: detect work directories, load .env file with LLM settings, load file filtering regexps
	_, perpetualDir, err := utils.FindProjectRoot(stdErrLogger)
	if err != nil {
		stdErrLogger.Panicln("Error finding project root directory:", err)
	}

	//load json config files for project and operations, will panic if it cannot be loaded or parsed
	config.LoadProjectConfig(perpetualDir, stdErrLogger)
	config.LoadOpAnnotateConfig(perpetualDir, stdErrLogger)
	config.LoadOpDocConfig(perpetualDir, stdErrLogger)
	config.LoadOpExplainConfig(perpetualDir, stdErrLogger)
	config.LoadOpImplementConfig(perpetualDir, stdErrLogger)
	config.LoadOpReportConfig(perpetualDir, stdErrLogger)

	//test load of project description file
	wrn := ""
	if descFile == "" {
		_, wrn, err = utils.LoadTextFile(filepath.Join(perpetualDir, config.ProjectDescriptionFile))
		if err != nil {
			if os.IsNotExist(err) {
				stdErrLogger.Infoln("Not loading missing project description file (description.md)")
			} else {
				stdErrLogger.Panicln("Failed to load project description file:", err)
			}
		}
		if wrn != "" {
			stdErrLogger.Warnf("%s: %s", config.ProjectDescriptionFile, wrn)
		}
	} else if strings.ToLower(descFile) != "disabled" {
		_, wrn, err = utils.LoadTextFile(descFile)
		if err != nil {
			stdErrLogger.Panicln("Failed to load project description file:", err)
		}
		if wrn != "" {
			stdErrLogger.Warnf("%s: %s", descFile, wrn)
		}
	} else {
		stdErrLogger.Infoln("Loading of project description file (description.md) is disabled")
	}

	//if we are only checking for .perpetual directory validity, output it here
	if projTest {
		fmt.Println(perpetualDir)
		return
	}
}
