package config

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

type configStorage struct {
	fileName  string
	cfgValues map[string]any
	logger    logging.ILogger
}

type configValue struct {
	config *configStorage
	value  any
	key    string
}

func (o *configStorage) get(key string) configValue {
	v, exist := o.cfgValues[key]
	if !exist {
		o.logger.Panicf("Config %s do not have value with key %s", o.fileName, key)
	}
	return configValue{config: o, value: v, key: key}
}

func as[T any](v configValue) T {
	r, ok := v.value.(T)
	if !ok {
		v.config.logger.Panicf("Config %s: failed to represent key %s as type %T", v.config.fileName, v.key, *new(T))
	}
	return r
}

func (o *configStorage) Regexp(key string) *regexp.Regexp {
	return as[*regexp.Regexp](o.get(key))
}

func (o *configStorage) RegexpArray(key string) []*regexp.Regexp {
	return utils.NewSlice(as[[]*regexp.Regexp](o.get(key))...)
}

func (o *configStorage) String(key string) string {
	return as[string](o.get(key))
}

func (o *configStorage) Tags(key string) utils.TagPair {
	sourceArray := o.get(key).value.([]any)
	if len(sourceArray) != 2 {
		panic(fmt.Errorf("invalid length of array %s: %d", key, len(sourceArray)))
	}
	return utils.TagPair{Left: sourceArray[0].(string), Right: sourceArray[1].(string)}
}

func (o *configStorage) Integer(key string) int {
	return int(as[float64](o.get(key)))
}

func (o *configStorage) Float(key string) float64 {
	return as[float64](o.get(key))
}

func (o *configStorage) TextMatcherString(key string) utils.TextMatcher[string] {
	return as[utils.TextMatcher[string]](o.get(key))
}

func (o *configStorage) TextMatcherInteger(key string) utils.TextMatcher[int] {
	return as[utils.TextMatcher[int]](o.get(key))
}

func LoadOpAnnotateConfig(baseDir string, logger logging.ILogger) Config {
	storageObject := map[string]any{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, OpAnnotateConfigFile), &storageObject); err != nil {
		logger.Panicf("Error loading %s config: %v", OpAnnotateConfigFile, err)
	}
	if err := processOpAnnotateConfig(storageObject); err != nil {
		logger.Panicf("Error processing %s config: %v", OpAnnotateConfigFile, err)
	}
	return &configStorage{cfgValues: storageObject, logger: logger, fileName: OpAnnotateConfigFile}
}

func LoadProjectConfig(baseDir string, logger logging.ILogger) Config {
	storageObject := map[string]any{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, ProjectConfigFile), &storageObject); err != nil {
		logger.Panicf("Error loading %s config: %v", ProjectConfigFile, err)
	}
	if err := processProjectConfig(storageObject); err != nil {
		logger.Panicf("Error processing %s config: %v", ProjectConfigFile, err)
	}
	return &configStorage{cfgValues: storageObject, logger: logger, fileName: ProjectConfigFile}
}

func LoadOpDocConfig(baseDir string, logger logging.ILogger) Config {
	storageObject := map[string]any{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, OpDocConfigFile), &storageObject); err != nil {
		logger.Panicf("Error loading %s config: %v", OpDocConfigFile, err)
	}
	if err := processOpDocConfig(storageObject); err != nil {
		logger.Panicf("Error processing %s config: %v", OpDocConfigFile, err)
	}
	return &configStorage{cfgValues: storageObject, logger: logger, fileName: OpDocConfigFile}
}

func LoadOpExplainConfig(baseDir string, logger logging.ILogger) Config {
	storageObject := map[string]any{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, OpExplainConfigFile), &storageObject); err != nil {
		logger.Panicf("Error loading %s config: %v", OpExplainConfigFile, err)
	}
	if err := processOpExplainConfig(storageObject); err != nil {
		logger.Panicf("Error processing %s config: %v", OpExplainConfigFile, err)
	}
	return &configStorage{cfgValues: storageObject, logger: logger, fileName: OpExplainConfigFile}
}

