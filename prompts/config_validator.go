package prompts

import "fmt"

func validateConfigAgainstTemplate(template, config map[string]interface{}) error {
	// Check that all required keys from template exist in config
	for key := range template {
		if _, exists := config[key]; !exists {
			// If key is required but missing, return error
			return fmt.Errorf("missing required key in config file: %s", key)
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
				if _, ok := value.(string); !ok {
					return fmt.Errorf("config key '%s' must be a string", key)
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

// validateAnnotateStage1Prompts validates that Stage1Prompts is a 2D array with correct structure
func validateAnnotateStage1Prompts(value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("%s must be an array", AnnotateStage1PromptNames)
	}

	if len(arr) == 0 {
		return fmt.Errorf("%s must contain at least one element", AnnotateStage1PromptNames)
	}

	for i, outer := range arr {
		innerArr, ok := outer.([]interface{})
		if !ok {
			return fmt.Errorf("%s[%d] must be an array", AnnotateStage1PromptNames, i)
		}

		if len(innerArr) != 2 {
			return fmt.Errorf("%s[%d] must contain exactly 2 elements", AnnotateStage1PromptNames, i)
		}

		for j, inner := range innerArr {
			if _, ok := inner.(string); !ok {
				return fmt.Errorf("%s[%d][%d] must be a string", AnnotateStage1PromptNames, i, j)
			}
		}
	}

	return nil
}

// validateEvenStringArray validates that array has even number of string elements
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

func ValidateAnnotateConfig(config map[string]interface{}) error {
	template := getDefaultAnnotateConfigTemplate()
	if err := validateConfigAgainstTemplate(template, config); err != nil {
		return err
	}

	if err := validateAnnotateStage1Prompts(config[AnnotateStage1PromptNames]); err != nil {
		return err
	}

	if err := validateEvenStringArray(config[FilenameTagsName], FilenameTagsName); err != nil {
		return err
	}

	if err := validateEvenStringArray(config[CodeTagsRxName], CodeTagsRxName); err != nil {
		return err
	}

	return nil
}
