package config

import (
	"regexp"
)

type Config interface {
	Integer(key string) int
	String(key string) string
	Regexp(key string) *regexp.Regexp
	Object(key string) map[string]interface{}
	StringArray(key string) []string
	StringArray2D(key string) [][]string
	RegexpArray(key string) []*regexp.Regexp
}