func LoadOpImplementConfig(baseDir string, logger logging.ILogger) Config {
	storageObject := map[string]any{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, OpImplementConfigFile), &storageObject); err != nil {
		logger.Panicf("Error loading %s config: %v", OpImplementConfigFile, err)
	}
	if err := processOpImplementConfig(storageObject); err != nil {
		logger.Panicf("Error processing %s config: %v", OpImplementConfigFile, err)
	}
	return &configStorage{cfgValues: storageObject, logger: logger, fileName: OpImplementConfigFile}
}

func LoadOpReportConfig(baseDir string, logger logging.ILogger) Config {
	storageObject := map[string]any{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, OpReportConfigFile), &storageObject); err != nil {
		logger.Panicf("Error loading %s config: %v", OpReportConfigFile, err)
	}
	if err := processOpReportConfig(storageObject); err != nil {
		logger.Panicf("Error processing %s config: %v", OpReportConfigFile, err)
	}
	return &configStorage{cfgValues: storageObject, logger: logger, fileName: OpReportConfigFile}
}

func processOpReportConfig(cfg map[string]any) error {
	//validate against config template
	template := GetReportConfigTemplate()
	if err := validateConfigAgainstTemplate(template, cfg); err != nil {
		return err
	}
	//validate arrays with value-pairs
	if err := validateEvenStringArray(cfg[K_ReportFilenameTags], K_ReportFilenameTags); err != nil {
		return err
	}
	return nil
}

func processOpImplementConfig(cfg map[string]any) error {
	//validate against config template
	template := GetImplementConfigTemplate()
	if err := validateConfigAgainstTemplate(template, cfg); err != nil {
		return err
	}
	//validate arrays with value-pairs
	if err := validateNonEmptyStringArray(cfg[K_ImplementCommentsRx], K_ImplementCommentsRx); err != nil {
		return err
	}
	//precompile regexps
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_ImplementCommentsRx]), K_ImplementCommentsRx); err != nil {
		return err
	} else {
		cfg[K_ImplementCommentsRx] = rxArr
	}
	if rx, err := regexp.Compile(cfg[K_ImplementFilenameEmbedRx].(string)); err != nil {
		return fmt.Errorf("%s must be a valid regexp: %s", K_ImplementFilenameEmbedRx, err)
	} else {
		cfg[K_ImplementFilenameEmbedRx] = rx
	}
	return nil
}

func processOpExplainConfig(cfg map[string]any) error {
	// validate against config template
	template := GetExplainConfigTemplate()
	if err := validateConfigAgainstTemplate(template, cfg); err != nil {
		return err
	}
	return nil
}

func processOpDocConfig(cfg map[string]any) error {
	//validate against config template
	template := GetDocConfigTemplate()
	if err := validateConfigAgainstTemplate(template, cfg); err != nil {
		return err
	}
	return nil
}

func processOpAnnotateConfig(cfg map[string]any) error {
	//validate against config template
	template := GetAnnotateConfigTemplate()
	if err := validateConfigAgainstTemplate(template, cfg); err != nil {
		return err
	}
	//custom validation of annotate-prompt array
	if err := validateOpAnnotateFilePrompts(cfg[K_AnnotateFilePrompts]); err != nil {
		return err
	}
	//convert K_AnnotateFilePrompts to the proper format
	matcher, err := utils.NewRxMatcher[string](2, cfg[K_AnnotateFilePrompts])
	if err != nil {
		return fmt.Errorf("failed to parse %s: %v", K_AnnotateFilePrompts, err)
	}
	//write back converted value for direct acceess
	cfg[K_AnnotateFilePrompts] = matcher
	return nil
}

func validateOpAnnotateFilePrompts(value any) error {
	arr, ok := value.([]any)
	if !ok {
		return fmt.Errorf("%s must be an array", K_AnnotateFilePrompts)
	}
	if len(arr) == 0 {
		return fmt.Errorf("%s must contain at least one element", K_AnnotateFilePrompts)
	}
	for i, outer := range arr {
		innerArr, ok := outer.([]any)
		if !ok {
			return fmt.Errorf("%s[%d] must be an array", K_AnnotateFilePrompts, i)
		}
		if len(innerArr) != 3 {
			return fmt.Errorf("%s[%d] must contain exactly 3 elements", K_AnnotateFilePrompts, i)
		}
		for j, inner := range innerArr {
			str, ok := inner.(string)
			if !ok {
				return fmt.Errorf("%s[%d][%d] must be a string", K_AnnotateFilePrompts, i, j)
			}
			if len(str) < 1 {
				return fmt.Errorf("%s[%d][%d] is empty", K_AnnotateFilePrompts, i, j)
			}
		}
	}
	return nil
}

