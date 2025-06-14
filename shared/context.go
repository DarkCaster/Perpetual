package shared

import (
	"strings"

	"github.com/DarkCaster/Perpetual/logging"
)

// This file contain shared service funcs used with various operations to manage context saving measures

func ValidateContextSavingValue(contextSavingMode string, logger logging.ILogger) string {
	contextSavingMode = strings.ToUpper(contextSavingMode)
	if contextSavingMode != "AUTO" && contextSavingMode != "OFF" && contextSavingMode != "MEDIUM" && contextSavingMode != "HIGH" {
		logger.Panicln("Invalid context saving mode value provided")
	}
	return contextSavingMode
}

func GetLocalSearchModeFromContextSavingValue(contextSavingMode string, requestedFilesCount, searchLimit int) int {
	var searchMode int
	switch contextSavingMode {
	case "HIGH":
		searchMode = 1
	case "MEDIUM":
		searchMode = 1
	case "OFF":
		searchMode = 0
	case "AUTO":
		fallthrough
	default:
		if requestedFilesCount <= searchLimit {
			//for low requested file count - use aggressive search mode
			searchMode = 0
		} else {
			//for high requested file count - use conservative search mode
			searchMode = 1
		}
	}
	return searchMode
}

func GetAnnotateAndEmbedCmdLineFlags(userFilterFile, contextSavingMode string) ([]string, []string) {
	op_annotate_params := []string{}
	op_embed_params := []string{}
	if userFilterFile != "" {
		op_annotate_params = append(op_annotate_params, "-x", userFilterFile)
		op_embed_params = append(op_embed_params, "-x", userFilterFile)
	}
	if contextSavingMode != "AUTO" {
		op_annotate_params = append(op_annotate_params, "-c", contextSavingMode)
	}
	return op_annotate_params, op_embed_params
}
