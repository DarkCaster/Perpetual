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
		result = AppendFileToMessage(result, projectRootDir, item, filenameTags, logger)
	}
	return result
}

func AppendFileToMessage(message Message, projectRootDir, file string, filenameTags []string, logger logging.ILogger) Message {
	text, err := utils.LoadTextFile(filepath.Join(projectRootDir, file))
	if err != nil {
		logger.Panicln("Failed to attach file to prompt:", err)
	}
	return AddFileFragment(message, file, text, filenameTags)
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

func ComposeMessageWithAnnotations(prompt string, targetFiles, filenameTags []string, annotations map[string]string, logger logging.ILogger) Message {
	request := AddPlainTextFragment(NewMessage(UserRequest), prompt)
	for _, item := range targetFiles {
		request = AddIndexFragment(request, item, filenameTags)
		if annotation, ok := annotations[item]; !ok || annotation == "" {
			annotation = "No annotation available"
		} else {
			request = AddPlainTextFragment(request, annotation)
		}
	}
	return request
}
