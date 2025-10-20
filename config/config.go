package config

import (
	"regexp"
)

// File path matcher holding some path-related data that can be used if path matched
type PathData[T any] interface {
	Match(path string) bool
	GetData() []T
}

// Container with PathData instances in same order as in config file, return associated data for the first patch match (if any)
type PathDataCollection[T any] interface {
	GetDataIfMatch(path string) (bool, []T)
}

type Config interface {
	Float(key string) float64
	Integer(key string) int
	String(key string) string
	Regexp(key string) *regexp.Regexp
	Object(key string) map[string]interface{}
	StringArray(key string) []string
	StringArray2D(key string) [][]string
	RegexpArray(key string) []*regexp.Regexp
	PathWithStringData(key string) PathDataCollection[string]
	PathWithIntegerData(key string) PathDataCollection[int]
}
