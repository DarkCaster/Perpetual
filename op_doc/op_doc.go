package op_doc

import (
	"flag"
	"path/filepath"
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_annotate"
	"github.com/DarkCaster/Perpetual/op_stash"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const OpName = "doc"
const OpDesc = "Create or rework documentation files (in markdown or plain-text format)"

func docFlags() *flag.FlagSet {
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)
	return flags
}

func Run(args []string, logger logging.ILogger) {
	var help, salvageFiles, verbose, trace, noAnnotate, forceUpload, includeTests bool
	var docFile, docExample, action, userFilterFile string

	flags := docFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.BoolVar(&noAnnotate, "n", false, "No annotate mode: skip re-annotating of changed files and use current annotations if any")
	flags.StringVar(&docFile, "r", "", "Target documentation file for processing")
	flags.StringVar(&docExample, "e", "", "Optional documentation file to use as an example/reference for style, structure and format, but not for content")
	flags.StringVar(&action, "a", "write", "Select action to perform (valid values: draft|write|refine)")
	flags.BoolVar(&forceUpload, "f", false, "Disable 'no-upload' file-filter and upload such files for review if reqested")
	flags.BoolVar(&salvageFiles, "s", false, "Try to salvage incorrect filenames on stage 1. Experimental, use in projects with a large number of files where LLM tends to make more mistakes when generating list of files to analyze")
	flags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from processing")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing")
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

	// Make main documentation file relative to the project root
	docFile, err = utils.MakePathRelative(projectRootDir, docFile, false)
	if err != nil {
		logger.Panicln("Requested file is not inside project root", docFile)
	}

	if docExample != "" {
		// Make example relative to the project root
		docExample, err = utils.MakePathRelative(projectRootDir, docExample, false)
		if err != nil {
			logger.Panicln("Example file is not inside project root", docExample)
		}
		// Try reading example to ensure it presence and it is a correct text file with valid encoding
		if _, err = utils.LoadTextFile(filepath.Join(projectRootDir, docExample)); err != nil {
			logger.Panicln("Failed to load example document:", err)
		}
	}

	// Try reading document to ensure it presence and it is a correct text file with valid encoding
	var docFiles []string
	if _, err = utils.LoadTextFile(filepath.Join(projectRootDir, docFile)); err == nil {
		docFiles = append(docFiles, docFile)
	} else if action != "DRAFT" {
		logger.Panicln("Failed to load target document:", err)
	}

	var docContent string
	if action == "DRAFT" {
		docContent = MDDocDraft
	} else {
		if !noAnnotate {
			logger.Debugln("Running 'annotate' operation to update file annotations")
			op_annotate.Run(nil, logger)
		}

		docConfig, err := config.LoadOpDocConfig(perpetualDir)
		if err != nil {
			logger.Panicf("Error loading op_doc config: %s", err)
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

		// Load annotations needed for stage1
		annotations, err := utils.GetAnnotations(filepath.Join(perpetualDir, utils.AnnotationsFileName), fileChecksums)
		if err != nil {
			logger.Panicln("Error reading annotations:", err)
		}

		// Run stage1 to find out what project-files contents we need to work on document
		requestedFiles := Stage1(projectRootDir,
			perpetualDir,
			docConfig,
			projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
			fileNames,
			annotations,
			docFile,
			docExample,
			action,
			salvageFiles,
			logger)

		// Check requested files for no-upload mark and filter it out
		var filteredRequestedFiles []string
		if forceUpload {
			filteredRequestedFiles = requestedFiles
		} else {
			for _, file := range requestedFiles {
				if found, err := utils.FindInRelativeFile(
					projectRootDir,
					file,
					docConfig.RegexpArray(config.K_NoUploadCommentsRx)); err == nil && !found {
					filteredRequestedFiles = append(filteredRequestedFiles, file)
				} else if found {
					logger.Warnln("Skipping file marked with 'no-upload' comment:", file)
				} else {
					logger.Errorln("Error searching for 'no-upload' comment in file:", file, err)
				}
			}
		}

		// Run stage2 to make changes to the document and save it to docContent
		docContent = Stage2(projectRootDir,
			perpetualDir,
			docConfig,
			projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
			fileNames,
			filteredRequestedFiles,
			annotations,
			docFile,
			docExample,
			action,
			logger)
	}

	// Add extra newline if not present
	if !strings.HasSuffix(docContent, "\n") {
		docContent += "\n"
	}

	docResults := make(map[string]string)
	docResults[docFile] = docContent

	// Create and apply stash from generated results
	newStashFileName := op_stash.CreateStash(docResults, docFiles, logger)
	op_stash.Run([]string{"-a", "-n", newStashFileName}, logger)
}
