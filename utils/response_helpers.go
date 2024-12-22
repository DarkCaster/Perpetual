package utils

import (
	"regexp"
	"strings"

	"github.com/DarkCaster/Perpetual/logging"
)

func FilterAndTrimResponses(responses []string, forbiddenTagPairs []*regexp.Regexp, logger logging.ILogger) []string {
	var finalResponses []string
	var evenIndexElements []*regexp.Regexp
	for i := 0; i < len(forbiddenTagPairs); i += 2 {
		evenIndexElements = append(evenIndexElements, forbiddenTagPairs[i])
	}
	var oddIndexElements []*regexp.Regexp
	for i := 1; i < len(forbiddenTagPairs); i += 2 {
		oddIndexElements = append(oddIndexElements, forbiddenTagPairs[i])
	}
	var checkUnique = make(map[string]bool)
	for i, variant := range responses {
		// Filter-out variants that contain code-blocks - this is not allowed
		if blocks, err := ParseMultiTaggedTextRx(variant, evenIndexElements, oddIndexElements, true); err != nil || len(blocks) > 0 {
			logger.Warnf("LLM response #%d contains not allowed tagged text or code blocks", i)
			continue
		}
		// Trim unneded symbols from both ends of annotation
		variant = strings.Trim(variant, " \t\n") //note: there is a space character first, do not remove it
		if len(variant) < 1 {
			logger.Warnf("LLM response #%d is empty", i)
			continue
		}
		if !checkUnique[variant] {
			finalResponses = append(finalResponses, variant)
			checkUnique[variant] = true
		}
	}
	return finalResponses
}
