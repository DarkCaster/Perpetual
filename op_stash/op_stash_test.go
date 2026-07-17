package op_stash

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/DarkCaster/Perpetual/llm"
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

func assertFileBytes(t *testing.T, projectRootDir, filename string, expected []byte) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(projectRootDir, filename))
	if err != nil {
		t.Fatalf("failed to read %q: %v", filename, err)
	}

	if string(data) != string(expected) {
		t.Fatalf("unexpected bytes for %q:\nexpected: %v\nactual:   %v", filename, expected, data)
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

	runStash(t, "-m", "apply", "-s", "modify")
	utils.RunGlobalCleanup()
	assertFileContents(t, projectRootDir, "file.txt", "modified\n")

	runStash(t, "-m", "rollback", "-s", "modify")
	utils.RunGlobalCleanup()
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

	runStash(t, "-m", "apply", "-s", "create")
	utils.RunGlobalCleanup()
	assertFileContents(t, projectRootDir, "new/file.txt", "created\n")

	runStash(t, "-m", "rollback", "-s", "create")
	utils.RunGlobalCleanup()
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

	runStash(t, "-m", "apply", "-s", "delete")
	utils.RunGlobalCleanup()
	assertFileNotExists(t, projectRootDir, "deleted.txt")

	runStash(t, "-m", "rollback", "-s", "delete")
	utils.RunGlobalCleanup()
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

	runStash(t, "-m", "apply", "-s", "single-target", "-o", "source.txt", "-t", "nested/target.txt")
	utils.RunGlobalCleanup()
	assertFileContents(t, projectRootDir, "nested/target.txt", "target contents\n")
	assertFileNotExists(t, projectRootDir, "source.txt")

	runStash(t, "-m", "rollback", "-s", "single-target", "-o", "source.txt", "-t", "nested/target.txt")
	utils.RunGlobalCleanup()
	assertFileNotExists(t, projectRootDir, "nested/target.txt")
}

func TestStashApplyAndRollbackPreserveUTF16Encoding(t *testing.T) {
	projectRootDir, stashDir := setupTempProject(t)

	writeTestStash(t, stashDir, "utf16", Stash{
		Version: StashVersion,
		Files: []FileEntry{
			{
				Filename: "encoded.txt",
				FileParams: utils.FileParams{
					ModernEncoding:        utils.UTF16BE,
					UsingFallbackEncoding: false,
				},
				Original: FileState{
					Exists:   true,
					Contents: "old",
				},
				Modified: FileState{
					Exists:   true,
					Contents: "new",
				},
			},
		},
	})

	runStash(t, "-m", "apply", "-s", "utf16")
	utils.RunGlobalCleanup()
	assertFileBytes(t, projectRootDir, "encoded.txt", []byte{
		0xFE, 0xFF,
		0x00, 0x6E,
		0x00, 0x65,
		0x00, 0x77,
	})

	runStash(t, "-m", "rollback", "-s", "utf16")
	utils.RunGlobalCleanup()
	assertFileBytes(t, projectRootDir, "encoded.txt", []byte{
		0xFE, 0xFF,
		0x00, 0x6F,
		0x00, 0x6C,
		0x00, 0x64,
	})
}

func TestStashApplyUsesFallbackEncoding(t *testing.T) {
	projectRootDir, stashDir := setupTempProject(t)
	t.Setenv("FALLBACK_TEXT_ENCODING", "windows-1252")

	writeTestStash(t, stashDir, "fallback", Stash{
		Version: StashVersion,
		Files: []FileEntry{
			{
				Filename: "fallback.txt",
				FileParams: utils.FileParams{
					ModernEncoding:        utils.UTF8,
					UsingFallbackEncoding: true,
				},
				Original: FileState{
					Exists: false,
				},
				Modified: FileState{
					Exists:   true,
					Contents: "café",
				},
			},
		},
	})

	runStash(t, "-m", "apply", "-s", "fallback")
	utils.RunGlobalCleanup()
	assertFileBytes(t, projectRootDir, "fallback.txt", []byte{0x63, 0x61, 0x66, 0xE9})
}

