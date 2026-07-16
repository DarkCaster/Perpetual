package op_implement

import (
	"os"
	"path/filepath"

	"github.com/DarkCaster/Perpetual/llm"
	"github.com/DarkCaster/Perpetual/utils"
)

// StateFileName is the name of the JSON file used to store intermediate state
// between the preparation stages (1-3) and the final implementation stage (4)
// when running the 'implement' operation in managed step-by-step mode.
const StateFileName = ".implement_state.json"

// state holds all the data required to resume the 'implement' operation at
// stage 4 after the preparation stages (1-3) have been completed and confirmed
// by the operator/agent.
type state struct {
	OtherFilesToModify  []string      `json:"other_files_to_modify,omitempty"`
	TargetFilesToModify []string      `json:"target_files_to_modify,omitempty"`
	FilesToDelete       []string      `json:"files_to_delete,omitempty"`
	Messages            []llm.Message `json:"messages,omitempty"`
}

// getStateFilePath returns the full path to the state file inside perpetualDir.
func getStateFilePath(perpetualDir string) string {
	return filepath.Join(perpetualDir, StateFileName)
}

// saveState writes the provided state to the state file inside perpetualDir.
func saveState(perpetualDir string, state state) error {
	return utils.SaveJsonFile(getStateFilePath(perpetualDir), state)
}

// loadState reads and validates the state file from perpetualDir, precache source files
func loadState(perpetualDir, projectRootDir string) (state, error) {
	var state state
	if err := utils.LoadJsonFile(getStateFilePath(perpetualDir), &state); err != nil {
		return state, err
	}
	// precache source files from state, best effort:
	// - ensure they will be handled same way as if we not used state file at all;
	// - ensure source files metadata (encoding) was properly detected and converted if needed;
	// - ensure source files do not change during stage 4;
	for _, file := range state.OtherFilesToModify {
		llm.PrecacheSourceFile(projectRootDir, file)
	}
	for _, file := range state.TargetFilesToModify {
		llm.PrecacheSourceFile(projectRootDir, file)
	}
	return state, nil
}

// removeState deletes the state file from perpetualDir if it exists.
func removeState(perpetualDir string) error {
	err := utils.RemoveFile(getStateFilePath(perpetualDir))
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}
