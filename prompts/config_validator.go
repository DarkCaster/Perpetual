package prompts

import (
	"fmt"
)

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

func validateOpAnnotateStage1Prompts(value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("%s must be an array", K_AnnotateStage1Prompts)
	}

	if len(arr) == 0 {
		return fmt.Errorf("%s must contain at least one element", K_AnnotateStage1Prompts)
	}

	for i, outer := range arr {
		innerArr, ok := outer.([]interface{})
		if !ok {
			return fmt.Errorf("%s[%d] must be an array", K_AnnotateStage1Prompts, i)
		}

		if len(innerArr) != 2 {
			return fmt.Errorf("%s[%d] must contain exactly 2 elements", K_AnnotateStage1Prompts, i)
		}

		for j, inner := range innerArr {
			str, ok := inner.(string)
			if !ok {
				return fmt.Errorf("%s[%d][%d] must be a string", K_AnnotateStage1Prompts, i, j)
			}
			if len(str) < 1 {
				return fmt.Errorf("%s[%d][%d] is empty", K_AnnotateStage1Prompts, i, j)
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

func ValidateOpAnnotateConfig(config map[string]interface{}) error {
	template := GetDefaultAnnotateConfigTemplate()
	if err := validateConfigAgainstTemplate(template, config); err != nil {
		return err
	}

	if err := validateOpAnnotateStage1Prompts(config[K_AnnotateStage1Prompts]); err != nil {
		return err
	}

	if err := validateEvenStringArray(config[K_FilenameTags], K_FilenameTags); err != nil {
		return err
	}

	if err := validateEvenStringArray(config[K_CodeTagsRx], K_CodeTagsRx); err != nil {
		return err
	}

	return nil
}

func ValidateOpImplementConfig(config map[string]interface{}) error {
	template := GetDefaultImplementConfigTemplate()
	if err := validateConfigAgainstTemplate(template, config); err != nil {
		return err
	}

	if err := validateEvenStringArray(config[K_CodeTagsRx], K_CodeTagsRx); err != nil {
		return err
	}

	if err := validateEvenStringArray(config[K_FilenameTags], K_FilenameTags); err != nil {
		return err
	}

	if err := validateEvenStringArray(config[K_FilenameTagsRx], K_FilenameTagsRx); err != nil {
		return err
	}

	if err := validateNonEmptyStringArray(config[K_ImplementCommentsRx], K_ImplementCommentsRx); err != nil {
		return err
	}

	if err := validateNonEmptyStringArray(config[K_NoUploadCommentsRx], K_NoUploadCommentsRx); err != nil {
		return err
	}

	return nil
}

func ValidateOpDocConfig(config map[string]interface{}) error {
	template := GetDefaultDocConfigTemplate()
	if err := validateConfigAgainstTemplate(template, config); err != nil {
		return err
	}

	if err := validateEvenStringArray(config[K_FilenameTags], K_FilenameTags); err != nil {
		return err
	}

	if err := validateEvenStringArray(config[K_FilenameTagsRx], K_FilenameTagsRx); err != nil {
		return err
	}

	if err := validateNonEmptyStringArray(config[K_NoUploadCommentsRx], K_NoUploadCommentsRx); err != nil {
		return err
	}

	return nil
}
