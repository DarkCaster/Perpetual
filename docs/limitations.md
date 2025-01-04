# Current Technical Limitations

This document outlines the current technical limitations. These limitations may be addressed in future updates.

## File Operations

Given `Perpetual`'s focus on direct codebase interaction and to maintain its simplicity, it has some limitations on file operations:

- Cannot delete files
- Cannot run external tools or commands on the user's system
- Cannot interact with version control systems (e.g., Git, SVN)
- Cannot install packages (e.g., npm, NuGet)

These limitations are in place to ensure a controlled and safe environment for code manipulation.

## Supported Source File Encoding

`Perpetual` only supports the following text encodings for source files:

- UTF-8 (with or without BOM)
- UTF-16 (LE and BE, with BOM)
- UTF-32 (LE and BE, with BOM)

When reading files, `Perpetual` performs the following operations:

1. Detects the file encoding
2. Converts the content to UTF-8 without BOM
3. Validates the UTF-8 encoding
4. Unsupported encodings will be treated as UTF-8 without BOM

Currently, **all files are written back as UTF-8 without BOM** (Byte Order Mark) to ensure consistency across the project. This may be improved in the future to write files back in their original encoding.

## Line Endings (CR LF)

`Perpetual` handles line endings in the following manner:

- **Reading**: Supports files with any line-ending style (CR, LF, or CRLF)
- **Writing**:
  - On Windows: Uses CRLF
  - On Linux: Uses LF
  - On macOS: Uses LF (Note: macOS is not officially supported, but will follow Linux behavior if compiled)

This behavior is similar to Git's `core.autocrlf = true` setting.

**Important**: Mixed line endings within a file are not preserved during modifications. This is because the LLM generates the entire source file content at once, potentially altering the original line ending style.

## Symlinks

`Perpetual` has specific limitations regarding symlinks:

- Files inside the project that contain symlinks within their relative path are ignored
- The base project directory cannot be a symlink
- Any parent directory of the project directory **can be a symlink**

These limitations are in place to enhance security and simplify implementation using Go. Future versions may improve symlink handling.

## Filename Casing

`Perpetual` enforces strict rules for filename casing to ensure consistency and prevent conflicts:

- Project files must not have the same file paths with different cases
  - This is particularly important for case-sensitive file systems (e.g., Linux)
  - Not applicable for Windows due to its case-insensitive file system

`Perpetual` checks for filename case collisions to align with existing project paths when necessary. This ensures consistency across different operating systems and prevents potential conflicts.

## Project Root Detection

`Perpetual` uses a specific method to detect the project root:

- Searches for a `.perpetual` directory starting from the current working directory
- Moves up the directory tree until it finds the `.perpetual` directory or reaches the file system root
- The project root cannot be a symlink

This approach ensures that `Perpetual` operates within the intended project scope.
