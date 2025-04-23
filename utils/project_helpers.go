package utils

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/DarkCaster/Perpetual/logging"
)

type annotationEntry struct {
	Filename   string `json:"filename"`
	Checksum   string `json:"checksum"`
	Annotation string `json:"annotation"`
}

type annotationEntries []annotationEntry

func (entries annotationEntries) Len() int {
	return len(entries)
}

func (entries annotationEntries) Less(i, j int) bool {
	return entries[i].Filename < entries[j].Filename
}

func (entries annotationEntries) Swap(i, j int) {
	entries[i], entries[j] = entries[j], entries[i]
}

type embeddingEntry struct {
	Filename string      `json:"filename"`
	Checksum string      `json:"checksum"`
	Vectors  [][]float64 `json:"vectors"`
}

type embeddingEntries []embeddingEntry

func (entries embeddingEntries) Len() int {
	return len(entries)
}

func (entries embeddingEntries) Less(i, j int) bool {
	return entries[i].Filename < entries[j].Filename
}

func (entries embeddingEntries) Swap(i, j int) {
	entries[i], entries[j] = entries[j], entries[i]
}

func SaveAnnotations(filePath string, checksums map[string]string, annotations map[string]string) error {
	var entries annotationEntries
	for filename, checksum := range checksums {
		annotation, ok := annotations[filename]
		if !ok {
			continue
		}
		entry := annotationEntry{
			Filename:   filename,
			Checksum:   checksum,
			Annotation: annotation,
		}
		entries = append(entries, entry)
	}
	sort.Sort(entries)
	err := SaveJsonFile(filePath, entries)
	if err != nil {
		return err
	}
	return nil
}

func GetAnnotations(filePath string, filenames []string) (map[string]string, error) {
	var annotations annotationEntries
	err := LoadJsonFile(filePath, &annotations)
	if err != nil {
		annotations = nil
	}

	result := make(map[string]string)

	for _, filename := range filenames {
		var annotation string
		found := false

		for _, entry := range annotations {
			if entry.Filename == filename {
				annotation = entry.Annotation
				found = true
				break
			}
		}

		if !found {
			continue
		}

		result[filename] = annotation
	}

	return result, nil
}

func GetChangedAnnotations(annotationsFilePath string, fileChecksums map[string]string) ([]string, error) {
	var annotations annotationEntries
	err := LoadJsonFile(annotationsFilePath, &annotations)
	if err != nil {
		annotations = nil
	}

	annotationChecksums := make(map[string]string)
	for _, entry := range annotations {
		annotationChecksums[entry.Filename] = entry.Checksum
	}

	var changedFiles []string
	for filename, checksum := range fileChecksums {
		annotationChecksum, ok := annotationChecksums[filename]
		if !ok || annotationChecksum != checksum {
			changedFiles = append(changedFiles, filename)
		}
	}

	sort.Strings(changedFiles)
	return changedFiles, nil
}

func GetChangedEmbeddings(embeddingsFilePath string, fileChecksums map[string]string) ([]string, error) {
	var embeddings embeddingEntries
	err := LoadJsonFile(embeddingsFilePath, &embeddings)
	if err != nil {
		embeddings = nil
	}

	embeddingChecksums := make(map[string]string)
	for _, entry := range embeddings {
		embeddingChecksums[entry.Filename] = entry.Checksum
	}

	var changedFiles []string
	for filename, checksum := range fileChecksums {
		embeddingChecksum, ok := embeddingChecksums[filename]
		if !ok || embeddingChecksum != checksum {
			changedFiles = append(changedFiles, filename)
		}
	}

	sort.Strings(changedFiles)
	return changedFiles, nil
}

func GetChecksumsFromAnnotations(annotationsFilePath string, files []string) map[string]string {
	var annotations annotationEntries
	err := LoadJsonFile(annotationsFilePath, &annotations)
	if err != nil {
		annotations = nil
	}

	annotationChecksums := make(map[string]string)
	for _, file := range files {
		annotationChecksums[file] = "error"
	}
	for _, entry := range annotations {
		annotationChecksums[entry.Filename] = entry.Checksum
	}

	return annotationChecksums
}

