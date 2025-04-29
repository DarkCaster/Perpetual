package op_embed

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	var help, force, dryRun, verbose, trace bool
	var requestedFile, userFilterFile string

	flags := embedFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.BoolVar(&force, "f", false, "Force regeneration of embeddings for all files, even if they are up to date")
	flags.BoolVar(&dryRun, "d", false, "Perform a dry run without actually generating embeddings, list of files that will be processed")
	flags.StringVar(&requestedFile, "r", "", "Only generate embeddings for single file provided with this flag, even if its embedding is already up to date (implies -f flag)")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if dryRun {
		logger = stdErrLogger
	}
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

	projectRootDir, perpetualDir, err := utils.FindProjectRoot(outerCallLogger)
	if err != nil {
		logger.Panicln("Error finding project root directory:", err)
	}

	globalConfigDir, err := utils.FindConfigDir()
	if err != nil {
		logger.Panicln("Error finding perpetual config directory:", err)
	}

	outerCallLogger.Infoln("Project root directory:", projectRootDir)
	outerCallLogger.Debugln("Perpetual directory:", perpetualDir)

	if innerCall {
		logger.Debugln("Not re-loading env files for inner call of embed operation")
	} else {
		utils.LoadEnvFiles(logger, filepath.Join(perpetualDir, utils.DotEnvFileName), filepath.Join(globalConfigDir, utils.DotEnvFileName))
	}

	projectConfig, err := config.LoadProjectConfig(perpetualDir)
	if err != nil {
		logger.Panicf("Error loading project config: %s", err)
	}

	// Create llm connector for generating embeddings early
	// So we can stop right here if embeddings not supported or disabled
	connector, err := llm.NewLLMConnector(OpName, "", "",
		projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
		map[string]interface{}{}, "", "",
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		if innerCall {
			logger.Debugln("Failed to create LLM connector for embed operation:", err)
			logger.Infoln("Embedding model is not configured or not available for selected LLM provider")
			return
		} else {
			logger.Panicln("Failed to create LLM connector:", err)
		}
	}

	var userBlacklist []*regexp.Regexp
	if userFilterFile != "" {
		userBlacklist, err = utils.AppendUserFilterFromFile(userFilterFile, userBlacklist)
		if err != nil {
			logger.Panicln("Error processing user blacklist-filter:", err)
		}
	}

	// Preparation of project files
	outerCallLogger.Infoln("Fetching project files")
	fileNames, _, err := utils.GetProjectFileList(
		projectRootDir,
		perpetualDir,
		projectConfig.RegexpArray(config.K_ProjectFilesWhitelist),
		projectConfig.RegexpArray(config.K_ProjectFilesBlacklist))

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

	logger.Infoln("Calculating checksums for project files")
	fileChecksums, err := utils.CalculateFilesChecksums(projectRootDir, fileNames)
	if err != nil {
		logger.Panicln("Error getting project-files checksums:", err)
	}

	embeddingsFilePath := filepath.Join(perpetualDir, utils.EmbeddingsFileName)

	//load old embeddings file
	logger.Traceln("Loading embeddings")
	embeddings, oldChecksums, vectorDimensions, err := utils.GetEmbeddings(embeddingsFilePath, fileNames)
	if err != nil {
		logger.Panicln("Failed to load embeddings:", err)
	}
	logger.Traceln("Done loading embeddings")

	var filesToEmbed []string
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
	} else if force {
		logger.Debugln("Removing obsolete embeddings storage file:", embeddingsFilePath)
		if err = utils.RemoveFile(embeddingsFilePath); err != nil {
			logger.Panicln("Failed to remove obsolete embeddings storage file:", err)
		}
		filesToEmbed = utils.NewSlice(fileNames...)
		sort.Strings(filesToEmbed)
		//clear-out previously loaded embeddings
		embeddings = make(map[string][][]float32)
		vectorDimensions = 0
	} else {
		logger.Traceln("Detecting changes")
		filesToEmbed = utils.GetChangedFiles(oldChecksums, fileChecksums)
		logger.Traceln("Done detecting changes")
	}

	if vectorDimensions > 0 {
		logger.Infoln("Vectors dimensions detected from existing embeddings:", vectorDimensions)
	}

	if vectorDimensions < 0 {
		logger.Panicln("Vectors dimensions inconsistency detected for existing embeddings, check your LLM embeddings configuration and rebuild all embeddings by running embed operation with -f flag")
	}

	//filter with user-blacklist, revert checksum for dropped files, so they can be reevaluated next time
	filesToEmbed, droppedFiles := utils.FilterFilesWithBlacklist(filesToEmbed, userBlacklist)
	if len(droppedFiles) > 0 {
		logger.Infoln("Number of files to embed, filtered by user-provided blacklist:", len(droppedFiles))
	}
	for _, file := range droppedFiles {
		fileChecksums[file] = oldChecksums[file]
		logger.Debugln("Filtered-out:", file)
	}

	if dryRun {
		logger.Infoln("Files to embed:")
		for _, file := range filesToEmbed {
			fmt.Println(file)
		}
		os.Exit(0)
	}

	logger.Infoln("Generating embeddings, file count:", len(filesToEmbed))

	if len(filesToEmbed) > 0 {
		logger.Infoln(connector.GetDebugString())

		// only rotate logfile for outer call if we have files to proceed
		if !innerCall {
			logger.Debugln("Rotating log file")
			if err := llm.RotateLLMRawLogFile(perpetualDir); err != nil {
				logger.Panicln("Failed to rotate log file:", err)
			}
		}
	}

	errorFlag := false
	changedFlag := false
	for i, filePath := range filesToEmbed {
		// Read file contents and generate embedding
		fileBytes, err := utils.LoadTextFile(filepath.Join(projectRootDir, filePath))
		if err != nil {
			logger.Errorf("Failed to read file %s: %s", filePath, err)
			continue
		}
		fileContents := string(fileBytes)
		onFailRetriesLeft := max(connector.GetOnFailureRetryLimit(), 1)
		for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
			logger.Infof("%d: %s", i+1, filePath)
			vectors, status, err := connector.CreateEmbeddings(llm.DocEmbed, fmt.Sprintf("file:%s", filePath), fileContents)
			// Check for general error on query
			if err != nil {
				logger.Errorf("LLM query failed with status %d, error: %s", status, err)
				if status == llm.QueryInitFailed || onFailRetriesLeft < 1 {
					fileChecksums[filePath] = oldChecksums[filePath]
					errorFlag = true
					break
				}
				continue
			}
			// Check for hitting token limit, ideally should not occur at all for embedding models
			if status == llm.QueryMaxTokens {
				logger.Errorln("LLM response(s) reached max tokens, that's probably an error with configuration of embedding model")
				// do not retry, go to the next file
				fileChecksums[filePath] = oldChecksums[filePath]
				errorFlag = true
				break
			}

			// Check vector dimensions consistency
			if vectorDimensions == 0 && len(vectors) > 0 && len(vectors[0]) > 0 {
				// This is the first valid vector, set the dimensions
				vectorDimensions = len(vectors[0])
				logger.Infoln("Vectors dimensions detected:", vectorDimensions)
			}

			if vectorDimensions > 0 && len(vectors) > 0 {
				for _, vector := range vectors {
					if len(vector) != vectorDimensions {
						logger.Panicf(
							"Vector dimensions mismatch for file %s: expected %d, got %d, please check your LLM configuration and rebuild all embeddings if needed by running embed operation with -f flag",
							filePath,
							vectorDimensions,
							len(vector))
					}
				}
			}

			embeddings[filePath] = vectors
			changedFlag = true
			break
		}
	}

	// Save updated embeddings
	if changedFlag {
		logger.Infoln("Saving embeddings")
		if err := utils.SaveEmbeddings(embeddingsFilePath, fileChecksums, embeddings); err != nil {
			logger.Panicln("Failed to save embeddings:", err)
		}
		logger.Traceln("Done saving embeddings")
	} else {
		logger.Infoln("Embeddings unchanged")
	}

	if errorFlag {
		logger.Panicln("Not all files were successfully processed. Run embed again to process failed files.")
	}
}

