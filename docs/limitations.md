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
- The project root directory cannot be a symlink
- Parent directories of the project root can be symlinks

These limitations are in place to enhance security and simplify implementation using Go. Future versions may improve symlink handling.

## Filename Casing

`Perpetual` enforces strict rules for filename casing to ensure consistency and prevent conflicts:

- Project files must not have the same file paths with different cases
  - This is particularly important for case-sensitive file systems (e.g., Linux)
  - Not applicable for Windows due to its case-insensitive file system

When handling filenames, `Perpetual` attempts to:

- Match the case of existing project files
- Create necessary directories with correct casing when applying changes

## Project Root Detection

`Perpetual` uses a specific method to detect the project root:

- Searches for a `.perpetual` directory starting from the current working directory
- Moves up the directory tree until it finds the `.perpetual` directory or reaches the file system root
- The project root directory cannot be a symlink

This approach ensures that `Perpetual` operates within the intended project scope.

## Project Size Limitations

Perpetual must balance comprehensive code analysis with the context window limitations of modern LLMs. To accomplish this, it uses a strategic approach for handling projects of various sizes.

### How Project Indexing Works

When working with your codebase, Perpetual doesn't attempt to feed all source code into the LLM at once, which would be:

- Impractical for small to medium projects
- Technically impossible for large projects due to token limits
- Unnecessarily expensive in terms of API usage

Instead, Perpetual uses a multi-stage approach:

1. **Annotation Phase**: The `annotate` operation generates concise summaries for each source file, capturing their purpose and functionality.
2. **Project Index**: These annotations form a project index that serves as a map of your codebase.
3. **Selective Loading**: For each operation, the LLM first reviews this index to identify which files are relevant to the current task.
4. **Focused Analysis**: Only then does the LLM examine the content of selected files in detail.

This approach allows Perpetual to work with significantly larger projects than would otherwise be possible.

### Practical Limitations

Despite these optimizations, there are still practical constraints:

- **Maximum Recommended Size**: Projects with more than 500 files may experience degraded performance. Beyond this threshold:
  - The project index itself can approach or exceed the context window size
  - LLM responses become less reliable as the complexity increases
  - You may hit rate limits with your LLM provider more frequently
  - Costs increase substantially

- **Performance Degradation Signs**: When working with larger projects, you might notice:
  - LLM hallucinations about non-existent files
  - Incorrect file selection for the tasks
  - Incomplete or inconsistent responses
  - Higher error rates during processing

### Mitigating Size Limitations

For larger projects, Perpetual offers several features to improve performance:

#### 1. Filename Salvaging (`-s` flag)

Operations like `implement`, `doc`, and `explain` support the `-s` flag which attempts to salvage incorrect filenames when the LLM hallucinates. This feature:

- Matches filenames by base name when paths are incorrect
- Helps recover from common LLM errors when listing project files
- Is particularly useful for projects with deep directory structures

#### 2. Context Saving Modes (`-c` flag)

Most operations support the `-c` flag to control annotation verbosity:

- `auto` (default): Applies context saving automatically based on file count
- `off`: Uses full annotations regardless of project size
- `medium`: Generates shorter annotations to save context space
- `high`: Creates minimal annotations focused only on critical details

For large projects, the system automatically applies:

- Medium context saving for 1000+ files
- High context saving for 2000+ files

**Important**: After changing the context saving mode, you must reannotate your project with the `-f` flag to regenerate all annotations with the new verbosity level.

#### 3. Multi-pass annotation

Perpetual employs a sophisticated multi-pass annotation system to optimize the quality and efficiency of file annotations, particularly valuable for large projects.

**Two-Stage Processing:**

- **First Stage**: Generates multiple annotation variants
- **Second Stage**: Applies intelligent selection or combination of these variants to create the final annotation

**Variant Selection Strategies:**
Perpetual supports several strategies for processing annotation variants, controlled by the LLM configuration:

- **Short Strategy**: Selects the most concise annotation variant, prioritizing token efficiency. This is the fallback strategy when other approaches fail and is particularly useful for very large projects where context space is at a premium.

- **Long Strategy**: Chooses the most detailed annotation variant, providing more comprehensive information when context space permits.

- **Combine Strategy**: Uses the LLM to intelligently merge multiple annotation variants, creating a synthesis that captures the most important information from each variant while avoiding redundancy.

- **Best Strategy**: Leverages the LLM's judgment to select the highest-quality annotation among the variants based on factors like informativeness, accuracy, and conciseness.

Multi-pass annotations must be enabled per-LLM basis using your `.env` configuration file or ENV variables.

#### 4. Using local LLM with Ollama to generate annotations

It is now possible to use local LLM and models like `qwen2.5-coder:7b-instruct` to generate annotations. It provides results good enough for use with Perpetual for most of the supported programming languages. This way you can save on costs using commercial LLM with larger projects.

#### 5. Selective File Processing

Consider these additional approaches for very large projects:

- Work with logical subsets of your project rather than the entire codebase
- Use operation-specific flags to exclude test files when appropriate
- Apply custom filters with the `-x` flag to focus on specific parts of your codebase

### Future Improvements

As LLM technology advances, we expect these limitations to become less restrictive. Future versions of Perpetual will leverage larger context windows and more efficient processing techniques as they become available, gradually improving performance with larger codebases.
