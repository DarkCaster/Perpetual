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

func ComposeMessageWithFilesOrText(projectRootDir, prompt string, body interface{}, filenameTags []string, logger logging.ILogger) (Message, bool) {
	if text, isText := body.(string); isText {
		if text == "" {
			return NewMessage(UserRequest), false
		}
		logger.Debugf("Creating message with text")
		return AddPlainTextFragment(AddPlainTextFragment(NewMessage(UserRequest), prompt), text), true
	} else if fileNames, isFnList := body.([]string); isFnList {
		if len(fileNames) < 1 {
			return NewMessage(UserRequest), false
		}
		logger.Debugf("Creating message with files")
		return ComposeMessageWithFiles(projectRootDir, prompt, fileNames, filenameTags, logger), true
	}

	logger.Panicln("Unsupported message body type")
	return NewMessage(UserRequest), false
}
