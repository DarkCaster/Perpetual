package llm

import (
	"path/filepath"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

func ComposeMessageWithFiles(projectRootDir, prompt string, targetFiles, filenameTags []string, logger logging.ILogger) Message {
	// Create message fragment with prompt
	result := AddPlainTextFragment(NewMessage(UserRequest), prompt)
	// Attach target-files contents
	for _, item := range targetFiles {
		contents, err := utils.LoadTextFile(filepath.Join(projectRootDir, item))
		if err != nil {
			logger.Panicln("Failed to attach file contents to prompt", err)
		}
		result = AddFileFragment(result, item, contents, filenameTags)
	}
	return result
}
