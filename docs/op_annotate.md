# Annotate Operation

The `annotate` operation is a crucial part of `Perpetual`. It generates annotations for a project's source code files, creating a summary of each file's contents and purpose. This operation is primarily used to maintain an up-to-date index of the project's structure and content, which is then utilized by other operations within `Perpetual`. The project index is stored in the `.perpetual` directory as `.annotations.json` and is only updated when necessary, saving you costs and time on LLM API access.

While the `annotate` operation is an essential component of the `Perpetual` workflow, it is not typically necessary to run it manually. Other operations, such as `implement`, `doc`, `explain`, and brief `report` generation, trigger the `annotate` operation when needed to ensure that the project's annotations are current before proceeding with their tasks. Some operations provide a no-annotate mode to skip this automatic update.

## Usage

To manually run the `annotate` operation, use the following command:

```sh
Perpetual annotate [flags]
```

The `annotate` operation supports several command-line flags to customize its behavior:

- `-c <mode>`: Context saving mode, reducing LLM context use for large projects. Valid values are: `auto`, `off`, `medium`, `high`. The default is `auto`, which automatically determines whether to use context saving based on project size. When context saving is activated, `annotate` generates shorter file annotations to save tokens in later operations.

- `-f`: Force annotation of all files, even for files whose annotations are up to date. This flag is useful when you want to regenerate all annotations, regardless of whether the files have changed since the last annotation.

- `-d`: Perform a dry run without actually generating annotations. This flag lists the files that would be annotated without making LLM requests or updating `.annotations.json`. In this mode, logging is redirected to stderr so that stdout can contain only the file list.

- `-df <file>`: Optional path to a project description file for adding into LLM context. Valid values are a file path or `disabled`. If not specified, the operation attempts to load the project description from `.perpetual/description.md`.

- `-h`: Display the help message, showing all available flags and their descriptions.

- `-r <file>`: Only annotate a single specified file, even if its annotation is already up to date. This flag implies the `-f` flag. Use this when you want to update the annotation for a specific file. It may be useful if annotating all changed project files in a batch hits LLM API limits.

- `-x <file>`: Specify a path to a user-supplied regex filter file for filtering out certain files from processing. See more info about using the filter [here](user_filter.md).

- `-v`: Enable debug logging. This flag increases the verbosity of the operation's output, providing more detailed information about the annotation process.

- `-vv`: Enable both debug and trace logging. This flag provides the highest level of verbosity, useful for troubleshooting or understanding the internal workings of the annotation process.

### Examples

1. **Annotate only new or changed files:**

   ```sh
   Perpetual annotate
   ```

2. **Force (re)annotation of all files:**

   ```sh
   Perpetual annotate -f
   ```

3. **Annotate a specific file:**

   ```sh
   Perpetual annotate -r path/to/file.go
   ```

4. **Annotate with a custom project description:**

   ```sh
   Perpetual annotate -df custom_description.md
   ```

5. **List files that would be annotated without sending them to the LLM:**

   ```sh
   Perpetual annotate -d
   ```

When run, the `annotate` operation processes the specified files, or all changed files if no specific file is given, and generates or updates their annotations. These annotations are then stored in the project's configuration directory (`.perpetual/.annotations.json`) for use by other `Perpetual` operations.

## Tailoring Annotation Generation for Specific Project Files

You can instruct the LLM to alter annotations for specific files in a way you prefer. This may help to produce better and more concise annotations, remove non-relevant information, or draw attention to a specific part of the code. This customization works differently with various models, with Anthropic models often providing good results for this purpose. To achieve this, add a note near the beginning of your source file, for example:

```go
// NOTE for summarization: the summary for this file must only say "This file contains `goPrompts` struct that implements the `prompts` interface. Do not use this file directly.", nothing else.
```

You may add similar notes to other parts of your code. The LLM will use these hints to alter the generated annotations as instructed.

## LLM Configuration

The `annotate` operation can be configured using environment variables defined in `.env` files or exported to the system environment. These variables allow you to customize the behavior of the LLM used for generating annotations.

