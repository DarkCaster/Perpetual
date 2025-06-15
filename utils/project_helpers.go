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
	Vectors  [][]float32 `json:"vectors"`
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

func SaveEmbeddings(filePath string, checksums map[string]string, embeddings map[string][][]float32) error {
	var entries embeddingEntries
	for filename, checksum := range checksums {
		vectors, ok := embeddings[filename]
		if !ok {
			continue
		}
		entry := embeddingEntry{
			Filename: filename,
			Checksum: checksum,
			Vectors:  vectors,
		}
		entries = append(entries, entry)
	}
	sort.Sort(entries)
	err := SaveMsgPackFile(filePath, entries)
	if err != nil {
		return err
	}
	return nil
}

func GetEmbeddings(filePath string, filenames []string) (map[string][][]float32, map[string]string, int, error) {
	var embeddings embeddingEntries
	err := LoadMsgPackFile(filePath, &embeddings)
	if err != nil {
		embeddings = nil
		if !os.IsNotExist(err) {
			return nil, nil, 0, err
		}
	}
	vectorDimensions := 0
	fileVectors := make(map[string][][]float32)
	fileChecksums := make(map[string]string)
	for _, filename := range filenames {
		//set initial state of requested file checksum to "error"
		fileChecksums[filename] = "error"
		//only set vectors for requested files, to lower ram usage
		var vectors [][]float32
		found := false
		for _, entry := range embeddings {
			if entry.Filename == filename {
				vectors = entry.Vectors
				found = true
				break
			}
		}
		if !found {
			continue
		}
		//set vectors
		fileVectors[filename] = vectors
		//detect vector dimensions
		if vectorDimensions == 0 && len(vectors) > 0 && len(vectors[0]) > 0 {
			vectorDimensions = len(vectors[0])
		}
		//check vectors consistency
		if vectorDimensions > 0 && len(vectors) > 0 {
			for _, vector := range vectors {
				if len(vector) != vectorDimensions {
					//inconsistent dimensions detected
					vectorDimensions = -1
					break
				}
			}
		}
	}
	//save all checksums present at embeddings-storage for convenience
	for _, entry := range embeddings {
		fileChecksums[entry.Filename] = entry.Checksum
	}
	return fileVectors, fileChecksums, vectorDimensions, nil
}

func GetChangedFiles(oldFileChecksums map[string]string, curFileChecksums map[string]string) []string {
	var changedFiles []string
	for filename, checksum := range curFileChecksums {
		oldChecksum, ok := oldFileChecksums[filename]
		if !ok || oldChecksum != checksum {
			changedFiles = append(changedFiles, filename)
		}
	}
	sort.Strings(changedFiles)
	return changedFiles
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
