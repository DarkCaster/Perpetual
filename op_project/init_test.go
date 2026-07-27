package op_project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

// testLogger is a minimal no-op implementation of logging.ILogger used for testing.
// Panic-level methods actually panic (after formatting the message), so that unexpected
// error conditions inside Init are still surfaced as test failures.
type testLogger struct{}

func (l *testLogger) Panicln(v ...any) {
	panic(fmt.Sprintln(v...))
}

func (l *testLogger) Panicf(format string, v ...any) {
	panic(fmt.Sprintf(format, v...))
}

func (l *testLogger) Errorln(v ...any)                           {}
func (l *testLogger) Errorf(format string, v ...any)             {}
func (l *testLogger) Warnln(v ...any)                            {}
func (l *testLogger) Warnf(format string, v ...any)              {}
func (l *testLogger) Infoln(v ...any)                            {}
func (l *testLogger) Infof(format string, v ...any)              {}
func (l *testLogger) Notifyln(v ...any)                          {}
func (l *testLogger) Notifyf(format string, v ...any)            {}
func (l *testLogger) Debugln(v ...any)                           {}
func (l *testLogger) Debugf(format string, v ...any)             {}
func (l *testLogger) Traceln(v ...any)                           {}
func (l *testLogger) Tracef(format string, v ...any)             {}
func (l *testLogger) EnableLevel(level logging.LogLevel)         {}
func (l *testLogger) DisableLevel(level logging.LogLevel)        {}
func (l *testLogger) IsLevelEnabled(level logging.LogLevel) bool { return true }
func (l *testLogger) Clone() logging.ILogger                     { return &testLogger{} }

// setPerpetualDirEnv points PERPETUAL_DIR env variable at the given directory for the
// duration of the test, restoring the previous value (or unsetting it) afterwards.
func setPerpetualDirEnv(t *testing.T, dir string) {
	t.Helper()
	origVal, hadEnv := os.LookupEnv("PERPETUAL_DIR")
	if err := os.Setenv("PERPETUAL_DIR", dir); err != nil {
		t.Fatalf("failed to set PERPETUAL_DIR env var: %v", err)
	}
	t.Cleanup(func() {
		if hadEnv {
			_ = os.Setenv("PERPETUAL_DIR", origVal)
		} else {
			_ = os.Unsetenv("PERPETUAL_DIR")
		}
	})
}

func TestInitUserFilterAppendedToBlacklist(t *testing.T) {
	tempDir := t.TempDir()
	perpetualDir := filepath.Join(tempDir, ".perpetual")

	setPerpetualDirEnv(t, perpetualDir)

	customFilters := []string{
		"^custom_filter_one$",
		"^custom_filter_two(\\\\|\\/).*\\.tmp$",
	}
	filterData, err := json.Marshal(customFilters)
	if err != nil {
		t.Fatalf("failed to marshal custom filters: %v", err)
	}
	filterFilePath := filepath.Join(tempDir, "user_filter.json")
	if err := os.WriteFile(filterFilePath, filterData, 0644); err != nil {
		t.Fatalf("failed to write user filter file: %v", err)
	}

	logger := &testLogger{}
	if err := Init("test-version", "go", false, "", filterFilePath, logger); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	projectConfigPath := filepath.Join(perpetualDir, config.ProjectConfigFile)
	rawConfig := map[string]any{}
	if err := utils.LoadJsonFile(projectConfigPath, &rawConfig); err != nil {
		t.Fatalf("failed to load generated project config: %v", err)
	}

	blacklistRaw, ok := rawConfig[config.K_ProjectFilesBlacklist]
	if !ok {
		t.Fatalf("project config does not contain %s key", config.K_ProjectFilesBlacklist)
	}
	blacklistArr, ok := blacklistRaw.([]any)
	if !ok {
		t.Fatalf("%s is not an array, got %T", config.K_ProjectFilesBlacklist, blacklistRaw)
	}

	if len(blacklistArr) < len(customFilters) {
		t.Fatalf("blacklist array (len=%d) is too short to contain appended custom filters (len=%d)", len(blacklistArr), len(customFilters))
	}

	tail := blacklistArr[len(blacklistArr)-len(customFilters):]
	for i, expected := range customFilters {
		actual, ok := tail[i].(string)
		if !ok {
			t.Fatalf("expected string at tail position %d, got %T", i, tail[i])
		}
		if actual != expected {
			t.Errorf("expected custom filter %q at tail position %d, got %q", expected, i, actual)
		}
	}
}

func TestInitDescriptionFileSaved(t *testing.T) {
	tempDir := t.TempDir()
	perpetualDir := filepath.Join(tempDir, ".perpetual")

	setPerpetualDirEnv(t, perpetualDir)

	descriptionContent := "# Test Project Description\n\nThis is a unique test description content used to verify that the -df flag properly copies the provided description file into the .perpetual directory.\n"
	descFilePath := filepath.Join(tempDir, "custom_description.md")
	if err := os.WriteFile(descFilePath, []byte(descriptionContent), 0644); err != nil {
		t.Fatalf("failed to write description file: %v", err)
	}

	logger := &testLogger{}
	if err := Init("test-version", "go", false, descFilePath, "", logger); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	savedDescPath := filepath.Join(perpetualDir, config.ProjectDescriptionFile)
	savedContent, _, err := utils.LoadTextFile(savedDescPath)
	if err != nil {
		t.Fatalf("failed to read saved description file: %v", err)
	}

	if savedContent != descriptionContent {
		t.Errorf("saved description content mismatch.\nExpected:\n%s\nGot:\n%s", descriptionContent, savedContent)
	}
}