Configuration files with the `.env` extension may be placed in:

- the project-local `.perpetual` directory;
- the global Perpetual configuration directory;
- or values may be exported directly to the process environment.

Environment variables already exported to the process environment have the highest priority. Project-local `.env` files are loaded before global `.env` files. Within each directory, `.env` files are loaded in alphabetical order, with previously loaded values taking precedence. Operation-specific variables have priority over provider-wide defaults when supported by the selected provider.

### Key configuration options

1. **LLM Provider:**

   - `LLM_PROVIDER_OP_ANNOTATE`: Specifies the LLM provider to use for the `annotate` operation.
   - `LLM_PROVIDER`: Fallback provider if `LLM_PROVIDER_OP_ANNOTATE` is not set.

   Supported providers are:

   - `anthropic`
   - `openai`
   - `ollama`
   - `generic`

   Numbered provider profiles are also supported, for example `ollama1` or `generic2`, using corresponding environment variable prefixes such as `OLLAMA1_` or `GENERIC2_`.

2. **Model Selection:**

   - `ANTHROPIC_MODEL_OP_ANNOTATE`
   - `OPENAI_MODEL_OP_ANNOTATE`
   - `OLLAMA_MODEL_OP_ANNOTATE`
   - `GENERIC_MODEL_OP_ANNOTATE`

   If an operation-specific model variable is not set, the provider may fall back to the general model variable, such as `ANTHROPIC_MODEL`, `OPENAI_MODEL`, `OLLAMA_MODEL`, or `GENERIC_MODEL`.

3. **Token Limits:**

   - `ANTHROPIC_MAX_TOKENS_OP_ANNOTATE`
   - `OPENAI_MAX_TOKENS_OP_ANNOTATE`
   - `OLLAMA_MAX_TOKENS_OP_ANNOTATE`
   - `GENERIC_MAX_TOKENS_OP_ANNOTATE`

   These variables set the maximum number of output tokens for an annotation response. Keep this value modest because annotations from all files are later combined into a larger project index. If the annotation response reaches the token limit, this usually indicates that the source file is too complex or that the annotation prompt needs to be made more restrictive. Adding a summarization note to the source file can help produce a smaller annotation.

4. **Retry Settings:**

   - `ANTHROPIC_ON_FAIL_RETRIES_OP_ANNOTATE`
   - `OPENAI_ON_FAIL_RETRIES_OP_ANNOTATE`
   - `OLLAMA_ON_FAIL_RETRIES_OP_ANNOTATE`
   - `GENERIC_ON_FAIL_RETRIES_OP_ANNOTATE`

   These specify the number of retries on LLM query failure for the `annotate` operation.

5. **Temperature:**

   - `ANTHROPIC_TEMPERATURE_OP_ANNOTATE`
   - `OPENAI_TEMPERATURE_OP_ANNOTATE`
   - `OLLAMA_TEMPERATURE_OP_ANNOTATE`
   - `GENERIC_TEMPERATURE_OP_ANNOTATE`

   These set the temperature for the LLM during annotation. Lower values usually produce more deterministic output; higher values may produce more varied summaries.

6. **Other LLM Parameters:**

   Depending on the selected provider, additional operation-specific variables may be available, such as:

   - `ANTHROPIC_TOP_K_OP_ANNOTATE`
   - `ANTHROPIC_TOP_P_OP_ANNOTATE`
   - `ANTHROPIC_THINK_TOKENS_OP_ANNOTATE`
   - `OPENAI_REASONING_EFFORT_OP_ANNOTATE`
   - `OPENAI_TOP_P_OP_ANNOTATE`
   - `OPENAI_SERVICE_TIER_OP_ANNOTATE`
   - `OLLAMA_TOP_K_OP_ANNOTATE`
   - `OLLAMA_TOP_P_OP_ANNOTATE`
   - `OLLAMA_SEED_OP_ANNOTATE`
   - `OLLAMA_REPEAT_PENALTY_OP_ANNOTATE`
   - `GENERIC_TOP_K_OP_ANNOTATE`
   - `GENERIC_TOP_P_OP_ANNOTATE`
   - `GENERIC_SEED_OP_ANNOTATE`
   - `GENERIC_REPEAT_PENALTY_OP_ANNOTATE`
   - `GENERIC_FREQ_PENALTY_OP_ANNOTATE`
   - `GENERIC_PRESENCE_PENALTY_OP_ANNOTATE`

   These parameters are provider- and model-dependent. They are most often useful for local or OpenAI-compatible providers, especially Ollama or Generic providers.

