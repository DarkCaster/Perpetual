package op_stash

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const StashVersion = 1

type Stash struct {
	Version int         `json:"version"`
	Files   []FileEntry `json:"files"`
}

type FileEntry struct {
	Filename string    `json:"filename"`
	Original FileState `json:"original"`
	Modified FileState `json:"modified"`
}

type FileState struct {
	Exists   bool   `json:"exists"`
	Contents string `json:"contents,omitempty"`
}

const OpName = "stash"
const OpDesc = "Rollback or re-apply generated code"

func stashFlags() *flag.FlagSet {
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)
	return flags
}

func validateStash(stash Stash) error {
	if stash.Version != StashVersion {
		return fmt.Errorf("unsupported stash version: %d, expected: %d", stash.Version, StashVersion)
	}
	return nil
}

func loadStash(stashFile string) (Stash, error) {
	var stash Stash
	err := utils.LoadJsonFile(stashFile, &stash)
	if err != nil {
		return stash, err
	}
	if err := validateStash(stash); err != nil {
		return stash, err
	}
	return stash, nil
}

func resolveTargetFile(projectRootDir, sourceFile, targetFile string) (string, error) {
	target := sourceFile
	if targetFile != "" {
		target = targetFile
	}
	return utils.MakePathRelative(projectRootDir, target, true)
}

func applyFileState(projectRootDir string, filename string, state FileState, logger logging.ILogger) {
	targetPath := filepath.Join(projectRootDir, filename)

	if !state.Exists {
		logger.Infoln("Deleting:", filename)
		if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
			logger.Errorf("Failed to delete file %s: %v", filename, err)
		}
		return
	}

	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		logger.Errorf("Failed to create directory %s: %v", targetDir, err)
		return
	}

	logger.Infoln(filename)
	wrn, err := utils.SaveTextFile(targetPath, state.Contents)
	if err != nil {
		logger.Errorf("Failed to save file %s: %v", filename, err)
	}
	if wrn != "" {
		logger.Warnf("%s: %s", filename, wrn)
	}
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

	projectRootDir, perpetualDir, err := utils.FindProjectRoot(outerCallLogger, innerCall)
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
	stash, err := loadStash(stashFile)
	if err != nil {
		outerCallLogger.Panicln("Error loading stash:", err)
	}

	if listFiles {
		logger.Infoln("Listing files in stash:", name)
		for _, entry := range stash.Files {
			if entry.Original.Exists {
				fmt.Println("Original:", entry.Filename)
			} else {
				fmt.Println("Original (absent):", entry.Filename)
			}
			if entry.Modified.Exists {
				fmt.Println("Modified:", entry.Filename)
			} else {
				fmt.Println("Modified (deleted):", entry.Filename)
			}
		}
		return
	}

	applyStates := func(useModifiedState bool) {
		for _, entry := range stash.Files {
			if fileName != "" && entry.Filename != fileName {
				continue
			}

			target, err := resolveTargetFile(projectRootDir, entry.Filename, targetFile)
			if err != nil {
				outerCallLogger.Panicln("Requested file is not inside project root", target)
			}

			state := entry.Original
			if useModifiedState {
				state = entry.Modified
			}

			applyFileState(projectRootDir, target, state, outerCallLogger)
		}
	}

	if apply {
		logger.Infoln("Applying changes")
		applyStates(true)
	} else if rollback {
		logger.Infoln("Rolling back changes")
		applyStates(false)
	}
}

// This function creates new stash from code generation results. Called internally
func CreateStash(results map[string]string, projectFiles []string, filesToDelete []string, logger logging.ILogger) string {
	logger.Traceln("CreateStash: Starting")
	defer logger.Traceln("CreateStash: Finished")

	projectRootDir, perpetualDir, err := utils.FindProjectRoot(logger, true)
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

	readOriginalState := func(filePath string) FileState {
		_, err := os.Stat(filepath.Join(projectRootDir, filePath))
		if err != nil {
			if os.IsNotExist(err) {
				return FileState{Exists: false}
			}
			logger.Panicf("Error checking project file for backing up %s: %v", filePath, err)
		}

		backup, _, err := llm.GetSourceFileFromCache(filePath)
		if err != nil {
			logger.Errorf("Error getting file from cache (will retry to read it directly): %v", err)
			backup, _, err = utils.LoadTextFile(filepath.Join(projectRootDir, filePath))
			if err != nil {
				logger.Panicf("Error reading project file for backing up %s: %v", filePath, err)
			}
		}

		return FileState{
			Exists:   true,
			Contents: backup,
		}
	}

	normalizeFilePath := func(filePathInitial string) string {
		// getting leading directories, case insensitive, trying to fix its cases from closest match from the projectFiles
		leadingDirs := utils.CaseInsensitiveLeadingDirectoriesSearch(filePathInitial, projectFiles)
		// recursively create leading dirs
		fileDir := ""
		for _, dir := range leadingDirs {
			fileDir = filepath.Join(fileDir, dir)
		}
		// getting final file path
		fileName := filepath.Base(filePathInitial)
		return filepath.Join(fileDir, fileName)
	}

	logger.Infoln("Creating new stash from generated results")
	stash := Stash{
		Version: StashVersion,
	}

	entriesByFile := make(map[string]int)

	addOrReplaceEntry := func(entry FileEntry) {
		if index, ok := entriesByFile[entry.Filename]; ok {
			stash.Files[index] = entry
			return
		}
		entriesByFile[entry.Filename] = len(stash.Files)
		stash.Files = append(stash.Files, entry)
	}

	for filePathInitial, fileContent := range results {
		filePathFinal := normalizeFilePath(filePathInitial)

		addOrReplaceEntry(FileEntry{
			Filename: filePathFinal,
			Original: readOriginalState(filePathFinal),
			Modified: FileState{
				Exists:   true,
				Contents: fileContent,
			},
		})
	}

	for _, filePathInitial := range filesToDelete {
		filePathFinal := normalizeFilePath(filePathInitial)

		if _, ok := entriesByFile[filePathFinal]; ok {
			logger.Warnln("File is both modified and deleted in stash, deletion will take precedence:", filePathFinal)
		}

		addOrReplaceEntry(FileEntry{
			Filename: filePathFinal,
			Original: readOriginalState(filePathFinal),
			Modified: FileState{
				Exists: false,
			},
		})
	}

	if len(stash.Files) > 0 {
		logger.Debugln("Files in stash:")
		for _, entry := range stash.Files {
			if entry.Original.Exists && entry.Modified.Exists {
				logger.Debugln("Modified:", entry.Filename)
			} else if !entry.Original.Exists && entry.Modified.Exists {
				logger.Debugln("Created:", entry.Filename)
			} else if entry.Original.Exists && !entry.Modified.Exists {
				logger.Debugln("Deleted:", entry.Filename)
			} else {
				logger.Debugln("Unchanged absent:", entry.Filename)
			}
		}
	}

	stashFileName := time.Now().Format("2006-01-02_15-04-05")
	err = utils.SaveJsonFile(filepath.Join(stashDir, stashFileName+".json"), stash)
	if err != nil {
		logger.Panicln("Error saving stash:", err)
	}

	return stashFileName
}
