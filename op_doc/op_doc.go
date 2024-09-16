package op_doc

import (
	"flag"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_annotate"
	"github.com/DarkCaster/Perpetual/op_stash"
	"github.com/DarkCaster/Perpetual/prompts"
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
	var help, verbose, trace, noAnnotate bool
	var docFile, action string

	flags := docFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.BoolVar(&noAnnotate, "n", false, "No annotate mode: skip re-annotating of changed files and use current annotations if any")
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

	promptsDir := filepath.Join(perpetualDir, prompts.PromptsDir)
	logger.Debugln("Prompts directory:", promptsDir)

	utils.LoadEnvFiles(logger, filepath.Join(perpetualDir, utils.DotEnvFileName), filepath.Join(globalConfigDir, utils.DotEnvFileName))

	// Make markdownFile relative to project root
	docFile, err = utils.MakePathRelative(projectRootDir, docFile, false)
	if err != nil {
		logger.Panicln("Requested file is not inside project root", docFile)
	}

	// Try reading document to ensure it presence and it is a correct text file with valid encoding
	var docFiles []string
	if _, err = utils.LoadTextFile(filepath.Join(projectRootDir, docFile)); err == nil {
		docFiles = append(docFiles, docFile)
	} else if action != "DRAFT" {
		logger.Panicln("Failed to load document:", err)
	}

	var docContent string
	if action == "DRAFT" {
		docContent = MDDocDraft
	} else {
		if !noAnnotate {
			logger.Debugln("Running 'annotate' operation to update file annotations")
			op_annotate.Run(nil, logger)
		}

		systemPrompt, err := utils.LoadTextFile(filepath.Join(promptsDir, prompts.SystemPromptFile))
		if err != nil {
			logger.Warnln("Failed to read system prompt:", err)
		}

		var filesToMdLangMappings [][2]string
		err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesToMarkdownLangMappingFileName), &filesToMdLangMappings)
		if err != nil {
			logger.Warnln("Error reading optional filename to markdown-lang mappings:", err)
		}

		var projectFilesWhitelist []string
		err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesWhitelistFileName), &projectFilesWhitelist)
		if err != nil {
			logger.Panicln("Error reading project-files whitelist regexps:", err)
		}

		var projectFilesBlacklist []string
		err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.ProjectFilesBlacklistFileName), &projectFilesBlacklist)
		if err != nil {
			logger.Panicln("Error reading project-files blacklist regexps:", err)
		}

		var noUploadRxStrings []string
		err = utils.LoadJsonFile(filepath.Join(perpetualDir, prompts.NoUploadCommentRXFileName), &noUploadRxStrings)
		if err != nil {
			logger.Panicln("Error reading no-upload regexps:", err)
		}

		var noUploadRegexps []*regexp.Regexp
		for _, rx := range noUploadRxStrings {
			crx, err := regexp.Compile(rx)
			if err != nil {
				logger.Panicln("Failed to compile 'no-upload' comment search regexp: ", err)
			}
			noUploadRegexps = append(noUploadRegexps, crx)
		}

		fileNameTagsRxStrings := utils.LoadStringPair(filepath.Join(perpetualDir, prompts.FileNameTagsRXFileName), 2, 2, 2, logger)
		fileNameTags := utils.LoadStringPair(filepath.Join(perpetualDir, prompts.FileNameTagsFileName), 2, 2, 2, logger)

		// Get project files, which names selected with whitelist regexps and filtered with blacklist regexps
		fileChecksums, fileNames, _, err := utils.GetProjectFileList(projectRootDir, perpetualDir, projectFilesWhitelist, projectFilesBlacklist)
		if err != nil {
			logger.Panicln("Error getting project file-list:", err)
		}

		// Load annotations needed for stage1
		annotations, err := utils.GetAnnotations(filepath.Join(perpetualDir, utils.AnnotationsFileName), fileChecksums)
		if err != nil {
			logger.Panicln("Error reading annotations:", err)
		}

		// Run stage1 to find out what project-files contents we need to work on document
		requestedFiles := Stage1(projectRootDir, perpetualDir, promptsDir, systemPrompt, filesToMdLangMappings, fileNameTagsRxStrings, fileNameTags, fileNames, annotations, docFile, action, logger)

		// Check requested files for no-upload mark and filter it out
		var filteredRequestedFiles []string
		for _, file := range requestedFiles {
			if found, err := utils.FindInRelativeFile(projectRootDir, file, noUploadRegexps); err == nil && !found {
				filteredRequestedFiles = append(filteredRequestedFiles, file)
			} else if found {
				logger.Warnln("Skipping file marked with 'no-upload' comment:", file)
			} else {
				logger.Errorln("Error searching for 'no-upload' comment in file:", file, err)
			}
		}

		// Run stage2 to make changes to the document and save it to docContent
		docContent = Stage2(projectRootDir, perpetualDir, promptsDir, systemPrompt, filesToMdLangMappings, fileNameTags, fileNames, filteredRequestedFiles, annotations, docFile, action, logger)
	}

	docResults := make(map[string]string)
	docResults[docFile] = docContent

	// Create and apply stash from generated results
	newStashFileName := op_stash.CreateStash(docResults, docFiles, logger)
	op_stash.Run([]string{"-a", "-n", newStashFileName}, logger)
}