### Example Configuration in `.env` File

```sh
LLM_PROVIDER="anthropic"
LLM_PROVIDER_OP_ANNOTATE="anthropic"

ANTHROPIC_API_KEY="<your api key goes here>"
ANTHROPIC_MODEL_OP_ANNOTATE="claude-3-haiku-20240307"
ANTHROPIC_MAX_TOKENS_OP_ANNOTATE="1024"
ANTHROPIC_TEMPERATURE_OP_ANNOTATE="0.5"
ANTHROPIC_ON_FAIL_RETRIES_OP_ANNOTATE="1"
```

This configuration uses the Anthropic provider with the Claude 3 Haiku model, sets a maximum of 1024 output tokens for annotations, uses a temperature of 0.5, and allows 1 retry on failure.

Note that if operation-specific variables with the `_OP_ANNOTATE` suffix are not set, the `annotate` operation will generally fall back to provider-wide variables for the chosen LLM provider. This allows you to set general defaults and override them specifically for annotation when needed.

**Warning:** The `annotate` operation processes project files selected by the project whitelist and blacklist, including source-code files marked with `###NOUPLOAD###` comments. To improve privacy, you may configure `Perpetual` to use a local LLM with Ollama for the `annotate` operation, or use a user filter file with `-x` to exclude sensitive files from annotation.

## Prompts Configuration

Customization of LLM prompts for the `annotate` operation is handled through the `.perpetual/op_annotate.json` configuration file. This file is populated by the `init` operation, which sets up default language-specific prompts tailored to your project's needs.

The key parameters within this configuration file are:

- **`system_prompt`**: The system prompt that establishes the LLM's role and instructions for the annotation task.

- **`system_prompt_ack`**: The acknowledgment message used when a provider needs to represent the system prompt as a user/assistant exchange.

- **`annotate_file_prompts`**: An array of file patterns and corresponding prompts for generating file annotations. Each entry contains exactly three strings:
  1. a regular expression to match file names;
  2. a prompt template for normal annotations;
  3. a prompt template for context-saving mode, usually producing shorter annotations.

- **`annotate_file_response`**: The simulated acknowledgment message used after the file annotation prompt and before the actual source file content is sent.

- **`annotate_task_prompt`**: Prompt used internally for generating task annotations from `###IMPLEMENT###` comments in source files. These task annotations may be used by other operations for local similarity search and file pre-selection.

- **`annotate_task_response`**: Simulated acknowledgment message for the task annotation prompt.

### Example `op_annotate.json` Configuration (partial, simplified)

```json
{
  "system_prompt": "You are a highly skilled technical documentation writer...",
  "system_prompt_ack": "Understood. I will respond accordingly in my subsequent replies.",
  "annotate_file_prompts": [
    [
      "(?i)^.*\\.go$",
      "Create a summary for the GO source file...",
      "Create a short summary for the GO source file..."
    ],
    [
      "(?i)^.*_test\\.go$",
      "Create a summary for the GO unit-tests source file...",
      "Create a short summary for the GO unit-tests source file..."
    ]
  ],
  "annotate_file_response": "Waiting for file contents",
  "annotate_task_prompt": "Create detailed summary of the tasks marked with \"###IMPLEMENT###\" comments...",
  "annotate_task_response": "Waiting for file contents"
}
```

The `annotate_file_prompts` entries are matched against project-relative file paths in the order they appear in the configuration. The first matching regular expression determines which prompts are used. When context saving is inactive, the normal annotation prompt is used. When context saving is active, the shorter prompt is used.

