package config

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/DarkCaster/Perpetual/utils"
)

type opImplementConfig struct {
	cfgValues map[string]interface{}
}

func LoadOpImplementConfig(baseDir string) (Config, error) {
	storageObject := map[string]interface{}{}
	if err := utils.LoadJsonFile(filepath.Join(baseDir, OpImplementConfigFile), &storageObject); err != nil {
		return nil, fmt.Errorf("error loading op_implement config: %s", err)
	}
	if err := processOpImplementConfig(storageObject); err != nil {
		return nil, fmt.Errorf("failed to validate op_implement config: %s", err)
	}
	return &opImplementConfig{cfgValues: storageObject}, nil
}

func processOpImplementConfig(cfg map[string]interface{}) error {
	//validate against config template
	template := GetImplementConfigTemplate()
	if err := validateConfigAgainstTemplate(template, cfg); err != nil {
		return err
	}
	//validate arrays with value-pairs
	if err := validateEvenStringArray(cfg[K_CodeTagsRx], K_CodeTagsRx); err != nil {
		return err
	}
	if err := validateEvenStringArray(cfg[K_FilenameTags], K_FilenameTags); err != nil {
		return err
	}
	if err := validateEvenStringArray(cfg[K_FilenameTagsRx], K_FilenameTagsRx); err != nil {
		return err
	}
	if err := validateNonEmptyStringArray(cfg[K_ImplementCommentsRx], K_ImplementCommentsRx); err != nil {
		return err
	}
	if err := validateNonEmptyStringArray(cfg[K_NoUploadCommentsRx], K_NoUploadCommentsRx); err != nil {
		return err
	}
	//precompile regexps
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_CodeTagsRx]), K_CodeTagsRx); err != nil {
		return err
	} else {
		cfg[K_CodeTagsRx] = rxArr
	}
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_FilenameTagsRx]), K_FilenameTagsRx); err != nil {
		return err
	} else {
		cfg[K_FilenameTagsRx] = rxArr
	}
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_ImplementCommentsRx]), K_ImplementCommentsRx); err != nil {
		return err
	} else {
		cfg[K_ImplementCommentsRx] = rxArr
	}
	if rxArr, err := compileRegexArray(interfaceToStringArray(cfg[K_NoUploadCommentsRx]), K_NoUploadCommentsRx); err != nil {
		return err
	} else {
		cfg[K_NoUploadCommentsRx] = rxArr
	}
	if rx, err := regexp.Compile(cfg[K_FilenameEmbedRx].(string)); err != nil {
		return fmt.Errorf("failed to compile %s regexp: %s", K_FilenameEmbedRx, err)
	} else {
		cfg[K_FilenameEmbedRx] = rx
	}
	return nil
}

func (o *opImplementConfig) Object(key string) map[string]interface{} {
	return o.cfgValues[key].(map[string]interface{})
}

func (o *opImplementConfig) Regexp(key string) *regexp.Regexp {
	return o.cfgValues[key].(*regexp.Regexp)
}

func (o *opImplementConfig) RegexpArray(key string) []*regexp.Regexp {
	return o.cfgValues[key].([]*regexp.Regexp)
}

func (o *opImplementConfig) String(key string) string {
	return o.cfgValues[key].(string)
}

func (o *opImplementConfig) StringArray(key string) []string {
	return interfaceToStringArray(o.cfgValues[key])
}

func (o *opImplementConfig) StringArray2D(key string) [][]string {
	return interfaceTo2DStringArray(o.cfgValues[key])
}