func TestStashTargetFileUsesStoredEncoding(t *testing.T) {
	projectRootDir, stashDir := setupTempProject(t)

	writeTestStash(t, stashDir, "encoded-target", Stash{
		Version: StashVersion,
		Files: []FileEntry{
			{
				Filename: "source.txt",
				FileParams: utils.FileParams{
					ModernEncoding:        utils.UTF8BOM,
					UsingFallbackEncoding: false,
				},
				Original: FileState{
					Exists: false,
				},
				Modified: FileState{
					Exists:   true,
					Contents: "target",
				},
			},
		},
	})

	runStash(t, "-m", "apply", "-s", "encoded-target", "-o", "source.txt", "-t", "nested/target.txt")
	utils.RunGlobalCleanup()
	assertFileBytes(t, projectRootDir, "nested/target.txt", []byte{
		0xEF, 0xBB, 0xBF,
		0x74, 0x61, 0x72, 0x67, 0x65, 0x74,
	})

	runStash(t, "-m", "rollback", "-s", "encoded-target", "-o", "source.txt", "-t", "nested/target.txt")
	utils.RunGlobalCleanup()
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
		runStash(t, "-m", "apply", "-s", "missing-version")
	})
	utils.RunGlobalCleanup()
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
		runStash(t, "-m", "apply", "-s", "unsupported-version")
	})
	utils.RunGlobalCleanup()
}

func TestCreateStashStoresFileEncodingParams(t *testing.T) {
	projectRootDir, stashDir := setupTempProject(t)

	modifiedPath := filepath.Join(projectRootDir, "modified.txt")
	deletedPath := filepath.Join(projectRootDir, "deleted.txt")

	if err := os.WriteFile(modifiedPath, []byte("modified original"), 0644); err != nil {
		t.Fatalf("failed to create modified test file: %v", err)
	}
	if err := os.WriteFile(deletedPath, []byte("deleted original"), 0644); err != nil {
		t.Fatalf("failed to create deleted test file: %v", err)
	}

	modifiedParams := utils.FileParams{
		ModernEncoding:        utils.UTF16LE,
		UsingFallbackEncoding: false,
	}
	deletedParams := utils.FileParams{
		ModernEncoding:        utils.UTF8,
		UsingFallbackEncoding: true,
	}

	// read file as if it was already read by llm (so it will not be re-read)
	llm.PrecacheSourceFile(projectRootDir, "modified.txt")
	llm.PrecacheSourceFile(projectRootDir, "deleted.txt")
	// fake target file encoding
	utils.SetFileParams(modifiedPath, modifiedParams)
	utils.SetFileParams(deletedPath, deletedParams)

	stashName := CreateStash(
		map[string]string{
			"modified.txt": "modified result",
			"created.txt":  "created result",
		},
		[]string{"modified.txt", "deleted.txt"},
		[]string{"deleted.txt"},
		newTestLogger(t),
	)

	stash, err := loadStash(filepath.Join(stashDir, stashName+".json"))
	if err != nil {
		t.Fatalf("failed to load created stash: %v", err)
	}

	entries := make(map[string]FileEntry, len(stash.Files))
	for _, entry := range stash.Files {
		entries[entry.Filename] = entry
	}

	modifiedEntry, ok := entries["modified.txt"]
	if !ok {
		t.Fatalf("created stash does not contain modified.txt")
	}
	if modifiedEntry.FileParams != modifiedParams {
		t.Fatalf("modified.txt params = %+v, expected %+v", modifiedEntry.FileParams, modifiedParams)
	}

	deletedEntry, ok := entries["deleted.txt"]
	if !ok {
		t.Fatalf("created stash does not contain deleted.txt")
	}
	if deletedEntry.FileParams != deletedParams {
		t.Fatalf("deleted.txt params = %+v, expected %+v", deletedEntry.FileParams, deletedParams)
	}

	createdEntry, ok := entries["created.txt"]
	if !ok {
		t.Fatalf("created stash does not contain created.txt")
	}
	defaultParams := utils.FileParams{
		ModernEncoding:        utils.UTF8,
		UsingFallbackEncoding: false,
	}
	if createdEntry.FileParams != defaultParams {
		t.Fatalf("created.txt params = %+v, expected default %+v", createdEntry.FileParams, defaultParams)
	}
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

	runStash(t, "-m", "apply", "-s", stashName)
	utils.RunGlobalCleanup()
	assertFileContents(t, projectRootDir, "existing.txt", "existing modified\n")
	assertFileContents(t, projectRootDir, "created.txt", "created contents\n")
	assertFileNotExists(t, projectRootDir, "delete.txt")

	runStash(t, "-m", "rollback", "-s", stashName)
	utils.RunGlobalCleanup()
	assertFileContents(t, projectRootDir, "existing.txt", "existing original\n")
	assertFileNotExists(t, projectRootDir, "created.txt")
	assertFileContents(t, projectRootDir, "delete.txt", "delete original\n")
}
