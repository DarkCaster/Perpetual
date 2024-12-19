package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/DarkCaster/Perpetual/logging"
)

const DotEnvFileName = ".env"
const AnnotationsFileName = ".annotations.json"
const StashesDirName = ".stash"

func FindProjectRoot(logger logging.ILogger) (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	projectRootDir, perpetualDir, err := findProjectRoot(cwd, logger)
	// Check if projectRootDir is a symbolic link
	if err == nil {
		fileInfo, err := os.Lstat(projectRootDir)
		if err != nil {
			return projectRootDir, perpetualDir, err
		}
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			return projectRootDir, perpetualDir, fmt.Errorf("dir is a symlink or reparse point: %s", projectRootDir)
		}
	}
	return projectRootDir, perpetualDir, err
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

func FindConfigDir() (string, error) {
	basedir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(basedir, "Perpetual"), nil
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
