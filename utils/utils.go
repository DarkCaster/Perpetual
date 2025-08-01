package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"

	"github.com/DarkCaster/Perpetual/logging"
)

const DotEnvSuffixName = ".env"
const AnnotationsFileName = ".annotations.json"
const EmbeddingsFileName = ".embeddings.msgpack"
const StashesDirName = ".stash"

func FindProjectRoot(logger logging.ILogger) (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	projectRootDir := ""
	perpetualDir := ""
	// Try to get perpetualDir from ENV, or use default discovery logic
	if envDir, errEnv := GetEnvString("PERPETUAL_DIR"); errEnv == nil {
		projectRootDir = cwd
		perpetualDir = envDir
	} else {
		projectRootDir, perpetualDir, err = findProjectRoot(cwd, logger)
	}
	if err == nil {
		// Check presence of perpetualDir directory
		info, err := os.Stat(perpetualDir)
		if err != nil {
			return projectRootDir, perpetualDir, err
		}
		if !info.IsDir() {
			return projectRootDir, perpetualDir, fmt.Errorf(".perpetual directory is not a directory: %s", perpetualDir)
		}
		// Check that projectRootDir directory is not a symlink
		fileInfo, err := os.Lstat(projectRootDir)
		if err != nil {
			return projectRootDir, perpetualDir, err
		}
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			return projectRootDir, perpetualDir, fmt.Errorf("project directory is a symlink or reparse point: %s", projectRootDir)
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

// Used to produce completely new slice from another slice or single elements, where we need it.
// This is a bit lame but simple approach that allow to mess safely with original slice afterwards.
func NewSlice[T any](vars ...T) []T {
	// Give extra capacity to support adding a few elements right after creating the slice without repacking
	extraCapacity := int(math.Ceil(math.Log2(float64(len(vars)+1)))) + 1
	result := make([]T, len(vars), len(vars)+extraCapacity)
	copy(result, vars)
	return result
}
