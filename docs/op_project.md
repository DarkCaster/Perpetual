# Project Operation

The `project` operation manages the Perpetual configuration for your project. It combines project initialization with configuration validation and file-handling utilities.

## Usage

To run the `project` operation, use the following command:

```sh
Perpetual project [flags]
```

The operation mode is selected with the `-m` flag, which must be provided with exactly one mode value.

### Mode Selection Flag

The `-m <mode>` flag selects the function to perform. The following modes are available:

- `init`: Initialize the `.perpetual` directory and write default configuration files for the selected programming language or project type. Requires the `-l` flag.

- `test`: Search for the `.perpetual` directory, starting from the current directory, and validate the JSON configurations inside it. On success, outputs the detected absolute path of the `.perpetual` directory.

- `list`: List all project files accessible by Perpetual, relative to the project root. This function respects the `-x` and `-u` flags for filtering.

- `check-read`: Test reading all project files as text. If any files cannot be read, their paths (relative to project root) are printed to stdout. Works with `-x` and `-u` flags.

- `check-ascii`: Read project files and verify they contain only ASCII characters (0-127). Files containing non-ASCII characters or unreadable files are reported to stdout. Works with `-x` and `-u` flags.

- `save-utf`: Read project files and convert files with non-UTF8/UTF16/UTF32 encoding to UTF8. Prints paths of converted files to stdout. Works with `-x` and `-u` flags.

### Additional Options

- `-h`: Display the help message showing all available flags and their descriptions.

- `-l <language>`: Select the programming language for `-m init`, used to set up project configuration with default LLM prompts (valid values: `go`, `dotnet`, `bash`, `python3`, `vb6`, `c`, `cpp`, `arduino`, `flutter`). Required when using `-m init`.

- `-ex`: Create env-file examples inside the `.perpetual` directory. Only used together with `-m init`.

- `-df <file>`: Specify an optional path to a project description file (valid values: `file-path` or `disabled`).
  - For most modes, if omitted, Perpetual tries to load `.perpetual/description.md` as part of its startup checks; a missing default description file is allowed. Use `disabled` to skip loading the project description file entirely.
  - When used together with `-m init`, the specified file is instead loaded and saved as the project's default description file (`.perpetual/description.md`) during initialization.

- `-u`: Exclude unit-test source files from processing when running project-file checks. Unit test files are **included by default**; this flag applies the project's test-file blacklist to filter them out. Not used with `-m init`.

- `-x <file>`: Specify a path to a user-supplied regex filter file.
  - For most modes, this filter is applied on top of the project blacklist to exclude additional files from processing.
  - When used together with `-m init`, the regex patterns from the file are instead appended to the generated `project_files_blacklist` in `project.json`.
  - See more info about using the filter [here](user_filter.md).

- `-v`: Enable debug logging for more detailed operation information.

- `-vv`: Enable both debug and trace logging for the highest level of verbosity.

## Initializing a Project (`-m init`)

The `init` mode is the starting point for using Perpetual with your project. It sets up the necessary configuration and directory structure required for Perpetual to function properly within your project environment.

### Example Usage

Run the following command **from the root directory of your project**:

```sh
Perpetual project -m init -l go
```

This command initializes Perpetual for a Go project, setting up the appropriate configuration files with Go-specific prompts and file-selection rules.

### Supported Languages

Perpetual currently supports the following programming languages and technologies:

1. **Go (`go`)**
2. **.NET (`dotnet`)** – Includes C#, VB.NET, XAML, Razor, CSS, JavaScript, and HTML files
3. **Bash (`bash`)** – Intended mostly for Linux shell scripting
4. **Python 3 (`python3`)**
5. **Visual Basic 6 (`vb6`)** – Legacy VB6 projects
6. **C (`c`)**
7. **C++ (`cpp`)**
8. **Arduino (`arduino`)** – Arduino sketches with C/C++ support
9. **Flutter (`flutter`)** – Flutter/Dart apps, including selected native platform files

When using the `-l` flag, provide the language identifier as shown above. The language value is case-insensitive. You can add support for your programming language or project manually by editing file-selection regexps and LLM prompts in the corresponding config files.

