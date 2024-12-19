package utils

import (
	"errors"
	"regexp"
	"strings"
)

func ParseMultiTaggedTextRx(sourceText string, startTags []*regexp.Regexp, endTags []*regexp.Regexp, ignoreUnclosedTagErrors bool) ([]string, error) {
	var result []string
	for {
		var startIndex []int
		for _, tag := range startTags {
			probe := tag.FindStringIndex(sourceText)
			if probe != nil {
				if startIndex == nil || startIndex[1] < probe[1] {
					startIndex = probe
				}
			}
		}

		rightMostPos := 0
		var endIndex []int
		for _, tag := range endTags {
			probe := tag.FindStringIndex(sourceText)
			if probe != nil {
				if endIndex == nil || endIndex[0] > probe[0] {
					endIndex = probe
				}
				if probe[1] > rightMostPos {
					rightMostPos = probe[1]
				}
			}
		}

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
			if !ignoreUnclosedTagErrors {
				return result, errors.New("unclosed tag")
			}
			rightMostPos = len(sourceText)
			endIndex = []int{rightMostPos, rightMostPos}
		}
		// Save the text between the tags
		taggedText := sourceText[startIndex[1]:endIndex[0]]
		result = append(result, taggedText)
		// Trim the processed part of the text
		sourceText = sourceText[rightMostPos:]
	}
	return result, nil
}

// Parse text enclosed with multiple start and end tags provided as regexps.
func ParseMultiTaggedText(sourceText string, startTagRegexps []string, endTagRegexps []string, ignoreUnclosedTagErrors bool) ([]string, error) {
	var startTags []*regexp.Regexp
	for _, startTagRegexp := range startTagRegexps {
		startTag, err := regexp.Compile(startTagRegexp)
		if err != nil {
			return nil, err
		}
		startTags = append(startTags, startTag)
	}

	var endTags []*regexp.Regexp
	for _, endTagRegexp := range endTagRegexps {
		endTag, err := regexp.Compile(endTagRegexp)
		if err != nil {
			return nil, err
		}
		endTags = append(endTags, endTag)
	}
	return ParseMultiTaggedTextRx(sourceText, startTags, endTags, ignoreUnclosedTagErrors)
}

// Parse text enclosed with start and end tags provided as regexps.
func ParseTaggedText(sourceText string, startTagRegexp string, endTagRegexp string, ignoreUnclosedTagErrors bool) ([]string, error) {
	return ParseMultiTaggedText(sourceText, []string{startTagRegexp}, []string{endTagRegexp}, ignoreUnclosedTagErrors)
}

func ParseTaggedTextRx(sourceText string, startTags *regexp.Regexp, endTags *regexp.Regexp, ignoreUnclosedTagErrors bool) ([]string, error) {
	return ParseMultiTaggedTextRx(sourceText, []*regexp.Regexp{startTags}, []*regexp.Regexp{endTags}, ignoreUnclosedTagErrors)
}

func ReplaceTagRx(text string, searchRegex *regexp.Regexp, replacement string) (string, error) {
	matches := searchRegex.FindAllStringIndex(text, -1)
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

// Search and replace all matches of searchRegex in the text
// Please, consider not to use it on big texts.
func ReplaceTag(text, searchRegex, replacement string) (string, error) {
	rx, err := regexp.Compile(searchRegex)
	if err != nil {
		return "", err
	}
	return ReplaceTagRx(text, rx, replacement)
}

func GetTextAfterFirstMatchRx(text string, searchRegexp *regexp.Regexp) (string, error) {
	match := searchRegexp.FindStringIndex(text)
	if match == nil {
		return text, nil
	}
	return text[match[1]:], nil
}

func GetTextAfterFirstMatch(text, searchRegexp string) (string, error) {
	rx, err := regexp.Compile(searchRegexp)
	if err != nil {
		return "", err
	}
	return GetTextAfterFirstMatchRx(text, rx)
}

func GetTextAfterFirstMatchesRx(text string, searchRegexps []*regexp.Regexp) (string, error) {
	for _, element := range searchRegexps {
		var err error
		text, err = GetTextAfterFirstMatchRx(text, element)
		if err != nil {
			return text, err
		}
	}
	return text, nil
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
