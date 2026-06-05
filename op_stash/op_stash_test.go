package op_stash

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/utils"
)

func setupTempProject(t *testing.T) (string, string) {
	t.Helper()

	projectRootDir := t.TempDir()
	perpetualDir := filepath.Join(projectRootDir, ".perpetual")
	stashDir := filepath.Join(perpetualDir, utils.StashesDirName)

	if err := os.MkdirAll(stashDir, 0755); err != nil {
		t.Fatalf("failed to create temporary stash directory: %v", err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	if err := os.Chdir(projectRootDir); err != nil {
		t.Fatalf("failed to change working directory to temporary project: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	t.Setenv("PERPETUAL_DIR", perpetualDir)

	return projectRootDir, stashDir
}

func newTestLogger(t *testing.T) logging.ILogger {
	t.Helper()

	logger, err := logging.NewSimpleLogger(logging.ErrorLevel)
	if err != nil {
		t.Fatalf("failed to create test logger: %v", err)
	}

	return logger
}

func writeTestStash(t *testing.T, stashDir, name string, stash Stash) {
	t.Helper()

	data, err := json.MarshalIndent(stash, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal test stash: %v", err)
	}

	if filepath.Ext(name) != ".json" {
		name += ".json"
	}

	if err := os.WriteFile(filepath.Join(stashDir, name), append(data, '\n'), 0644); err != nil {
		t.Fatalf("failed to write test stash: %v", err)
	}
}

func writeRawTestStash(t *testing.T, stashDir, name, contents string) {
	t.Helper()

	if filepath.Ext(name) != ".json" {
		name += ".json"
	}

	if err := os.WriteFile(filepath.Join(stashDir, name), []byte(contents), 0644); err != nil {
		t.Fatalf("failed to write raw test stash: %v", err)
	}
}

func runStash(t *testing.T, args ...string) {
	t.Helper()
	Run(args, false, newTestLogger(t))
}

func assertFileContents(t *testing.T, projectRootDir, filename, expected string) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(projectRootDir, filename))
	if err != nil {
		t.Fatalf("failed to read %q: %v", filename, err)
	}

	if string(data) != expected {
		t.Fatalf("unexpected contents for %q:\nexpected: %q\nactual:   %q", filename, expected, string(data))
	}
}

func assertFileNotExists(t *testing.T, projectRootDir, filename string) {
	t.Helper()

	_, err := os.Stat(filepath.Join(projectRootDir, filename))
	if err == nil {
		t.Fatalf("expected %q to not exist", filename)
	}
	if !os.IsNotExist(err) {
		t.Fatalf("unexpected error while checking %q: %v", filename, err)
	}
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected function to panic")
		}
	}()

	fn()
}

func TestStashApplyAndRollbackModification(t *testing.T) {
	projectRootDir, stashDir := setupTempProject(t)

	if err := os.WriteFile(filepath.Join(projectRootDir, "file.txt"), []byte("original\n"), 0644); err != nil {
		t.Fatalf("failed to create project file: %v", err)
	}

	writeTestStash(t, stashDir, "modify", Stash{
		Version: StashVersion,
		Files: []FileEntry{
			{
				Filename: "file.txt",
				Original: FileState{
					Exists:   true,
					Contents: "original\n",
				},
				Modified: FileState{
					Exists:   true,
					Contents: "modified\n",
				},
			},
		},
	})

	runStash(t, "-a", "-n", "modify")
	assertFileContents(t, projectRootDir, "file.txt", "modified\n")

	runStash(t, "-r", "-n", "modify")
	assertFileContents(t, projectRootDir, "file.txt", "original\n")
}

func TestStashRollbackDeletesFileCreatedByApply(t *testing.T) {
	projectRootDir, stashDir := setupTempProject(t)

	writeTestStash(t, stashDir, "create", Stash{
		Version: StashVersion,
		Files: []FileEntry{
			{
				Filename: "new/file.txt",
				Original: FileState{
					Exists: false,
				},
				Modified: FileState{
					Exists:   true,
					Contents: "created\n",
				},
			},
		},
	})

	assertFileNotExists(t, projectRootDir, "new/file.txt")

	runStash(t, "-a", "-n", "create")
	assertFileContents(t, projectRootDir, "new/file.txt", "created\n")

	runStash(t, "-r", "-n", "create")
	assertFileNotExists(t, projectRootDir, "new/file.txt")
}

