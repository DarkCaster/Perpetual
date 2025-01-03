# Annotate Operation

The `annotate` operation is a crucial part of `Perpetual`. It generates annotations for project source-code files, creating a summary of each file's contents and purpose. This operation is primarily used to maintain an up-to-date index of the project's structure and content, which is then utilized by other operations within `Perpetual`. The project index is stored in the `.perpetual` directory as `annotations.json` and is only updated when necessary, saving you costs and time on LLM API access.

While the `annotate` operation is an essential component of the `Perpetual` workflow, it is not typically necessary to run it manually. Other operations, such as the `implement` operation, automatically trigger the `annotate` operation when needed to ensure that the project's annotations are current before proceeding with their tasks.

## Usage

To manually run the `annotate` operation, use the following command:

```sh
Perpetual annotate [flags]
```

The `annotate` operation supports several command-line flags to customize its behavior:

- `-f`: Force annotation of all files, even for files whose annotations are up to date. This flag is useful when you want to regenerate all annotations, regardless of whether the files have changed since the last annotation.

- `-d`: Perform a dry run without actually generating annotations. This flag will list the files that would be annotated, without making LLM requests and updating annotations.

- `-h`: Display the help message, showing all available flags and their descriptions.

- `-r <file>`: Only annotate a single specified file, even if its annotation is already up to date. This flag implies the `-f` flag. Use this when you want to update the annotation for a specific file. It may be useful if annotating all changed project files in a batch hits LLM API limits.

- `-v`: Enable debug logging. This flag increases the verbosity of the operation's output, providing more detailed information about the annotation process.

- `-vv`: Enable both debug and trace logging. This flag provides the highest level of verbosity, useful for troubleshooting or understanding the internal workings of the annotation process.

### Examples:

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

When run, the `annotate` operation will process the specified files (or all changed files if no specific file is given) and generate or update their annotations. These annotations are then stored in the project's configuration directory (`.perpetual/annotations.json`) for use by other `Perpetual` operations.

## Tailoring Annotation Generation for Specific Project Files

You can instruct the LLM to alter annotations for specific files in a way you prefer. This may help to produce better and more concise annotations, remove non-relevant information, or draw attention to a specific part of the code. This customization works differently with various models, with Anthropic models providing the best results for this purpose. To achieve this, add the following comment near the beginning of your source file (example):

```go
// NOTE for summarization: the summary for this file must only say "This file contains `GoPrompts` struct that implement `Prompts` interface. Consider not to use methods from this file directly.", nothing else.
```

You may add similar notes to other parts of your code. The LLM will use these hints to alter the generated annotations as specifically instructed.

## Configuration

The `annotate` operation can be configured using environment variables defined in the `.env` file. These variables allow you to customize the behavior of the LLM (Large Language Model) used for generating annotations. Here are the key configuration options that affect the `annotate` operation:

1. **LLM Provider:**
   - `LLM_PROVIDER_OP_ANNOTATE`: Specifies the LLM provider to use for the `annotate` operation. If not set, it falls back to the general `LLM_PROVIDER`.
   - `LLM_PROVIDER_OP_ANNOTATE_POST`: Specifies the LLM provider for post-annotation processing. If not set, it falls back to the general `LLM_PROVIDER`.

2. **Model Selection:**
   - `ANTHROPIC_MODEL_OP_ANNOTATE`: Specifies the Anthropic model to use for annotation (e.g., "claude-3-haiku-20240307").
   - `OPENAI_MODEL_OP_ANNOTATE`: Specifies the OpenAI model to use for annotation.
   - `OLLAMA_MODEL_OP_ANNOTATE`: Specifies the Ollama model to use for annotation.

3. **Token Limits:**
   - `ANTHROPIC_MAX_TOKENS_OP_ANNOTATE`, `OPENAI_MAX_TOKENS_OP_ANNOTATE`, `OLLAMA_MAX_TOKENS_OP_ANNOTATE`: Set the maximum number of tokens for the annotation response. The default is often set to 512 for annotations. Consider not using large values here because annotations from all files are joined together into the larger project index. Therefore, individual file annotations should remain small, and 512 is a reasonable limit.

4. **Retry Settings:**
   - `ANTHROPIC_ON_FAIL_RETRIES_OP_ANNOTATE`, `OPENAI_ON_FAIL_RETRIES_OP_ANNOTATE`, `OLLAMA_ON_FAIL_RETRIES_OP_ANNOTATE`: Specify the number of retries on failure for the `annotate` operation.

5. **Temperature:**
   - `ANTHROPIC_TEMPERATURE_OP_ANNOTATE`, `OPENAI_TEMPERATURE_OP_ANNOTATE`, `OLLAMA_TEMPERATURE_OP_ANNOTATE`: Set the temperature for the LLM during annotation. This affects the randomness of the output.

6. **Other LLM Parameters:**
   - `TOP_K`, `TOP_P`, `SEED`, `REPEAT_PENALTY`, `FREQ_PENALTY`, `PRESENCE_PENALTY`: These parameters can be set specifically for the `annotate` operation by appending `_OP_ANNOTATE` to the variable name (e.g., `ANTHROPIC_TOP_K_OP_ANNOTATE`). They are mostly useful for the local Ollama provider and are not needed for OpenAI or Anthropic models.

### Example Configuration in `.env` File:

```sh
LLM_PROVIDER="anthropic"
LLM_PROVIDER_OP_ANNOTATE="anthropic"
LLM_PROVIDER_OP_ANNOTATE_POST="anthropic"

ANTHROPIC_MODEL_OP_ANNOTATE="claude-3-haiku-20240307"
ANTHROPIC_MODEL_OP_ANNOTATE_POST="claude-3-haiku-20240307"
ANTHROPIC_MAX_TOKENS_OP_ANNOTATE="512"
ANTHROPIC_TEMPERATURE_OP_ANNOTATE="0.5"
ANTHROPIC_ON_FAIL_RETRIES_OP_ANNOTATE="3"
ANTHROPIC_TOP_K_OP_ANNOTATE="40"
ANTHROPIC_TOP_P_OP_ANNOTATE="0.9"
ANTHROPIC_TOP_K="20"
```

This configuration uses the Anthropic provider with the Claude 3 Haiku model, sets a maximum of 512 tokens for annotations, uses a temperature of 0.5, and allows up to 3 retries on failure. Additionally, it sets `TOP_K` to 40 and `TOP_P` to 0.9 specifically for the `annotate` operation.

Note that if operation-specific variables (with the `_OP_ANNOTATE` suffix) are not set, the `annotate` operation will fall back to the general variables for the chosen LLM provider. This allows for flexible configuration where you can set general defaults and override them specifically for the `annotate` operation if needed.

**Warning:** The `annotate` operation will process all files, including source-code files marked with `###NOUPLOAD###` comments. To improve your privacy, you may configure `Perpetual` to use a local LLM with Ollama for the `annotate` operation only.
