package config

import (
	"testing"
)

func TestValidateConfigAgainstTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template map[string]any
		config   map[string]any
		wantErr  bool
	}{
		{
			name: "valid config with all required keys",
			template: map[string]any{
				"key1": "string",
				"key2": "string",
				"key3": []string{"arr"},
			},
			config: map[string]any{
				"key1": "value1",
				"key2": "value2",
				"key3": []any{"value3"},
			},
			wantErr: false,
		},
		{
			name: "missing required key",
			template: map[string]any{
				"required": "yes",
			},
			config:  map[string]any{},
			wantErr: true,
		},
		{
			name:     "extra key",
			template: map[string]any{},
			config: map[string]any{
				"extra": "yes",
			},
			wantErr: true,
		},
		{
			name: "wrong type for string value",
			template: map[string]any{
				"key": "string",
			},
			config: map[string]any{
				"key": []any{},
			},
			wantErr: true,
		},
		{
			name: "wrong type for array value",
			template: map[string]any{
				"key": []string{},
			},
			config: map[string]any{
				"key": "not an array",
			},
			wantErr: true,
		},
		{
			name: "wrong type for integer value",
			template: map[string]any{
				"key": 1,
			},
			config: map[string]any{
				"key": 1.1,
			},
			wantErr: true,
		},
		{
			name: "wrong type for float value",
			template: map[string]any{
				"key": 1.1,
			},
			config: map[string]any{
				"key": 1,
			},
			wantErr: true,
		},
		{
			name: "nil type for integer value",
			template: map[string]any{
				"key": 1,
			},
			config: map[string]any{
				"key": nil,
			},
			wantErr: true,
		},
		{
			name: "nil type for float value",
			template: map[string]any{
				"key": 1.1,
			},
			config: map[string]any{
				"key": nil,
			},
			wantErr: true,
		},
		{
			name: "nil type for array value",
			template: map[string]any{
				"key": []string{},
			},
			config: map[string]any{
				"key": nil,
			},
			wantErr: true,
		},
		{
			name: "nil type for string value",
			template: map[string]any{
				"key": "string",
			},
			config: map[string]any{
				"key": nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfigAgainstTemplate(tt.template, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfigAgainstTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAnnotateConfigTemplateDoesNotContainObsoleteVariantKeys(t *testing.T) {
	template := GetAnnotateConfigTemplate()

	obsoleteKeys := []string{
		"stage2_prompt_variant",
		"stage2_prompt_combine",
		"stage2_prompt_best",
	}

	for _, key := range obsoleteKeys {
		if _, ok := template[key]; ok {
			t.Errorf("GetAnnotateConfigTemplate() contains obsolete key %q", key)
		}
	}

	requiredKeys := []string{
		K_SystemPrompt,
		K_SystemPromptAck,
		K_AnnotateTaskPrompt,
		K_AnnotateTaskResponse,
		K_AnnotateFilePrompts,
		K_AnnotateFileResponse,
	}

	for _, key := range requiredKeys {
		if _, ok := template[key]; !ok {
			t.Errorf("GetAnnotateConfigTemplate() missing required key %q", key)
		}
	}
}

func TestOperationConfigTemplatesDoNotContainJsonModeKeys(t *testing.T) {
	stage1OutputKeys := []string{
		"stage1_output_schema",
		"stage1_output_schema_name",
		"stage1_output_schema_desc",
		"stage1_output_key",
	}

	tests := []struct {
		name         string
		templateName string
		template     map[string]any
		obsoleteKeys []string
	}{
		{
			name:         "implement",
			templateName: "GetImplementConfigTemplate",
			template:     GetImplementConfigTemplate(),
			obsoleteKeys: append([]string{
				"stage1_analysis_json_mode_prompt",
				"stage1_task_analysis_json_mode_prompt",
				"stage3_planning_json_mode_prompt",
				"stage3_task_planning_json_mode_prompt",
				"stage3_planning_lite_json_mode_prompt",
				"stage3_output_schema",
				"stage3_output_schema_name",
				"stage3_output_schema_desc",
				"stage3_output_key",
			}, stage1OutputKeys...),
		},
		{
			name:         "doc",
			templateName: "GetDocConfigTemplate",
			template:     GetDocConfigTemplate(),
			obsoleteKeys: append([]string{
				"stage1_refine_json_mode_prompt",
				"stage1_write_json_mode_prompt",
			}, stage1OutputKeys...),
		},
		{
			name:         "explain",
			templateName: "GetExplainConfigTemplate",
			template:     GetExplainConfigTemplate(),
			obsoleteKeys: append([]string{
				"stage1_question_json_mode_prompt",
			}, stage1OutputKeys...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range tt.obsoleteKeys {
				if _, ok := tt.template[key]; ok {
					t.Errorf("%s() contains obsolete JSON-mode key %q", tt.templateName, key)
				}
			}
		})
	}
}

func TestValidateOpAnnotateFilePrompts(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{
			name:    "not an array",
			value:   "string",
			wantErr: true,
		},
		{
			name:    "empty array",
			value:   []any{},
			wantErr: true,
		},
		{
			name:    "inner element not an array",
			value:   []any{"not array"},
			wantErr: true,
		},
		{
			name:    "inner array wrong length",
			value:   []any{[]any{"one"}},
			wantErr: true,
		},
		{
			name:    "inner element not string",
			value:   []any{[]any{1, "two", "three"}},
			wantErr: true,
		},
		{
			name: "valid structure",
			value: []any{
				[]any{"prompt1", "response1", "response1short"},
				[]any{"prompt2", "response2", "response2short"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOpAnnotateFilePrompts(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAnnotateFilePrompts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEvenStringArray(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		arrName string
		wantErr bool
	}{
		{
			name:    "not an array",
			value:   "string",
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "odd number of elements",
			value:   []any{"one", "two", "three"},
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "non-string element",
			value:   []any{"one", 2},
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "valid even string array",
			value:   []any{"one", "two", "three", "four"},
			arrName: "test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEvenStringArray(tt.value, tt.arrName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEvenStringArray() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNonEmptyStringArray(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		arrName string
		wantErr bool
	}{
		{
			name:    "not an array",
			value:   "string",
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "empty array",
			value:   []any{},
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "empty-string only element",
			value:   []any{""},
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "empty-string element",
			value:   []any{"valid", ""},
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "non-string element",
			value:   []any{"valid", 123, "invalid"},
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "valid single element",
			value:   []any{"valid"},
			arrName: "test",
			wantErr: false,
		},
		{
			name:    "valid multiple elements",
			value:   []any{"one", "two", "three"},
			arrName: "test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNonEmptyStringArray(tt.value, tt.arrName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNonEmptyStringArray() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCompileRegexArray(t *testing.T) {
	tests := []struct {
		name    string
		source  []string
		wantErr bool
	}{
		{
			name:    "valid regex patterns",
			source:  []string{`^[a-z]+$`, `^[0-9]+$`},
			wantErr: false,
		},
		{
			name:    "invalid regex pattern",
			source:  []string{`^[a-z+$`},
			wantErr: true,
		},
		{
			name:    "empty regex array",
			source:  []string{},
			wantErr: false,
		},
		{
			name:    "nil value",
			source:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compileRegexArray(tt.source, "test")
			if (err != nil) != tt.wantErr {
				t.Errorf("compileRegexArray() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
