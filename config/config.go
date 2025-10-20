package config

import (
	"regexp"

	"github.com/DarkCaster/Perpetual/utils"
)

type Config interface {
	Float(key string) float64
	Integer(key string) int
	String(key string) string
	Regexp(key string) *regexp.Regexp
	Object(key string) map[string]interface{}
	Tags(key string) utils.TagPair
	RegexpArray(key string) []*regexp.Regexp
	TextMatcherString(key string) utils.TextMatcher[string]
	TextMatcherInteger(key string) utils.TextMatcher[int]
}
