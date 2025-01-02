package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/DarkCaster/Perpetual/logging"
)

type AnnotationEntry struct {
	Filename   string `json:"filename"`
	Checksum   string `json:"checksum"`
	Annotation string `json:"annotation"`
}

type AnnotationEntries []AnnotationEntry

func (entries AnnotationEntries) Len() int {
	return len(entries)
}

func (entries AnnotationEntries) Less(i, j int) bool {
	return entries[i].Filename < entries[j].Filename
}

func (entries AnnotationEntries) Swap(i, j int) {
	entries[i], entries[j] = entries[j], entries[i]
}

func SaveAnnotations(filePath string, checksums map[string]string, annotations map[string]string) error {
	var entries AnnotationEntries
	for filename, checksum := range checksums {
		annotation, ok := annotations[filename]
		if !ok {
			continue
		}
		entry := AnnotationEntry{
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

func GetAnnotations(filePath string, fileChecksums map[string]string) (map[string]string, error) {
	var annotations AnnotationEntries
	err := LoadJsonFile(filePath, &annotations)
	if err != nil {
		annotations = nil
	}

	result := make(map[string]string)

	for filename := range fileChecksums {
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

func GetChangedFiles(filePath string, fileChecksums map[string]string) ([]string, error) {
	var annotations AnnotationEntries
	err := LoadJsonFile(filePath, &annotations)
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

// Recursively get project files, starting from projectRootDir
// Return values:
// - map with checksums of files filtered with whitelist and blacklist
// - filenames filtered with whitelist and blacklist (relative to projectRootDir)
// - all filenames processed (relative to projectRootDir)
// - error, if any
func GetProjectFileList(projectRootDir string, perpetualDir string, projectFilesWhitelist []*regexp.Regexp, projectFilesBlacklist []*regexp.Regexp) (map[string]string, []string, []string, error) {
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
		return nil, nil, nil, err
	}
	sort.Strings(allFiles)

	// Generate filtered list of files
	var files []string
	for _, searchRx := range projectFilesWhitelist {
		for _, file := range allFiles {
			if searchRx.MatchString(file) {
				dropFile := false
				for _, dropRx := range projectFilesBlacklist {
					if dropRx.MatchString(file) {
						dropFile = true
						break
					}
				}
				if !dropFile {
					files = append(files, file)
				}
			}
		}
	}

	fileChecksums := make(map[string]string)
	for _, file := range files {
		filePath := filepath.Join(projectRootDir, file)
		checksum, err := CalculateSHA256(filePath)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error calculating checksum for file %s: %s", file, err)
		}
		fileChecksums[file] = checksum
	}

	return fileChecksums, files, allFiles, nil
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
				} else {
					logger.Warnln("Filtering-out file, it is not found among project file-list:", file)
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