func processProjectConfig(cfg map[string]any) error {
	//validate against config template
	template := GetProjectConfigTemplate()
	if err := validateConfigAgainstTemplate(template, cfg); err != nil {
		return err
	}

	//convert and Validate K_ProjectMdCodeMappings to the format usable for config
	codeBlocksMatcher, err := utils.NewRxMatcher[string](1, cfg[K_ProjectMdCodeMappings])
	if err != nil {
		return fmt.Errorf("failed to parse %s: %v", K_ProjectMdCodeMappings, err)
	}
	//write back converted value for direct acceess
	cfg[K_ProjectMdCodeMappings] = codeBlocksMatcher

	//convert and Validate K_ProjectFilesIncrModeMinLen to the format usable for config
	fileLenMatcher, err := utils.NewRxMatcher[int](1, cfg[K_ProjectFilesIncrModeMinLen])
	if err != nil {
		return fmt.Errorf("failed to parse %s: %v", K_ProjectFilesIncrModeMinLen, err)
	}
	//write back converted value for direct acceess
	cfg[K_ProjectFilesIncrModeMinLen] = fileLenMatcher

	//precompile regexps
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_ProjectFilesBlacklist]), K_ProjectFilesBlacklist); err != nil {
		return err
	} else {
		cfg[K_ProjectFilesBlacklist] = rxArr
	}
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_ProjectFilesWhitelist]), K_ProjectFilesWhitelist); err != nil {
		return err
	} else {
		cfg[K_ProjectFilesWhitelist] = rxArr
	}
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_ProjectTestFilesBlacklist]), K_ProjectTestFilesBlacklist); err != nil {
		return err
	} else {
		cfg[K_ProjectTestFilesBlacklist] = rxArr
	}
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_ProjectFilesIncrModeRx]), K_ProjectFilesIncrModeRx); err != nil {
		return err
	} else {
		cfg[K_ProjectFilesIncrModeRx] = rxArr
	}

	//validate arrays with value-pairs
	if err := validateEvenStringArray(cfg[K_ProjectFilenameTags], K_ProjectFilenameTags); err != nil {
		return err
	}
	if err := validateEvenStringArray(cfg[K_ProjectDeleteTags], K_ProjectDeleteTags); err != nil {
		return err
	}
	if err := validateEvenStringArray(cfg[K_ProjectFilenameTagsRx], K_ProjectFilenameTagsRx); err != nil {
		return err
	}
	if err := validateEvenStringArray(cfg[K_ProjectDeleteTagsRx], K_ProjectDeleteTagsRx); err != nil {
		return err
	}
	if err := validateNonEmptyStringArray(cfg[K_ProjectNoUploadCommentsRx], K_ProjectNoUploadCommentsRx); err != nil {
		return err
	}
	if err := validateEvenStringArray(cfg[K_ProjectCodeTagsRx], K_ProjectCodeTagsRx); err != nil {
		return err
	}
	//precompile regexps
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_ProjectFilenameTagsRx]), K_ProjectFilenameTagsRx); err != nil {
		return err
	} else {
		cfg[K_ProjectFilenameTagsRx] = rxArr
	}
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_ProjectDeleteTagsRx]), K_ProjectDeleteTagsRx); err != nil {
		return err
	} else {
		cfg[K_ProjectDeleteTagsRx] = rxArr
	}
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_ProjectNoUploadCommentsRx]), K_ProjectNoUploadCommentsRx); err != nil {
		return err
	} else {
		cfg[K_ProjectNoUploadCommentsRx] = rxArr
	}
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_ProjectCodeTagsRx]), K_ProjectCodeTagsRx); err != nil {
		return err
	} else {
		cfg[K_ProjectCodeTagsRx] = rxArr
	}
	return nil
}
