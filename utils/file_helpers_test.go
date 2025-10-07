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
			if !equalStringSlices(dirNames, tc.expectedDirNames) {
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

func TestDetectUTFBOM(t *testing.T) {
	testCases := []struct {
		name           string
		input          []byte
		expectedOutput utfEncoding
		expectedLength int
	}{
		{
			name:           "UTF-8 BOM",
			input:          []byte{0xEF, 0xBB, 0xBF, 0x68, 0x65, 0x6C, 0x6C, 0x6F},
			expectedOutput: UTF8BOM,
			expectedLength: 3,
		},
		{
			name:           "UTF-16LE BOM",
			input:          []byte{0xFF, 0xFE, 0x68, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F, 0x00},
			expectedOutput: UTF16LE,
			expectedLength: 2,
		},
		{
			name:           "UTF-16BE BOM",
			input:          []byte{0xFE, 0xFF, 0x00, 0x68, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F},
			expectedOutput: UTF16BE,
			expectedLength: 2,
		},
		{
			name:           "UTF-32LE BOM",
			input:          []byte{0xFF, 0xFE, 0x00, 0x00, 0x68, 0x00, 0x00, 0x00, 0x65, 0x00, 0x00, 0x00},
			expectedOutput: UTF32LE,
			expectedLength: 4,
		},
		{
			name:           "UTF-32BE BOM",
			input:          []byte{0x00, 0x00, 0xFE, 0xFF, 0x00, 0x00, 0x00, 0x68, 0x00, 0x00, 0x00, 0x65},
			expectedOutput: UTF32BE,
			expectedLength: 4,
		},
		{
			name:           "No BOM",
			input:          []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F},
			expectedOutput: UTF8,
			expectedLength: 0,
		},
		{
			name:           "Empty input",
			input:          []byte{},
			expectedOutput: UTF8,
			expectedLength: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, length := detectUTFEncoding(tc.input)
			if result != tc.expectedOutput || length != tc.expectedLength {
				t.Errorf("Expected (%v, %d), but got (%v, %d)", tc.expectedOutput, tc.expectedLength, result, length)
			}
		})
	}
}

func TestConvertToBOMLessUTF8(t *testing.T) {
	testCases := []struct {
		name           string
		input          []byte
		expectedOutput []byte
	}{
		{
			name:           "UTF-8 BOM",
			input:          []byte{0xEF, 0xBB, 0xBF, 0x68, 0x65, 0x6C, 0x6C, 0x6F},
			expectedOutput: []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F},
		},
		{
			name:           "UTF-16LE BOM",
			input:          []byte{0xFF, 0xFE, 0x68, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F, 0x00},
			expectedOutput: []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F},
		},
		{
			name:           "UTF-16BE BOM",
			input:          []byte{0xFE, 0xFF, 0x00, 0x68, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F},
			expectedOutput: []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F},
		},
		{
			name:           "UTF-32LE BOM",
			input:          []byte{0xFF, 0xFE, 0x00, 0x00, 0x68, 0x00, 0x00, 0x00, 0x65, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6F, 0x00, 0x00, 0x00},
			expectedOutput: []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F},
		},
		{
			name:           "UTF-32BE BOM",
			input:          []byte{0x00, 0x00, 0xFE, 0xFF, 0x00, 0x00, 0x00, 0x68, 0x00, 0x00, 0x00, 0x65, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6F},
			expectedOutput: []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F},
		},
		{
			name:           "No BOM",
			input:          []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F},
			expectedOutput: []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F},
		},
		{
			name:           "Empty input",
			input:          []byte{},
			expectedOutput: []byte{},
		},
		{
			name:           "Malformed UTF-16LE",
			input:          []byte{0xFF, 0xFE, 0x00},
			expectedOutput: []byte{0xEF, 0xBF, 0xBD},
		},
		{
			name:           "Malformed UTF-16BE",
			input:          []byte{0xFE, 0xFF, 0x00},
			expectedOutput: []byte{0xEF, 0xBF, 0xBD},
		},
		{
			name:           "Malformed UTF-32LE",
			input:          []byte{0xFF, 0xFE, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectedOutput: []byte{0xEF, 0xBF, 0xBD},
		},
		{
			name:           "Malformed UTF-32BE",
			input:          []byte{0x00, 0x00, 0xFE, 0xFF, 0x00, 0x00, 0x00},
			expectedOutput: []byte{0xEF, 0xBF, 0xBD},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := convertToBOMLessEncoding(tc.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !equalByteSlices(output, tc.expectedOutput) {
				t.Errorf("Expected %v, but got %v", tc.expectedOutput, output)
			}
		})
	}
}

