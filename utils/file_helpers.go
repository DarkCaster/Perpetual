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

type utfEncoding int

const (
	UTF8 utfEncoding = iota
	UTF8BOM
	UTF16LE
	UTF16BE
	UTF32LE
	UTF32BE
)

type fileParams struct {
	ModernEncoding        utfEncoding
	UsingFallbackEncoding bool
}

var projectFilesParams = map[string]fileParams{}

func detectUTFEncoding(data []byte) (utfEncoding, int) {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return UTF8BOM, 3
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
	return UTF8, 0
}

func convertFromUTFEncoding(data []byte, encoding utfEncoding, bomLen int) (string, error) {
	var err error = nil
	var runes []byte = nil
	//convert source data to the sequence of native UTF8 runes
	switch encoding {
	case UTF8:
		runes = NewSlice(data...) //already UTF8
	case UTF8BOM:
		runes = NewSlice(data[bomLen:]...) //just cut the BOM
	case UTF16LE:
		runes, err = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder().Bytes(NewSlice(data[bomLen:]...))
	case UTF16BE:
		runes, err = unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder().Bytes(NewSlice(data[bomLen:]...))
	case UTF32LE:
		runes, err = utf32.UTF32(utf32.LittleEndian, utf32.IgnoreBOM).NewDecoder().Bytes(NewSlice(data[bomLen:]...))
	case UTF32BE:
		runes, err = utf32.UTF32(utf32.BigEndian, utf32.IgnoreBOM).NewDecoder().Bytes(NewSlice(data[bomLen:]...))
	default:
		return "", fmt.Errorf("unsupported encoding")
	}
	//conversion failed
	if err != nil {
		return "", err
	}
	//here we have slice with native UTF8 runes, check it for validity
	if err = CheckUTF8(runes); err != nil {
		return "", err
	}
	//return final string
	return string(runes), nil
}

func convertFromFallbackEncoding(data []byte, encoding encoding.Encoding) (string, error) {
	if encoding == nil {
		return "", errors.New("fallback encoding is not defined")
	}
	runes, err := encoding.NewDecoder().Bytes(NewSlice(data...))
	//conversion failed
	if err != nil {
		return "", err
	}
	//here we have slice with native UTF8 runes, check it for validity
	if err = CheckUTF8(runes); err != nil {
		return "", err
	}
	//return final string
	return string(runes), nil
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

func LoadTextStdin() (string, string, error) {
	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", "", err
	}
	//get encodings
	fallbackEncoding, encErr := GetEncodingFromEnv()
	encoding, bomLen := detectUTFEncoding(bytes)
	//try converting to string using UTF encoding
	text, convErr := convertFromUTFEncoding(bytes, encoding, bomLen)
	if convErr != nil && encErr != nil {
		return "", "", convErr
	}
	wrn := ""
	//try converting to string using fallback encoding
	if convErr != nil {
		wrn = fmt.Sprintf("warning: %v, using fallback encoding", convErr)
		text, convErr = convertFromFallbackEncoding(bytes, fallbackEncoding)
		if convErr != nil {
			return "", wrn, fmt.Errorf("convert from fallback encoding failed: %v", convErr)
		}
	}
	return strings.ReplaceAll(text, "\r\n", "\n"), wrn, nil
}

func LoadTextFile(filePath string) (string, string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", err
	}
	//get encodings
	fallbackEncoding, encErr := GetEncodingFromEnv()
	encoding, bomLen := detectUTFEncoding(bytes)
	params := fileParams{ModernEncoding: encoding, UsingFallbackEncoding: false}
	//try converting to string using UTF encoding
	text, convErr := convertFromUTFEncoding(bytes, encoding, bomLen)
	if convErr != nil && encErr != nil {
		return "", "", convErr
	}
	wrn := ""
	//try converting to string using fallback encoding
	if convErr != nil {
		params.UsingFallbackEncoding = true
		wrn = fmt.Sprintf("warning: %v, using fallback encoding", convErr)
		text, convErr = convertFromFallbackEncoding(bytes, fallbackEncoding)
		if convErr != nil {
			return "", wrn, fmt.Errorf("convert from fallback encoding failed: %v", convErr)
		}
	}
	//store file-params for use with writing same file
	projectFilesParams[filePath] = params
	return strings.ReplaceAll(text, "\r\n", "\n"), wrn, nil
}

