package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode/utf8"
)

func checkUTF8(data []byte) error {
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError {
			return errors.New("invalid UTF8 encoding detected")
		}
		data = data[size:]
	}
	return nil
}

func LoadTextFile(filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	err = checkUTF8(bytes)
	if err != nil {
		return "", err
	}
	//TODO: use better method to convert windows line-endings
	return strings.ReplaceAll(string(bytes), "\r\n", "\n"), nil
}

func LoadJsonFile(filePath string, v any) error {
	str, err := LoadTextFile(filePath)
	if err != nil {
		return err
	}
	err = json.NewDecoder(strings.NewReader(str)).Decode(&v)
	if err != nil {
		return err
	}
	return nil
}

func SaveTextFile(filePath string, text string) error {
	//TODO: use better method to convert windows line-endings
	if runtime.GOOS == "windows" {
		text = strings.ReplaceAll(text, "\n", "\r\n")
	}
	err := os.WriteFile(filePath, []byte(text), 0644)
	if err != nil {
		return err
	}
	return nil
}

func AppendToTextFile(filePath string, text string) error {
	//TODO: use better method to convert windows line-endings
	if runtime.GOOS == "windows" {
		text = strings.ReplaceAll(text, "\n", "\r\n")
	}
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(text)
	if err != nil {
		return err
	}
	return nil
}

func SaveJsonFile(filePath string, v any) error {
	var writer bytes.Buffer
	encoder := json.NewEncoder(&writer)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		return err
	}
	return SaveTextFile(filePath, writer.String())
}

func FindInFile(filePath string, regexps []*regexp.Regexp) (bool, error) {
	strData, err := LoadTextFile(filePath)
	if err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(strings.NewReader(strData))
	for scanner.Scan() {
		for _, rx := range regexps {
			if rx.MatchString(scanner.Text()) {
				return true, nil
			}
		}
	}
	return false, nil
}

func MakePathRelative(basePath string, filePath string, cdIntoBasePath bool) (string, error) {
	// Change workdir if requested
	if cdIntoBasePath {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		if cwd != basePath {
			err = os.Chdir(basePath)
			if err != nil {
				return "", err
			}
			defer os.Chdir(cwd)
		}
	}
	// Get absolute paths for filePath and basePath
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	// Check if absFilePath is a path inside absBasePath
	relPath, err := filepath.Rel(basePath, absFilePath)
	if err != nil {
		return "", err
	}
	// If relPath doesn't start with ".." or is not an empty string, it means absFilePath is inside absBasePath
	if !strings.HasPrefix(relPath, "..") && relPath != "" {
		return relPath, nil
	}
	// If absFilePath is not inside absBasePath, return error
	return "", fmt.Errorf("file path %s is not inside base path %s", filePath, basePath)
}

func CheckFilenameCaseCollisions(fileNames []string) bool {
	// Create a map to store file names in lowercase
	lowercaseNames := make(map[string]bool)
	// Iterate over the file names
	for _, fileName := range fileNames {
		// Convert the file name to lowercase
		lowercaseName := strings.ToLower(fileName)
		// Check if the lowercase name already exists in the map
		if _, exists := lowercaseNames[lowercaseName]; exists {
			// Case collision detected
			return false
		}
		// Add the lowercase name to the map
		lowercaseNames[lowercaseName] = true
	}
	// No case collisions detected
	return true
}

func CheckForPathSeparatorsInFilenames(fileNames []string) bool {
	// Iterate over filenames
	for _, fileName := range fileNames {
		// Split full filename to its components
		components := strings.Split(fileName, string(os.PathSeparator))
		// Check that each component does not contain following characters: / or \
		for _, component := range components {
			if strings.ContainsAny(component, "/\\") {
				return false
			}
		}
	}
	return true
}

func CaseInsensitiveFileSearch(targetFile string, fileNames []string) (string, bool) {
	// Convert targetFile to lowercase for case-insensitive search
	targetFileLower := strings.ToLower(targetFile)
	// Iterate over fileNames and compare with targetFileLower
	for _, fileName := range fileNames {
		if strings.ToLower(fileName) == targetFileLower {
			// Return the original fileName from the fileNames slice
			return fileName, true
		}
	}
	// If no match found, return an error
	return targetFile, false
}

func CaseInsensitiveLeadingDirectoriesSearch(targetFile string, projectFileNames []string) []string {
	if targetFile == "" {
		return []string{}
	}
	// Create a slice to store project directory names
	var projectDirNames []string
	// Iterate over projectFileNames and extract directory names
	for _, fileName := range projectFileNames {
		dirName := filepath.Dir(fileName)
		projectDirNames = append(projectDirNames, dirName)
	}
	// Create the target directory name from targetFile
	targetDir := filepath.Dir(targetFile)
	// Attempt to find and correct the case of targetDir using projectDirNames
	targetDir, _ = CaseInsensitiveFileSearch(targetDir, projectDirNames)
	// Split the corrected targetDir into path components
	pathComponents := strings.Split(targetDir, string(os.PathSeparator))
	if len(pathComponents) > 0 && pathComponents[0] == "." {
		pathComponents = pathComponents[1:]
	}
	return pathComponents
}

func ConvertFilePathToOSFormat(targetFile string) string {
	if runtime.GOOS == "windows" {
		return convertFilePathToOSFormat(targetFile, "/", string(os.PathSeparator))
	} else {
		return convertFilePathToOSFormat(targetFile, "\\", string(os.PathSeparator))
	}
}

func convertFilePathToOSFormat(targetFile string, invalidSeparator string, validSeparator string) string {
	return strings.ReplaceAll(targetFile, invalidSeparator, validSeparator)
}
