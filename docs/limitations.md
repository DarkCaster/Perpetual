# Current Technical Limitations

This document outlines the current technical limitations. These limitations may be addressed in future updates.

## File Operations

Given `Perpetual`'s focus on direct codebase interaction and maintaining simplicity, it has some limitations on file operations:

- Cannot delete project files
- Cannot run external tools or commands on the user's system
- Cannot directly interact with version control systems (e.g., Git, SVN)
- Cannot install packages (e.g., npm, NuGet)
- Cannot automatically format, lint, build, or test generated code

These limitations are in place to ensure a controlled and safe environment for code manipulation.

`Perpetual` can create and modify files during implementation, but generated changes are applied through the stash mechanism. Because file deletion is not supported, rolling back a stash restores backed-up original files but does not remove newly created files that had no original version.

## Supported Source File Encoding

`Perpetual` supports the following text encodings for source files:

- UTF-8, with or without BOM
- UTF-16 LE and BE, with BOM
- UTF-32 LE and BE, with BOM

Additionally, if a file cannot be decoded as one of the above UTF encodings, `Perpetual` will attempt to use a fallback encoding. The fallback encoding can be set via the `FALLBACK_TEXT_ENCODING` environment variable. If it is not set, the default is `windows-1252`. The fallback encoding must be supported by the `golang.org/x/text/encoding/ianaindex` package.

When reading files, `Perpetual` performs the following operations:

1. Detects the file encoding by checking for BOMs and, if not found, assumes UTF-8 without BOM
2. Converts the content to UTF-8 for internal processing
3. Validates the UTF-8 encoding
4. If UTF-8 validation fails and a fallback encoding is available, converts using the fallback encoding

When writing files, `Perpetual` attempts to use the same encoding that was used when reading the file:

- If the file was originally read as one of the supported UTF encodings, it will be written back in that same encoding, including BOM if originally present
- If the file was read using the fallback encoding, it will be written back using the fallback encoding
- New files, or files that were not previously loaded by `Perpetual` during the current run, are written as plain UTF-8

Encoding preservation is therefore best-effort and depends on `Perpetual` having read the file during the current process. For example, applying or rolling back an old stash in a separate invocation may write files as UTF-8 if the original encoding parameters are no longer available.

This behavior minimizes unnecessary encoding changes in typical edit workflows.

## Line Endings (CR LF)

`Perpetual` handles line endings in the following manner:

- **Reading**: Converts CRLF to LF during file loading for internal processing
- **Writing**:
  - On Windows: Uses CRLF
  - On Linux and other non-Windows platforms: Uses LF
  - On macOS: Uses LF if compiled and run there, although macOS is not officially supported

This behavior is similar to Git's `core.autocrlf = true` setting.

**Important**: Mixed line endings within a file are not preserved during modifications. Files are normalized during loading and saving, and generated or modified content is written back using the platform-specific line ending style. This applies both to full-file generation and to incremental search-and-replace modifications.

## Symlinks

`Perpetual` has specific limitations regarding symlinks:

- Symlinked files and symlinked directories inside the project are not followed or processed by project file scanning
- The project root directory cannot be a symlink
- Parent directories of the project root can be symlinks

These limitations are in place to enhance security and simplify implementation using Go. Future versions may improve symlink handling.

## Filename Casing and Path Names

`Perpetual` enforces strict rules for filename casing to ensure consistency and prevent conflicts:

- Project files processed by `Perpetual` must not have the same file paths with different casing
  - This is particularly important for case-sensitive file systems, such as typical Linux file systems
  - On typical Windows file systems, such conflicts are usually prevented by the OS

When handling filenames, `Perpetual` attempts to:

- Match the case of existing project files using case-insensitive search
- Create necessary directories with correct casing when applying changes
- Detect and prevent case collisions in project file lists

