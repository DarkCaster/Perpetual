package op_init

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const OpName = "init"
const OpDesc = "Initialize new .perpetual directory, will store project configuration"

func initFlags() *flag.FlagSet {
	return flag.NewFlagSet(OpName, flag.ExitOnError)
}

func Run(args []string, logger logging.ILogger) {
	lang := ""
	var help, verbose, trace bool
	// Parse flags for the "init" operation
	initFlags := initFlags()
	initFlags.BoolVar(&help, "h", false, "Show usage")
	initFlags.StringVar(&lang, "l", "", "Select programming language for setting up default LLM prompts (valid values: go|dotnetfw|bash|python3|vb6|...)")
	initFlags.BoolVar(&verbose, "v", false, "Enable debug logging")
	initFlags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	initFlags.Parse(args)

	if verbose {
		logger.SetLevel(logging.DebugLevel)
	}
	if trace {
		logger.SetLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'init' operation")
	logger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", initFlags)
	}

	if lang == "" {
		usage.PrintOperationUsage("language must be provided", initFlags)
	}

	logger.Debugln("Parsed language:", lang)

	cwd, err := os.Getwd()
	if err != nil {
		logger.Panicln("Error getting current working directory:", err)
	}

	perpetualDir := filepath.Join(cwd, ".perpetual")
	_, err = os.Stat(perpetualDir)
	if err == nil {
		logger.Warnln("Directory .perpetual already exists")
	} else if !os.IsNotExist(err) {
		logger.Panicln("Error checking for .perpetual directory:", err)
	} else {
		logger.Traceln("Creating .perpetual directory")
		err = os.Mkdir(perpetualDir, 0755)
		if err != nil {
			logger.Panicln("Error creating .perpetual directory:", err)
		}
	}

	// Create config directory at globalConfigDir with all leading directories with default permissions
	globalConfigDir, err := utils.FindConfigDir()
	if err != nil {
		logger.Panicln("Error finding perpetual config directory:", err)
	}
	logger.Traceln("Creating global config directory")
	err = os.MkdirAll(globalConfigDir, 0755)
	if err != nil {
		logger.Panicln("Error creating global config directory:", err)
	}

	// Save default global config file if missing
	globalConfigText := "# This is global .env file for Perpetual, values declared here have the lowest priority and will be used last"
	globalConfigFile := filepath.Join(globalConfigDir, utils.DotEnvFileName)
	if _, err := os.Stat(globalConfigFile); os.IsNotExist(err) {
		logger.Traceln("Creating default global config file")
		err = utils.SaveTextFile(globalConfigFile, globalConfigText)
		if err != nil {
			logger.Panicln("Error creating default global config file:", err)
		}
	}

	// Create a .gitignore file in the .perpetual directory
	logger.Traceln("Creating .gitignore file")

	gitignoreText := fmt.Sprintf("/%s\n/%s\n/%s\n/%s\n", utils.DotEnvFileName, utils.AnnotationsFileName, llm.LLMRawLogFile, utils.StashesDirName)
	err = utils.SaveTextFile(filepath.Join(perpetualDir, ".gitignore"), gitignoreText)
	if err != nil {
		logger.Panicln("Error creating .gitignore file:", err)
	}

	logger.Traceln("Creating env example file")
	err = utils.SaveTextFile(filepath.Join(perpetualDir, DotEnvExampleFileName), DotEnvExample)
	if err != nil {
		logger.Panicln("Error creating env example file:", err)
	}

	// Create a prompt-files based on the selected language
	logger.Debugln("Creating prompt-files")
	promptsObj, err := newPrompts(lang)
	if err != nil {
		usage.PrintOperationUsage(err.Error(), initFlags)
	}

	saveConfig := func(filePath string, v any) {
		logger.Traceln("Saving config:", filePath)
		err = utils.SaveJsonFile(filepath.Join(perpetualDir, filePath), v)
		if err != nil {
			logger.Panicln(err)
		}
	}

	// Save operation-config files
	saveConfig(config.OpAnnotateConfigFile, promptsObj.GetAnnotateConfig())
	saveConfig(config.OpImplementConfigFile, promptsObj.GetImplementConfig())
	saveConfig(config.OpDocConfigFile, promptsObj.GetDocConfig())

	// Save project-config file
	saveConfig(config.ProjectConfigFile, promptsObj.GetProjectConfig())
}
