package op_project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_implement"
	"github.com/DarkCaster/Perpetual/utils"
)

func Init(version, lang string, envExamples bool, logger logging.ILogger) error {
	if lang == "" {
		return errors.New("language must be provided")
	}

	logger.Debugln("Parsed language:", lang)

	cwd, err := os.Getwd()
	if err != nil {
		logger.Panicln("Error getting current working directory:", err)
	}

	perpetualDir := ""
	if envDir, errEnv := utils.GetEnvString("PERPETUAL_DIR"); errEnv == nil {
		perpetualDir = envDir
	} else {
		perpetualDir = filepath.Join(cwd, ".perpetual")
	}

	info, err := os.Stat(perpetualDir)
	if err == nil {
		if !info.IsDir() {
			logger.Panicln(".perpetual already exists and it is not a directory!")
		}
		logger.Warnln("Directory .perpetual already exists, replacing configs:", perpetualDir)
	} else if !os.IsNotExist(err) {
		logger.Panicln("Error checking for .perpetual directory:", err)
	} else {
		logger.Infoln("Creating .perpetual directory:", perpetualDir)
		err = os.Mkdir(perpetualDir, 0755)
		if err != nil {
			logger.Panicln("Error creating .perpetual directory:", err)
		}
	}

	const DotEnvMaskName = "*.env"

	// Create a .gitignore file in the .perpetual directory
	logger.Traceln("Creating .gitignore file")

	gitignoreText := fmt.Sprintf("/%s\n/%s\n/%s\n/%s\n/%s*\n/%s\n/%s\n", DotEnvMaskName, utils.AnnotationsFileName, utils.EmbeddingsFileName, utils.LockFileName, llm.LLMRawLogFile, utils.StashesDirName, op_implement.StateFileName)
	_, err = utils.SaveTextFile(filepath.Join(perpetualDir, ".gitignore"), gitignoreText)
	if err != nil {
		logger.Panicln("Error creating .gitignore file:", err)
	}

	if envExamples {
		logger.Traceln("Creating env example files")

		for _, example := range GetEnvExampleCatalogWithVersion(version) {
			if _, err = utils.SaveTextFile(filepath.Join(perpetualDir, example.Filename), example.Content); err != nil {
				logger.Panicln("Error creating env example file:", err)
			}
		}
	}

	logger.Tracef("Creating %s file", descriptionTemplateFileName)
	if _, err = utils.SaveTextFile(filepath.Join(perpetualDir, descriptionTemplateFileName), descriptionTemplate); err != nil {
		logger.Panicf("Error creating %s file: %v", descriptionTemplateFileName, err)
	}

	// Create a prompt-files based on the selected language
	logger.Debugln("Creating prompt-files")
	promptsObj, err := newPrompts(lang)
	if err != nil {
		return err
	}

	saveConfig := func(filePath string, v any) {
		logger.Traceln("Saving config:", filePath)
		err = utils.SaveJsonFile(filepath.Join(perpetualDir, filePath), v)
		if err != nil {
			logger.Panicln(err)
		}
	}

	// Save operation-config files
	saveConfig(config.OpAnnotateConfigFile, promptsObj.GetAnnotateConfig())
	saveConfig(config.OpImplementConfigFile, promptsObj.GetImplementConfig())
	saveConfig(config.OpDocConfigFile, promptsObj.GetDocConfig())
	saveConfig(config.OpReportConfigFile, promptsObj.GetReportConfig())
	saveConfig(config.OpExplainConfigFile, promptsObj.GetExplainConfig())

	// Save project-config file
	saveConfig(config.ProjectConfigFile, promptsObj.GetProjectConfig())

	obsoleteFiles := []string{
		"filename_embed_regexp.json",
		"filename_tags_regexps.json",
		"filename_tags.json",
		"no_upload_comment_regexps.json",
		"output_tags_regexps.json",
		"project_files_blacklist.json",
		"project_files_to_markdown_lang_mappings.json",
		"project_files_whitelist.json",
		"reasonings_tags_regexps.json",
		"reasonings_tags.json",
	}

	for _, file := range obsoleteFiles {
		filePath := filepath.Join(perpetualDir, file)
		if _, err := os.Stat(filePath); err == nil {
			logger.Warnln("Obsolete config file found (remove it manually):", file)
		}
	}

	obsoleteDirs := []string{"prompts"}

	for _, dir := range obsoleteDirs {
		dirPath := filepath.Join(perpetualDir, dir)
		if _, err := os.Stat(dirPath); err == nil {
			logger.Warnln("Obsolete directory found (remove it manually):", dir)
		}
	}

	return nil
}
