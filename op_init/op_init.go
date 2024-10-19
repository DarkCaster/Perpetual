package op_init

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/prompts"
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

	// Create a prompts directory (if it doesn't exist)
	promptsDir := filepath.Join(perpetualDir, prompts.PromptsDir)
	_, err = os.Stat(promptsDir)
	if os.IsNotExist(err) {
		logger.Traceln("Creating prompts directory")
		err = os.Mkdir(promptsDir, 0755)
		if err != nil {
			logger.Panicln("Error creating prompts directory:", err)
		}
	} else if err != nil {
		logger.Panicln("Error checking prompts directory:", err)
	}

	// Create a prompt-files based on the selected language
	logger.Debugln("Creating prompt-files")
	promptsObj, err := prompts.NewPrompts(lang)
	if err != nil {
		usage.PrintOperationUsage(err.Error(), initFlags)
	}

	savePrompt := func(filePath string, prompt string) {
		logger.Traceln("Saving prompt:", filePath)
		err = utils.SaveTextFile(filepath.Join(promptsDir, filePath), prompt)
		if err != nil {
			logger.Panicln(err)
		}
	}

	// Save the system prompt to a file
	savePrompt(prompts.SystemPromptFile, promptsObj.GetSystemPrompt())

	// Save annotate-operation prompts
	logger.Traceln("Saving prompt:", prompts.AnnotatePromptFile)
	err = utils.SaveJsonFile(filepath.Join(promptsDir, prompts.AnnotatePromptFile), promptsObj.GetAnnotatePrompt())
	if err != nil {
		logger.Panicln("error writing prompt-JSON file: ", err)
	}
	savePrompt(prompts.AIAnnotateResponseFile, promptsObj.GetAIAnnotateResponse())

	// Save implement-operation stage1 prompts
	savePrompt(prompts.ImplementStage1ProjectIndexPromptFile, promptsObj.GetImplementStage1ProjectIndexPrompt())
	savePrompt(prompts.AIImplementStage1ProjectIndexResponseFile, promptsObj.GetAIImplementStage1ProjectIndexResponse())
	savePrompt(prompts.ImplementStage1SourceAnalysisPromptFile, promptsObj.GetImplementStage1SourceAnalysisPrompt())

	// Save implement-operation stage2 prompts
	savePrompt(prompts.ImplementStage2ProjectCodePromptFile, promptsObj.GetImplementStage2ProjectCodePrompt())
	savePrompt(prompts.AIImplementStage2ProjectCodeResponseFile, promptsObj.GetAIImplementStage2ProjectCodeResponse())
	savePrompt(prompts.ImplementStage2FilesToChangePromptFile, promptsObj.GetImplementStage2FilesToChangePrompt())
	savePrompt(prompts.ImplementStage2FilesToChangeExtendedPromptFile, promptsObj.GetImplementStage2FilesToChangeExtendedPrompt())

	// Save implement-operation stage2 no-planning prompts
	savePrompt(prompts.ImplementStage2NoPlanningPromptFile, promptsObj.GetImplementStage2NoPlanningPrompt())
	savePrompt(prompts.AIImplementStage2NoPlanningResponseFile, promptsObj.GetAIImplementStage2NoPlanningResponse())

	// Save implement-operation stage3 prompts
	savePrompt(prompts.ImplementStage3ChangesDonePromptFile, promptsObj.GetImplementStage3ChangesDonePrompt())
	savePrompt(prompts.AIImplementStage3ChangesDoneResponseFile, promptsObj.GetAIImplementStage3ChangesDoneResponse())
	savePrompt(prompts.ImplementStage3ProcessFilePromptFile, promptsObj.GetImplementStage3ProcessFilePrompt())
	savePrompt(prompts.ImplementStage3ContinuePromptFile, promptsObj.GetImplementStage3ContinuePrompt())

	// Save doc-operation prompts
	savePrompt(prompts.DocProjectIndexPromptFile, promptsObj.GetDocProjectIndexPrompt())
	savePrompt(prompts.AIDocProjectIndexResponseFile, promptsObj.GetAIDocProjectIndexResponse())
	savePrompt(prompts.DocProjectCodePromptFile, promptsObj.GetDocProjectCodePrompt())
	savePrompt(prompts.AIDocProjectCodeResponseFile, promptsObj.GetAIDocProjectCodeResponse())
	savePrompt(prompts.DocExamplePromptFile, promptsObj.GetDocExamplePrompt())
	savePrompt(prompts.AIDocExampleResponseFile, promptsObj.GetAIDocExampleResponse())
	savePrompt(prompts.DocStage1WritePromptFile, promptsObj.GetDocStage1WritePrompt())
	savePrompt(prompts.DocStage1RefinePromptFile, promptsObj.GetDocStage1RefinePrompt())
	savePrompt(prompts.DocStage2WritePromptFile, promptsObj.GetDocStage2WritePrompt())
	savePrompt(prompts.DocStage2RefinePromptFile, promptsObj.GetDocStage2RefinePrompt())
	savePrompt(prompts.DocStage2ContinuePromptFile, promptsObj.GetDocStage2ContinuePrompt())

	// Save project files search white-list regexps to a json
	logger.Debugln("Creating helper regexps and tags definitions")
	saveJson := func(filePath string, v any) {
		logger.Traceln("Saving json:", filePath)
		err = utils.SaveJsonFile(filepath.Join(perpetualDir, filePath), v)
		if err != nil {
			logger.Panicln(err)
		}
	}

	logger.Traceln("Saving project files whitelist regexps")
	saveJson(prompts.ProjectFilesWhitelistFileName, promptsObj.GetProjectFilesWhitelist())
	saveJson(prompts.ProjectFilesToMarkdownLangMappingFileName, promptsObj.GetProjectFilesToMarkdownMappings())

	logger.Traceln("Saving project files blacklist regexps")
	saveJson(prompts.ProjectFilesBlacklistFileName, promptsObj.GetProjectFilesBlacklist())

	logger.Traceln("Saving implement-operation comment regexps")
	saveJson(prompts.OpImplementCommentRXFileName, promptsObj.GetImplementCommentRegexps())

	logger.Traceln("Saving no-upload comment regexps")
	saveJson(prompts.NoUploadCommentRXFileName, promptsObj.GetNoUploadCommentRegexps())

	logger.Traceln("Saving file-name tags regexps")
	saveJson(prompts.FileNameTagsRXFileName, promptsObj.GetFileNameTagsRegexps())

	logger.Traceln("Saving file-name tags")
	saveJson(prompts.FileNameTagsFileName, promptsObj.GetFileNameTags())

	logger.Traceln("Saving file-name embed regexp")
	saveJson(prompts.FileNameEmbedRXFileName, promptsObj.GetFileNameEmbedRegex())

	logger.Traceln("Saving output-tags regexps")
	saveJson(prompts.OutputTagsRXFileName, promptsObj.GetOutputTagsRegexps())

	logger.Traceln("Saving reasonings-tags regexps")
	saveJson(prompts.ReasoningsTagsRXFileName, promptsObj.GetReasoningsTagsRegexps())

	logger.Traceln("Saving reasonings-tags")
	saveJson(prompts.ReasoningsTagsFileName, promptsObj.GetReasoningsTags())
}
