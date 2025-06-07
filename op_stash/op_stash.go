package op_stash

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

type Stash struct {
	OriginalFiles []FileEntry `json:"original_files"`
	ModifiedFiles []FileEntry `json:"modified_files"`
}

type FileEntry struct {
	Filename string `json:"filename"`
	Contents string `json:"contents"`
}

const OpName = "stash"
const OpDesc = "Rollback or re-apply generated code"

func stashFlags() *flag.FlagSet {
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)
	return flags
}

func Run(args []string, innerCall bool, logger logging.ILogger) {
	var help, list, verbose, apply, rollback, trace, listFiles bool
	var name, fileName, targetFile string

	// Parse flags for the "stash" operation
	flags := stashFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.BoolVar(&list, "l", false, "List current stashes")
	flags.BoolVar(&apply, "a", false, "Apply changes of a stash")
	flags.BoolVar(&rollback, "r", false, "Rollback changes of a stash")
	flags.BoolVar(&listFiles, "lf", false, "List files in a stash")
	flags.StringVar(&name, "n", "latest", "Set stash name to apply or revert")
	flags.StringVar(&fileName, "f", "", "Select single file to apply or revert from stash")
	flags.StringVar(&targetFile, "t", "", "Target file where selected single file from stash will be saved, relative to project root")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if verbose {
		logger.EnableLevel(logging.DebugLevel)
	}
	if trace {
		logger.EnableLevel(logging.DebugLevel)
		logger.EnableLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'stash' operation")
	logger.Traceln("Args:", args)

	if !apply && !rollback && !list && !listFiles {
		help = true
	}

	if help {
		usage.PrintOperationUsage("", flags)
		return
	}

	outerCallLogger := logger.Clone()
	if innerCall {
		outerCallLogger.DisableLevel(logging.ErrorLevel)
		outerCallLogger.DisableLevel(logging.WarnLevel)
		outerCallLogger.DisableLevel(logging.InfoLevel)
	}

	projectRootDir, perpetualDir, err := utils.FindProjectRoot(outerCallLogger)
	if err != nil {
		outerCallLogger.Panicln("Error finding project root directory:", err)
	}

	outerCallLogger.Infoln("Project root directory:", projectRootDir)
	outerCallLogger.Debugln("Perpetual directory:", perpetualDir)

	stashDir := filepath.Join(perpetualDir, utils.StashesDirName)

	if _, err := os.Stat(stashDir); os.IsNotExist(err) {
		err := os.Mkdir(stashDir, 0755)
		if err != nil {
			outerCallLogger.Panicln("Error creating stash directory:", err)
		}
	}

	stashes, err := os.ReadDir(stashDir)
	if err != nil {
		outerCallLogger.Panicln("Error reading stash directory:", err)
	}

	fileStashes := make([]os.DirEntry, 0, len(stashes))
	for _, entry := range stashes {
		if entry.Type().IsRegular() {
			fileStashes = append(fileStashes, entry)
		}
	}
	stashes = fileStashes

	// Check if no stash files are present
	if len(stashes) == 0 {
		logger.Infoln("No stashes found.")
		return
	}

	if list {
		// Print stashes directly to console
		for _, stash := range stashes {
			fmt.Println(strings.TrimSuffix(stash.Name(), ".json"))
		}
		return
	}

	if name == "latest" {
		name = stashes[len(stashes)-1].Name()
	}
	logger.Infoln("Processing stash:", name)
	// Add .json extension if not present
	if filepath.Ext(name) != ".json" {
		name += ".json"
	}
	stashFile := filepath.Join(stashDir, name)
	if _, err := os.Stat(stashFile); os.IsNotExist(err) {
		outerCallLogger.Panicln("Stash not found:", name)
	}
	var stash Stash
	err = utils.LoadJsonFile(stashFile, &stash)
	if err != nil {
		outerCallLogger.Panicln("Error loading stash:", err)
	}

	if listFiles {
		logger.Infoln("Listing files in stash:", name)
		for _, entry := range stash.OriginalFiles {
			fmt.Println("Original:", entry.Filename)
		}
		for _, entry := range stash.ModifiedFiles {
			fmt.Println("Modified:", entry.Filename)
		}
		return
	}

	writeFiles := func(entries []FileEntry) {
		for _, entry := range entries {
			if fileName != "" && entry.Filename != fileName {
				continue
			}
			target := entry.Filename
			if fileName != "" && targetFile != "" {
				target = targetFile
				target, err := utils.MakePathRelative(projectRootDir, target, true)
				if err != nil {
					outerCallLogger.Panicln("Requested file is not inside project root", target)
				}
			}
			// Get base dir
			targetDir := filepath.Dir(target)
			// Split the corrected targetDir into path components
			pathComponents := strings.Split(targetDir, string(os.PathSeparator))
			if len(pathComponents) > 0 && pathComponents[0] == "." {
				pathComponents = pathComponents[1:]
			}
			// Recursively create leading dirs
			fileDir := ""
			for _, dir := range pathComponents {
				fileDir = filepath.Join(fileDir, dir)
				err := os.Mkdir(filepath.Join(projectRootDir, fileDir), 0755)
				if err != nil && !os.IsExist(err) {
					outerCallLogger.Errorln("Failed to create directory:", fileDir, err)
				}
			}
			// Write file
			outerCallLogger.Infoln(target)
			err := utils.SaveTextFile(filepath.Join(projectRootDir, target), entry.Contents)
			if err != nil {
				outerCallLogger.Errorln("Failed to save file:", err)
			}
		}
	}

	if apply {
		logger.Infoln("Applying changes")
		writeFiles(stash.ModifiedFiles)
	} else if rollback {
		logger.Infoln("Rolling back changes")
		writeFiles(stash.OriginalFiles)
	}
}

// This function creates new stash from code generation results
func CreateStash(results map[string]string, projectFiles []string, logger logging.ILogger) string {
	logger.Traceln("CreateStash: Starting")
	defer logger.Traceln("CreateStash: Finished")

	projectRootDir, perpetualDir, err := utils.FindProjectRoot(logger)
	if err != nil {
		logger.Panicln("Error finding project root directory:", err)
	}

	stashDir := filepath.Join(perpetualDir, utils.StashesDirName)

	if _, err := os.Stat(stashDir); os.IsNotExist(err) {
		err := os.Mkdir(stashDir, 0755)
		if err != nil {
			logger.Panicln("Error creating stash directory:", err)
		}
	}

	logger.Infoln("Creating new stash from generated results")
	var stash Stash
	for filePathInitial, fileContent := range results {
		// getting leading directories, case insensitive, trying to fix its cases from closest match from the projectFiles
		leadingDirs := utils.CaseInsensitiveLeadingDirectoriesSearch(filePathInitial, projectFiles)
		// recursively create leading dirs
		fileDir := ""
		for _, dir := range leadingDirs {
			fileDir = filepath.Join(fileDir, dir)
		}
		// getting final file path
		fileName := filepath.Base(filePathInitial)
		filePathFinal := filepath.Join(fileDir, fileName)
		// Check file exist on disk
		_, err := os.Stat(filepath.Join(projectRootDir, filePathFinal))
		if err == nil {
			// Read file
			backup, err := utils.LoadTextFile(filepath.Join(projectRootDir, filePathFinal))
			if err != nil {
				logger.Errorf("Error reading project file for backing up:", err)
			}
			// Store file content to stash original files
			stash.OriginalFiles = append(stash.OriginalFiles, FileEntry{Filename: filePathFinal, Contents: backup})
		}
		// Store file content to stash modified files
		stash.ModifiedFiles = append(stash.ModifiedFiles, FileEntry{Filename: filePathFinal, Contents: fileContent})
	}

	if len(stash.OriginalFiles) > 0 {
		logger.Debugln("Files backed up:")
		for _, entry := range stash.OriginalFiles {
			logger.Debugln(entry.Filename)
		}
	}

	stashFileName := time.Now().Format("2006-01-02_15-04-05")
	err = utils.SaveJsonFile(filepath.Join(stashDir, stashFileName+".json"), stash)
	if err != nil {
		logger.Panicln("Error saving stash:", err)
	}

	return stashFileName
}
