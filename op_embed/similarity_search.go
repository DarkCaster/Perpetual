package op_embed

import (
	"math"
)

/*func SimilaritySearchStage(limit int, ratio float64, perpetualDir string, searchQueries, searchTags, sourceFiles, preSelectedFiles []string, logger logging.ILogger) []string {
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

	logger.Infoln(connector.GetDebugString())
}*/

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
