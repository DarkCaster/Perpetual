package config

import "regexp"

type Config interface {
	String(key string) string
	Regexp(key string) regexp.Regexp
	Object(key string) map[string]interface{}
	Array(key string) []string
	Array2D(key string) [][]string
}
