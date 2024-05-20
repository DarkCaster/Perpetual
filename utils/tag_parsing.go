package utils

import (
	"errors"
	"regexp"
	"strings"
)

// Parse multiline text enclosed with start and end tags provided as regexps.
// Good for extracting data from whole text file output from LLM
func ParseTaggedText(sourceText string, startTagRegex string, endTagRegex string) ([]string, error) {
	startTag, err := regexp.Compile(startTagRegex)
	if err != nil {
		return nil, err
	}
	endTag, err := regexp.Compile(endTagRegex)
	if err != nil {
		return nil, err
	}
	var result []string
	for {
		startIndex := startTag.FindStringIndex(sourceText)
		endIndex := endTag.FindStringIndex(sourceText)
		// If neither start nor end tags are found, exit the loop
		if startIndex == nil && endIndex == nil {
			break
		}
		// If only the end tag is found or the end tag comes before the start tag, it's an error
		if startIndex == nil || (endIndex != nil && endIndex[0] < startIndex[1]) {
			return result, errors.New("invalid tag order")
		}
		// If only the start tag is found, it's an error
		if endIndex == nil {
			return result, errors.New("unclosed tag")
		}
		// Save the text between the tags
		taggedText := sourceText[startIndex[1]:endIndex[0]]
		result = append(result, taggedText)
		// Trim the processed part of the text
		sourceText = sourceText[endIndex[1]:]
	}
	return result, nil
}

// Search and replace all matches of searchRegex in the text
// Please, consider not to use it on big texts.
func ReplaceTag(text, searchRegex, replacement string) (string, error) {
	rx, err := regexp.Compile(searchRegex)
	if err != nil {
		return "", err
	}
	matches := rx.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return "", errors.New("failed to find tag to replace inside source text")
	}
	var parts []string
	start := 0
	for _, match := range matches {
		parts = append(parts, text[start:match[0]])
		parts = append(parts, replacement)
		start = match[1]
	}
	parts = append(parts, text[start:])
	return strings.Join(parts, ""), nil
}

func GetTextAfterFirstMatch(text, searchRegexp string) (string, error) {
	rx, err := regexp.Compile(searchRegexp)
	if err != nil {
		return "", err
	}
	match := rx.FindStringIndex(text)
	if match == nil {
		return text, nil
	}
	return text[match[1]:], nil
}

func GetTextAfterFirstMatches(text string, searchRegexps []string) (string, error) {
	for _, element := range searchRegexps {
		var err error
		text, err = GetTextAfterFirstMatch(text, element)
		if err != nil {
			return text, err
		}
	}
	return text, nil
}

func GetTextBeforeLastMatch(text, searchRegexp string) (string, error) {
	rx, err := regexp.Compile(searchRegexp)
	if err != nil {
		return "", err
	}
	matches := rx.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return text, nil
	}
	lastMatch := matches[len(matches)-1]
	return text[:lastMatch[0]], nil
}

func GetTextBeforeLastMatches(text string, searchRegexps []string) (string, error) {
	for _, element := range searchRegexps {
		var err error
		text, err = GetTextBeforeLastMatch(text, element)
		if err != nil {
			return text, err
		}
	}
	return text, nil
}
