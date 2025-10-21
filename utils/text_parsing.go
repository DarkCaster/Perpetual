package utils

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type TagPair struct {
	Left  string
	Right string
}

// Ordered container of [regexp + associated data] records in the same order as in config file
// Return associated data when one of the regexps matched (trying in the same order as in config file)
type TextMatcher[T any] interface {
	TryMatch(path string) (bool, []T)
}

type rxDataPair[T any] struct {
	Rx   *regexp.Regexp
	Data []T
}

type rxMatcher[T any] struct {
	Pairs []*rxDataPair[T]
}

func (c *rxMatcher[T]) TryMatch(path string) (bool, []T) {
	for _, p := range c.Pairs {
		if p.Rx.MatchString(path) {
			return true, p.Data
		}
	}
	return false, nil
}

func newRxDataPair[T any](rxStr string, opts []any) (*rxDataPair[T], error) {
	if len(opts) == 0 {
		return nil, fmt.Errorf("no data provided along with regexp: %s", rxStr)
	}
	data := make([]T, len(opts))
	tType := reflect.TypeFor[T]()
	for i, opt := range opts {
		ok := false
		v := reflect.ValueOf(opt)
		switch v.Kind() {
		// manual conversion of floats, because JSON deserialize numbers as float
		case reflect.Float64, reflect.Float32:
			if ok = v.CanConvert(tType); ok {
				data[i] = v.Convert(tType).Interface().(T)
			}
		// direct assignment for other types
		default:
			data[i], ok = opt.(T)
		}
		if !ok {
			return nil, fmt.Errorf("array element at position %d is not a valid value of type %s", i+2, tType.String())
		}
	}
	if rx, err := regexp.Compile(rxStr); err == nil {
		return &rxDataPair[T]{Rx: rx, Data: data}, nil
	} else {
		return nil, fmt.Errorf("failed to compile regexp: %s: %v", rxStr, err)
	}
}

func NewRxMatcher[T any](optsCount int, source any) (TextMatcher[T], error) {
	if optsCount < 1 {
		return nil, fmt.Errorf("options count must be > 0")
	}
	sourceArray, ok := source.([]any)
	if !ok {
		return nil, fmt.Errorf("value is not an a 2d array")
	}
	collection := make([]*rxDataPair[T], len(sourceArray))
	for i, el := range sourceArray {
		if innerArr, ok := el.([]any); ok {
			if len(innerArr) != optsCount+1 {
				typeName := reflect.TypeFor[T]().String()
				return nil, fmt.Errorf("inner array at pos %d must contain regexp followed by exactly %d parameters of type %s", i+1, optsCount, typeName)
			}
			if rxStr, ok := innerArr[0].(string); ok {
				if pair, err := newRxDataPair[T](rxStr, innerArr[1:]); err == nil {
					collection[i] = pair
				} else {
					return nil, fmt.Errorf("inner array at pos %d processing failed: %v", i+1, err)
				}
			} else {
				return nil, fmt.Errorf("inner array at pos %d first element is not a regexp", i+1)
			}
		} else {
			typeName := reflect.TypeFor[T]().String()
			return nil, fmt.Errorf("inner array at pos %d is not a valid array with regexp + corresponding options of type %s", i+1, typeName)
		}
	}
	return &rxMatcher[T]{Pairs: collection}, nil
}

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

func GetTextBeforeFirstMatchRx(text string, searchRegexp *regexp.Regexp) string {
	match := searchRegexp.FindStringIndex(text)
	if match == nil {
		//we need to return empty string here, so GetTextBeforeFirstMatchRx + GetTextAfterFirstMatchRx will complement each other
		return ""
	}
	return text[:match[0]]
}

func GetTextAfterFirstMatchRx(text string, searchRegexp *regexp.Regexp) string {
	match := searchRegexp.FindStringIndex(text)
	if match == nil {
		return text
	}
	return text[match[1]:]
}

func GetTextAfterFirstMatchesRx(text string, searchRegexps []*regexp.Regexp) string {
	for _, element := range searchRegexps {
		text = GetTextAfterFirstMatchRx(text, element)
	}
	return text
}

func GetTextAfterLastMatchRx(text string, searchRegexp *regexp.Regexp) string {
	matches := searchRegexp.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		//we need to return empty string here, so GetTextBeforeLastMatchRx + GetTextAfterLastMatchRx will complement each other
		return ""
	}
	lastMatch := matches[len(matches)-1]
	return text[lastMatch[1]:]
}

func GetTextBeforeLastMatchRx(text string, searchRegexp *regexp.Regexp) string {
	matches := searchRegexp.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return text
	}
	lastMatch := matches[len(matches)-1]
	return text[:lastMatch[0]]
}

func GetTextBeforeLastMatchesRx(text string, searchRegexps []*regexp.Regexp) string {
	for _, element := range searchRegexps {
		text = GetTextBeforeLastMatchRx(text, element)
	}
	return text
}

func GetEvenRegexps(arr []*regexp.Regexp) []*regexp.Regexp {
	var evenIndexElements []*regexp.Regexp
	for i := 0; i < len(arr); i += 2 {
		evenIndexElements = append(evenIndexElements, arr[i])
	}
	return evenIndexElements
}

func GetOddRegexps(arr []*regexp.Regexp) []*regexp.Regexp {
	var oddIndexElements []*regexp.Regexp
	for i := 1; i < len(arr); i += 2 {
		oddIndexElements = append(oddIndexElements, arr[i])
	}
	return oddIndexElements
}

func SplitTextToChunks(sourceText string, chunkSize, chunkOverlap int) []string {
	if len(sourceText) < 1 {
		return []string{}
	}

	if chunkOverlap >= chunkSize {
		return []string{sourceText}
	}

	sourceRunes := []rune(sourceText)
	// single chunk text
	if len(sourceRunes) <= chunkSize {
		return []string{sourceText}
	}

	result := []string{}
	//copy source runes chunk by chunk, mind the overlap from start
	//leave the last chunk untouched
	for len(sourceRunes) > chunkSize {
		//copy chunk from start
		result = append(result, string(sourceRunes[:chunkSize]))
		skip := chunkSize - chunkOverlap
		if len(sourceRunes)-skip < chunkSize {
			skip = len(sourceRunes) - chunkSize
		}
		sourceRunes = sourceRunes[skip:]
	}
	//sourceRunes slice now contains full chunk, add it and return
	result = append(result, string(sourceRunes))
	return result
}
