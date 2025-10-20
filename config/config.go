package config

import (
	"regexp"
)

// Ordered container of [regexp + associated data] records in the same order as in config file
// Return associated data when one of the regexps matched (trying in the same order as in config file)
type TextMatcher[T any] interface {
	TryMatch(path string) (bool, []T)
}

type Config interface {
	Float(key string) float64
	Integer(key string) int
	String(key string) string
	Regexp(key string) *regexp.Regexp
	Object(key string) map[string]interface{}
	StringArray(key string) []string
	RegexpArray(key string) []*regexp.Regexp
	TextMatcherString(key string) TextMatcher[string]
	TextMatcherInteger(key string) TextMatcher[int]
}
