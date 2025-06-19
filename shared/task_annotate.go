package shared

import "github.com/DarkCaster/Perpetual/logging"

// This func used to extract `IMPLEMENT` comments from source code file,
// and create short task-annotation from it that will be used to pre-select files using local similarity search.
func TaskAnnotate(
	perpetualDir string,
	targetFiles []string,
	passCount int,
	logger logging.ILogger) []string {

	return []string{}
}
