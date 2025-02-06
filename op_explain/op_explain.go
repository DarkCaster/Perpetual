package op_explain

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/DarkCaster/Perpetual/config"
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
	projectRootDir, perpetualDir, err := utils.FindProjectRoot(logger)
	if err != nil {
		logger.Panicln("Error finding project root directory:", err)
	}

	logger.Infoln("Project root directory:", projectRootDir)
	logger.Debugln("Perpetual directory:", perpetualDir)

	//explainConfig
	_, err = config.LoadOpExplainConfig(perpetualDir)
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

	if !noAnnotate {
		logger.Debugln("Running 'annotate' operation to update file annotations")
		op_annotate.Run(nil, logger)
	}

	// Load annotations
	//annotations
	_, err = utils.GetAnnotations(filepath.Join(perpetualDir, utils.AnnotationsFileName), fileChecksums)
	if err != nil {
		logger.Panicln("Error loading annotations:", err)
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
