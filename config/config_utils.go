package config

import (
	"fmt"
	"reflect"
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
			} else if _, isInteger := templateVal.(int); isInteger {
				// Validate value is an integer
				fNum, ok := value.(float64)
				if !ok {
					return fmt.Errorf("config key '%s' is not a number but must be an integer", key)
				}
				if fNum != float64(int64(fNum)) {
					return fmt.Errorf("config key '%s' must be an integer number", key)
				}
			} else if _, isFloat := templateVal.(float64); isFloat {
				// Validate value is an integer
				_, ok := value.(float64)
				if !ok {
					return fmt.Errorf("config key '%s' must be a float number", key)
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
	target := make([]string, len(sourceArray), cap(sourceArray))
	for i, element := range sourceArray {
		target[i] = element.(string)
	}
	return target
}

func interfaceTo2DStringArray(source interface{}) [][]string {
	sourceArray := source.([]interface{})
	target := make([][]string, len(sourceArray), cap(sourceArray))
	for i, element := range sourceArray {
		target[i] = interfaceToStringArray(element)
	}
	return target
}

type rxDataPair[T any] struct {
	Rx   *regexp.Regexp
	Data []T
}

func (p *rxDataPair[T]) Match(path string) bool {
	return p.Rx.MatchString(path)
}

func (p *rxDataPair[T]) GetData() []T {
	return p.Data
}

type rxDataCollection[T any] struct {
	Pairs []*rxDataPair[T]
}

func (c *rxDataCollection[T]) GetDataIfMatch(path string) (bool, []T) {
	for _, p := range c.Pairs {
		if p.Match(path) {
			return true, p.Data
		}
	}
	return false, nil
}

func newRxDataPair[T any](rxStr string, opts []any) (*rxDataPair[T], error) {
	if len(opts) == 0 {
		return nil, fmt.Errorf("no data provided along with path-regexp: %s", rxStr)
	}
	data := make([]T, len(opts))
	for i := 0; i < len(opts); i++ {
		if val, ok := opts[i].(T); ok {
			data[i] = val
		} else {
			typeName := reflect.TypeFor[T]().String()
			return nil, fmt.Errorf("array element at position %d is not a valid value of type %s", i+2, typeName)
		}
	}
	if rx, err := regexp.Compile(rxStr); err == nil {
		return &rxDataPair[T]{Rx: rx, Data: data}, nil
	} else {
		return nil, fmt.Errorf("failed to compile path-regexp: %s: %v", rxStr, err)
	}
}

// source must be 2-d array from json, inner array first element is a regex string, next elements is a path options data, and it must be assignable to T type
func newRxDataCollection[T any](optsCount int, source []any) (PathDataCollection[T], error) {
	if optsCount < 1 {
		return nil, fmt.Errorf("options count must be > 0")
	}
	collection := make([]*rxDataPair[T], len(source))
	for i, el := range source {
		if innerArr, ok := el.([]any); ok {
			if len(innerArr) != optsCount+1 {
				typeName := reflect.TypeFor[T]().String()
				return nil, fmt.Errorf("inner array at pos %d must contain path-regexp followed by exactly %d parameters of type %s", i+1, optsCount, typeName)
			}
			if rxStr, ok := innerArr[0].(string); ok {
				if pair, err := newRxDataPair[T](rxStr, innerArr[1:]); err == nil {
					collection[i] = pair
				} else {
					return nil, fmt.Errorf("inner array at pos %d processing failed: %v", i+1, err)
				}
			} else {
				return nil, fmt.Errorf("inner array at pos %d first element is not a string path-regexp", i+1)
			}
		} else {
			typeName := reflect.TypeFor[T]().String()
			return nil, fmt.Errorf("inner array at pos %d is not a valid array of path-regexp with corresponding options of type %s", i+1, typeName)
		}
	}
	return &rxDataCollection[T]{Pairs: collection}, nil
}
