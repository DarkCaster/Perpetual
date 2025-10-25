package llm

import (
	"fmt"
	"path/filepath"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

// NOTE: make cache management threadsafe if needed
var sourceFileCache map[string]string

func ComposeMessageWithSourceFiles(projectRootDir, prompt string, targetFiles []string, filenameTags utils.TagPair, logger logging.ILogger) Message {
	// Create message fragment with prompt
	result := AddPlainTextFragment(NewMessage(UserRequest), prompt)
	// Attach target-files contents
	for _, item := range targetFiles {
		result = AppendSourceFileToMessage(result, projectRootDir, item, filenameTags, logger)
	}
	return result
}

func AppendSourceFileToMessage(message Message, projectRootDir, file string, filenameTags utils.TagPair, logger logging.ILogger) Message {
	text, _, err := utils.LoadTextFile(filepath.Join(projectRootDir, file))
	if err != nil {
		logger.Panicln("Failed to attach file to prompt:", err)
	}
	sourceFileCache[file] = text
	return AddFileFragment(message, file, text, filenameTags)
}

func GetSourceFileFromCache(file string) (string, int, error) {
	if text, ok := sourceFileCache[file]; ok {
		return text, len(text), nil
	}
	return "", 0, fmt.Errorf("file %s not found in cache", file)
}

func ComposeMessageWithAnnotations(prompt string, targetFiles []string, filenameTags utils.TagPair, annotations map[string]string, logger logging.ILogger) Message {
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

func ComposeMessageWithFilesOrText(projectRootDir, prompt string, body interface{}, filenameTags utils.TagPair, logger logging.ILogger) (Message, bool) {
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
		return ComposeMessageWithSourceFiles(projectRootDir, prompt, fileNames, filenameTags, logger), true
	}

	logger.Panicln("Unsupported message body type")
	return NewMessage(UserRequest), false
}
