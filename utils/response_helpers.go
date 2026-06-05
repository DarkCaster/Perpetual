package utils

import (
	"regexp"
	"strings"

	"github.com/DarkCaster/Perpetual/logging"
)

func FilterAndTrimResponse(response string, forbiddenTagPairs []*regexp.Regexp, logger logging.ILogger) string {
	var evenIndexElements []*regexp.Regexp
	for i := 0; i < len(forbiddenTagPairs); i += 2 {
		evenIndexElements = append(evenIndexElements, forbiddenTagPairs[i])
	}
	var oddIndexElements []*regexp.Regexp
	for i := 1; i < len(forbiddenTagPairs); i += 2 {
		oddIndexElements = append(oddIndexElements, forbiddenTagPairs[i])
	}
	// Filter-out variants that contain code-blocks - this is not allowed
	if blocks, err := ParseMultiTaggedTextRx(response, evenIndexElements, oddIndexElements, true); err != nil || len(blocks) > 0 {
		logger.Warnln("LLM response contains not allowed tagged text or code blocks")
		return ""
	}
	// Trim unneded symbols from both ends of annotation
	response = strings.Trim(response, " \t\n") //note: there is a space character first, do not remove it
	if len(response) < 1 {
		logger.Warnln("LLM response is empty")
		return ""
	}
	return response
}
