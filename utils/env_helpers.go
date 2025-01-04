package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/joho/godotenv"
)

func BackupEnvVars(vars ...string) map[string]string {
	result := make(map[string]string)
	for _, name := range vars {
		value, err := GetEnvString(name)
		if err != nil {
			continue
		}
		result[name] = value
	}
	return result
}

func UnsetEnvVars(vars ...string) error {
	var lastErr error = nil
	for _, name := range vars {
		if err := os.Unsetenv(name); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func RestoreEnvVars(backup map[string]string) error {
	var lastErr error = nil
	for key, value := range backup {
		if err := os.Setenv(key, value); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func GetEnvString(vars ...string) (string, error) {
	for _, v := range vars {
		if value := os.Getenv(v); value != "" {
			return value, nil
		}
	}
	return "", fmt.Errorf("none of the environment variables were found: %v", vars)
}

func GetEnvUpperString(vars ...string) (string, error) {
	result, err := GetEnvString(vars...)
	if err != nil {
		return result, err
	}
	return strings.ToUpper(result), err
}

func GetEnvInt(vars ...string) (int, error) {
	for _, v := range vars {
		if value := os.Getenv(v); value != "" {
			if intValue, err := strconv.Atoi(value); err == nil {
				return intValue, nil
			}
		}
	}
	return 0, fmt.Errorf("none of the environment variables were found or could be converted to int: %v", vars)
}

func GetEnvFloat(vars ...string) (float64, error) {
	for _, v := range vars {
		if value := os.Getenv(v); value != "" {
			if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
				return floatValue, nil
			}
		}
	}
	return 0, fmt.Errorf("none of the environment variables were found or could be converted to float: %v", vars)
}

func LoadEnvFiles(logger logging.ILogger, filePaths ...string) {
	for _, filePath := range filePaths {
		err := godotenv.Load(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Warnln("Not loading missing env file:", filePath)
			} else {
				logger.Panicf("Failed to load env-file %s: %s", filePath, err)
			}
		} else {
			logger.Infoln("Loaded env file:", filePath)
		}
	}
}