In addition, file and directory names must not contain path separator characters as literal characters. In practice, `Perpetual` rejects project file path components containing `/` or `\`. This avoids ambiguous path handling across different operating systems.

## Project Root Detection

`Perpetual` uses a specific method to detect the project root:

- Searches for a `.perpetual` directory starting from the current working directory
- Moves up the directory tree until it finds the `.perpetual` directory or reaches the file system root
- The project root directory cannot be a symlink
- Can be overridden using the `PERPETUAL_DIR` environment variable

When `PERPETUAL_DIR` is used, it defines the Perpetual configuration directory, and the current working directory is treated as the project root.

Because project root detection happens before `.env` files are loaded, `PERPETUAL_DIR` must be set in the process environment before running `perpetual`; placing it only inside a project `.env` file is not sufficient for root discovery.

This approach ensures that `Perpetual` operates within the intended project scope.

## Project Size Limitations

`Perpetual` must balance comprehensive code analysis with the context window limitations of modern LLMs. To accomplish this, it uses a staged approach for handling projects of various sizes.

### How Project Indexing Works

When working with your codebase, `Perpetual` does not attempt to feed all source code into the LLM at once, which would be:

- Impractical for small to medium projects
- Technically impossible for large projects due to token limits
- Unnecessarily expensive in terms of API usage

Instead, `Perpetual` uses a staged workflow:

1. **Annotation Phase**: The `annotate` operation generates concise summaries for source files, capturing their purpose and functionality.
2. **Project Index**: These annotations form a project index that serves as a map of your codebase.
3. **File Selection**: For operations such as `implement`, `doc`, and `explain`, the LLM reviews this index to identify files relevant to the current task.
4. **Optional Local Similarity Search**: If embeddings are configured, local semantic search can pre-select files before stage 1 or add related files that the LLM may have missed.
5. **Focused Analysis**: Only then does the LLM examine the content of selected files in detail.

This approach allows `Perpetual` to work with significantly larger projects than would otherwise be possible.

### Practical Limitations

Despite these optimizations, there are still practical constraints:

- **No fixed hard file-count limit**: The maximum workable project size depends on model context size, annotation quality, file count, file sizes, prompt configuration, and provider limits.
- **Default context-saving thresholds**: Newly generated default project configuration enables medium context-saving behavior around several hundred files and high context-saving behavior for very large projects. The default thresholds are configurable in the project config.
- **Very large projects can still degrade**:
  - The project index itself can approach or exceed the context window size
  - LLM responses become less reliable as complexity increases
  - You may hit rate limits with your LLM provider more frequently
  - Costs increase substantially

### Performance Degradation Signs

When working with larger projects, you might notice:

- LLM hallucinations about non-existent files
- Incorrect file selection for tasks
- Relevant files being missed during file selection
- Incomplete or inconsistent responses
- Higher error rates during processing
- More frequent token-limit or context-limit failures

### Mitigating Size Limitations

For larger projects, `Perpetual` offers several features to improve performance.

#### 1. Context Saving Modes (`-c` flag)

Most high-level operations support the `-c` flag to control context usage:

- `auto` (default): Applies context saving automatically based on file-count thresholds defined in project configuration. Recommended.
- `off`: Disables context saving regardless of project size.
- `medium`: Uses medium context saving measures. This is generally suitable for larger projects but is not recommended for small projects.
- `high`: Uses stronger context saving measures. This may reduce result quality and is recommended only for large projects.

Each operation or sub-operation may have its own context saving behavior. Currently:

- `annotate` generates shorter annotations when context saving is active.
- `implement`, `doc`, and `explain` can use project-file pre-selection with local similarity search to reduce the annotation count sent to the LLM during stage 1.
- Additional local similarity search can also add relevant files after LLM file selection.

**Important**: If manually changing context saving mode, it is recommended to reannotate your project with the `-f` flag to regenerate all annotations with the selected verbosity level. It is also recommended to reannotate project files if the project reaches thresholds where automatic context saving changes from regular annotations to shorter annotations.

Currently, the context saving mode used for existing annotations is not saved. If you want to set it manually, you need to use `-c` on each invocation of the `perpetual` utility.

#### 2. Local Similarity Search

Operations like `implement`, `doc`, and `explain` may use additional local search with embeddings to add files to review that the LLM may have missed, or to locally reduce the annotation count sent to the LLM when context saving is enabled.

Local similarity search:

- Uses embeddings to find files semantically related to the current task
- Supports both aggressive and conservative file selection strategies
- Helps reduce the number of files the LLM needs to process
- Is particularly useful for projects with many files where only a subset is relevant

To use similarity search, you need to enable embeddings support for your provider. You can use embeddings generated by one provider with tasks processed by another provider, because searching with embeddings is performed locally. You must rebuild embeddings if you change the embeddings model or its settings. See the [`embed`](op_embed.md) operation documentation for more information.

If embeddings are disabled or not configured, local similarity search features are disabled. This affects context saving for `implement`, `doc`, and `explain`, and can also worsen results because `Perpetual` cannot locally add semantically related files that the LLM may have missed.

#### 3. Multi-pass File Selection

Operations like `implement`, `doc`, and `explain` support multi-pass file selection to select relevant files in multiple passes and improve the quality of the final result.

Use multi-pass selection with the `-sp` flag. This results in more API calls and token usage, but may improve quality for complex tasks. When context saving is active, multiple passes can also include different randomized subsets of files during pre-selection.

#### 4. Using a Local LLM with Ollama to Generate Annotations

It is possible to use a local LLM and models such as `qwen3:8b` or `qwen3:14b` to generate annotations. These models may provide results good enough to be used with `Perpetual` for many supported programming languages.

This allows you to save costs by using a local LLM in large projects while reserving more expensive cloud-based models for final implementation, documentation, or explanation stages.

#### 5. Selective File Processing

Consider these additional approaches for very large projects:

- Work with logical subsets of your project rather than the entire codebase
- Do not include unit-test files in processing unless needed; use the `-u` flag to include them where supported
- Apply custom filters with the `-x` flag to focus on specific parts of your codebase
- Use project whitelist and blacklist configuration to exclude generated files, vendored dependencies, build artifacts, and other low-value files

### Future Improvements

As LLM technology advances, these limitations are expected to become less restrictive. Future versions of `Perpetual` may leverage larger context windows, better local search, improved file selection, and more efficient processing techniques to improve performance with larger codebases.
