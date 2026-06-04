# Misc Operation

The `misc` operation provides various helper functions for project validation and file handling that are not covered by other operations. This operation is particularly useful for troubleshooting, project setup verification, and file system maintenance tasks.

The `misc` operation is designed with clear separation of output: all human-readable logging goes to stderr, while machine-parsable output (such as file lists or error reports) goes to stdout. This makes it suitable for scripting and automation purposes.

## Usage

To run the `misc` operation, use the following command:

```sh
Perpetual misc [flags]
```

Exactly one main function flag must be provided.

The `misc` operation supports several command-line flags to access its different functions:

### Main Function Flags

- `-p`: Search for the `.perpetual` directory starting from the current directory and validate JSON configurations inside it. Outputs the detected path of the `.perpetual` directory on success.

- `-l`: List all project files accessible by Perpetual, relative to the project root. This function respects the `-x` and `-u` flags for filtering.

- `-fc`: Test reading all project files as text. If any files cannot be read, their paths (relative to project root) are printed to stdout. Works with `-x` and `-u` flags.

- `-fa`: Read project files and verify they contain only ASCII characters (0-127). Files containing non-ASCII characters or unreadable files are reported to stdout. Works with `-x` and `-u` flags.

- `-fs`: Read project files and convert files with non-UTF8/UTF16/UTF32 encoding to UTF8. Prints paths of converted files to stdout. Works with `-x` and `-u` flags.

### Additional Options

- `-h`: Display the help message showing all available flags and their descriptions.

- `-df <file>`: Specify an optional path to a project description file. Use `disabled` to skip loading the project description file entirely. If omitted, Perpetual tries to load `.perpetual/description.md`; a missing default description file is allowed.

- `-u`: Include unit test source files in processing. By default, unit test files are excluded according to the project test-file blacklist.

- `-x <file>`: Specify a path to a user-supplied regex filter file for excluding certain files from processing. See more info about using the filter [here](user_filter.md).

- `-v`: Enable debug logging for more detailed operation information.

- `-vv`: Enable both debug and trace logging for the highest level of verbosity.

## Functions

### Project Validation (`-p`)

The project validation function performs several important checks:

1. **Directory Discovery**: Searches for the `.perpetual` directory starting from the current directory and moving up through parent directories until found or the filesystem root is reached. If `PERPETUAL_DIR` is set, that directory is used instead.

2. **Configuration Validation**: Loads and validates all JSON configuration files (project config and operation configs) to ensure they are properly formatted and contain valid settings.

3. **Environment Setup**: Loads environment variables from `.env` files in both the project's `.perpetual` directory and the global configuration directory. Files are loaded alphabetically inside each directory, with project-local files loaded before global files.

4. **Project Description Check**: Loads the project description file according to the `-df` option. Missing default `.perpetual/description.md` is not treated as an error.

On success, this function outputs the detected path of the `.perpetual` directory to stdout, making it useful for scripting and automation.

### File Listing (`-l`)

The file listing function provides a comprehensive view of project files:

1. **File Discovery**: Recursively scans the project directory to find all files, excluding the `.perpetual` directory and its contents.

2. **Filter Application**: Applies project whitelist and blacklist filters as defined in the project configuration.

3. **Case Sensitivity Check**: Validates that no filename case collisions exist within the project.

4. **Path Validation**: Ensures filenames and directory names don't contain invalid path separator characters.

5. **Additional Filtering**: Applies user-supplied filters (`-x`) and unit test exclusions (`-u` flag) as requested.

The output is a sorted list of relative file paths, one per line, suitable for piping to other commands or processing in scripts.

### File Readability Check (`-fc`)

This function tests the readability of all project files as text:

1. **Encoding Detection**: Automatically detects UTF encodings using BOM markers where present, and otherwise validates content as UTF-8.

2. **Fallback Handling**: Uses fallback encoding (default: `windows-1252`, configurable via `FALLBACK_TEXT_ENCODING`) for files that cannot be decoded with standard UTF encodings.

3. **Error Reporting**: Outputs paths of files that cannot be read successfully, along with detailed error information in stderr logs.

This is particularly useful for identifying files with corrupted encodings or binary files that were mistakenly included in the project.

### ASCII Content Validation (`-fa`)

The ASCII validation function ensures files contain only ASCII characters:

1. **Character Scanning**: Examines each readable file character by character, tracking line and position information for non-ASCII characters.

2. **Comprehensive Checking**: Validates that all characters fall within the ASCII range (0-127).

3. **Detailed Reporting**: For files containing non-ASCII characters, provides the exact byte position, line number, and character position where the violation occurs.

4. **Error Output**: Prints paths of non-compliant or unreadable files to stdout with detailed diagnostic information in stderr.

This function is essential for projects that require strict ASCII-only source code file content. It is also suitable for detecting text inconsistencies that arise when using AI to edit files.

### File Encoding Conversion (`-fs`)

The encoding conversion function modernizes file encodings:

1. **Encoding Analysis**: Detects current file encoding using BOM patterns and UTF-8 validation.

2. **Selective Conversion**: Only converts files that were read with fallback encoding warnings.

3. **UTF-8 Standardization**: Converts affected files to standard UTF-8 encoding without BOM.

4. **Change Reporting**: Outputs paths of converted files to stdout, allowing users to track which files were modified.

This function helps resolve compatibility issues with files that are not encoded as UTF-8 or another supported UTF encoding.

## Examples

1. **Validate project setup and get perpetual directory path:**

   ```sh
   Perpetual misc -p
   ```

2. **List all project files including unit tests:**

   ```sh
   Perpetual misc -l -u
   ```

3. **Check file readability with custom filters:**

   ```sh
   Perpetual misc -fc -x custom_filters.json
   ```

4. **Verify ASCII-only content with debug logging:**

   ```sh
   Perpetual misc -fa -v
   ```

5. **Convert non-UTF files to UTF-8:**

   ```sh
   Perpetual misc -fs
   ```

6. **List files without loading project description:**

   ```sh
   Perpetual misc -l -df disabled
   ```

## Notes

The `misc` operation shares the same startup checks as other operations, including:

- Automatic discovery of the project root and `.perpetual` directory
- Loading of environment variables from appropriate locations
- Validation of all configuration files
- Application of project-specific file filters and settings

To use the `misc` operation, a project configuration must already exist. You can initialize a new configuration by running `Perpetual init -l <lang>`. For more information, see the op_init documentation (op_init.md).