### Details

The `init` mode performs several tasks to set up your project for use with Perpetual:

1. **Creates a local Perpetual configuration directory**. By default this is `.perpetual` in the current directory (or the directory pointed to by the `PERPETUAL_DIR` environment variable, if set). If the directory already exists, it is reused.
2. **Prepares the project configuration**, optionally appending user-supplied filter patterns (from `-x`) to the generated `project_files_blacklist`.
3. **Optionally loads a project description file** (via `-df`) and saves it as `.perpetual/description.md`.
4. **Creates a `.gitignore` file** within the `.perpetual` directory to exclude generated Perpetual state files and local `.env` files from version control.
5. **Writes the project description file** if one was supplied via `-df`, or otherwise creates a `description.md.template` file that can be used as a starting point for writing one manually.
6. **Generates JSON configuration files** for the selected programming language, including default prompts, regexps, tags, context-saving settings, incremental-mode settings, and file-selection rules.
7. **Creates example `.env.example` files** for Perpetual and each supported LLM provider, but only when the `-ex` flag is used: `.env.example`, `ollama.env.example`, `openai.env.example`, `anthropic.env.example`, and `generic.env.example`.
8. **Warns about obsolete configuration files and directories** if any are found in the `.perpetual` directory.

When run inside an already initialized project, `-m init` will overwrite the generated project-local config files, `.gitignore`, and description template (and the example environment files when `-ex` is used). Back up any manual changes before running it again.

**Note:** The `init` mode respects the `PERPETUAL_DIR` environment variable. If set, it uses the specified directory instead of creating `.perpetual` in the current directory.

### Directory and File Structure

After running `-m init`, the following project-local structure is created by default:

```text
<project_root>/
└── .perpetual/
    ├── .gitignore
    ├── description.md.template
    ├── op_annotate.json
    ├── op_implement.json
    ├── op_doc.json
    ├── op_explain.json
    ├── op_report.json
    └── project.json
```

If a project description file was supplied via `-df`, it is saved as `description.md` instead of generating the `description.md.template` file:

```text
<project_root>/
└── .perpetual/
    ├── description.md
    └── ...
```

When the `-ex` flag is used, the following example environment files are additionally created inside the `.perpetual` directory:

```text
<project_root>/
└── .perpetual/
    ├── .env.example
    ├── ollama.env.example
    ├── openai.env.example
    ├── anthropic.env.example
    └── generic.env.example
```

In addition to project-local configuration, Perpetual can also read environment values from a global, OS-specific user config directory when loading settings, for example:

```text
~/.config/Perpetual/
└── *.env
```

on Linux, or:

```text
<User profile dir>\AppData\Roaming\Perpetual\
└── *.env
```

on Windows. The `init` mode itself does not create this global directory or any global `.env` file.

### Customizable Files

Most files generated by `-m init` can be customized to fine-tune Perpetual's behavior for your project:

1. **`.env` files**

   Perpetual does not create a project-local `.env` file by default. To configure environment variables, run `-m init` with the `-ex` flag to generate example files, copy one or more of them to files ending with `.env`, then edit them as needed.

   For example:

   - Copy `.env.example` to `.env` for general provider-selection settings.
   - Copy `openai.env.example` to `openai.env` for OpenAI-specific settings.
   - Copy `ollama.env.example` to `ollama.env` for Ollama-specific settings.

   You can merge settings into a single `.env` file or keep them split across multiple `*.env` files.

   Perpetual loads environment values with the following precedence:

   - Existing system environment variables have the highest priority.
   - Project-local env files: `<project_root>/.perpetual/*.env`
   - Global env files: e.g. `~/.config/Perpetual/*.env` on Linux or `<User profile dir>\AppData\Roaming\Perpetual\*.env` on Windows

   Files are loaded in alphabetical order within each directory. Values loaded earlier take precedence over values loaded later.

