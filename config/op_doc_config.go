package config

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/DarkCaster/Perpetual/utils"
)

type opDocConfig struct {
	cfgValues map[string]interface{}
}

func LoadOpDocConfig(baseDir string) (Config, error) {
	storageObject := map[string]interface{}{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, OpDocConfigFile), &storageObject); err != nil {
		return nil, fmt.Errorf("error loading op_doc config: %s", err)
	}
	if err := processOpDocConfig(storageObject); err != nil {
		return nil, fmt.Errorf("failed to validate op_doc config: %s", err)
	}
	return &opDocConfig{cfgValues: storageObject}, nil
}

func processOpDocConfig(cfg map[string]interface{}) error {
	//validate against config template
	template := GetDocConfigTemplate()
	if err := validateConfigAgainstTemplate(template, cfg); err != nil {
		return err
	}
	//validate arrays with value-pairs
	if err := validateEvenStringArray(cfg[K_FilenameTags], K_FilenameTags); err != nil {
		return err
	}
	if err := validateEvenStringArray(cfg[K_FilenameTagsRx], K_FilenameTagsRx); err != nil {
		return err
	}
	if err := validateNonEmptyStringArray(cfg[K_NoUploadCommentsRx], K_NoUploadCommentsRx); err != nil {
		return err
	}
	//precompile regexps
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_FilenameTagsRx]), K_FilenameTagsRx); err != nil {
		return err
	} else {
		cfg[K_FilenameTagsRx] = rxArr
	}
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_NoUploadCommentsRx]), K_NoUploadCommentsRx); err != nil {
		return err
	} else {
		cfg[K_NoUploadCommentsRx] = rxArr
	}
	return nil
}

func (o *opDocConfig) Object(key string) map[string]interface{} {
	return o.cfgValues[key].(map[string]interface{})
}

func (o *opDocConfig) Regexp(key string) *regexp.Regexp {
	return o.cfgValues[key].(*regexp.Regexp)
}

func (o *opDocConfig) RegexpArray(key string) []*regexp.Regexp {
	return o.cfgValues[key].([]*regexp.Regexp)
}

func (o *opDocConfig) String(key string) string {
	return o.cfgValues[key].(string)
}

func (o *opDocConfig) StringArray(key string) []string {
	return interfaceToStringArray(o.cfgValues[key])
}

func (o *opDocConfig) StringArray2D(key string) [][]string {
	return interfaceTo2DStringArray(o.cfgValues[key])
}
