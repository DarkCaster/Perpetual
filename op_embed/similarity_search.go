package op_embed

import (
	"math"
	"path/filepath"
	"sort"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

func SimilaritySearchStage(limit int, ratio float64, perpetualDir string, searchQueries, searchTags, sourceFiles, preSelectedFiles []string, logger logging.ILogger) []string {
	if limit < 1 {
		logger.Infoln("Local similarity search is disabled")
		return preSelectedFiles
	}

	//generate embeddings for search queries
	searchVectors := [][]float32{}
	for i, query := range searchQueries {
		vectors, err := GenerateEmbeddings(searchTags[i], query, logger)
		if err != nil {
			logger.Debugln("Failed to generate embeddings for search queries:", err)
			logger.Infoln("LLM embeddings for local similarity search is not configured or failed")
			return preSelectedFiles
		}
		searchVectors = append(searchVectors, vectors...)
	}

	logger.Traceln("Loading embeddings")
	embeddings, _, vectorDimensions, err := utils.GetEmbeddings(filepath.Join(perpetualDir, utils.EmbeddingsFileName), sourceFiles)
	if err != nil {
		logger.Panicln("Failed to load embeddings:", err)
	}
	logger.Traceln("Done loading embeddings")

	if vectorDimensions < 0 {
		logger.Panicln("Vectors dimensions inconsistency detected for existing embeddings, check your LLM embeddings configuration and rebuild all embeddings by running embed operation with -f flag")
	}

	//get similarity results for search queries
	similarityResults := SimilaritySearch(searchVectors, embeddings)

	//calculate result limit
	resultsDistribution := make([]int, len(similarityResults))

	//helper for (re)calculating resultsDistribution for all or some elements of resultsDistribution:
	redistributeResultsLimit := func(start, count int) {
		pos := start
		if pos >= len(resultsDistribution) {
			return
		}
		//fair slots distribution across resultsDistribution elements
		for ; count > 0; count-- {
			resultsDistribution[pos] += 1
			pos++
			if pos >= len(resultsDistribution) {
				pos = start
			}
		}
		//ensure each result-distribution counter entry has at least 1 result
		for i := start; i < len(resultsDistribution); i++ {
			if resultsDistribution[i] < 1 {
				resultsDistribution[i] = 1
			}
		}
	}

	//initial results distribution
	redistributeResultsLimit(0, int(math.Min(math.Ceil(float64(len(preSelectedFiles))*ratio), float64(limit))))

	selectedFiles := []string{}
	for i, result := range similarityResults {
		//invalidate scores for files that already in preSelectedFiles
		for _, filename := range preSelectedFiles {
			result[filename] = -math.MaxFloat32
		}
		//invalidate scores for files that already selected
		for _, filename := range selectedFiles {
			result[filename] = -math.MaxFloat32
		}
		sortedResult := sortFilesByScore(result)
		added := 0
		//select top N results according to previously calculated resultsDistribution count
		for r := 0; r < resultsDistribution[i] && r < len(sortedResult); r++ {
			//TODO: select files with scores > treshold value. currently it is hardcoded as 0
			if result[sortedResult[r]] > 0 {
				added++
				selectedFiles = append(selectedFiles, sortedResult[r])
			}
		}
		//calculate how much more extra slots we have
		extra := resultsDistribution[i] - added
		if extra > 0 {
			//redistribute extra slots for use with other similarityResults
			redistributeResultsLimit(i+1, extra)
		}
	}

	return selectedFiles
}

func sortFilesByScore(sourceMap map[string]float32) []string {
	keys := make([]string, 0, len(sourceMap))
	for key := range sourceMap {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return sourceMap[keys[i]] > sourceMap[keys[j]]
	})
	return keys
}

func SimilaritySearch(searchVector [][]float32, filesSourceVectors map[string][][]float32) []map[string]float32 {
	scoresBySearchVector := []map[string]float32{}
	// iterate over search vectors
	for _, searchVector := range searchVector {
		scores := make(map[string]float32)
		// iterate over source files
		for filename, sourceVectors := range filesSourceVectors {
			// iterate over source vectors
			for _, sourceVector := range sourceVectors {
				//perform cosine similarity search of searchVector against sourceVector
				if score, ok := cosineSearch(searchVector, sourceVector); ok {
					//update score for this file if it higher than prevously recorded
					if oldScore, ok := scores[filename]; !ok || oldScore < score {
						scores[filename] = score
					}
				}
			}
		}
		scoresBySearchVector = append(scoresBySearchVector, scores)
	}
	return scoresBySearchVector
}

// TODO: optimize ? maybe, use 32bit math ?
// based on example from here: https://github.com/tmc/langchaingo/blob/main/examples/cybertron-embedding-example/cybertron-embedding.go
func cosineSearch(x, y []float32) (float32, bool) {
	if len(x) != len(y) {
		return -math.MaxFloat32, false
	}
	var dot, nx, ny float64
	for i := range x {
		nx += float64(x[i]) * float64(x[i])
		ny += float64(y[i]) * float64(y[i])
		dot += float64(x[i]) * float64(y[i])
	}
	score := dot / (math.Sqrt(nx) * math.Sqrt(ny))
	//should not happen, but still...
	if score > math.MaxFloat32 {
		return math.MaxFloat32, true
	}
	if score < -math.MaxFloat32 {
		return -math.MaxFloat32, true
	}
	return float32(score), true
}
