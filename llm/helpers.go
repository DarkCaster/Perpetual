package llm

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

var sourceFileCacheLock sync.Mutex //just in case, for possible future multi-threaded access
var sourceFileCache map[string]string = make(map[string]string)

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
		logger.Panicf("Failed to attach file '%s' to prompt: %v", file, err)
	}
	sourceFileCacheLock.Lock()
	sourceFileCache[file] = text
	sourceFileCacheLock.Unlock()
	return AddFileFragment(message, file, text, filenameTags)
}

func PrecacheSourceFile(projectRootDir, file string) {
	sourceFileCacheLock.Lock()
	defer sourceFileCacheLock.Unlock()
	if _, exist := sourceFileCache[file]; !exist {
		if text, _, err := utils.LoadTextFile(filepath.Join(projectRootDir, file)); err == nil {
			sourceFileCache[file] = text
		}
	}
}

func GetSourceFileFromCache(file string) (string, int, error) {
	sourceFileCacheLock.Lock()
	defer sourceFileCacheLock.Unlock()
	if text, exist := sourceFileCache[file]; exist {
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

func ComposeMessageWithFilesOrText(projectRootDir, prompt string, body any, filenameTags utils.TagPair, logger logging.ILogger) (Message, bool) {
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