func TestStashRollbackRecreatesFileDeletedByApply(t *testing.T) {
	projectRootDir, stashDir := setupTempProject(t)

	if err := os.WriteFile(filepath.Join(projectRootDir, "deleted.txt"), []byte("backup\n"), 0644); err != nil {
		t.Fatalf("failed to create project file: %v", err)
	}

	writeTestStash(t, stashDir, "delete", Stash{
		Version: StashVersion,
		Files: []FileEntry{
			{
				Filename: "deleted.txt",
				Original: FileState{
					Exists:   true,
					Contents: "backup\n",
				},
				Modified: FileState{
					Exists: false,
				},
			},
		},
	})

	runStash(t, "-a", "-n", "delete")
	assertFileNotExists(t, projectRootDir, "deleted.txt")

	runStash(t, "-r", "-n", "delete")
	assertFileContents(t, projectRootDir, "deleted.txt", "backup\n")
}

func TestStashSingleFileTargetCanBeCreatedAndDeletedOnRollback(t *testing.T) {
	projectRootDir, stashDir := setupTempProject(t)

	writeTestStash(t, stashDir, "single-target", Stash{
		Version: StashVersion,
		Files: []FileEntry{
			{
				Filename: "source.txt",
				Original: FileState{
					Exists: false,
				},
				Modified: FileState{
					Exists:   true,
					Contents: "target contents\n",
				},
			},
		},
	})

	runStash(t, "-a", "-n", "single-target", "-f", "source.txt", "-t", "nested/target.txt")
	assertFileContents(t, projectRootDir, "nested/target.txt", "target contents\n")
	assertFileNotExists(t, projectRootDir, "source.txt")

	runStash(t, "-r", "-n", "single-target", "-f", "source.txt", "-t", "nested/target.txt")
	assertFileNotExists(t, projectRootDir, "nested/target.txt")
}

func TestStashRejectsMissingVersion(t *testing.T) {
	_, stashDir := setupTempProject(t)

	writeRawTestStash(t, stashDir, "missing-version", `{
  "original_files": [
    {
      "filename": "file.txt",
      "contents": "old"
    }
  ],
  "modified_files": [
    {
      "filename": "file.txt",
      "contents": "new"
    }
  ]
}
`)

	assertPanics(t, func() {
		runStash(t, "-a", "-n", "missing-version")
	})
}

func TestStashRejectsUnsupportedVersion(t *testing.T) {
	_, stashDir := setupTempProject(t)

	writeTestStash(t, stashDir, "unsupported-version", Stash{
		Version: StashVersion + 1,
		Files: []FileEntry{
			{
				Filename: "file.txt",
				Original: FileState{
					Exists: false,
				},
				Modified: FileState{
					Exists:   true,
					Contents: "contents\n",
				},
			},
		},
	})

	assertPanics(t, func() {
		runStash(t, "-a", "-n", "unsupported-version")
	})
}

func TestCreateStashIncludesFilesToDelete(t *testing.T) {
	projectRootDir, _ := setupTempProject(t)

	if err := os.WriteFile(filepath.Join(projectRootDir, "existing.txt"), []byte("existing original\n"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectRootDir, "delete.txt"), []byte("delete original\n"), 0644); err != nil {
		t.Fatalf("failed to create delete file: %v", err)
	}

	stashName := CreateStash(
		map[string]string{
			"existing.txt": "existing modified\n",
			"created.txt":  "created contents\n",
		},
		[]string{"existing.txt", "delete.txt"},
		[]string{"delete.txt"},
		newTestLogger(t),
	)

	runStash(t, "-a", "-n", stashName)
	assertFileContents(t, projectRootDir, "existing.txt", "existing modified\n")
	assertFileContents(t, projectRootDir, "created.txt", "created contents\n")
	assertFileNotExists(t, projectRootDir, "delete.txt")

	runStash(t, "-r", "-n", stashName)
	assertFileContents(t, projectRootDir, "existing.txt", "existing original\n")
	assertFileNotExists(t, projectRootDir, "created.txt")
	assertFileContents(t, projectRootDir, "delete.txt", "delete original\n")
}
