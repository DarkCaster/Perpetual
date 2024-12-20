package config

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/DarkCaster/Perpetual/utils"
)

type projectConfig struct {
	cfgValues map[string]interface{}
}

func LoadProjectConfig(baseDir string) (Config, error) {
	storageObject := map[string]interface{}{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, ProjectConfigFile), &storageObject); err != nil {
		return nil, err
	}
	if err := processProjectConfig(storageObject); err != nil {
		return nil, err
	}
	return &projectConfig{cfgValues: storageObject}, nil
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
	return nil
}

func (o *projectConfig) Object(key string) map[string]interface{} {
	return o.cfgValues[key].(map[string]interface{})
}

func (o *projectConfig) Regexp(key string) *regexp.Regexp {
	return o.cfgValues[key].(*regexp.Regexp)
}

func (o *projectConfig) RegexpArray(key string) []*regexp.Regexp {
	return o.cfgValues[key].([]*regexp.Regexp)
}

func (o *projectConfig) String(key string) string {
	return o.cfgValues[key].(string)
}

func (o *projectConfig) StringArray(key string) []string {
	return interfaceToStringArray(o.cfgValues[key])
}

func (o *projectConfig) StringArray2D(key string) [][]string {
	return interfaceTo2DStringArray(o.cfgValues[key])
}
