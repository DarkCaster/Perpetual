package utils

import (
	"testing"
)

func TestCaseInsensitiveFileSearch(t *testing.T) {
	testCases := []struct {
		name          string
		targetFile    string
		fileNames     []string
		expectedFile  string
		expectedFound bool
	}{
		{
			name:          "Target file found",
			targetFile:    "file.go",
			fileNames:     []string{"file.go", "other.go", "dir/file.go"},
			expectedFile:  "file.go",
			expectedFound: true,
		},
		{
			name:          "Target file found in directory",
			targetFile:    "dir/file.go",
			fileNames:     []string{"file.go", "other.go", "dir/file.go"},
			expectedFile:  "dir/file.go",
			expectedFound: true,
		},
		{
			name:          "Target file found different case",
			targetFile:    "file.go",
			fileNames:     []string{"FILE.go", "other.go", "dir/file.go"},
			expectedFile:  "FILE.go",
			expectedFound: true,
		},
		{
			name:          "Target file found different case in directory",
			targetFile:    "DIR/file.go",
			fileNames:     []string{"FILE.go", "other.go", "dIr/file.go"},
			expectedFile:  "dIr/file.go",
			expectedFound: true,
		},
		{
			name:          "Target file not found",
			targetFile:    "missing.go",
			fileNames:     []string{"file.go", "other.go", "dir/file.go"},
			expectedFile:  "missing.go",
			expectedFound: false,
		},
		{
			name:          "Empty file names",
			targetFile:    "file.go",
			fileNames:     []string{},
			expectedFile:  "file.go",
			expectedFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, found := CaseInsensitiveFileSearch(tc.targetFile, tc.fileNames)
			if file != tc.expectedFile || found != tc.expectedFound {
				t.Errorf("CaseInsensitiveFileSearch(%q, %v) = (%q, %v), expected (%q, %v)", tc.targetFile, tc.fileNames, file, found, tc.expectedFile, tc.expectedFound)
			}
		})
	}
}

