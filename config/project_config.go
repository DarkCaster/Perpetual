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
		return nil, fmt.Errorf("error loading project config: %s", err)
	}
	if err := processProjectConfig(storageObject); err != nil {
		return nil, fmt.Errorf("failed to validate project config: %s", err)
	}
	return &projectConfig{cfgValues: storageObject}, nil
}

func processProjectConfig(cfg map[string]interface{}) error {
	//validate against config template
	template := GetProjectConfigTemplate()
	if err := validateConfigAgainstTemplate(template, cfg); err != nil {
		return err
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
