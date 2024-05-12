package op_implement

import (
	"os"
	"path/filepath"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

// Stage4 saves the modified file contents from the results map to the actual files on disk
func Stage4(projectRootDir string, results map[string]string, projectFiles []string, logger logging.ILogger) {
	logger.Traceln("Stage4: Starting")
	defer logger.Traceln("Stage4: Finished")

	logger.Infoln("Running stage4: applying results")
	for filePathInitial, fileContent := range results {
		// getting leading directories, case insensitive, trying to fix its cases from closest match from the projectFiles
		leadingDirs := utils.CaseInsensitiveLeadingDirectoriesSearch(filePathInitial, projectFiles)
		// recursively create leading dirs
		fileDir := ""
		for _, dir := range leadingDirs {
			fileDir = filepath.Join(fileDir, dir)
			err := os.Mkdir(filepath.Join(projectRootDir, fileDir), 0755)
			if err != nil && !os.IsExist(err) {
				logger.Errorln("Failed to create directory:", fileDir, err)
			}
		}
		// getting final file path
		fileName := filepath.Base(filePathInitial)
		filePathFinal := filepath.Join(fileDir, fileName)
		logger.Infoln(filePathFinal)
		err := utils.SaveTextFile(filepath.Join(projectRootDir, filePathFinal), fileContent)
		if err != nil {
			logger.Errorln("Failed to save file:", err)
		}
	}

}
