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
	savePrompt(prompts.AnnotatePromptFile, promptsObj.GetAnnotatePrompt())
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

	// Save project files search white-list regexps to a json
	logger.Debugln("Creating helper regexps and tags definitions")
	logger.Traceln("Saving project files whitelist regexps")
	err = utils.SaveJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesWhitelistFileName), promptsObj.GetProjectFilesWhitelist())
	if err != nil {
		logger.Panicln("error writing project files whitelist regexps to JSON file: ", err)
	}
	err = utils.SaveJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesToMarkdownLangMappingFileName), promptsObj.GetProjectFilesToMarkdownMappings())
	if err != nil {
		logger.Panicln("error writing project-files to markdown-lang mapping regexps to JSON file: ", err)
	}
	logger.Traceln("Saving project files blacklist regexps")
	err = utils.SaveJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesBlacklistFileName), promptsObj.GetProjectFilesBlacklist())
	if err != nil {
		logger.Panicln("error writing project files blacklist regexps to JSON file: ", err)
	}
	// Save implement-operation comment search regexps to a json
	logger.Traceln("Saving implement-operation comment regexps")
	err = utils.SaveJsonFile(filepath.Join(perpetualDir, prompts.OpImplementCommentRXFileName), promptsObj.GetImplementCommentRegexps())
	if err != nil {
		logger.Panicln("error writing implement-operation comment regexps to JSON file: ", err)
	}
	// Save no-upload comment regexps to a json
	logger.Traceln("Saving no-upload comment regexps")
	err = utils.SaveJsonFile(filepath.Join(perpetualDir, prompts.NoUploadCommentRXFileName), promptsObj.GetNoUploadCommentRegexps())
	if err != nil {
		logger.Panicln("error writing noupload comment regexps to JSON file: ", err)
	}
	// Save file-name tags regexps to a json
	logger.Traceln("Saving file-name tags regexps")
	err = utils.SaveJsonFile(filepath.Join(perpetualDir, prompts.FileNameTagsRXFileName), promptsObj.GetFileNameTagsRegexps())
	if err != nil {
		logger.Panicln("error writing file-name tags regexps to JSON file: ", err)
	}
	// Save file-name tags to a json
	logger.Traceln("Saving file-name tags")
	err = utils.SaveJsonFile(filepath.Join(perpetualDir, prompts.FileNameTagsFileName), promptsObj.GetFileNameTags())
	if err != nil {
		logger.Panicln("error writing file-name tags to JSON file: ", err)
	}
	// Save file-name embed regexp
	logger.Traceln("Saving file-name embed regexp")
	err = utils.SaveJsonFile(filepath.Join(perpetualDir, prompts.FileNameEmbedRXFileName), promptsObj.GetFileNameEmbedRegex())
	if err != nil {
		logger.Panicln("error writing file-name embed regexp JSON file: ", err)
	}
	// Save output tags to a json
	logger.Traceln("Saving output-tags regexps")
	err = utils.SaveJsonFile(filepath.Join(perpetualDir, prompts.OutputTagsRXFileName), promptsObj.GetOutputTagsRegexps())
	if err != nil {
		logger.Panicln("error writing output tags to JSON file: ", err)
	}
	// Save reasonings tags regexps to a json
	logger.Traceln("Saving reasonings-tags regexps")
	err = utils.SaveJsonFile(filepath.Join(perpetualDir, prompts.ReasoningsTagsRXFileName), promptsObj.GetReasoningsTagsRegexps())
	if err != nil {
		logger.Panicln("error writing reasonings tags regexps to JSON file: ", err)
	}
	// Save reasonings tags to a json
	logger.Traceln("Saving reasonings-tags")
	err = utils.SaveJsonFile(filepath.Join(perpetualDir, prompts.ReasoningsTagsFileName), promptsObj.GetReasoningsTags())
	if err != nil {
		logger.Panicln("error writing reasonings tags to JSON file: ", err)
	}
}
