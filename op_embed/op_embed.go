package op_embed

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

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
	var help, force, dryRun, verbose, trace, readQuestion, includeTests bool
	var inputFile, requestedFile, userFilterFile string
	var searchLimit int

	flags := embedFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.BoolVar(&force, "f", false, "Force regeneration of embeddings for all files, even if they are up to date")
	flags.BoolVar(&dryRun, "d", false, "Perform a dry run without actually generating embeddings, list of files that will be processed")
	flags.StringVar(&requestedFile, "r", "", "Only generate embeddings for single file provided with this flag, even if its embedding is already up to date (implies -f flag)")
	flags.BoolVar(&readQuestion, "q", false, "Read question from stdin and find files relevant to it")
	flags.StringVar(&inputFile, "i", "", "Read question from file, plain text or markdown format (implies -q flag)")
	flags.IntVar(&searchLimit, "s", 5, "Limit on the number of files returned that are relevant to the question")
	flags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from processing")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if inputFile != "" {
		readQuestion = true
	}
	if dryRun || readQuestion {
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
		utils.LoadEnvFiles(logger, perpetualDir, globalConfigDir)
	}

	projectConfig := config.LoadProjectConfig(perpetualDir, logger)

	// Create llm connector for generating embeddings early
	// So we can stop right here if embeddings not supported or disabled
	connector, err := llm.NewLLMConnector(OpName, "", "",
		projectConfig.TextMatcherString(config.K_ProjectMdCodeMappings),
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
	if readQuestion && !includeTests {
		userBlacklist = append(userBlacklist, projectConfig.RegexpArray(config.K_ProjectTestFilesBlacklist)...)
	}
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

	// Read input from file or stdin
	var question string
	if readQuestion {
		if inputFile != "" {
			data, wrn, err := utils.LoadTextFile(inputFile)
			if err != nil {
				logger.Panicln("Error reading input file:", err)
			}
			if wrn != "" {
				logger.Warnf("%s: %s", inputFile, wrn)
			}
			question = data
		} else {
			logger.Infoln("Reading question from stdin")
			data, wrn, err := utils.LoadTextStdin()
			if err != nil {
				logger.Panicln("Error reading from stdin:", err)
			}
			if wrn != "" {
				logger.Warnf("stdin: %s", wrn)
			}
			question = string(data)
		}
		// Trim excess line breaks at both sides of question, and stop on empty input
		question = strings.Trim(question, "\n")
		if len(question) < 1 {
			logger.Panicln("Question is empty, cannot continue")
		}
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
		logger.Infoln("Number of files to embed, filtered by blacklists:", len(droppedFiles))
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
		return
	}

	logger.Infoln("Generating embeddings, file count:", len(filesToEmbed))

	if len(filesToEmbed) > 0 {
		// only rotate logfile for outer call if we have files to proceed
		if !innerCall {
			logger.Debugln("Rotating log file")
			if err := llm.RotateLLMRawLogFile(perpetualDir); err != nil {
				logger.Panicln("Failed to rotate log file:", err)
			}
		}
		debugString := connector.GetDebugString()
		logger.Notifyln(debugString)
		llm.GetSimpleRawMessageLogger(perpetualDir)(fmt.Sprintf("=== Embed (files): %s\n\n\n", debugString))
	}

	errorFlag := false
	changedFlag := false
	for i, filePath := range filesToEmbed {
		// Read file contents and generate embedding
		fileBytes, wrn, err := utils.LoadTextFile(filepath.Join(projectRootDir, filePath))
		if err != nil {
			logger.Errorf("Failed to read file %s: %s", filePath, err)
			continue
		}
		if wrn != "" {
			logger.Warnf("%s: %s", filePath, wrn)
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

	if question == "" {
		return
	}

	//file-list for search
	fileNames, droppedFiles = utils.FilterFilesWithBlacklist(fileNames, userBlacklist)
	if len(droppedFiles) > 0 {
		logger.Infoln("Number of files to search, filtered by blacklists:", len(droppedFiles))
	}
	for _, file := range droppedFiles {
		logger.Debugln("Filtered-out:", file)
	}

	results := SimilaritySearchStage(
		SelectAggressive,
		searchLimit,
		perpetualDir,
		[]string{question},
		[]string{"question"},
		fileNames,
		[]string{},
		logger)
	for _, file := range results {
		fmt.Println(file)
	}
}

var searchCacheLock sync.Mutex // Just as precaution for future
var searchVectorsCache map[string][][]float32 = map[string][][]float32{}

func TryReadFromCache(data string) [][]float32 {
	searchCacheLock.Lock()
	defer searchCacheLock.Unlock()
	if result, ok := searchVectorsCache[data]; ok {
		return result
	}
	return nil
}

func ClearCache() {
	searchCacheLock.Lock()
	defer searchCacheLock.Unlock()
	searchVectorsCache = map[string][][]float32{}
}

func WriteToCache(data string, vectors [][]float32) {
	searchCacheLock.Lock()
	defer searchCacheLock.Unlock()
	searchVectorsCache[data] = vectors
}

// Called internally to generate embeddings for local similarity search queries
func generateEmbeddings(tag, content string, logger logging.ILogger) ([][]float32, float32, error) {
	logger.Debugln("Running GenerateVectors")

	silentLogger := logger.Clone()
	silentLogger.DisableLevel(logging.ErrorLevel)
	silentLogger.DisableLevel(logging.WarnLevel)
	silentLogger.DisableLevel(logging.InfoLevel)

	_, perpetualDir, err := utils.FindProjectRoot(silentLogger)
	if err != nil {
		logger.Panicln("Error finding project root directory:", err)
	}

	projectConfig := config.LoadProjectConfig(perpetualDir, logger)

	// Create llm connector for generating embeddings
	connector, err := llm.NewLLMConnector(OpName, "", "",
		projectConfig.TextMatcherString(config.K_ProjectMdCodeMappings),
		map[string]interface{}{}, "", "",
		llm.GetSimpleRawMessageLogger(perpetualDir))
	if err != nil {
		return [][]float32{}, 0, err
	}

	logger.Notifyln("Generating search query embeddings for:", tag)

	debugString := connector.GetDebugString()
	logger.Notifyln(debugString)
	llm.GetSimpleRawMessageLogger(perpetualDir)(fmt.Sprintf("=== Embed (search query): %s\n\n\n", debugString))

	// Try cached vectors first
	if cachedVectors := TryReadFromCache(content); cachedVectors != nil {
		logger.Notifyln("Using cached search query embeddings for:", tag)
		llm.GetSimpleRawMessageLogger(perpetualDir)(
			fmt.Sprintf("Using cached search query embeddings for %s, chunk/vector count: %d\n\n\n", tag, len(cachedVectors)))
		return cachedVectors, connector.GetEmbedScoreThreshold(), nil
	}

	onFailRetriesLeft := max(connector.GetOnFailureRetryLimit(), 1)
	for ; onFailRetriesLeft >= 0; onFailRetriesLeft-- {
		vectors, status, err := connector.CreateEmbeddings(llm.SearchEmbed, tag, content)
		// Check for general error on query
		if err != nil {
			logger.Errorf("LLM query failed with status %d, error: %s", status, err)
			if status == llm.QueryInitFailed || onFailRetriesLeft < 1 {
				return [][]float32{}, 0, err
			}
			continue
		}
		// Check for hitting token limit, ideally should not occur at all for embedding models
		if status == llm.QueryMaxTokens {
			err := "LLM response(s) reached max tokens, that's probably an error with configuration of embedding model"
			logger.Errorf(err)
			return [][]float32{}, 0, errors.New(err)
		}
		// Save search query vectors to cache
		WriteToCache(content, vectors)
		return vectors, connector.GetEmbedScoreThreshold(), nil
	}

	return [][]float32{}, 0, errors.New("unknown error")
}

func CheckEmbedSupport() bool {
	_, err := llm.NewLLMConnector(OpName, "", "", nil, map[string]interface{}{}, "", "", func(v ...any) {})
	return err == nil
}
