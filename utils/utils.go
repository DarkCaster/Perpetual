package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/joho/godotenv"
)

const DotEnvFileName = ".env"
const AnnotationsFileName = ".annotations.json"
const StashesDirName = ".stash"

func FindProjectRoot(logger logging.ILogger) (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	return findProjectRoot(cwd, logger)
}

func findProjectRoot(startDir string, logger logging.ILogger) (string, string, error) {
	perpetualDir := filepath.Join(startDir, ".perpetual")
	_, err := os.Stat(perpetualDir)
	if err == nil {
		// .perpetual directory found in the current directory
		return startDir, perpetualDir, nil
	} else if !os.IsNotExist(err) {
		// Error other than directory not found
		return "", "", err
	}

	// .perpetual directory not found, check the parent directory
	parentDir := filepath.Dir(startDir)
	if parentDir == startDir {
		// Reached the root directory
		return "", "", os.ErrNotExist
	}

	logger.Warnln("Directory .perpetual not found in", startDir)
	return findProjectRoot(parentDir, logger)
}

func LoadEnvFile(filePath string) error {
	return godotenv.Load(filePath)
}

func CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	checksum := hash.Sum(nil)
	return fmt.Sprintf("%x", checksum), nil
}
