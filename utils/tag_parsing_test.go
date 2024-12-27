package utils

import (
	"errors"
	"regexp"
	"testing"
)

func TestReplaceTag(t *testing.T) {
	testCases := []struct {
		name        string
		text        string
		tagRegex    string
		replacement string
		expected    string
		shouldError bool
	}{
		{
			name:        "No tags",
			text:        "This is a plain text without tags.",
			tagRegex:    "<tag>(.*?)</tag>",
			replacement: "REPLACED",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "Single tag",
			text:        "This is a <tag>sample</tag> text.",
			tagRegex:    "<tag>(.*?)</tag>",
			replacement: "REPLACED",
			expected:    "This is a REPLACED text.",
			shouldError: false,
		},
		{
			name:        "Single tag alt",
			text:        "This is a \"###SAMPLE###\" text.",
			tagRegex:    "###SAMPLE###",
			replacement: "REPLACED",
			expected:    "This is a \"REPLACED\" text.",
			shouldError: false,
		},
		{
			name:        "Multiple tags",
			text:        "This is a <tag>sample</tag> text with <tag>multiple</tag> tags.",
			tagRegex:    "<tag>(.*?)</tag>",
			replacement: "REPLACED",
			expected:    "This is a REPLACED text with REPLACED tags.",
			shouldError: false,
		},
		{
			name:        "Multiple tags alt",
			text:        "This is a \"###SAMPLE###\" text with \"###SAMPLE###\" tags.",
			tagRegex:    "###SAMPLE###",
			replacement: "REPLACED",
			expected:    "This is a \"REPLACED\" text with \"REPLACED\" tags.",
			shouldError: false,
		},
		{
			name:        "Tags with same replace",
			text:        "This is a \"###SAMPLE###\" text with \"###SAMPLE###\" tags.",
			tagRegex:    "###SAMPLE###",
			replacement: "###SAMPLE###",
			expected:    "This is a \"###SAMPLE###\" text with \"###SAMPLE###\" tags.",
			shouldError: false,
		},
		{
			name:        "Tag with no content",
			text:        "This is a <tag></tag> text.",
			tagRegex:    "<tag>(.*?)</tag>",
			replacement: "REPLACED",
			expected:    "This is a REPLACED text.",
			shouldError: false,
		},
		{
			name:        "Tag regex with no groups",
			text:        "This is a <tag>sample</tag> text.",
			tagRegex:    "<tag>[^<>]*</tag>",
			replacement: "REPLACED",
			expected:    "This is a REPLACED text.",
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ReplaceTagRx(tc.text, regexp.MustCompile(tc.tagRegex), tc.replacement)
			if tc.shouldError && err == nil {
				t.Errorf("Expected error, but got nil")
			} else if !tc.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if !tc.shouldError && result != tc.expected {
				t.Errorf("Expected '%s', but got '%s'", tc.expected, result)
			}
		})
	}
}

