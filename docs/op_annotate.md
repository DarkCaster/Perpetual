# Annotate Operation

The `annotate` operation is a crucial part of `Perpetual`. It generates annotations for a project's source code files, creating a summary of each file's contents and purpose. This operation is primarily used to maintain an up-to-date index of the project's structure and content, which is then utilized by other operations within `Perpetual`. The project index is stored in the `.perpetual` directory as `.annotations.json` and is only updated when necessary, saving you costs and time on LLM API access.

While the `annotate` operation is an essential component of the `Perpetual` workflow, it is not typically necessary to run it manually. Other operations, such as the `implement` operation, automatically trigger the `annotate` operation when needed to ensure that the project's annotations are current before proceeding with their tasks.

## Usage

To manually run the `annotate` operation, use the following command:

```sh
Perpetual annotate [flags]
```

The `annotate` operation supports several command-line flags to customize its behavior:

- `-c <mode>`: Context saving mode, reduce LLM context use for large projects. Valid values are: `auto`, `off`, `medium`, `high`. The default is `auto`, which automatically determines the appropriate context saving level based on project size. When context saving is activated, `annotate` generates shorter file annotations to save tokens on stage 1 for other operations.

- `-f`: Force annotation of all files, even for files whose annotations are up to date. This flag is useful when you want to regenerate all annotations, regardless of whether the files have changed since the last annotation.

- `-d`: Perform a dry run without actually generating annotations. This flag will list the files that would be annotated without making LLM requests and updating annotations.

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

When run, the `annotate` operation will process the specified files (or all changed files if no specific file is given) and generate or update their annotations. These annotations are then stored in the project's configuration directory (`.perpetual/.annotations.json`) for use by other `Perpetual` operations.

## Tailoring Annotation Generation for Specific Project Files

You can instruct the LLM to alter annotations for specific files in a way you prefer. This may help to produce better and more concise annotations, remove non-relevant information, or draw attention to a specific part of the code. This customization works differently with various models, with Anthropic models providing the best results for this purpose. To achieve this, add the following comment near the beginning of your source file (example):

```go
// NOTE for summarization: the summary for this file must only say "This file contains `GoPrompts` struct that implement `Prompts` interface. Consider not to use methods from this file directly.", nothing else.
```

You may add similar notes to other parts of your code. The LLM will use these hints to alter the generated annotations as specifically instructed.

## LLM Configuration

The `annotate` operation can be configured using environment variables defined in the `.env` file. These variables allow you to customize the behavior of the LLM (Large Language Model) used for generating annotations. Here are the key configuration options that affect the `annotate` operation:

1. **LLM Provider:**
   - `LLM_PROVIDER_OP_ANNOTATE`: Specifies the LLM provider to use for the `annotate` operation. If not set, it falls back to the general `LLM_PROVIDER`.
   - `LLM_PROVIDER_OP_ANNOTATE_POST`: Specifies the LLM provider for post-annotation processing when multiple variants are generated. If not set, it falls back to the general `LLM_PROVIDER`.

2. **Model Selection:**
   - `ANTHROPIC_MODEL_OP_ANNOTATE`: Specifies the Anthropic model to use for annotation (e.g., "claude-3-haiku-20240307").
   - `OPENAI_MODEL_OP_ANNOTATE`: Specifies the OpenAI model to use for annotation.
   - `OLLAMA_MODEL_OP_ANNOTATE`: Specifies the Ollama model to use for annotation.
   - `GENERIC_MODEL_OP_ANNOTATE`: Specifies the Generic provider (OpenAI compatible) model to use for annotation.

3. **Token Limits:**
   - `ANTHROPIC_MAX_TOKENS_OP_ANNOTATE`, `OPENAI_MAX_TOKENS_OP_ANNOTATE`, `OLLAMA_MAX_TOKENS_OP_ANNOTATE`, `GENERIC_MAX_TOKENS_OP_ANNOTATE`: Set the maximum number of tokens for the annotation response. The default is often set to 768 for annotations. Consider not using large values here because annotations from all files are joined together into the larger project index. Therefore, individual file annotations should remain small, and 768 is a reasonable limit. When hitting the token limit, this indicates that the source code file is too complex and you need to add some notes for summarization to make the annotation for this file smaller.

