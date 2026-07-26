package op_project

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

func loadConfigs(descFile string, logger logging.ILogger) (string, string, config.Config) {
	// Initialize: detect work directories, load .env file with LLM settings, load file filtering regexps
	projectRootDir, perpetualDir, err := utils.FindProjectRoot(logger, false)
	if err != nil {
		logger.Panicln("Error finding project root directory:", err)
	}

	globalConfigDir, err := utils.FindConfigDir()
	if err != nil {
		logger.Panicln("Error finding perpetual config directory:", err)
	}

	logger.Infoln("Project root directory:", projectRootDir)
	logger.Debugln("Perpetual directory:", perpetualDir)

	utils.LoadEnvFiles(logger, perpetualDir, globalConfigDir)

	//load json config files for project and operations, will panic if it cannot be loaded or parsed
	projectConfig := config.LoadProjectConfig(perpetualDir, logger)
	config.LoadOpAnnotateConfig(perpetualDir, logger)
	config.LoadOpDocConfig(perpetualDir, logger)
	config.LoadOpExplainConfig(perpetualDir, logger)
	config.LoadOpImplementConfig(perpetualDir, logger)
	config.LoadOpReportConfig(perpetualDir, logger)

	//test load of project description file
	wrn := ""
	if descFile == "" {
		_, wrn, err = utils.LoadTextFile(filepath.Join(perpetualDir, config.ProjectDescriptionFile))
		if err != nil {
			if os.IsNotExist(err) {
				logger.Infoln("Not loading missing project description file (description.md)")
			} else {
				logger.Panicln("Failed to load project description file:", err)
			}
		}
		if wrn != "" {
			logger.Warnf("%s: %s", config.ProjectDescriptionFile, wrn)
		}
	} else if strings.ToLower(descFile) != "disabled" {
		_, wrn, err = utils.LoadTextFile(descFile)
		if err != nil {
			logger.Panicln("Failed to load project description file:", err)
		}
		if wrn != "" {
			logger.Warnf("%s: %s", descFile, wrn)
		}
	} else {
		logger.Infoln("Loading of project description file (description.md) is disabled")
	}

	return projectRootDir, perpetualDir, projectConfig
}

func CheckProjectConfig(descFile string, logger logging.ILogger) {
	_, perpetualDir, _ := loadConfigs(descFile, logger)
	fmt.Println(perpetualDir)
}

func FetchProjectFiles(descFile, userFilterFile string, includeTests, listFiles bool, logger logging.ILogger) (string, []string) {
	projectRootDir, perpetualDir, projectConfig := loadConfigs(descFile, logger)

	// Preparation of project files
	logger.Infoln("Fetching project files")
	fileNames, _, err := utils.GetProjectFileList(
		projectRootDir,
		perpetualDir,
		projectConfig.RegexpArray(config.K_ProjectFilesWhitelist),
		projectConfig.RegexpArray(config.K_ProjectFilesBlacklist))

	if err != nil {
		logger.Panicln("Error getting project file-list:", err)
	}

	// Check fileNames array for case collisions
	if !utils.CheckFilenameCaseCollisions(fileNames) {
		//list current files
		if listFiles {
			for _, file := range fileNames {
				fmt.Println(file)
			}
		}
		logger.Panicln("Filename case collisions detected in project files")
	}

	// File names and dir-names must not contain path separators characters
	if !utils.CheckForPathSeparatorsInFilenames(fileNames) {
		//list current files
		if listFiles {
			for _, file := range fileNames {
				fmt.Println(file)
			}
		}
		logger.Panicln("Invalid characters detected in project filenames or directories: / and \\ characters are not allowed!")
	}

	// Filter project files with unittest- and user- filters
	var userBlacklist []*regexp.Regexp
	if userFilterFile != "" {
		userBlacklist, err = utils.AppendUserFilterFromFile(userFilterFile, userBlacklist)
		if err != nil {
			logger.Panicln("Error processing user blacklist-filter:", err)
		}
	}

	if !includeTests {
		userBlacklist = append(userBlacklist, projectConfig.RegexpArray(config.K_ProjectTestFilesBlacklist)...)
	}

	fileNames, droppedFiles := utils.FilterFilesWithBlacklist(fileNames, userBlacklist)
	if len(droppedFiles) > 0 {
		logger.Infoln("Number of blacklisted files with unit-tests and/or user-provided filters:", len(droppedFiles))
		slices.Sort(droppedFiles)
		for _, file := range droppedFiles {
			logger.Debugln("Filtered-out:", file)
		}
	}

	//list currently available files
	slices.Sort(fileNames)
	if listFiles {
		for _, file := range fileNames {
			fmt.Println(file)
		}
	}

	return projectRootDir, fileNames
}

