package utils

import (
	"errors"
	"regexp/syntax"
	"testing"
)

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

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
			result, err := ReplaceTag(tc.text, tc.tagRegex, tc.replacement)
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
		expectedError  error
	}{
		{
			name:           "Single tag",
			sourceText:     "Line with <tag>text</tag>",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: []string{"text"},
			expectedError:  nil,
		},
		{
			name:           "Multiple tags",
			sourceText:     "Line with <tag>text</tag> and <tag>more text</tag>",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: []string{"text", "more text"},
			expectedError:  nil,
		},
		{
			name:           "No tags",
			sourceText:     "Line without tags",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: []string{},
			expectedError:  nil,
		},
		{
			name:           "Unclosed tag",
			sourceText:     "Line with <tag>unclosed",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: nil,
			expectedError:  errors.New("unclosed tag"),
		},
		{
			name:           "Invalid tag order",
			sourceText:     "Line with </tag>invalid<tag>",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: nil,
			expectedError:  errors.New("invalid tag order"),
		},
		{
			name:           "Multiple lines",
			sourceText:     "Line 1 <tag>text</tag>\nLine 2 <tag>more text</tag>",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: []string{"text", "more text"},
			expectedError:  nil,
		},
		{
			name:           "Invalid start tag regex",
			sourceText:     "Line with <tag>text</tag>",
			startTagRegex:  "[",
			endTagRegex:    "</tag>",
			expectedResult: nil,
			expectedError:  errors.New("error parsing regexp: missing closing ]: `[`"),
		},
		{
			name:           "Invalid end tag regex",
			sourceText:     "Line with <tag>text</tag>",
			startTagRegex:  "<tag>",
			endTagRegex:    "]",
			expectedResult: nil,
			expectedError:  errors.New("unclosed tag"),
		},
		{
			name:           "Second match tag missing",
			sourceText:     "Line 1 <tag>text</tag><tag>more text",
			startTagRegex:  "<tag>",
			endTagRegex:    "</tag>",
			expectedResult: []string{"text"},
			expectedError:  errors.New("unclosed tag"),
		},
		{
			name:           "LLM single output with header and footer",
			sourceText:     "Blah blah:\n<output>data\n</output>\n Blah blah",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n"},
			expectedError:  nil,
		},
		{
			name:           "LLM single output with header only",
			sourceText:     "Blah blah:\n<output>data\n</output>\n",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n"},
			expectedError:  nil,
		},
		{
			name:           "LLM single output with footer only",
			sourceText:     "<output>data\n</output>\nBlah\nblah\n",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n"},
			expectedError:  nil,
		},
		{
			name:           "LLM single output with header only and no newline at the end",
			sourceText:     "Blah blah:\n<output>data\n</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n"},
			expectedError:  nil,
		},
		{
			name:           "LLM single inline output with header only",
			sourceText:     "Blah blah:\n<output>data</output>\n",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data"},
			expectedError:  nil,
		},
		{
			name:           "LLM single inline output with header only and no newline at the end",
			sourceText:     "Blah blah:\n<output>data</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data"},
			expectedError:  nil,
		},
		{
			name:           "LLM single output no header",
			sourceText:     "<output>\ndata\n</output>\n",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n"},
			expectedError:  nil,
		},
		{
			name:           "LLM single output no header and no newline at the end",
			sourceText:     "<output>\ndata</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data"},
			expectedError:  nil,
		},
		{
			name:           "LLM single output without newline without header and no newline at the end",
			sourceText:     "<output>data\n</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n"},
			expectedError:  nil,
		},
		{
			name:           "LLM suspicious close tag inside",
			sourceText:     "<output>data\nthis </output> - is not a real close tag</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\nthis </output> - is not a real close tag"},
			expectedError:  nil,
		},
		{
			name:           "LLM suspicious open tag inside",
			sourceText:     "<output>data\nthis <output> - is not a real open tag</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\nthis <output> - is not a real open tag"},
			expectedError:  nil,
		},
		{
			name:           "LLM suspicious open close tags inside",
			sourceText:     "<output>data\nthis <output></output> - are not a real tags</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\nthis <output></output> - are not a real tags"},
			expectedError:  nil,
		},
		{
			name:           "LLM inline suspicious open close tags inside",
			sourceText:     "<output><output></output></output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"<output></output>"},
			expectedError:  nil,
		},
		{
			name:           "LLM multiple outputs with newlines between outputs",
			sourceText:     "<output>data\n</output>\n<output>blah\n</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n", "blah\n"},
			expectedError:  nil,
		},
		{
			name:           "LLM multiple outputs with inline heading text",
			sourceText:     "blah<output>data</output>\nblah<output>extra</output>\nblah",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data", "extra"},
			expectedError:  nil,
		},
		{
			name:           "LLM multiple outputs with inline heading text and spaces",
			sourceText:     "blah <output>data</output>\nblah <output>extra</output>\nblah",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data", "extra"},
			expectedError:  nil,
		},
		{
			name:           "LLM multiple outputs with inline heading text and spaces",
			sourceText:     "blah<output>data</output> \nblah<output>extra</output> \nblah",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data", "extra"},
			expectedError:  nil,
		},
		{
			name:           "LLM multiple outputs with surrounding text",
			sourceText:     "blah\n<output>\ndata\n</output>\nblah\n\n<output>\nextra\n\n</output>",
			startTagRegex:  "(?m)\\s*<output>\\n?",
			endTagRegex:    "(?m)<\\/output>\\s*($|\\n)",
			expectedResult: []string{"data\n", "extra\n\n"},
			expectedError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseTaggedText(tc.sourceText, tc.startTagRegex, tc.endTagRegex)
			if !equalSlices(result, tc.expectedResult) {
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
		{
			name:           "Invalid regex",
			text:           "Hello, world!",
			searchRegex:    "[a-",
			expectedResult: "",
			expectedError:  &syntax.Error{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GetTextAfterFirstMatch(tc.text, tc.searchRegex)
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
		{
			name:           "Invalid regex",
			text:           "Hello, world!",
			searchRegex:    "[a-",
			expectedResult: "",
			expectedError:  &syntax.Error{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GetTextBeforeLastMatch(tc.text, tc.searchRegex)
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
