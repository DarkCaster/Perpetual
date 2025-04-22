package op_embed

import (
	"flag"
	"path/filepath"
	"sort"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const OpName = "embed"
const OpDesc = "Generate embeddings for project files to enable semantic search"

func embedFlags() *flag.FlagSet {
	return flag.NewFlagSet(OpName, flag.ExitOnError)
}

func Run(args []string, innerCall bool, logger, stdErrLogger logging.ILogger) {
	var help, force, verbose, trace, includeTests bool
	var requestedFile, userFilterFile string

	flags := embedFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.BoolVar(&force, "f", false, "Force regeneration of embeddings for all files, even if they are up to date")
	flags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from processing")
	flags.StringVar(&requestedFile, "r", "", "Only generate embeddings for single file provided with this flag, even if its embedding is already up to date (implies -f flag)")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if verbose {
		logger.EnableLevel(logging.DebugLevel)
	}
	if trace {
		logger.EnableLevel(logging.DebugLevel)
		logger.EnableLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'embed' operation")
	logger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", flags)
	}

	outerCallLogger := logger.Clone()
	if innerCall {
		outerCallLogger.DisableLevel(logging.ErrorLevel)
		outerCallLogger.DisableLevel(logging.WarnLevel)
		outerCallLogger.DisableLevel(logging.InfoLevel)
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
	logger.Infoln("Fetching project files")
	fileNames, _, err := utils.GetProjectFileList(
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

	// Calculate file checksums
	logger.Infoln("Calculating checksums for project files")
	fileChecksums, err := utils.CalculateFilesChecksums(projectRootDir, fileNames)
	if err != nil {
		logger.Panicln("Error getting project-files checksums:", err)
	}

	embeddingsFilePath := filepath.Join(perpetualDir, utils.EmbeddingsFileName)

	var filesToEmbed []string
	if !force {
		filesToEmbed, err = utils.GetChangedEmbeddings(embeddingsFilePath, fileChecksums)
		if err != nil {
			logger.Panicln("error getting changed files:", err)
		}
	} else {
		if requestedFile != "" {
			// Check if requested file is within fileNames array
			requestedFile, err := utils.MakePathRelative(projectRootDir, requestedFile, false)
			if err != nil {
				logger.Panicln("Requested file is not inside project root", requestedFile)
			}
			requestedFile, found := utils.CaseInsensitiveFileSearch(requestedFile, fileNames)
			if !found {
				logger.Panicln("Requested file not found in project")
			}
			filesToEmbed = utils.NewSlice(requestedFile)
		} else {
			filesToEmbed = utils.NewSlice(fileNames...)
			sort.Strings(filesToEmbed)
		}
	}

	//STRIP THE REST to separate funcsion

	// Create llm connector for generating embeddings
	connector, err := llm.NewLLMConnector(OpName, "", "",
		projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
		map[string]interface{}{}, "", "",
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		logger.Panicln("Failed to create LLM connector:", err)
	}

	// TODO: Generate embeddings for each file
	// This would involve:
	// 1. Reading file contents
	// 2. Generating embeddings (using a vector embedding model)
	// 3. Storing embeddings in a structured format

	logger.Infoln("Generating embeddings, file count:", len(filesToEmbed))

	if len(filesToEmbed) > 0 && connector.GetVariantCount() <= 1 {
		logger.Infoln(connector.GetDebugString())
	}

	// Placeholder for embeddings generation
	// In a real implementation, this would call an embedding service or library
	embeddings := make(map[string][]float32)
	for _, file := range filesToEmbed {
		logger.Debugln("Processing file:", file)

		// Read file contents
		/*fileContent*/
		_, err := utils.LoadTextFile(filepath.Join(projectRootDir, file))
		if err != nil {
			logger.Errorf("Failed to read file %s: %s", file, err)
			continue
		}

		// Get file annotation
		//annotation := annotations[file]

		// Combine content and annotation for better semantic representation
		//combinedText := annotation + "\n" + fileContent

		// TODO: Generate actual embeddings
		// This is a placeholder - in a real implementation, you would call an embedding model
		// embeddings[file] = generateEmbedding(combinedText)

		// For now, just create a dummy embedding
		dummyEmbedding := make([]float32, 5)
		embeddings[file] = dummyEmbedding
	}

	// TODO: Save embeddings to file
	logger.Infoln("Saving embeddings to", embeddingsFilePath)

	// Placeholder for saving embeddings
	// In a real implementation, this would serialize the embeddings to JSON
	// and save them to the specified file

	logger.Infoln("Embedding generation complete")
}
