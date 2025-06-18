package shared

import (
	"math/rand"
	"slices"
	"sort"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_embed"
	"github.com/DarkCaster/Perpetual/utils"
)

func Stage1Preselect(
	perpetualDir string,
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

	// Calculate how many files we need
	filesToRequest := int((float64(len(projectFiles)) / 100.0) * percentToSelect)
	if filesToRequest < 10 {
		logger.Warnf("File pre-selection percentage is too low for the count of project-files: %f%", percentToSelect)
		return createErrorResult()
	}

	// Files to randomize
	filesToRandomize := int((float64(filesToRequest) / 100.0) * percentToRandomize)
	filesToRequest -= filesToRandomize
	if filesToRequest < 5 {
		logger.Warnf("File randomization percentage is too big for the count of project-files: %f%", percentToRandomize)
		filesToRequest += filesToRandomize
		filesToRandomize = 0
	}

	// Prepare for local similarity search
	searchQueries, searchTags := op_embed.GetQueriesForSimilaritySearch(query, targetFiles, annotations)
	//TODO: extract tasks as separate queries for targetFiles with implement comments, do not use targetFiles at GetQueriesForSimilaritySearch call

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
	logger.Infof("Context saving enabled, pre-selecting %d files and %d random files (%d in total) for %d passes",
		len(similarFiles),
		filesToRandomize,
		len(similarFiles)+filesToRandomize,
		passCount)

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

	return results
}
