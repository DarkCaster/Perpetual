package op_doc

import (
	"flag"
	"path/filepath"
	"strings"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_annotate"
	"github.com/DarkCaster/Perpetual/op_stash"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const OpName = "doc"
const OpDesc = "Create or rework markdown docs"

func docFlags() *flag.FlagSet {
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)
	return flags
}

func Run(args []string, logger logging.ILogger) {
	var help, verbose, trace bool
	var docFile, action string

	flags := docFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.StringVar(&docFile, "r", "", "Markdown file for processing")
	flags.StringVar(&action, "a", "write", "Select action to perform (valid values: draft|write|refine)")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if verbose {
		logger.SetLevel(logging.DebugLevel)
	}
	if trace {
		logger.SetLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'doc' operation")
	logger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", flags)
	}

	if docFile == "" {
		logger.Panicln("Markdown file not specified. Use -r flag to specify the file.")
	}

	action = strings.ToUpper(action)
	if action != "DRAFT" && action != "WRITE" && action != "REFINE" {
		logger.Panicln("Invalid action provided")
	}

	// Find project root and perpetual directories
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

	utils.LoadEnvFiles(logger, filepath.Join(perpetualDir, utils.DotEnvFileName), filepath.Join(globalConfigDir, utils.DotEnvFileName))

	// Make markdownFile relative to project root
	docFile, err = utils.MakePathRelative(projectRootDir, docFile, false)
	if err != nil {
		logger.Panicln("Requested file is not inside project root", docFile)
	}

	docResults := make(map[string]string)
	var docFiles []string

	// Check if docFile exists, try to read its initial content
	var docContent string
	fullPath := filepath.Join(projectRootDir, docFile)
	if docContent, err = utils.LoadTextFile(fullPath); err == nil {
		docFiles = append(docFiles, docFile)
	}

	if strings.ToUpper(action) == "DRAFT" {
		docContent = MDDocDraft
	} else {
		logger.Debugln("Running 'annotate' operation to update file annotations")
		op_annotate.Run(nil, logger)

	}

	docResults[docFile] = docContent

	// Create and apply stash from generated results
	newStashFileName := op_stash.CreateStash(docResults, docFiles, logger)
	op_stash.Run([]string{"-a", "-n", newStashFileName}, logger)
}