func GetEncodingByName(name string) (encoding.Encoding, error) {
	enc, err := ianaindex.IANA.Encoding(name)
	if err != nil {
		return nil, fmt.Errorf("unknown fallback encoding %s: %v", name, err)
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

func LoadJsonFile(filePath string, v any) error {
	str, _, err := LoadTextFile(filePath)
	if err != nil {
		return err
	}
	err = json.NewDecoder(strings.NewReader(str)).Decode(&v)
	if err != nil {
		return err
	}
	return nil
}

func convertToUTFEncoding(data []byte, encoding utfEncoding) ([]byte, error) {
	var err error = nil
	var result []byte = nil
	//convert source data to the sequence of native UTF8 runes
	switch encoding {
	case UTF8:
		result = NewSlice(data...) //already UTF8
	case UTF8BOM:
		result, err = unicode.UTF8BOM.NewEncoder().Bytes(NewSlice(data...))
	case UTF16LE:
		result, err = unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM).NewEncoder().Bytes(NewSlice(data...))
	case UTF16BE:
		result, err = unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM).NewEncoder().Bytes(NewSlice(data...))
	case UTF32LE:
		result, err = utf32.UTF32(utf32.LittleEndian, utf32.ExpectBOM).NewEncoder().Bytes(NewSlice(data...))
	case UTF32BE:
		result, err = utf32.UTF32(utf32.BigEndian, utf32.ExpectBOM).NewEncoder().Bytes(NewSlice(data...))
	default:
		return nil, fmt.Errorf("unsupported encoding")
	}
	//conversion failed
	if err != nil {
		return nil, err
	}
	//return final converted data
	return result, nil
}

func SaveTextFile(filePath string, text string) (string, error) {
	//TODO: use better method to convert windows line-endings
	if runtime.GOOS == "windows" {
		text = strings.ReplaceAll(text, "\n", "\r\n")
	}
	data := []byte(text)
	wrn := ""
	//get file parameters, try converting file to the encoding used when reading
	if params, ok := projectFilesParams[filePath]; ok {
		if params.UsingFallbackEncoding {
			encoding, err := GetEncodingFromEnv()
			if err != nil {
				wrn = fmt.Sprintf("warning: %v, failed to get fallback encoding, using plain UTF8", err)
			} else {
				convData, err := encoding.NewEncoder().Bytes(NewSlice(data...))
				if err != nil {
					wrn = fmt.Sprintf("warning: %v, failed to convert file to fallback encoding, using plain UTF8", err)
				} else {
					data = convData
				}
			}
		} else {
			convData, err := convertToUTFEncoding(data, params.ModernEncoding)
			if err != nil {
				wrn = fmt.Sprintf("warning: %v, failed to convert file to source encoding, using plain UTF8", err)
			} else {
				data = convData
			}
		}
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return wrn, err
	}
	return wrn, nil
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
	_, err = SaveTextFile(filePath, writer.String())
	return err
}

func FindInFile(filePath string, regexps []*regexp.Regexp) (bool, string, error) {
	strData, wrn, err := LoadTextFile(filePath)
	if err != nil {
		return false, wrn, err
	}
	scanner := bufio.NewScanner(strings.NewReader(strData))
	for scanner.Scan() {
		for _, rx := range regexps {
			if rx.MatchString(scanner.Text()) {
				return true, wrn, nil
			}
		}
	}
	return false, wrn, nil
}

func FindInRelativeFile(projectRootDir string, relativeFilePath string, regexps []*regexp.Regexp) (bool, string, error) {
	filePath, err := MakePathRelative(projectRootDir, relativeFilePath, true)
	if err != nil {
		return false, "", err
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
