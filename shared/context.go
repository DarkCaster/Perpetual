package shared

import (
	"strings"

	"github.com/DarkCaster/Perpetual/config"
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

func GetLocalSearchLimitsForContextSaving(contextSavingMode string, projectFileCount int, projectConfig config.Config) (float64, float64) {
	switch contextSavingMode {
	case "HIGH":
		return projectConfig.Float(config.K_ProjectHighContextSavingSelectPercent),
			projectConfig.Float(config.K_ProjectHighContextSavingRandomPercent)
	case "MEDIUM":
		return projectConfig.Float(config.K_ProjectMediumContextSavingSelectPercent),
			projectConfig.Float(config.K_ProjectMediumContextSavingRandomPercent)
	case "OFF":
		return 0, 0
	case "AUTO":
		fallthrough
	default:
		if projectFileCount > projectConfig.Integer(config.K_ProjectHighContextSavingFileCount) {
			return projectConfig.Float(config.K_ProjectHighContextSavingSelectPercent),
				projectConfig.Float(config.K_ProjectHighContextSavingRandomPercent)
		} else if projectFileCount > projectConfig.Integer(config.K_ProjectMediumContextSavingFileCount) {
			return projectConfig.Float(config.K_ProjectMediumContextSavingSelectPercent),
				projectConfig.Float(config.K_ProjectMediumContextSavingRandomPercent)
		}
	}

	if projectFileCount > projectConfig.Integer(config.K_ProjectHighContextSavingFileCount) {
		return projectConfig.Float(config.K_ProjectHighContextSavingSelectPercent),
			projectConfig.Float(config.K_ProjectHighContextSavingRandomPercent)
	} else if projectFileCount > projectConfig.Integer(config.K_ProjectMediumContextSavingFileCount) {
		return projectConfig.Float(config.K_ProjectMediumContextSavingSelectPercent),
			projectConfig.Float(config.K_ProjectMediumContextSavingRandomPercent)
	}
	return 0, 0
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
