package shared

import (
	"github.com/DarkCaster/Perpetual/logging"
)

func Stage1Preselect(
	percentToSelect, percentToRandomize float64,
	projectFiles []string,
	query string,
	targetFiles []string,
	logger logging.ILogger) []string {
	if percentToSelect < 1 {
		logger.Infof("Context saving disabled, pre-selecting all available project files: %d", len(projectFiles))
		return projectFiles
	}

	return projectFiles
}