4. **Retry Settings:**
   - `ANTHROPIC_ON_FAIL_RETRIES_OP_ANNOTATE`, `OPENAI_ON_FAIL_RETRIES_OP_ANNOTATE`, `OLLAMA_ON_FAIL_RETRIES_OP_ANNOTATE`, `GENERIC_ON_FAIL_RETRIES_OP_ANNOTATE`: Specify the number of retries on failure for the `annotate` operation.

5. **Temperature:**
   - `ANTHROPIC_TEMPERATURE_OP_ANNOTATE`, `OPENAI_TEMPERATURE_OP_ANNOTATE`, `OLLAMA_TEMPERATURE_OP_ANNOTATE`, `GENERIC_TEMPERATURE_OP_ANNOTATE`: Set the temperature for the LLM during annotation. This affects the randomness of the output.

6. **Variant Generation and Selection:**
   - `ANTHROPIC_VARIANT_COUNT_OP_ANNOTATE`, `OPENAI_VARIANT_COUNT_OP_ANNOTATE`, `OLLAMA_VARIANT_COUNT_OP_ANNOTATE`, `GENERIC_VARIANT_COUNT_OP_ANNOTATE`: Number of annotation variants to generate.
   - `ANTHROPIC_VARIANT_SELECTION_OP_ANNOTATE`, `OPENAI_VARIANT_SELECTION_OP_ANNOTATE`, `OLLAMA_VARIANT_SELECTION_OP_ANNOTATE`, `GENERIC_VARIANT_SELECTION_OP_ANNOTATE`: Strategy for selecting or combining variants ("short", "long", "combine", "best").

7. **Other LLM Parameters:**
   - `TOP_K`, `TOP_P`, `SEED`, `REPEAT_PENALTY`, `FREQ_PENALTY`, `PRESENCE_PENALTY`: These parameters can be set specifically for the `annotate` operation by appending `_OP_ANNOTATE` to the variable name (e.g., `ANTHROPIC_TOP_K_OP_ANNOTATE`). They are mostly useful for the local Ollama provider and are not needed for OpenAI or Anthropic models.

### Example Configuration in `.env` File

```sh
LLM_PROVIDER="anthropic"
LLM_PROVIDER_OP_ANNOTATE="anthropic"
LLM_PROVIDER_OP_ANNOTATE_POST="anthropic"

ANTHROPIC_MODEL_OP_ANNOTATE="claude-3-haiku-20240307"
ANTHROPIC_MODEL_OP_ANNOTATE_POST="claude-3-haiku-20240307"
ANTHROPIC_MAX_TOKENS_OP_ANNOTATE="768"
ANTHROPIC_MAX_TOKENS_OP_ANNOTATE_POST="768"
ANTHROPIC_TEMPERATURE_OP_ANNOTATE="0.5"
ANTHROPIC_ON_FAIL_RETRIES_OP_ANNOTATE="1"
ANTHROPIC_VARIANT_COUNT_OP_ANNOTATE="1"
ANTHROPIC_VARIANT_SELECTION_OP_ANNOTATE="short"
```

This configuration uses the Anthropic provider with the Claude 3 Haiku model, sets a maximum of 768 tokens for annotations, uses a temperature of 0.5, and allows 1 retry on failure. It generates 1 annotation variant with the "short" selection strategy.

Note that if operation-specific variables (with the `_OP_ANNOTATE` suffix) are not set, the `annotate` operation will fall back to the general variables for the chosen LLM provider. This allows for flexible configuration where you can set general defaults and override them specifically for the `annotate` operation if needed.

**Warning:** The `annotate` operation will process all files, including source-code files marked with `###NOUPLOAD###` comments. To improve your privacy, you may configure `Perpetual` to use a local LLM with Ollama for the `annotate` operation only.

## Prompts Configuration

Customization of LLM prompts for the `annotate` operation is handled through the `.perpetual/op_annotate.json` configuration file. This file is populated using the `init` operation, which sets up default language-specific prompts tailored to your project's needs. The key parameters within this configuration file include:

