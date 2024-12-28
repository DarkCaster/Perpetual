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
		text, err := utils.LoadTextFile(filepath.Join(projectRootDir, item))
		if err != nil {
			logger.Panicln("Failed to attach file to prompt:", err)
		}
		result = AddFileFragment(result, item, text, filenameTags)
	}
	return result
}

func ComposeMessageFromPromptAndTextFile(projectRootDir, prompt, targetFile string, logger logging.ILogger) Message {
	result := AddPlainTextFragment(NewMessage(UserRequest), prompt)
	text, err := utils.LoadTextFile(filepath.Join(projectRootDir, targetFile))
	if err != nil {
		logger.Panicln("Failed to load document:", err)
	}
	result = AddPlainTextFragment(result, text)
	return result
}