func TestCheckUTF8(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectError bool
	}{
		{
			name:        "Valid UTF-8",
			input:       []byte("Hello, 世界"),
			expectError: false,
		},
		{
			name:        "Empty input",
			input:       []byte{},
			expectError: false,
		},
		{
			name:        "Invalid UTF-8",
			input:       []byte{0xFF, 0xFE, 0x00},
			expectError: true,
		},
		{
			name:        "Malformed UTF-8 from TestConvertToBOMLessUTF8",
			input:       []byte{0xEF, 0xBF, 0xBD},
			expectError: true,
		},
		{
			name:        "Incomplete UTF-8 sequence",
			input:       []byte{0xE2, 0x82}, // Incomplete Euro sign
			expectError: true,
		},
		{
			name:        "Valid UTF-8 with multi-byte characters",
			input:       []byte("こんにちは"), // Japanese "Hello"
			expectError: false,
		},
		{
			name:        "Mixed valid and invalid UTF-8",
			input:       []byte("Hello\xFFWorld"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckUTF8(tc.input)
			if tc.expectError && err == nil {
				t.Errorf("Expected an error, but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
		})
	}
}

func TestGetEncodingByName(t *testing.T) {
	testCases := []struct {
		name        string
		encoding    string
		expectError bool
	}{
		{
			name:        "Windows-1252 (Western European)",
			encoding:    "windows-1252",
			expectError: false,
		},
		{
			name:        "ISO-8859-1 (Latin-1)",
			encoding:    "ISO-8859-1",
			expectError: false,
		},
		{
			name:        "Windows-1251 (Cyrillic)",
			encoding:    "windows-1251",
			expectError: false,
		},
		{
			name:        "Shift_JIS (Japanese)",
			encoding:    "Shift_JIS",
			expectError: false,
		},
		{
			name:        "EUC-KR (Korean)",
			encoding:    "EUC-KR",
			expectError: false,
		},
		{
			name:        "ISO-8859-2 (Central European)",
			encoding:    "ISO-8859-2",
			expectError: false,
		},
		{
			name:        "Windows-1250 (Central European)",
			encoding:    "windows-1250",
			expectError: false,
		},
		{
			name:        "Big5 (Traditional Chinese)",
			encoding:    "Big5",
			expectError: false,
		},
		{
			name:        "ISO-8859-15 (Latin-9)",
			encoding:    "ISO-8859-15",
			expectError: false,
		},
		{
			name:        "Windows-1256 (Arabic)",
			encoding:    "windows-1256",
			expectError: false,
		},
		{
			name:        "ISO-8859-7 (Greek)",
			encoding:    "ISO-8859-7",
			expectError: false,
		},
		{
			name:        "Windows-1253 (Greek)",
			encoding:    "windows-1253",
			expectError: false,
		},
		{
			name:        "ISO-8859-8 (Hebrew)",
			encoding:    "ISO-8859-8",
			expectError: false,
		},
		{
			name:        "Windows-1255 (Hebrew)",
			encoding:    "windows-1255",
			expectError: false,
		},
		{
			name:        "ISO-8859-9 (Turkish)",
			encoding:    "ISO-8859-9",
			expectError: false,
		},
		{
			name:        "Windows-1254 (Turkish)",
			encoding:    "windows-1254",
			expectError: false,
		},
		{
			name:        "EUC-JP (Japanese)",
			encoding:    "EUC-JP",
			expectError: false,
		},
		{
			name:        "ISO-2022-JP (Japanese)",
			encoding:    "ISO-2022-JP",
			expectError: false,
		},
		{
			name:        "KOI8-R (Russian)",
			encoding:    "KOI8-R",
			expectError: false,
		},
		{
			name:        "Invalid encoding",
			encoding:    "INVALID-ENCODING-XYZ",
			expectError: true,
		},
		{
			name:        "Empty encoding name",
			encoding:    "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enc, err := GetEncodingByName(tc.encoding)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error for encoding %q, but got nil", tc.encoding)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for encoding %q, but got: %v", tc.encoding, err)
				}
				if enc == nil {
					t.Errorf("Expected non-nil encoding for %q, but got nil", tc.encoding)
				}
			}
		})
	}
}
