package op_init

import (
	"fmt"
	"strings"
)

type prompts interface {
	// Project config - for selecting project files for proccessing, setting correct file type at markdown code-blocks
	GetProjectConfig() map[string]interface{}
	// Configs for operations
	GetAnnotateConfig() map[string]interface{}
	GetImplementConfig() map[string]interface{}
	GetDocConfig() map[string]interface{}
	GetReportConfig() map[string]interface{}
	GetExplainConfig() map[string]interface{}
}

// Create particular Prompts implementation depending on requested language
func newPrompts(targetLang string) (prompts, error) {
	targetLang = strings.ToUpper(targetLang)

	switch targetLang {
	case "GO":
		return &goPrompts{}, nil
	case "DOTNETFW":
		return &dotNetFWPrompts{}, nil
	case "BASH":
		return &bashPrompts{}, nil
	case "PYTHON3":
		return &py3Prompts{}, nil
	case "VB6":
		return &vb6Prompts{}, nil
	case "C":
		return &cPrompts{}, nil
	case "CPP":
		return &cppPrompts{}, nil
	case "ARDUINO":
		return &arduinoPrompts{}, nil
	case "FLUTTER":
		return &flutterPrompts{}, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", targetLang)
	}
}
