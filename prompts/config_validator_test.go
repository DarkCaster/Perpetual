package prompts

import (
	"testing"
)

func TestValidateConfigAgainstTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template map[string]interface{}
		config   map[string]interface{}
		wantErr  bool
	}{
		{
			name: "valid config with all required keys",
			template: map[string]interface{}{
				"key1": "string",
				"key2": "string",
				"key3": []string{"arr"},
				"key4": map[string]interface{}{},
			},
			config: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": []interface{}{"value3"},
				"key4": map[string]interface{}{"nested": "value"},
			},
			wantErr: false,
		},
		{
			name: "missing required key",
			template: map[string]interface{}{
				"required": "yes",
			},
			config:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "wrong type for string value",
			template: map[string]interface{}{
				"key": "string",
			},
			config: map[string]interface{}{
				"key": []interface{}{},
			},
			wantErr: true,
		},
		{
			name: "wrong type for array value",
			template: map[string]interface{}{
				"key": []string{},
			},
			config: map[string]interface{}{
				"key": "not an array",
			},
			wantErr: true,
		},
		{
			name: "wrong type for object value",
			template: map[string]interface{}{
				"key": map[string]interface{}{},
			},
			config: map[string]interface{}{
				"key": "not an object",
			},
			wantErr: true,
		},
		{
			name: "nil type for object value",
			template: map[string]interface{}{
				"key": map[string]interface{}{},
			},
			config: map[string]interface{}{
				"key": nil,
			},
			wantErr: true,
		},
		{
			name: "nil type for array value",
			template: map[string]interface{}{
				"key": []string{},
			},
			config: map[string]interface{}{
				"key": nil,
			},
			wantErr: true,
		},
		{
			name: "nil type for string value",
			template: map[string]interface{}{
				"key": "string",
			},
			config: map[string]interface{}{
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

func TestValidateOpAnnotateStage1Prompts(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "not an array",
			value:   "string",
			wantErr: true,
		},
		{
			name:    "empty array",
			value:   []interface{}{},
			wantErr: true,
		},
		{
			name:    "inner element not an array",
			value:   []interface{}{"not array"},
			wantErr: true,
		},
		{
			name:    "inner array wrong length",
			value:   []interface{}{[]interface{}{"one"}},
			wantErr: true,
		},
		{
			name:    "inner element not string",
			value:   []interface{}{[]interface{}{1, "two"}},
			wantErr: true,
		},
		{
			name: "valid structure",
			value: []interface{}{
				[]interface{}{"prompt1", "response1"},
				[]interface{}{"prompt2", "response2"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOpAnnotateStage1Prompts(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAnnotateStage1Prompts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEvenStringArray(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
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
			value:   []interface{}{"one", "two", "three"},
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "non-string element",
			value:   []interface{}{"one", 2},
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "valid even string array",
			value:   []interface{}{"one", "two", "three", "four"},
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
		value   interface{}
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
			value:   []interface{}{},
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "empty-string only element",
			value:   []interface{}{""},
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "empty-string element",
			value:   []interface{}{"valid", ""},
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "non-string element",
			value:   []interface{}{"valid", 123, "invalid"},
			arrName: "test",
			wantErr: true,
		},
		{
			name:    "valid single element",
			value:   []interface{}{"valid"},
			arrName: "test",
			wantErr: false,
		},
		{
			name:    "valid multiple elements",
			value:   []interface{}{"one", "two", "three"},
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