- **`system_prompt`**: The system prompt that establishes the LLM's role and instructions for the annotation task.

- **`system_prompt_ack`**: The acknowledgment message that the LLM should use to confirm understanding of the system prompt.

- **`stage1_prompts`**: An array of file patterns and corresponding prompts for generating annotations. Each entry contains three elements: a regular expression to match file names, a prompt template for normal annotations, and a prompt template for context-saving mode (shorter annotations).

- **`stage1_response`**: The acknowledgment message used by the LLM after receiving the file to annotate.

- **`stage2_prompt_variant`**: Prompt used to request additional annotation variants when multiple variants are being generated.

- **`stage2_prompt_combine`**: Prompt used to combine multiple annotation variants into a final version.

- **`stage2_prompt_best`**: Prompt used to select the best annotation variant from multiple options.

- **`annotate_task_prompt`**: Prompt used for generating task annotations from `###IMPLEMENT###` comments in source files.

- **`annotate_task_response`**: Acknowledgment message for the task annotation prompt.

### Example `op_annotate.json` Configuration (partial, simplified)

```json
{
  "system_prompt": "You are a highly skilled technical documentation writer...",
  "system_prompt_ack": "I understand...",
  "stage1_prompts": [
    ["(?i)^.*\\.go$", "Create a summary for the GO source file...", "Create a short summary for the GO source file..."],
    ["(?i)^.*_test\\.go$", "Create a summary for the GO unit-tests source file...", "Create a short summary for the GO unit-tests source file..."]
  ],
  "stage1_response": "Waiting for file contents",
  "stage2_prompt_variant": "Create another summary variant",
  "stage2_prompt_combine": "Evaluate the summaries you have created and rework them into a final summary...",
  "stage2_prompt_best": "Evaluate the summaries you have created and choose summary variant that better matches...",
  "annotate_task_prompt": "Create detailed summary of the tasks marked with \"###IMPLEMENT###\" comments...",
  "annotate_task_response": "Waiting for file contents"
}
```

## Workflow

1. **Initialization:**
   - The `annotate` operation begins by parsing command-line flags to determine its behavior.
   - It locates the project's root directory and the `.perpetual` configuration directory.
   - Environment variables are loaded from `.env` files to configure the core LLM parameters.
   - Configuration and prompts are loaded from the `.perpetual/op_annotate.json` and `.perpetual/project.json` files.

2. **File Discovery:**
   - The operation scans the project directory to identify source code files to annotate, applying whitelist and blacklist regex patterns from the project configuration.
   - It calculates SHA-256 checksums for these files to track changes since the last annotation.
   - Files are sorted by size (smallest first) before processing to optimize LLM context usage.

3. **Annotation Decision:**
   - Based on the provided flags (`-f`, `-d`, `-r`), the operation determines which files require annotation.
   - Files that haven't changed since the last annotation are skipped unless forced.
   - In dry-run mode (`-d`), it lists the files that would be annotated without making any changes.

4. **Annotation Generation:**
   - For each selected file, the operation:
     - Selects an appropriate prompt from `stage1_prompts` based on file pattern matching and context saving mode
     - Sends the prompt and file content to the LLM with proper file tagging
     - Handles retries on failure according to the configured retry limit
     - Generates multiple variants if configured (`variant_count` > 1)
     - Processes variants according to the selected strategy:
       - **short**: Selects the shortest variant
       - **long**: Selects the longest variant
       - **combine**: Uses post-processing LLM to combine variants into final annotation
       - **best**: Uses post-processing LLM to select the best variant

5. **Error Handling:**
   - If a file fails to be annotated after all retries, its original checksum is preserved
   - Processing continues with remaining files
   - An error flag is set to indicate partial failure

6. **Annotation Storage:**
   - Generated annotations are merged with existing annotations and saved to `.perpetual/.annotations.json`
   - File checksums are updated to reflect the latest state of each successfully annotated file

If any file fails to be annotated after the specified number of retries, the operation will continue processing other files but will exit with an error at the end, indicating that not all files were successfully annotated. Running the `annotate` operation again will attempt to process the failed files.
