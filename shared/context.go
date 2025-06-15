package shared

import (
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_embed"
)

// This file contain service functions used with various operations to manage context saving measures

func ValidateContextSavingValue(contextSavingMode string, logger logging.ILogger) string {
	contextSavingMode = strings.ToUpper(contextSavingMode)
	if contextSavingMode != "AUTO" && contextSavingMode != "OFF" && contextSavingMode != "MEDIUM" && contextSavingMode != "HIGH" {
		logger.Panicln("Invalid context saving mode value provided")
	}
	return contextSavingMode
}

func GetLocalSearchModeFromContextSavingValue(contextSavingMode string, requestedFilesCount, searchLimit int) op_embed.SimilarFileSelectMode {
	var searchMode op_embed.SimilarFileSelectMode
	switch contextSavingMode {
	case "HIGH":
		searchMode = op_embed.SelectConservative
	case "MEDIUM":
		searchMode = op_embed.SelectConservative
	case "OFF":
		searchMode = op_embed.SelectAggressive
	case "AUTO":
		fallthrough
	default:
		if requestedFilesCount <= searchLimit {
			//for low requested file count - use aggressive search mode
			searchMode = op_embed.SelectAggressive
		} else {
			//for high requested file count - use conservative search mode
			searchMode = op_embed.SelectConservative
		}
	}
	return searchMode
}

// Get percentage of files to select from project-files for processing at stage 1 to lower LLM context pressure
// Return 2 values:
// Percentage of files to select from project-files.
// Percentage of files that will be randomized on selection in order to mitigate possible selection error.
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
