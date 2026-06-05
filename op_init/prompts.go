package op_init

import (
	"fmt"
	"strings"
)

type prompts interface {
	// Project config - for selecting project files for proccessing, setting correct file type at markdown code-blocks
	GetProjectConfig() map[string]any
	// Configs for operations
	GetAnnotateConfig() map[string]any
	GetImplementConfig() map[string]any
	GetDocConfig() map[string]any
	GetReportConfig() map[string]any
	GetExplainConfig() map[string]any
}

// Create particular Prompts implementation depending on requested language
func newPrompts(targetLang string) (prompts, error) {
	targetLang = strings.ToUpper(targetLang)

	switch targetLang {
	case "GO":
		return &goPrompts{}, nil
	case "DOTNET":
		return &dotNetPrompts{}, nil
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
