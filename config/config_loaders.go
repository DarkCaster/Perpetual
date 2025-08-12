package config

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/DarkCaster/Perpetual/utils"
)

type opConfig struct {
	cfgValues map[string]interface{}
}

func (o *opConfig) Object(key string) map[string]interface{} {
	return o.cfgValues[key].(map[string]interface{})
}

func (o *opConfig) Regexp(key string) *regexp.Regexp {
	return o.cfgValues[key].(*regexp.Regexp)
}

func (o *opConfig) RegexpArray(key string) []*regexp.Regexp {
	return utils.NewSlice(o.cfgValues[key].([]*regexp.Regexp)...)
}

func (o *opConfig) String(key string) string {
	return o.cfgValues[key].(string)
}

func (o *opConfig) StringArray(key string) []string {
	return interfaceToStringArray(o.cfgValues[key])
}

func (o *opConfig) StringArray2D(key string) [][]string {
	return interfaceTo2DStringArray(o.cfgValues[key])
}

func (o *opConfig) Integer(key string) int {
	return int(o.cfgValues[key].(float64))
}

func (o *opConfig) Float(key string) float64 {
	return o.cfgValues[key].(float64)
}

func LoadOpAnnotateConfig(baseDir string) (Config, error) {
	storageObject := map[string]interface{}{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, OpAnnotateConfigFile), &storageObject); err != nil {
		return nil, err
	}
	if err := processOpAnnotateConfig(storageObject); err != nil {
		return nil, err
	}
	return &opConfig{cfgValues: storageObject}, nil
}

func LoadProjectConfig(baseDir string) (Config, error) {
	storageObject := map[string]interface{}{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, ProjectConfigFile), &storageObject); err != nil {
		return nil, err
	}
	if err := processProjectConfig(storageObject); err != nil {
		return nil, err
	}
	return &opConfig{cfgValues: storageObject}, nil
}

func LoadOpDocConfig(baseDir string) (Config, error) {
	storageObject := map[string]interface{}{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, OpDocConfigFile), &storageObject); err != nil {
		return nil, err
	}
	if err := processOpDocConfig(storageObject); err != nil {
		return nil, err
	}
	return &opConfig{cfgValues: storageObject}, nil
}

func LoadOpExplainConfig(baseDir string) (Config, error) {
	storageObject := map[string]interface{}{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, OpExplainConfigFile), &storageObject); err != nil {
		return nil, err
	}
	if err := processOpExplainConfig(storageObject); err != nil {
		return nil, err
	}
	return &opConfig{cfgValues: storageObject}, nil
}

func LoadOpImplementConfig(baseDir string) (Config, error) {
	storageObject := map[string]interface{}{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, OpImplementConfigFile), &storageObject); err != nil {
		return nil, err
	}
	if err := processOpImplementConfig(storageObject); err != nil {
		return nil, err
	}
	return &opConfig{cfgValues: storageObject}, nil
}

func LoadOpReportConfig(baseDir string) (Config, error) {
	storageObject := map[string]interface{}{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, OpReportConfigFile), &storageObject); err != nil {
		return nil, err
	}
	if err := processOpReportConfig(storageObject); err != nil {
		return nil, err
	}
	return &opConfig{cfgValues: storageObject}, nil
}

func processOpReportConfig(cfg map[string]interface{}) error {
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

func processOpImplementConfig(cfg map[string]interface{}) error {
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
	if rx, err := regexp.Compile(cfg[K_FilenameEmbedRx].(string)); err != nil {
		return fmt.Errorf("%s must be a valid regexp: %s", K_FilenameEmbedRx, err)
	} else {
		cfg[K_FilenameEmbedRx] = rx
	}
	return nil
}

func processOpExplainConfig(cfg map[string]interface{}) error {
	// validate against config template
	template := GetExplainConfigTemplate()
	if err := validateConfigAgainstTemplate(template, cfg); err != nil {
		return err
	}
	return nil
}

func processOpDocConfig(cfg map[string]interface{}) error {
	//validate against config template
	template := GetDocConfigTemplate()
	if err := validateConfigAgainstTemplate(template, cfg); err != nil {
		return err
	}
	return nil
}

func processOpAnnotateConfig(cfg map[string]interface{}) error {
	//validate against config template
	template := GetAnnotateConfigTemplate()
	if err := validateConfigAgainstTemplate(template, cfg); err != nil {
		return err
	}
	//custom validation of annotate-prompt array
	if err := validateOpAnnotateStage1Prompts(cfg[K_AnnotateStage1Prompts]); err != nil {
		return err
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
		if len(innerArr) != 3 {
			return fmt.Errorf("%s[%d] must contain exactly 3 elements", K_AnnotateStage1Prompts, i)
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

func processProjectConfig(cfg map[string]interface{}) error {
	//validate against config template
	template := GetProjectConfigTemplate()
	if err := validateConfigAgainstTemplate(template, cfg); err != nil {
		return err
	}
	// Validate K_ProjectMdCodeMappings
	mappings := interfaceTo2DStringArray(cfg[K_ProjectMdCodeMappings])
	for i, mapping := range mappings {
		if len(mapping) != 2 {
			return fmt.Errorf("%s[%d] must contain exactly 2 elements", K_ProjectMdCodeMappings, i)
		}
		// Try compiling first element as regexp
		if _, err := regexp.Compile(mapping[0]); err != nil {
			return fmt.Errorf("%s[%d][0] must be a valid regexp: %s", K_ProjectMdCodeMappings, i, err)
		}
	}
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
	//validate arrays with value-pairs
	if err := validateEvenStringArray(cfg[K_ProjectFilenameTags], K_ProjectFilenameTags); err != nil {
		return err
	}
	if err := validateEvenStringArray(cfg[K_ProjectFilenameTagsRx], K_ProjectFilenameTagsRx); err != nil {
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