func GetChecksumsFromEmbeddings(embeddingsFilePath string, files []string) map[string]string {
	var embeddings embeddingEntries
	err := LoadJsonFile(embeddingsFilePath, &embeddings)
	if err != nil {
		embeddings = nil
	}

	embeddingChecksums := make(map[string]string)
	for _, file := range files {
		embeddingChecksums[file] = "error"
	}
	for _, entry := range embeddings {
		embeddingChecksums[entry.Filename] = entry.Checksum
	}

	return embeddingChecksums
}

// Recursively get project files, starting from projectRootDir
// Return values:
// - filenames filtered with whitelist and blacklist (relative to projectRootDir)
// - all filenames processed (relative to projectRootDir)
// - error, if any
func GetProjectFileList(projectRootDir string, perpetualDir string, projectFilesWhitelist []*regexp.Regexp, projectFilesBlacklist []*regexp.Regexp) ([]string, []string, error) {
	var allFiles []string

	// Recursively get all files at projectRootDir and make it names relative to projectRootDir
	err := filepath.Walk(projectRootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == perpetualDir || strings.HasPrefix(path, perpetualDir+string(os.PathSeparator)) {
			return filepath.SkipDir
		}
		relPath, err := filepath.Rel(projectRootDir, path)
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			allFiles = append(allFiles, relPath)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(allFiles)

	// Generate filtered list of files
	files, _ := FilterFilesWithWhitelist(allFiles, projectFilesWhitelist)
	files, _ = FilterFilesWithBlacklist(files, projectFilesBlacklist)

	return files, allFiles, nil
}

func CalculateFilesChecksums(projectRootDir string, files []string) (map[string]string, error) {
	fileChecksums := make(map[string]string)
	for _, file := range files {
		filePath := filepath.Join(projectRootDir, file)
		checksum, err := CalculateSHA256(filePath)
		if err != nil {
			return nil, fmt.Errorf("error calculating checksum for file %s: %s", file, err)
		}
		fileChecksums[file] = checksum
	}
	return fileChecksums, nil
}

func GetFileSizes(projectRootDir string, files []string) map[string]int {
	fileSizes := make(map[string]int)
	for _, file := range files {
		filePath := filepath.Join(projectRootDir, file)
		text, err := LoadTextFile(filePath)
		if err == nil {
			fileSizes[file] = len(text)
		} else {
			fileSizes[file] = math.MaxInt
		}
	}
	return fileSizes
}

func FilterFilesWithWhitelist(sourceFiles []string, whitelist []*regexp.Regexp) ([]string, []string) {
	filteredFiles := make([]string, 0, len(sourceFiles))
	droppedFiles := make([]string, 0, len(sourceFiles))
	for _, file := range sourceFiles {
		dropFile := true
		for _, searchRx := range whitelist {
			if searchRx.MatchString(file) {
				dropFile = false
				break
			}
		}
		if dropFile {
			droppedFiles = append(droppedFiles, file)
		} else {
			filteredFiles = append(filteredFiles, file)
		}
	}
	return filteredFiles, droppedFiles
}

func FilterFilesWithBlacklist(sourceFiles []string, blacklist []*regexp.Regexp) ([]string, []string) {
	filteredFiles := make([]string, 0, len(sourceFiles))
	droppedFiles := make([]string, 0, len(sourceFiles))
	for _, file := range sourceFiles {
		dropFile := false
		for _, searchRx := range blacklist {
			if searchRx.MatchString(file) {
				dropFile = true
				break
			}
		}
		if dropFile {
			droppedFiles = append(droppedFiles, file)
		} else {
			filteredFiles = append(filteredFiles, file)
		}
	}
	return filteredFiles, droppedFiles
}

func FilterNoUploadProjectFiles(projectRootDir string, sourceFiles []string, noUploadRegexps []*regexp.Regexp, allowMissingFiles bool, logger logging.ILogger) []string {
	var results []string
	for _, file := range sourceFiles {
		if found, err := FindInRelativeFile(projectRootDir, file, noUploadRegexps); (err == nil || (allowMissingFiles && os.IsNotExist(err))) && !found {
			results = append(results, file)
		} else if found {
			logger.Warnln("Skipping file marked as 'no-upload':", file)
		} else {
			logger.Errorln("Error searching for 'no-upload' comment in file:", file, err)
		}
	}
	return results
}

func tryToSalvageFilename(projectFiles []string, fileToRecover string, logger logging.ILogger) (string, bool) {
	filename := strings.ToLower(filepath.Base(fileToRecover))
	var matches []string

	// Find all files that end with the same filename
	for _, projectFile := range projectFiles {
		if strings.ToLower(filepath.Base(projectFile)) == filename {
			matches = append(matches, projectFile)
		}
	}

	if len(matches) == 1 {
		logger.Infoln("Salvaged filename:", matches[0], "from:", fileToRecover)
		return matches[0], true
	} else if len(matches) > 1 {
		logger.Warnln("Multiple matches found while salvaging filename:", fileToRecover)
	} else {
		logger.Warnln("No matches found while salvaging filename:", fileToRecover)
	}
	return "", false
}

func FilterRequestedProjectFiles(projectRootDir string, llmRequestedFiles []string, userRequestedFiles []string, projectFiles []string, logger logging.ILogger) []string {
	var filteredResult []string
	logger.Debugln("Unfiltered file-list requested by LLM:", llmRequestedFiles)
	logger.Infoln("Files requested by LLM:")
	for _, check := range llmRequestedFiles {
		//Remove new line from the end of filename, if present
		if check != "" && check[len(check)-1] == '\n' {
			check = check[:len(check)-1]
		}
		//Remove \r from the end of filename, if present
		if check != "" && check[len(check)-1] == '\r' {
			check = check[:len(check)-1]
		}
		//Replace possibly-invalid path separators
		check = ConvertFilePathToOSFormat(check)
		//Make file path relative to project root
		file, err := MakePathRelative(projectRootDir, check, true)
		if err != nil {
			logger.Errorln("Failed to validate filename requested by LLM:", check)
			continue
		}
		//Filter-out file if it is among files reqested by user, also fix case if so
		file, found := CaseInsensitiveFileSearch(file, userRequestedFiles)
		if found {
			logger.Warnln("Filtering-out file, it is already requested by user:", file)
		} else {
			file, found := CaseInsensitiveFileSearch(file, filteredResult)
			if found {
				logger.Warnln("Filtering-out file, it is already processed or having name-case conflict:", file)
			} else {
				file, found := CaseInsensitiveFileSearch(file, projectFiles)
				if found {
					filteredResult = append(filteredResult, file)
					logger.Infoln(file)
				} else if file, found = tryToSalvageFilename(projectFiles, file, logger); found {
					_, found1 := CaseInsensitiveFileSearch(file, userRequestedFiles)
					_, found2 := CaseInsensitiveFileSearch(file, filteredResult)
					if found1 || found2 {
						logger.Warnln("Filtering-out salvaged file, because it is already in filtered or user files", file)
					} else {
						filteredResult = append(filteredResult, file)
					}
				}
			}
		}
	}

	return filteredResult
}

func AppendUserFilterFromFile(userFilterFile string, sourceFilter []*regexp.Regexp) ([]*regexp.Regexp, error) {
	var userFilesBlacklist []string
	if err := LoadJsonFile(userFilterFile, &userFilesBlacklist); err != nil {
		return nil, err
	}
	//compile to regex
	var rx []*regexp.Regexp
	for i, str := range userFilesBlacklist {
		r, err := regexp.Compile(str)
		if err != nil {
			return nil, fmt.Errorf("error compiling regexp from filter-list at pos [%d]: %s", i, err)
		}
		rx = append(rx, r)
	}
	return append(sourceFilter, rx...), nil
}
