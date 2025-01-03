package config

import (
	"fmt"
	"regexp"
)

func validateConfigAgainstTemplate(template, config map[string]interface{}) error {
	// Check that all required keys from template exist in config
	for key := range template {
		if _, exists := config[key]; !exists {
			// If key is required but missing, return error
			return fmt.Errorf("missing required key in config file: %s", key)
		}
	}

	// Check that all required keys from config exist in template
	for key := range config {
		if _, exists := template[key]; !exists {
			// If key is required but missing, return error
			return fmt.Errorf("extra key in config file: %s", key)
		}
	}

	// Validate config values are of correct type
	for key, value := range config {
		// Check if template expects an array for this key
		if templateVal, exists := template[key]; exists && templateVal != nil {
			if _, isArray := templateVal.([]string); isArray {
				// Validate value is an array
				if _, ok := value.([]interface{}); !ok {
					return fmt.Errorf("config key '%s' must be an array", key)
				}
			} else if _, isString := templateVal.(string); isString {
				// Validate value is a string
				str, ok := value.(string)
				if !ok {
					return fmt.Errorf("config key '%s' must be a string", key)
				}
				if len(str) < 1 {
					return fmt.Errorf("config key '%s' is empty", key)
				}
			} else if _, isObject := templateVal.(map[string]interface{}); isObject {
				// Validate value is an object
				if _, ok := value.(map[string]interface{}); !ok {
					return fmt.Errorf("config key '%s' must be an object", key)
				}
			}
		}
	}

	return nil
}

func validateEvenStringArray(value interface{}, name string) error {
	arr, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("%s must be an array", name)
	}

	if len(arr)%2 != 0 {
		return fmt.Errorf("%s must contain even number of elements", name)
	}

	for i, v := range arr {
		if _, ok := v.(string); !ok {
			return fmt.Errorf("%s[%d] must be a string", name, i)
		}
	}

	return nil
}

func validateNonEmptyStringArray(value interface{}, name string) error {
	arr, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("%s must be an array", name)
	}

	if len(arr) < 1 {
		return fmt.Errorf("%s must contain at least one element", name)
	}

	for i, v := range arr {
		str, ok := v.(string)
		if !ok {
			return fmt.Errorf("%s[%d] must be a string", name, i)
		}
		if len(str) < 1 {
			return fmt.Errorf("%s[%d] is empty", name, i)
		}
	}

	return nil
}

func compileRegexArray(source []string, name string) ([]*regexp.Regexp, error) {
	result := make([]*regexp.Regexp, len(source))
	for i, rxStr := range source {
		rx, err := regexp.Compile(rxStr)
		if err != nil {
			return nil, fmt.Errorf("failed to compile %s[%d] regexp: %s", name, i, err)
		}
		result[i] = rx
	}
	return result, nil
}

func interfaceToStringArray(source interface{}) []string {
	sourceArray := source.([]interface{})
	target := make([]string, len(sourceArray))
	for i, element := range sourceArray {
		target[i] = element.(string)
	}
	return target
}

func interfaceTo2DStringArray(source interface{}) [][]string {
	sourceArray := source.([]interface{})
	target := make([][]string, len(sourceArray))
	for i, element := range sourceArray {
		target[i] = interfaceToStringArray(element)
	}
	return target
}
