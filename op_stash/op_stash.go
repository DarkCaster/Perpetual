package op_stash

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func Run(args []string, logger logging.ILogger) {
	var help, list, verbose, apply, rollback, trace bool
	var name string

	// Parse flags for the "stash" operation
	flags := stashFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.BoolVar(&list, "l", false, "List current stashes")
	flags.BoolVar(&apply, "a", false, "Apply changes of a stash")
	flags.BoolVar(&rollback, "r", false, "Rollback changes of a stash")
	flags.StringVar(&name, "n", "latest", "Set stash name to apply or revert")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if verbose {
		logger.SetLevel(logging.DebugLevel)
	}
	if trace {
		logger.SetLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'stash' operation")
	logger.Traceln("Args:", args)

	if !apply && !rollback && !list {
		help = true
	}

	if help {
		usage.PrintOperationUsage("", flags)
		return
	}

	_, perpetualDir, err := utils.FindProjectRoot(logger)
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

	stashes, err := os.ReadDir(stashDir)
	if err != nil {
		logger.Panicln("Error reading stash directory:", err)
	}

	fileStashes := make([]os.DirEntry, 0)
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

	if apply || rollback {
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
			logger.Panicln("Stash not found:", name)
		}
		var stash Stash
		err := utils.LoadJsonFile(stashFile, &stash)
		if err != nil {
			logger.Panicln("Error loading stash:", err)
		}

		if apply {
			logger.Infoln("Applied stashed changes:", name)
		} else {
			logger.Infoln("Stashed changes rolled back:", name)
		}

		return
	}
}