func GenerateEmbeddings(mode llm.EmbedMode, tag, content string, logger logging.ILogger) ([][]float32, error) {
	logger.Debugln("Running GenerateVectors")

	silentLogger := logger.Clone()
	silentLogger.DisableLevel(logging.ErrorLevel)
	silentLogger.DisableLevel(logging.WarnLevel)
	silentLogger.DisableLevel(logging.InfoLevel)

	_, perpetualDir, err := utils.FindProjectRoot(silentLogger)
	if err != nil {
		logger.Panicln("Error finding project root directory:", err)
	}

	projectConfig, err := config.LoadProjectConfig(perpetualDir)
	if err != nil {
		logger.Panicf("Error loading project config: %s", err)
	}

	// Create llm connector for generating embeddings
	connector, err := llm.NewLLMConnector(OpName, "", "",
		projectConfig.StringArray2D(config.K_ProjectMdCodeMappings),
		map[string]interface{}{}, "", "",
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		return [][]float32{}, err
	}

	switch mode {
	case llm.DocEmbed:
		logger.Infoln("Generating document embeddings for:", tag)
	case llm.SearchEmbed:
		logger.Infoln("Generating search query embeddings for:", tag)
	default:
		logger.Infoln("Generating embeddings for:", tag)
	}

	logger.Infoln(connector.GetDebugString())

	onFailRetriesLeft := max(connector.GetOnFailureRetryLimit(), 1)
	for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
		vectors, status, err := connector.CreateEmbeddings(mode, tag, content)
		// Check for general error on query
		if err != nil {
			logger.Errorf("LLM query failed with status %d, error: %s", status, err)
			if status == llm.QueryInitFailed || onFailRetriesLeft < 1 {
				return [][]float32{}, err
			}
			continue
		}
		// Check for hitting token limit, ideally should not occur at all for embedding models
		if status == llm.QueryMaxTokens {
			err := "LLM response(s) reached max tokens, that's probably an error with configuration of embedding model"
			logger.Errorf(err)
			return [][]float32{}, errors.New(err)
		}
		return vectors, nil
	}

	return [][]float32{}, errors.New("unknown error")
}