func TestParseTaggedText(t *testing.T) {
	testCases := []struct {
		name           string
		sourceText     string
		startTagRegex  string
		endTagRegex    string
		expectedResult []string
		ignUncTagErr   bool
		expectedError  error
	}{
		{
			name:           "Single tag",
			sourceText:     "Line with <tag>text</tag>",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: []string{"text"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "Multiple tags",
			sourceText:     "Line with <tag>text</tag> and <tag>more text</tag>",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: []string{"text", "more text"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "No tags",
			sourceText:     "Line without tags",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: []string{},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "Unclosed tag",
			sourceText:     "Line with <tag>unclosed",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: nil,
			ignUncTagErr:   false,
			expectedError:  errors.New("unclosed tag"),
		},
		{
			name:           "Unclosed tag ignored",
			sourceText:     "Line with <tag>unclosed",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: []string{"unclosed"},
			ignUncTagErr:   true,
			expectedError:  nil,
		},
		{
			name:           "Invalid tag order",
			sourceText:     "Line with </tag>invalid<tag>",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: nil,
			ignUncTagErr:   false,
			expectedError:  errors.New("invalid tag order"),
		},
		{
			name:           "Multiple lines",
			sourceText:     "Line 1 <tag>text</tag>\nLine 2 <tag>more text</tag>",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: []string{"text", "more text"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "Second match tag missing",
			sourceText:     "Line 1 <tag>text</tag><tag>more text",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: []string{"text"},
			ignUncTagErr:   false,
			expectedError:  errors.New("unclosed tag"),
		},
		{
			name:           "Second match tag missing ignored",
			sourceText:     "Line 1 <tag>text</tag><tag>more text",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: []string{"text", "more text"},
			ignUncTagErr:   true,
			expectedError:  nil,
		},
		{
			name:           "LLM single output with header and footer",
			sourceText:     "Blah blah:\n<output>data\n</output>\n Blah blah",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM single output with header only",
			sourceText:     "Blah blah:\n<output>data\n</output>\n",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM single output with footer only",
			sourceText:     "<output>data\n</output>\nBlah\nblah\n",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM single output with header only and no newline at the end",
			sourceText:     "Blah blah:\n<output>data\n</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM single inline output with header only",
			sourceText:     "Blah blah:\n<output>data</output>\n",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM single inline output with header only and no newline at the end",
			sourceText:     "Blah blah:\n<output>data</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM single output no header",
			sourceText:     "<output>\ndata\n</output>\n",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM single output no header and no newline at the end",
			sourceText:     "<output>\ndata</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM single output without newline without header and no newline at the end",
			sourceText:     "<output>data\n</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM suspicious close tag inside",
			sourceText:     "<output>data\nthis </output> - is not a real close tag</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\nthis </output> - is not a real close tag"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM suspicious open tag inside",
			sourceText:     "<output>data\nthis <output> - is not a real open tag</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\nthis <output> - is not a real open tag"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM suspicious open close tags inside",
			sourceText:     "<output>data\nthis <output></output> - are not a real tags</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\nthis <output></output> - are not a real tags"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM inline suspicious open close tags inside",
			sourceText:     "<output><output></output></output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"<output></output>"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM multiple outputs with newlines between outputs",
			sourceText:     "<output>data\n</output>\n<output>blah\n</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n", "blah\n"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM multiple outputs with inline heading text",
			sourceText:     "blah<output>data</output>\nblah<output>extra</output>\nblah",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data", "extra"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM multiple outputs with inline heading text and spaces",
			sourceText:     "blah <output>data</output>\nblah <output>extra</output>\nblah",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data", "extra"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM multiple outputs with inline heading text and spaces",
			sourceText:     "blah<output>data</output> \nblah<output>extra</output> \nblah",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data", "extra"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "LLM multiple outputs with surrounding text",
			sourceText:     "blah\n<output>\ndata\n</output>\nblah\n\n<output>\nextra\n\n</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n", "extra\n\n"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseTaggedTextRx(tc.sourceText, regexp.MustCompile(tc.startTagRegex), regexp.MustCompile(tc.endTagRegex), tc.ignUncTagErr)
			if !equalStringSlices(result, tc.expectedResult) {
				t.Errorf("Expected result %v, but got %v", tc.expectedResult, result)
			}
			if tc.expectedError == nil && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tc.expectedError != nil && (err == nil || tc.expectedError.Error() != err.Error()) {
				t.Errorf("Expected error '%v', but got '%v'", tc.expectedError, err)
			}
		})
	}
}

func TestParseMultiTaggedText(t *testing.T) {
	testCases := []struct {
		name           string
		sourceText     string
		startTagRegex  []string
		endTagRegex    []string
		expectedResult []string
		ignUncTagErr   bool
		expectedError  error
	}{
		{
			name:           "Single 1-st tag, 1 match",
			sourceText:     "Line with <tag>text</tag>",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: []string{"text"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "Single 2-nd tag, 1 match",
			sourceText:     "Line with <x>text</x>",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: []string{"text"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "Single 1-st tag, 2 matches",
			sourceText:     "Line with <tag><x>text</x></tag>",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: []string{"text"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "Single 2-nd tag, 2 matches",
			sourceText:     "Line with <x><tag>text</tag></x>",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: []string{"text"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},

		{
			name:           "Single 1-st tag, 2 matches, 1-st tag reverse order",
			sourceText:     "Line with <x><tag>text</x></tag>",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: []string{"text"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "Single 1-st tag, 2 matches, 2-nd tag reverse order",
			sourceText:     "Line with <tag><x>text</tag></x>",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: []string{"text"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},

		{
			name:           "Multiple tags",
			sourceText:     "Line with <tag>text</tag> and <tag>more text</tag>",
			startTagRegex:  []string{"<tag>"},
			endTagRegex:    []string{"</tag>"},
			expectedResult: []string{"text", "more text"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "Multiple tags, 2 matches",
			sourceText:     "Line with <tag><x>text</x></tag> and <tag><x>more text</x></tag>",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: []string{"text", "more text"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},

		{
			name:           "Multiple tags, 2 matches, reverse order",
			sourceText:     "Line with <x><tag>text</tag></x> and <tag><x>more text</x></tag>",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: []string{"text", "more text"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},

		{
			name:           "Multiple tags, 2 matches, partially reverse order",
			sourceText:     "Line with <tag><x>text</tag></x> and <x><tag>more text</x></tag>",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: []string{"text", "more text"},
			ignUncTagErr:   false,
			expectedError:  nil,
		},

		{
			name:           "No tags",
			sourceText:     "Line without tags",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: []string{},
			ignUncTagErr:   false,
			expectedError:  nil,
		},
		{
			name:           "Unclosed tag",
			sourceText:     "Line with <tag><x>unclosed",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: nil,
			ignUncTagErr:   false,
			expectedError:  errors.New("unclosed tag"),
		},
		{
			name:           "Unclosed tag ignored",
			sourceText:     "Line with <tag><x>unclosed",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: []string{"unclosed"},
			ignUncTagErr:   true,
			expectedError:  nil,
		},
		{
			name:           "Invalid tag order",
			sourceText:     "Line with </tag></x>invalid<x><tag>",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: nil,
			ignUncTagErr:   false,
			expectedError:  errors.New("invalid tag order"),
		},
		{
			name:           "Second match tag missing",
			sourceText:     "Line 1 <tag><x>text</tag><tag>more text",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: []string{"text"},
			ignUncTagErr:   false,
			expectedError:  errors.New("unclosed tag"),
		},
		{
			name:           "Second match tag missing ignore",
			sourceText:     "Line 1 <tag><x>text</tag><tag>more text",
			startTagRegex:  []string{"<tag>", "<x>"},
			endTagRegex:    []string{"</tag>", "</x>"},
			expectedResult: []string{"text", "more text"},
			ignUncTagErr:   true,
			expectedError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var startRx []*regexp.Regexp
			for _, str := range tc.startTagRegex {
				startRx = append(startRx, regexp.MustCompile(str))
			}
			var endRx []*regexp.Regexp
			for _, str := range tc.endTagRegex {
				endRx = append(endRx, regexp.MustCompile(str))
			}
			result, err := ParseMultiTaggedTextRx(tc.sourceText, startRx, endRx, tc.ignUncTagErr)
			if !equalStringSlices(result, tc.expectedResult) {
				t.Errorf("Expected result %v, but got %v", tc.expectedResult, result)
			}
			if tc.expectedError == nil && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tc.expectedError != nil && (err == nil || tc.expectedError.Error() != err.Error()) {
				t.Errorf("Expected error '%v', but got '%v'", tc.expectedError, err)
			}
		})
	}
}

func TestGetTextAfterFirstMatch(t *testing.T) {
	testCases := []struct {
		name           string
		text           string
		searchRegex    string
		expectedResult string
		expectedError  error
	}{
		{
			name:           "Match found",
			text:           "Hello, world! This is a test.",
			searchRegex:    "world",
			expectedResult: "! This is a test.",
			expectedError:  nil,
		},
		{
			name:           "No match found",
			text:           "This is a test without a match.",
			searchRegex:    "<tag>",
			expectedResult: "This is a test without a match.",
			expectedError:  nil,
		},
		{
			name:           "Empty text",
			text:           "",
			searchRegex:    "match",
			expectedResult: "",
			expectedError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GetTextAfterFirstMatchRx(tc.text, regexp.MustCompile(tc.searchRegex))
			if tc.expectedError == nil && err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if tc.expectedError != nil && err == nil {
				t.Errorf("Expected error: %v, but got nil", tc.expectedError)
			} else if tc.expectedError != nil && err != nil && errors.Is(err, tc.expectedError) {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedError, err)
			} else if result != tc.expectedResult {
				t.Errorf("Expected result: %q, but got: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestGetTextAfterFirstMatches(t *testing.T) {
	testCases := []struct {
		name           string
		text           string
		searchRegexps  []string
		expectedResult string
		expectedError  error
	}{
		{
			name:           "All matches found",
			text:           "Hello, world! This is a test.",
			searchRegexps:  []string{"world", "This"},
			expectedResult: " is a test.",
			expectedError:  nil,
		},
		{
			name:           "1-st match found",
			text:           "Hello, world! This is a test.",
			searchRegexps:  []string{"world", "xxx"},
			expectedResult: "! This is a test.",
			expectedError:  nil,
		},
		{
			name:           "2-nd match found",
			text:           "Hello, world! This is a test.",
			searchRegexps:  []string{"none", "This"},
			expectedResult: " is a test.",
			expectedError:  nil,
		},
		{
			name:           "No match found",
			text:           "This is a test without a match.",
			searchRegexps:  []string{"<tag>", "<another_tag>"},
			expectedResult: "This is a test without a match.",
			expectedError:  nil,
		},
		{
			name:           "Empty text",
			text:           "",
			searchRegexps:  []string{"match1", "match2"},
			expectedResult: "",
			expectedError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var rx []*regexp.Regexp
			for _, str := range tc.searchRegexps {
				rx = append(rx, regexp.MustCompile(str))
			}
			result, err := GetTextAfterFirstMatchesRx(tc.text, rx)
			if tc.expectedError == nil && err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if tc.expectedError != nil && err == nil {
				t.Errorf("Expected error: %v, but got nil", tc.expectedError)
			} else if tc.expectedError != nil && err != nil && errors.Is(err, tc.expectedError) {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedError, err)
			} else if result != tc.expectedResult {
				t.Errorf("Expected result: %q, but got: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestGetTextBeforeLastMatch(t *testing.T) {
	testCases := []struct {
		name           string
		text           string
		searchRegex    string
		expectedResult string
		expectedError  error
	}{
		{
			name:           "Match found",
			text:           "Hello, world! This is a test.",
			searchRegex:    "is",
			expectedResult: "Hello, world! This ",
			expectedError:  nil,
		},
		{
			name:           "No match found",
			text:           "This is a test without a match.",
			searchRegex:    "nothing",
			expectedResult: "This is a test without a match.",
			expectedError:  nil,
		},
		{
			name:           "Empty text",
			text:           "",
			searchRegex:    "match",
			expectedResult: "",
			expectedError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GetTextBeforeLastMatchRx(tc.text, regexp.MustCompile(tc.searchRegex))
			if tc.expectedError == nil && err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if tc.expectedError != nil && err == nil {
				t.Errorf("Expected error: %v, but got nil", tc.expectedError)
			} else if tc.expectedError != nil && err != nil && errors.Is(err, tc.expectedError) {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedError, err)
			} else if result != tc.expectedResult {
				t.Errorf("Expected result: %q, but got: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestGetTextBeforeLastMatches(t *testing.T) {
	testCases := []struct {
		name           string
		text           string
		searchRegexps  []string
		expectedResult string
		expectedError  error
	}{
		{
			name:           "All matches found",
			text:           "Hello, world! This is a test.",
			searchRegexps:  []string{"world", "This"},
			expectedResult: "Hello, ",
			expectedError:  nil,
		},
		{
			name:           "1-st match found",
			text:           "Hello, world! This is a test.",
			searchRegexps:  []string{"world", "xxx"},
			expectedResult: "Hello, ",
			expectedError:  nil,
		},
		{
			name:           "2-nd match found",
			text:           "Hello, world! This is a test.",
			searchRegexps:  []string{"none", "This"},
			expectedResult: "Hello, world! ",
			expectedError:  nil,
		},
		{
			name:           "No match found",
			text:           "This is a test without a match.",
			searchRegexps:  []string{"<tag>", "<another_tag>"},
			expectedResult: "This is a test without a match.",
			expectedError:  nil,
		},
		{
			name:           "Empty text",
			text:           "",
			searchRegexps:  []string{"match1", "match2"},
			expectedResult: "",
			expectedError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var rx []*regexp.Regexp
			for _, str := range tc.searchRegexps {
				rx = append(rx, regexp.MustCompile(str))
			}
			result, err := GetTextBeforeLastMatchesRx(tc.text, rx)
			if tc.expectedError == nil && err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if tc.expectedError != nil && err == nil {
				t.Errorf("Expected error: %v, but got nil", tc.expectedError)
			} else if tc.expectedError != nil && err != nil && errors.Is(err, tc.expectedError) {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedError, err)
			} else if result != tc.expectedResult {
				t.Errorf("Expected result: %q, but got: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestGetEvenRegexps(t *testing.T) {
	testCases := []struct {
		name     string
		input    []*regexp.Regexp
		expected []*regexp.Regexp
	}{
		{
			name:     "Empty input",
			input:    []*regexp.Regexp{},
			expected: []*regexp.Regexp{},
		},
		{
			name:     "Single element",
			input:    []*regexp.Regexp{regexp.MustCompile("a")},
			expected: []*regexp.Regexp{regexp.MustCompile("a")},
		},
		{
			name:     "Even number of elements",
			input:    []*regexp.Regexp{regexp.MustCompile("a"), regexp.MustCompile("b"), regexp.MustCompile("c"), regexp.MustCompile("d")},
			expected: []*regexp.Regexp{regexp.MustCompile("a"), regexp.MustCompile("c")},
		},
		{
			name:     "Odd number of elements",
			input:    []*regexp.Regexp{regexp.MustCompile("a"), regexp.MustCompile("b"), regexp.MustCompile("c")},
			expected: []*regexp.Regexp{regexp.MustCompile("a"), regexp.MustCompile("c")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetEvenRegexps(tc.input)
			if !equalRegexSlices(result, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
}

func TestGetOddRegexps(t *testing.T) {
	testCases := []struct {
		name     string
		input    []*regexp.Regexp
		expected []*regexp.Regexp
	}{
		{
			name:     "Empty input",
			input:    []*regexp.Regexp{},
			expected: []*regexp.Regexp{},
		},
		{
			name:     "Single element",
			input:    []*regexp.Regexp{regexp.MustCompile("a")},
			expected: []*regexp.Regexp{},
		},
		{
			name:     "Even number of elements",
			input:    []*regexp.Regexp{regexp.MustCompile("a"), regexp.MustCompile("b"), regexp.MustCompile("c"), regexp.MustCompile("d")},
			expected: []*regexp.Regexp{regexp.MustCompile("b"), regexp.MustCompile("d")},
		},
		{
			name:     "Odd number of elements",
			input:    []*regexp.Regexp{regexp.MustCompile("a"), regexp.MustCompile("b"), regexp.MustCompile("c")},
			expected: []*regexp.Regexp{regexp.MustCompile("b")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetOddRegexps(tc.input)
			if !equalRegexSlices(result, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
}

// Helper function to compare slices of regexps
func equalRegexSlices(a, b []*regexp.Regexp) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].String() != b[i].String() {
			return false
		}
	}
	return true
}
