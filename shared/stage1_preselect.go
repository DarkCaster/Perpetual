package shared

import (
	"math/rand"
	"sort"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_embed"
)

func Stage1Preselect(
	perpetualDir string,
	percentToSelect, percentToRandomize float64,
	projectFiles []string,
	query string,
	targetFiles []string,
	annotations map[string]string,
	logger logging.ILogger) []string {

	// Add trace and debug logging
	logger.Traceln("Stage1Preselect: Starting")
	defer logger.Traceln("Stage1Preselect: Finished")

	if percentToSelect < 1 {
		logger.Infof("Context saving disabled, pre-selecting all available project files: %d", len(projectFiles))
		return projectFiles
	}

	// Calculate how many files we need
	filesToRequest := int((float64(len(projectFiles)) / 100.0) * percentToSelect)
	if filesToRequest < 10 {
		logger.Warnf("File pre-selection percentage is too low for the count of project-files: %f%", percentToSelect)
		return projectFiles
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
		return projectFiles
	}

	// Remove similarFiles and targetFiles from projectFiles slice
	unusedProjectFiles := []string{}
	for _, file := range projectFiles {
		found := false
		for _, similarFile := range similarFiles {
			if file == similarFile {
				found = true
				break
			}
		}
		if found {
			continue
		}
		for _, targetFile := range targetFiles {
			if file == targetFile {
				found = true
				break
			}
		}
		if found {
			continue
		}
		unusedProjectFiles = append(unusedProjectFiles, file)
	}

	// Get filesToRandomize random files from unusedProjectFiles
	randomFiles := []string{}
	if filesToRandomize > 0 && len(unusedProjectFiles) > 0 {
		// Don't request more files than available
		if filesToRandomize > len(unusedProjectFiles) {
			filesToRandomize = len(unusedProjectFiles)
		}
		// Shuffle unusedProjectFiles array
		rand.Shuffle(len(unusedProjectFiles), func(i, j int) {
			unusedProjectFiles[i], unusedProjectFiles[j] = unusedProjectFiles[j], unusedProjectFiles[i]
		})
		// Select first filesToRandomize files
		randomFiles = unusedProjectFiles[:filesToRandomize]
	}

	logger.Infof("Context saving enabled, pre-selected %d files and %d random files (%d in total)",
		len(similarFiles),
		len(randomFiles),
		len(similarFiles)+len(randomFiles))

	// Sort and return result
	result := append(similarFiles, randomFiles...)
	sort.Strings(result)
	return result
}
