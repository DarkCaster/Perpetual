package shared

import (
	"fmt"
	"math/rand"
	"slices"
	"sort"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_embed"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage1Preselect(
	perpetualDir string,
	projectRootDir string,
	percentToSelect, percentToRandomize float64,
	projectFiles []string,
	query string,
	targetFiles []string,
	annotations map[string]string,
	passCount int,
	logger logging.ILogger) [][]string {

	createErrorResult := func() [][]string {
		results := [][]string{}
		for range passCount {
			results = append(results, projectFiles)
		}
		return results
	}

	// Add trace and debug logging
	logger.Traceln("Stage1Preselect: Starting")
	defer logger.Traceln("Stage1Preselect: Finished")

	if percentToSelect < 1 {
		logger.Infof("Context saving disabled, pre-selecting all available project files: %d", len(projectFiles))
		return createErrorResult()
	}

	if !op_embed.CheckEmbedSupport() {
		logger.Infof("Context saving disabled: embeddings not available, pre-selecting all available project files: %d", len(projectFiles))
		return createErrorResult()
	}

	// Calculate how many files we need
	filesToRequest := int((float64(len(projectFiles)) / 100.0) * percentToSelect)
	if filesToRequest < 10 {
		logger.Warnf("File pre-selection percentage is too low for the count of project-files: %0.1f%%", percentToSelect)
		return createErrorResult()
	}

	// Files to randomize
	filesToRandomize := int((float64(filesToRequest) / 100.0) * percentToRandomize)
	filesToRequest -= filesToRandomize
	if filesToRequest < 5 {
		logger.Warnf("File randomization percentage is too big for the count of project-files: %0.1f%%", percentToRandomize)
		filesToRequest += filesToRandomize
		filesToRandomize = 0
	}

	logger.Notifyf("Running pre-select stage: picking %0.1f%% of project-files (%0.1f%% of them randomly selected)", percentToSelect, percentToRandomize)

	// Prepare for local similarity search
	var searchQueries []string
	var searchTags []string
	// Compose query
	if query != "" {
		searchQueries = append(searchQueries, query)
		searchTags = append(searchTags, "query")
	}

	// Compose task annotations
	if len(targetFiles) > 0 {
		taskAnnotations := TaskAnnotate(targetFiles, logger)
		for i, task := range taskAnnotations {
			searchQueries = append(searchQueries, task)
			searchTags = append(searchTags, fmt.Sprintf("task:%s", targetFiles[i]))
		}
	}

	// make actual similarity search more silent, because it will spam a lot of unneded info
	silentLogger := logger.Clone()
	silentLogger.DisableLevel(logging.InfoLevel)
	similarFiles := op_embed.SimilaritySearchStage(
		op_embed.SelectAggressive,
		filesToRequest,
		perpetualDir,
		searchQueries,
		searchTags,
		projectFiles,
		targetFiles,
		silentLogger)
	if len(similarFiles) < 1 {
		logger.Warnln("Context saving disabled: local search returned no results")
		return createErrorResult()
	}

	// Remove similarFiles and targetFiles from projectFiles slice
	unusedProjectFiles := []string{}
	for _, file := range projectFiles {
		if !slices.Contains(similarFiles, file) && !slices.Contains(targetFiles, file) {
			unusedProjectFiles = append(unusedProjectFiles, file)
		}
	}

	// Make slices with different set of random files
	results := [][]string{}
	// Don't request more files than available
	filesToRandomize = min(filesToRandomize, len(unusedProjectFiles))
	for range passCount {
		// Shuffle unusedProjectFiles array
		if filesToRandomize > 0 {
			rand.Shuffle(len(unusedProjectFiles), func(i, j int) {
				unusedProjectFiles[i], unusedProjectFiles[j] = unusedProjectFiles[j], unusedProjectFiles[i]
			})
		}
		// Create final result for use with stage 1 pass
		result := append(utils.NewSlice(similarFiles...), unusedProjectFiles[:filesToRandomize]...)
		sort.Strings(result)
		results = append(results, result)
	}

	logger.Infof("Context saving enabled: pre-selected %d files and %d random files (%d in total) for %d passes",
		len(similarFiles),
		filesToRandomize,
		len(similarFiles)+filesToRandomize,
		passCount)

	return results
}

func MergeFileLists(fileLists [][]string, logger logging.ILogger) []string {
	result := []string{}
	resultMap := make(map[string]bool)
	for _, fileList := range fileLists {
		for _, file := range fileList {
			if resultMap[file] {
				continue
			}
			resultMap[file] = true
			result = append(result, file)
		}
	}
	sort.Strings(result)
	logger.Infoln("Files requested by LLM (merged list):")
	for _, file := range result {
		logger.Infoln(file)
	}
	return result
}