2. **Operation-specific JSON configs**

   Files like `op_annotate.json`, `op_implement.json`, `op_doc.json`, `op_explain.json`, and `op_report.json` contain prompts, regex patterns, tags, and other settings Perpetual uses when interacting with your source code. You can adjust these to change how Perpetual processes files and communicates with the LLM.

3. **`project.json`**

   This file controls project-wide behavior, including:

   - **`project_files_whitelist`**: Regex patterns to include relevant files.
   - **`project_files_blacklist`**: Regex patterns to exclude files from processing (may include patterns appended from a `-x` user filter file supplied at `init` time).
   - **`project_test_files_blacklist`**: Regex patterns to exclude test files when operations are run without test inclusion.
   - **`files_to_md_code_mappings`**: Maps file path patterns to Markdown code-block languages.
   - **`filename_tags`** and **`filename_tags_rx`**: Tags and regexps used when sending and parsing filenames.
   - **`delete_tags`** and **`delete_tags_rx`**: Tags and regexps used when parsing file-deletion requests from LLM responses.
   - **`code_tags_rx`**: Regexps used to parse code blocks from LLM responses.
   - **`noupload_comments_rx`**: Regexps for detecting files marked as not uploadable.
   - **Project index and description prompts**: Prompt and response text used when providing project structure or project description context to the LLM.
   - **Context-saving settings**: Thresholds and percentages used to reduce LLM context usage for larger projects.
   - **Incremental file-modification settings**: Regexps and minimum file-size mappings used by implementation mode when generating compact search-and-replace changes.

4. **`description.md` / `description.md.template`**

   `description.md.template` is a template for an optional project description, created when `-m init` is run without a `-df` file. To use it, copy or rename it to `.perpetual/description.md` and edit it for your project. If a description file is supplied via `-df` at `init` time, it is saved directly as `.perpetual/description.md` instead. The resulting `description.md` can provide additional context to operations such as `implement`, `explain`, `doc`, `annotate`, and `report`.

These configuration files can be committed to version control to share and track changes across your team. Local `.env` files should usually not be committed.

### Files Not Intended for User Modification

The following are managed automatically by Perpetual and should not be edited manually:

1. **`.annotations.json`**: Stores source-code annotations for your project. Updated via the `annotate` operation.
2. **`.embeddings.msgpack`**: Stores source-code vector embeddings for your project. Updated via the `embed` operation.
3. **`.stash/`**: Contains code backups created during operations. Managed by the `stash` operation.
4. **`.message_log.txt*`**: Logs interactions with the LLM provider for debugging purposes, including rotated log files.

### Obsolete Files and Directories

When running `-m init`, Perpetual checks the `.perpetual` directory for deprecated items left over from older versions. It does not remove them automatically; instead, it prints a warning so you can delete them manually.

- **Obsolete Files**
  - `filename_embed_regexp.json`
  - `filename_tags_regexps.json`
  - `filename_tags.json`
  - `no_upload_comment_regexps.json`
  - `output_tags_regexps.json`
  - `project_files_blacklist.json`
  - `project_files_to_markdown_lang_mappings.json`
  - `project_files_whitelist.json`
  - `reasonings_tags_regexps.json`
  - `reasonings_tags.json`

- **Obsolete Directories**
  - `prompts`

## Validating and Inspecting a Project (`-m test`, `list`, `check-read`, `check-ascii`, `save-utf`)

The remaining modes provide helper functions for project validation and file handling that are not covered by other operations. These modes are particularly useful for troubleshooting, project setup verification, and file system maintenance tasks. Unlike `-m init`, they require that a project has already been initialized (see above).

All of these modes share the same startup checks:

1. **Directory Discovery**: Searches for the `.perpetual` directory starting from the current directory and moving up through parent directories until found or the filesystem root is reached. If `PERPETUAL_DIR` is set, that directory is used instead. The discovered `.perpetual` path must be a directory, and the project root must not be a symlink or reparse point.

2. **Environment Setup**: Loads environment variables from `.env` files in both the project's `.perpetual` directory and the global configuration directory. Files are loaded alphabetically inside each directory, with project-local files loaded before global files. Existing environment variables are not overwritten, so values exported before running Perpetual have priority.

