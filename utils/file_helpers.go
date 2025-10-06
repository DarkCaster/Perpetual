package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
)

type textEncoding int

const (
	Other textEncoding = iota
	UTF8
	UTF16LE
	UTF16BE
	UTF32LE
	UTF32BE
)

func detectUTFEncoding(data []byte) (textEncoding, int) {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return UTF8, 3
	}
	if len(data) >= 2 {
		if data[0] == 0xFF && data[1] == 0xFE {
			if len(data) >= 4 && data[2] == 0x00 && data[3] == 0x00 {
				return UTF32LE, 4
			}
			return UTF16LE, 2
		}
		if data[0] == 0xFE && data[1] == 0xFF {
			return UTF16BE, 2
		}
	}
	if len(data) >= 4 {
		if data[0] == 0x00 && data[1] == 0x00 && data[2] == 0xFE && data[3] == 0xFF {
			return UTF32BE, 4
		}
	}
	return Other, 0
}

func convertToBOMLessUTF8(data []byte) ([]byte, error) {
	encoding, bomLen := detectUTFEncoding(data)
	switch encoding {
	case Other:
		return data, nil
	case UTF8:
		return data[bomLen:], nil
	case UTF16LE:
		decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
		return decoder.Bytes(data[bomLen:])
	case UTF16BE:
		decoder := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()
		return decoder.Bytes(data[bomLen:])
	case UTF32LE:
		decoder := utf32.UTF32(utf32.LittleEndian, utf32.IgnoreBOM).NewDecoder()
		return decoder.Bytes(data[bomLen:])
	case UTF32BE:
		decoder := utf32.UTF32(utf32.BigEndian, utf32.IgnoreBOM).NewDecoder()
		return decoder.Bytes(data[bomLen:])
	default:
		return nil, fmt.Errorf("unsupported encoding")
	}
}

func CheckUTF8(data []byte) error {
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError {
			return errors.New("invalid UTF8 encoding detected")
		}
		data = data[size:]
	}
	return nil
}

func LoadTextStdin() (string, error) {
	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return LoadTextData(bytes)
}

func LoadTextFile(filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return LoadTextData(bytes)
}

func GetEncodingByName(name string) (encoding.Encoding, error) {
	enc, err := ianaindex.IANA.Encoding(name)
	if err != nil {
		return nil, fmt.Errorf("unknown fallback encoding: %s", name)
	}
	if enc == nil {
		return nil, fmt.Errorf("fallback encoding %s is not supported", name)
	}
	return enc, nil
}

func GetEncodingFromEnv() (encoding.Encoding, error) {
	name, err := GetEnvString("FALLBACK_TEXT_ENCODING")
	if err != nil {
		name = "windows-1252"
	}
	return GetEncodingByName(name)
}

func LoadTextData(bytes []byte) (string, error) {
	bytes, err := convertToBOMLessUTF8(bytes)
	if err != nil {
		return "", err
	}
	err = CheckUTF8(bytes)
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

func WriteTextStdout(text string) error {
	//TODO: use better method to convert windows line-endings
	if runtime.GOOS == "windows" {
		text = strings.ReplaceAll(text, "\n", "\r\n")
	}
	_, err := fmt.Print(text)
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

func RemoveFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}

func RotateFiles(baseFilePath string, count int) error {
	// Check if the base file exists
	if _, err := os.Stat(baseFilePath); os.IsNotExist(err) {
		// Base file doesn't exist, nothing to rotate
		return nil
	}

	// Remove the oldest rotation file if it exists
	oldestFile := fmt.Sprintf("%s.%d", baseFilePath, count-1)
	_ = os.Remove(oldestFile) // Ignore error if file doesn't exist

	// Shift all existing rotation files by one position
	for i := count - 2; i >= 0; i-- {
		oldFile := fmt.Sprintf("%s.%d", baseFilePath, i)
		newFile := fmt.Sprintf("%s.%d", baseFilePath, i+1)

		// Check if the old file exists before attempting to rename
		if _, err := os.Stat(oldFile); err == nil {
			if err := os.Rename(oldFile, newFile); err != nil {
				return fmt.Errorf("failed to rename %s to %s: %w", oldFile, newFile, err)
			}
		}
	}

	// Rename the base file to .0
	newFile := fmt.Sprintf("%s.0", baseFilePath)
	if err := os.Rename(baseFilePath, newFile); err != nil {
		return fmt.Errorf("failed to rename %s to %s: %w", baseFilePath, newFile, err)
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

func FindInRelativeFile(projectRootDir string, relativeFilePath string, regexps []*regexp.Regexp) (bool, error) {
	filePath, err := MakePathRelative(projectRootDir, relativeFilePath, true)
	if err != nil {
		return false, err
	}
	return FindInFile(filepath.Join(projectRootDir, filePath), regexps)
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

func SaveMsgPackFile(filePath string, v any) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	bFile := bufio.NewWriter(file)
	encoder := msgpack.GetEncoder()
	encoder.Reset(bFile)
	err = encoder.Encode(v)
	msgpack.PutEncoder(encoder)
	if err != nil {
		return err
	}
	return bFile.Flush()
}

func LoadMsgPackFile(filePath string, v any) error {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := msgpack.GetDecoder()
	decoder.UsePreallocateValues(true)
	decoder.Reset(file)
	err = decoder.Decode(&v)
	msgpack.PutDecoder(decoder)
	return err
}
