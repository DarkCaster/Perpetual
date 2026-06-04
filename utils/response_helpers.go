package utils

import (
	"encoding/json"
	"fmt"
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

func ParseListFromJSON(jsonData, key string) ([]string, error) {
	var bJsonData = []byte(jsonData)
	if err := CheckUTF8(bJsonData); err != nil {
		return nil, err
	}
	objMap := map[string]interface{}{}
	if err := json.Unmarshal(bJsonData, &objMap); err != nil {
		return nil, err
	}
	outputObj, exists := objMap[key]
	if !exists {
		return nil, fmt.Errorf("cannot find object with key %s in deserialized json", key)
	}
	outputArray, ok := outputObj.([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert output object to array")
	}
	target := make([]string, len(outputArray))
	for i, element := range outputArray {
		target[i] = fmt.Sprint(element)
	}
	return target, nil
}
