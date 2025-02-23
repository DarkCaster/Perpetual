package utils

import (
	"regexp"
	"testing"
)

func TestFilterFilesWithWhitelist(t *testing.T) {
	testCases := []struct {
		name             string
		sourceFiles      []string
		whitelist        []*regexp.Regexp
		expectedFiltered []string
		expectedDropped  []string
	}{
		{
			name:             "No whitelist, all files dropped",
			sourceFiles:      []string{"file1.txt", "file2.log"},
			whitelist:        []*regexp.Regexp{},
			expectedFiltered: []string{},
			expectedDropped:  []string{"file1.txt", "file2.log"},
		},
		{
			name:             "Single regex match",
			sourceFiles:      []string{"file1.txt", "file2.log"},
			whitelist:        []*regexp.Regexp{regexp.MustCompile(`\.txt$`)},
			expectedFiltered: []string{"file1.txt"},
			expectedDropped:  []string{"file2.log"},
		},
		{
			name:             "Multiple regex matches",
			sourceFiles:      []string{"file1.txt", "file2.log", "readme.md"},
			whitelist:        []*regexp.Regexp{regexp.MustCompile(`\.txt$`), regexp.MustCompile(`\.md$`)},
			expectedFiltered: []string{"file1.txt", "readme.md"},
			expectedDropped:  []string{"file2.log"},
		},
		{
			name:             "All files match",
			sourceFiles:      []string{"file1.txt", "file2.log", "readme.md"},
			whitelist:        []*regexp.Regexp{regexp.MustCompile(`.*`)},
			expectedFiltered: []string{"file1.txt", "file2.log", "readme.md"},
			expectedDropped:  []string{},
		},
		{
			name:             "No files match",
			sourceFiles:      []string{"file1.txt", "file2.log"},
			whitelist:        []*regexp.Regexp{regexp.MustCompile(`\.md$`)},
			expectedFiltered: []string{},
			expectedDropped:  []string{"file1.txt", "file2.log"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filtered, dropped := FilterFilesWithWhitelist(tc.sourceFiles, tc.whitelist)
			if !equalStringSlices(filtered, tc.expectedFiltered) {
				t.Errorf("Expected filtered %v, got %v", tc.expectedFiltered, filtered)
			}
			if !equalStringSlices(dropped, tc.expectedDropped) {
				t.Errorf("Expected dropped %v, got %v", tc.expectedDropped, dropped)
			}
		})
	}
}

func TestFilterFilesWithBlacklist(t *testing.T) {
	testCases := []struct {
		name             string
		sourceFiles      []string
		blacklist        []*regexp.Regexp
		expectedFiltered []string
		expectedDropped  []string
	}{
		{
			name:             "No blacklist, no files dropped",
			sourceFiles:      []string{"file1.txt", "file2.log"},
			blacklist:        []*regexp.Regexp{},
			expectedFiltered: []string{"file1.txt", "file2.log"},
			expectedDropped:  []string{},
		},
		{
			name:             "Single regex match",
			sourceFiles:      []string{"file1.txt", "file2.log"},
			blacklist:        []*regexp.Regexp{regexp.MustCompile(`\.txt$`)},
			expectedFiltered: []string{"file2.log"},
			expectedDropped:  []string{"file1.txt"},
		},
		{
			name:             "Multiple regex matches",
			sourceFiles:      []string{"file1.txt", "file2.log", "readme.md"},
			blacklist:        []*regexp.Regexp{regexp.MustCompile(`\.txt$`), regexp.MustCompile(`\.md$`)},
			expectedFiltered: []string{"file2.log"},
			expectedDropped:  []string{"file1.txt", "readme.md"},
		},
		{
			name:             "All files match",
			sourceFiles:      []string{"file1.txt", "file2.log", "readme.md"},
			blacklist:        []*regexp.Regexp{regexp.MustCompile(`.*`)},
			expectedFiltered: []string{},
			expectedDropped:  []string{"file1.txt", "file2.log", "readme.md"},
		},
		{
			name:             "No files match",
			sourceFiles:      []string{"file1.txt", "file2.log"},
			blacklist:        []*regexp.Regexp{regexp.MustCompile(`\.md$`)},
			expectedFiltered: []string{"file1.txt", "file2.log"},
			expectedDropped:  []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filtered, dropped := FilterFilesWithBlacklist(tc.sourceFiles, tc.blacklist)
			if !equalStringSlices(filtered, tc.expectedFiltered) {
				t.Errorf("Expected filtered %v, got %v", tc.expectedFiltered, filtered)
			}
			if !equalStringSlices(dropped, tc.expectedDropped) {
				t.Errorf("Expected dropped %v, got %v", tc.expectedDropped, dropped)
			}
		})
	}
}