func TestCaseInsensitiveLeadingDirectoriesSearch(t *testing.T) {
	testCases := []struct {
		name             string
		targetFile       string
		projectFileNames []string
		expectedDirNames []string
	}{
		{
			name:             "Simple case",
			targetFile:       "path/to/file.txt",
			projectFileNames: []string{"path/to/file1.txt", "path/to/dir/file2.txt"},
			expectedDirNames: []string{"path", "to"},
		},
		{
			name:             "Simple case where first file in root dir",
			targetFile:       "path/to/file.txt",
			projectFileNames: []string{"file1.txt", "path/to/ile2.txt"},
			expectedDirNames: []string{"path", "to"},
		},
		{
			name:             "Simple case where one of the file is empty",
			targetFile:       "path/to/file.txt",
			projectFileNames: []string{"", "path/to/ile2.txt"},
			expectedDirNames: []string{"path", "to"},
		},
		{
			name:             "Simple case where second file in root dir",
			targetFile:       "path/to/file.txt",
			projectFileNames: []string{"path/to/file1.txt", "file2.txt"},
			expectedDirNames: []string{"path", "to"},
		},
		{
			name:             "Case insensitive",
			targetFile:       "PATH/TO/FILE.TXT",
			projectFileNames: []string{"path/tO/file1.txt", "path/to/dir/file2.txt"},
			expectedDirNames: []string{"path", "tO"},
		},
		{
			name:             "Case insensitive match to second file",
			targetFile:       "PATH/TO/FILE.TXT",
			projectFileNames: []string{"path/tO/dir/file1.txt", "path/To/file2.txt"},
			expectedDirNames: []string{"path", "To"},
		},
		{
			name:             "Case insensitive where first file is empty",
			targetFile:       "PATH/TO/FILE.TXT",
			projectFileNames: []string{"", "path/To/file2.txt"},
			expectedDirNames: []string{"path", "To"},
		},
		{
			name:             "Different case in project files",
			targetFile:       "path/to/file.txt",
			projectFileNames: []string{"PATH/to/file1.txt", "path/TO/dir/file2.txt"},
			expectedDirNames: []string{"PATH", "to"},
		},
		{
			name:             "Empty target file",
			targetFile:       "",
			projectFileNames: []string{"path/to/file1.txt", "path/to/dir/file2.txt"},
			expectedDirNames: []string{},
		},
		{
			name:             "No matching directories",
			targetFile:       "oTher/pAth/file.txt",
			projectFileNames: []string{"path/to/file1.txt", "path/to/dir/file2.txt"},
			expectedDirNames: []string{"oTher", "pAth"},
		},
		{
			name:             "Simple case with all files in root dir",
			targetFile:       "file.txt",
			projectFileNames: []string{"file1.txt", "file2.txt"},
			expectedDirNames: []string{},
		},
		{
			name:             "Simple case with all files in root dir and one file is empty",
			targetFile:       "file.txt",
			projectFileNames: []string{"file1.txt", ""},
			expectedDirNames: []string{},
		},
		{
			name:             "Simple case with all files in root dir and files match",
			targetFile:       "file.txt",
			projectFileNames: []string{"file.txt", "file2.txt"},
			expectedDirNames: []string{},
		},
		{
			name:             "Simple case with all files in root dir and files match with different case",
			targetFile:       "file.txt",
			projectFileNames: []string{"File.txt", "file2.txt"},
			expectedDirNames: []string{},
		},
		{
			name:             "Simple case with project files in root dir and search file in subdir",
			targetFile:       "suBDir/file.txt",
			projectFileNames: []string{"file1.txt", "file2.txt"},
			expectedDirNames: []string{"suBDir"},
		},
		{
			name:             "Empty target file with all files in root dir",
			targetFile:       "",
			projectFileNames: []string{"file1.txt", "file2.txt"},
			expectedDirNames: []string{},
		},
		{
			name:             "Empty target file with one of the file also empty",
			targetFile:       "",
			projectFileNames: []string{"file1.txt", ""},
			expectedDirNames: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dirNames := CaseInsensitiveLeadingDirectoriesSearch(tc.targetFile, tc.projectFileNames)
			if !equalSlices(dirNames, tc.expectedDirNames) {
				t.Errorf("Expected %v, got %v", tc.expectedDirNames, dirNames)
			}
		})
	}
}

func TestConvertFilePathToOSFormat(t *testing.T) {
	testCases := []struct {
		name           string
		targetFile     string
		invalidSep     string
		validSep       string
		expectedResult string
	}{
		{
			name:           "Windows path with forward slashes",
			targetFile:     "C:/Users/username/file.txt",
			invalidSep:     "/",
			validSep:       "\\",
			expectedResult: "C:\\Users\\username\\file.txt",
		},
		{
			name:           "Unix path with forward slashes",
			targetFile:     "/home/username/file.txt",
			invalidSep:     "\\",
			validSep:       "/",
			expectedResult: "/home/username/file.txt",
		},
		{
			name:           "Unix path with backslashes",
			targetFile:     "\\home\\username\\file.txt",
			invalidSep:     "\\",
			validSep:       "/",
			expectedResult: "/home/username/file.txt",
		},
		{
			name:           "Mixed path separators",
			targetFile:     "C:/Users\\username/file.txt",
			invalidSep:     "/",
			validSep:       "\\",
			expectedResult: "C:\\Users\\username\\file.txt",
		},
		{
			name:           "No separators",
			targetFile:     "file.txt",
			invalidSep:     "/",
			validSep:       "\\",
			expectedResult: "file.txt",
		},
		{
			name:           "Empty path",
			targetFile:     "",
			invalidSep:     "/",
			validSep:       "\\",
			expectedResult: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := convertFilePathToOSFormat(tc.targetFile, tc.invalidSep, tc.validSep)
			if result != tc.expectedResult {
				t.Errorf("Expected %q, but got %q", tc.expectedResult, result)
			}
		})
	}
}