3. **Configuration Validation**: Loads and validates all JSON configuration files (project config and operation configs) to ensure they are properly formatted and contain valid settings.

4. **Project Description Check**: Loads the project description file according to the `-df` option. A missing default `.perpetual/description.md` is not treated as an error.

### Project Validation (`-m test`)

On success, this mode performs all of the startup checks above and outputs the detected path of the `.perpetual` directory to stdout, making it useful for scripting and automation.

### File Listing (`-m list`)

The file listing function provides a comprehensive view of project files:

1. **File Discovery**: Recursively scans the project directory to find all files, excluding the `.perpetual` directory and its contents.

2. **Filter Application**: Applies project whitelist and blacklist filters as defined in the project configuration.

3. **Case Sensitivity Check**: Validates that no filename case collisions exist within the project.

4. **Path Validation**: Ensures filenames and directory names don't contain invalid path separator characters.

5. **Additional Filtering**: Applies the user-supplied filter (`-x`) and the test-file blacklist (`-u`), if requested.

The output is a sorted list of relative file paths, one per line, suitable for piping to other commands or processing in scripts.

### File Readability Check (`-m check-read`)

This function tests the readability of all project files as text:

1. **Encoding Detection**: Automatically detects UTF encodings using BOM markers where present, and otherwise validates content as UTF-8.

2. **Fallback Handling**: Uses fallback encoding (default: `windows-1252`, configurable via `FALLBACK_TEXT_ENCODING`) for files that cannot be decoded with standard UTF encodings.

3. **Error Reporting**: Outputs paths of files that cannot be read successfully, along with detailed error information in stderr logs.

This is particularly useful for identifying files with corrupted encodings or binary files that were mistakenly included in the project.

### ASCII Content Validation (`-m check-ascii`)

The ASCII validation function ensures files contain only ASCII characters:

1. **Character Scanning**: Examines each readable file character by character, tracking line and position information for non-ASCII characters.

2. **Comprehensive Checking**: Validates that all characters fall within the ASCII range (0-127).

3. **Detailed Reporting**: For files containing non-ASCII characters, provides the exact byte position, line number, and character position where the violation occurs.

4. **Error Output**: Prints paths of non-compliant or unreadable files to stdout with detailed diagnostic information in stderr.

This function is essential for projects that require strict ASCII-only source code file content. It is also suitable for detecting text inconsistencies that arise when using AI to edit files.

### File Encoding Conversion (`-m save-utf`)

The encoding conversion function modernizes file encodings:

1. **Encoding Analysis**: Detects current file encoding using BOM patterns and UTF-8 validation.

2. **Selective Conversion**: Only converts files that were read with fallback encoding warnings. UTF-8, UTF-8 with BOM, UTF-16, and UTF-32 files that decode successfully as supported UTF encodings are not converted by this mode.

3. **UTF-8 Standardization**: Converts affected files to standard UTF-8 encoding without BOM.

4. **Change Reporting**: Outputs paths of converted files to stdout, allowing users to track which files were modified.

This function helps resolve compatibility issues with files that are not encoded as UTF-8 or another supported UTF encoding.

## Examples

1. **Initialize a new project for Go, including provider .env examples:**

   ```sh
   Perpetual project -m init -l go -ex
   ```

2. **Initialize a project and import an existing project description:**

   ```sh
   Perpetual project -m init -l cpp -df ./docs/project_overview.md
   ```

3. **Validate project setup and get the `.perpetual` directory path:**

   ```sh
   Perpetual project -m test
   ```

4. **List all project files, excluding unit tests:**

   ```sh
   Perpetual project -m list -u
   ```

5. **Check file readability with custom filters:**

   ```sh
   Perpetual project -m check-read -x custom_filters.json
   ```

6. **Verify ASCII-only content with debug logging:**

   ```sh
   Perpetual project -m check-ascii -v
   ```

7. **Convert non-UTF files to UTF-8:**

   ```sh
   Perpetual project -m save-utf
   ```

8. **List files without loading the project description:**

   ```sh
   Perpetual project -m list -df disabled
   ```
