package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

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

func LoadEnvFiles(filePaths ...string) (bool, error) {
	failedCount := 0
	for _, filePath := range filePaths {
		err := godotenv.Load(filePath)
		if err != nil {
			failedCount++
			if !os.IsNotExist(err) {
				return false, fmt.Errorf("Error loading env file %s: %s", filePath, err)
			}
		}
	}
	return failedCount < len(filePaths), nil
}
