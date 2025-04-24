package utils

import (
	"testing"
)

func TestSplitTextToChunks(t *testing.T) {
	testCases := []struct {
		name           string
		sourceText     string
		chunkSize      int
		chunkOverlap   int
		expectedChunks []string
	}{
		{
			name:           "Empty text",
			sourceText:     "",
			chunkSize:      10,
			chunkOverlap:   0,
			expectedChunks: []string{},
		},
		{
			name:           "Text smaller than chunk size, no chunk overlap",
			sourceText:     "Small text",
			chunkSize:      20,
			chunkOverlap:   0,
			expectedChunks: []string{"Small text"},
		},
		{
			name:           "Text smaller than chunk size, with chunk overlap",
			sourceText:     "Small text",
			chunkSize:      20,
			chunkOverlap:   3,
			expectedChunks: []string{"Small text"},
		},
		{
			name:           "Text smaller than chunk size, with big chunk overlap",
			sourceText:     "Small text",
			chunkSize:      20,
			chunkOverlap:   19,
			expectedChunks: []string{"Small text"},
		},
		{
			name:           "Text equal to chunk size",
			sourceText:     "Exactly 10",
			chunkSize:      10,
			chunkOverlap:   0,
			expectedChunks: []string{"Exactly 10"},
		},
		{
			name:           "Text larger than chunk size, no overlap",
			sourceText:     "This is a longer text that needs to be split",
			chunkSize:      10,
			chunkOverlap:   0,
			expectedChunks: []string{"This is a ", "longer tex", "t that nee", "ds to be s", "o be split"},
		},
		{
			name:           "Text larger than chunk size, with overlap",
			sourceText:     "This is a longer text that needs to be split",
			chunkSize:      10,
			chunkOverlap:   3,
			expectedChunks: []string{"This is a ", " a longer ", "er text th", " that need", "eeds to be", "o be split"},
		},
		{
			name:           "Text with overlap equal to chunk size",
			sourceText:     "This is a longer text",
			chunkSize:      10,
			chunkOverlap:   10,
			expectedChunks: []string{"This is a longer text"},
		},
		{
			name:           "Text with overlap greater than chunk size",
			sourceText:     "This is a longer text",
			chunkSize:      10,
			chunkOverlap:   15,
			expectedChunks: []string{"This is a longer text"},
		},
		{
			name:           "Unicode text",
			sourceText:     "こんにちは世界! 你好世界！ Hello world!",
			chunkSize:      10,
			chunkOverlap:   3,
			expectedChunks: []string{"こんにちは世界! 你", "! 你好世界！ He", " Hello wor", "llo world!"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SplitTextToChunks(tc.sourceText, tc.chunkSize, tc.chunkOverlap)

			if len(result) != len(tc.expectedChunks) {
				t.Errorf("Expected %d chunks, got %d", len(tc.expectedChunks), len(result))
				return
			}

			for i, chunk := range result {
				if chunk != tc.expectedChunks[i] {
					t.Errorf("Chunk %d mismatch:\nExpected: %q\nGot:      %q", i, tc.expectedChunks[i], chunk)
				}
			}
		})
	}
}