func CheckFilesRead(descFile, userFilterFile string, includeTests bool, logger logging.ILogger) {
	projectRootDir, fileNames := FetchProjectFiles(descFile, userFilterFile, includeTests, false, logger)

	logger.Debugln("Reading project files")
	fileContent := map[string]string{}
	failedFiles := []string{}
	warnedFiles := []string{}
	for _, file := range fileNames {
		logger.Traceln("Reading file:", file)
		content, wrn, err := utils.LoadTextFile(filepath.Join(projectRootDir, file))
		if err != nil {
			logger.Errorf("%s: %v", file, err)
			failedFiles = append(failedFiles, file)
			continue
		}
		fileContent[file] = content
		if wrn != "" {
			logger.Warnf("%s: %s", file, wrn)
			warnedFiles = append(warnedFiles, file)
		}
	}

	slices.Sort(failedFiles)
	for _, file := range failedFiles {
		fmt.Println(file)
	}
	if len(failedFiles) > 0 {
		logger.Panicln("Reading of some project files was unsuccessful")
	}
}

func CheckFilesReadAsAscii(descFile, userFilterFile string, includeTests bool, logger logging.ILogger) {
	projectRootDir, fileNames := FetchProjectFiles(descFile, userFilterFile, includeTests, false, logger)

	logger.Debugln("Reading project files")
	fileContent := map[string]string{}
	failedFiles := []string{}
	warnedFiles := []string{}
	for _, file := range fileNames {
		logger.Traceln("Reading file:", file)
		content, wrn, err := utils.LoadTextFile(filepath.Join(projectRootDir, file))
		if err != nil {
			logger.Errorf("%s: %v", file, err)
			failedFiles = append(failedFiles, file)
			continue
		}
		fileContent[file] = content
		if wrn != "" {
			logger.Warnf("%s: %s", file, wrn)
			warnedFiles = append(warnedFiles, file)
		}
	}

	loadedFiles := slices.Collect(maps.Keys(fileContent))
	slices.Sort(loadedFiles)
	for _, file := range loadedFiles {
		// Check if content only contains ASCII characters (0-127)
		line := 1
		linePos := 1
		content := fileContent[file]
		for b, r := range content {
			if r == '\n' {
				line++
				linePos = 1
			}
			if r > 127 {
				logger.Warnf("%s: non-ASCII character found at byte %d (line %d, pos %d)", file, b, line, linePos)
				failedFiles = append(failedFiles, file)
				break
			}
			linePos++
		}
	}

	slices.Sort(failedFiles)
	for _, file := range failedFiles {
		fmt.Println(file)
	}
	if len(failedFiles) > 0 {
		logger.Panicln("Some files contain non ASCII content, or cannot be read as text at all")
	}
}

func CheckFilesAndSaveAsUTF(descFile, userFilterFile string, includeTests bool, logger logging.ILogger) {
	projectRootDir, fileNames := FetchProjectFiles(descFile, userFilterFile, includeTests, false, logger)

	logger.Debugln("Reading project files")
	fileContent := map[string]string{}
	failedFiles := []string{}
	warnedFiles := []string{}
	for _, file := range fileNames {
		logger.Traceln("Reading file:", file)
		content, wrn, err := utils.LoadTextFile(filepath.Join(projectRootDir, file))
		if err != nil {
			logger.Errorf("%s: %v", file, err)
			failedFiles = append(failedFiles, file)
			continue
		}
		fileContent[file] = content
		if wrn != "" {
			logger.Warnf("%s: %s", file, wrn)
			warnedFiles = append(warnedFiles, file)
		}
	}

	slices.Sort(warnedFiles)
	for _, file := range warnedFiles {
		if err := utils.SaveTextFileAsUTF8(filepath.Join(projectRootDir, file), fileContent[file]); err != nil {
			logger.Panicln("Failed to save text file as UTF8:", err)
		}
		fmt.Println(file)
	}
}