Related project-level configuration is stored in `.perpetual/project.json`. Important keys for annotation include:

- **`project_files_whitelist`**: Regex list selecting files that belong to the project source set.

- **`project_files_blacklist`**: Regex list excluding files from the project source set.

- **`project_description_prompt`** and **`project_description_response`**: Prompt and simulated response used when adding the project description to annotation context.

- **`filename_tags`**: Tags used to wrap file names when sending files to the LLM.

- **`files_to_md_code_mappings`**: Optional filename-to-Markdown-code-language mappings used when rendering source files in LLM prompts.

- **`code_tags_rx`**: Regex tag pairs used to detect and reject annotation responses that contain unwanted tagged text or code block wrapping.

- **Context-saving thresholds and percentages**, such as `medium_context_saving_file_count` and `high_context_saving_file_count`, which control automatic context-saving behavior.

## Workflow

1. **Initialization:**
   - The `annotate` operation begins by parsing command-line flags to determine its behavior.
   - It validates the context-saving mode.
   - It locates the project's root directory and the `.perpetual` configuration directory.
   - Environment variables are loaded from `.env` files, unless `annotate` is being called internally by another operation.
   - Configuration and prompts are loaded from `.perpetual/op_annotate.json` and `.perpetual/project.json`.
   - The project description is loaded from `.perpetual/description.md`, from a custom file specified with the `-df` flag, or skipped if `-df disabled` is used.

2. **File Discovery:**
   - The operation scans the project directory to identify source files to annotate.
   - The `.perpetual` directory is excluded from scanning.
   - Project whitelist and blacklist regex patterns from `project.json` are applied.
   - The resulting file list is checked for case-insensitive filename collisions.
   - File and directory names are checked to ensure they do not contain invalid path separator characters.
   - SHA-256 checksums are calculated for the selected project files to track changes since the last annotation.

3. **Annotation Decision:**
   - Based on the provided flags (`-f`, `-d`, `-r`), the operation determines which files require annotation.
   - Files that have not changed since the last annotation are skipped unless forced.
   - If `-r` is used, the requested path is resolved relative to the project root and matched against known project files.
   - If a user regex filter file is provided with `-x`, matching files are filtered out from the files selected for annotation.
   - In dry-run mode (`-d`), the operation lists the files that would be annotated and exits without making LLM requests.

4. **Context-Saving Selection:**
   - In `auto` mode, the operation switches to short annotation prompts when the project file count reaches the configured medium context-saving threshold.
   - In `medium` or `high` mode, short annotation prompts are always used.
   - In `off` mode, normal annotation prompts are used.

5. **Annotation Generation:**
   - The LLM connector is created for the `annotate` operation.
   - If files need annotation and this is a top-level `annotate` call, the raw LLM message log is rotated.
   - Files selected for annotation are sorted by size, smallest first.
   - For each selected file, the operation:
     - selects an appropriate prompt from `annotate_file_prompts` based on file pattern matching and context-saving mode;
     - reads the source file contents;
     - builds a message chain that may include the project description if available;
     - sends the prompt and file content to the LLM with file name tags and Markdown code block formatting;
     - handles retries on failure according to the configured retry limit;
     - treats token-limit responses as failures for that file;
     - filters and trims the LLM response;
     - stores the first valid filtered response as the file annotation.

6. **Error Handling:**
   - If a file fails to be annotated after all retries, its original checksum is preserved so it will be retried in a later run.
   - Processing continues with remaining files.
   - An error flag is set to indicate partial failure.
   - Empty responses, invalid responses, and responses containing forbidden code block or tag patterns are treated as failures for the affected file.

7. **Annotation Storage:**
   - Existing annotations are loaded from `.perpetual/.annotations.json`.
   - Newly generated annotations are merged with existing annotations.
   - Updated annotations and checksums are saved back to `.perpetual/.annotations.json`.
   - Only successfully annotated files receive updated checksums.

If any file fails to be annotated after the specified number of retries, the operation continues processing other files but exits with an error at the end, indicating that not all files were successfully annotated. Running the `annotate` operation again will attempt to process the failed files.
