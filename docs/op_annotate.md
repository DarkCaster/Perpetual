# Annotate operation

The `annotate` operation is a crucial part of the `Perpetual`. It generates annotations for project source-code files, creating a summary of each file's contents and purpose. This operation is primarily used to maintain an up-to-date index of the project's structure and content, which is then utilized by other operations within the `Perpetual`. Project index is stored inside `.perpetual` directory and it only updated when necessary saving your costs and time on LLM API access.

While the `annotate` operation is an essential component of the `Perpetual` workflow, it is not typically necessary to run it manually. Other operations, such as the `implement` operation, automatically trigger the `annotate` operation when needed to ensure that the project's annotations are current before proceeding with their tasks.

## Usage

To manually run the `annotate` operation, use the following command:

```shell
./Perpetual annotate [flags]
```

The `annotate` operation supports several command-line flags to customize its behavior:

- `-f`: Force annotation of all files, even for files which annotations are up to date. This flag is useful when you want to regenerate all annotations, regardless of whether the files have changed since the last annotation.

- `-d`: Perform a dry run without actually generating annotations. This flag will list the files that would be annotated, without making LLM requests and updating annotations.

- `-h`: Display the help message, showing all available flags and their descriptions.

- `-r <file>`: Only annotate a single specified file, even if its annotation is already up to date. This flag implies the `-f` flag. Use this when you want to update the annotation for a specific file. May be useful if annotating all changed project files in a batch hits LLM API limits.

- `-v`: Enable debug logging. This flag increases the verbosity of the operation's output, providing more detailed information about the annotation process.

- `-vv`: Enable both debug and trace logging. This flag provides the highest level of verbosity, useful for troubleshooting or understanding the internal workings of the annotation process.

Examples:

1. Annotate only new or changed files:

   ```shell
   ./Perpetual annotate
   ```

2. Force (re)annotation of all files:

   ```shell
   ./Perpetual annotate -f
   ```

3. Annotate a specific file:

   ```shell
   ./Perpetual annotate -r path/to/file.go
   ```

When run, the `annotate` operation will process the specified files (or all changed files if no specific file is given) and generate or update their annotations. These annotations are then stored in the project's configuration directory for use by other `Perpetual` operations.

## Configuration

The `annotate` operation can be configured using environment variables defined in the `.env` file. These variables allow you to customize the behavior of the LLM (Large Language Model) used for generating annotations. Here are the key configuration options that affect the `annotate` operation:

1. LLM Provider:
   - `LLM_PROVIDER_OP_ANNOTATE`: Specifies the LLM provider to use for the `annotate` operation. If not set, it falls back to the general `LLM_PROVIDER`.

2. Model Selection:
   - `ANTHROPIC_MODEL_OP_ANNOTATE`: Specifies the Anthropic model to use for annotation (e.g., "claude-3-haiku-20240307").
   - `OPENAI_MODEL_OP_ANNOTATE`: Specifies the OpenAI model to use for annotation.
   - `OLLAMA_MODEL_OP_ANNOTATE`: Specifies the Ollama model to use for annotation.

3. Token Limits:
   - `ANTHROPIC_MAX_TOKENS_OP_ANNOTATE`, `OPENAI_MAX_TOKENS_OP_ANNOTATE`, `OLLAMA_MAX_TOKENS_OP_ANNOTATE`: Set the maximum number of tokens for the annotation response. Default is often set to 512 for annotations. Consider not to use large values here, because annotations from all files joined together into the bigger project index, so separate file-annotation should be small, 512 is a reasonable limit.

4. Retry Settings:
   - `ANTHROPIC_ON_FAIL_RETRIES_OP_ANNOTATE`, `OPENAI_ON_FAIL_RETRIES_OP_ANNOTATE`, `OLLAMA_ON_FAIL_RETRIES_OP_ANNOTATE`: Specify the number of retries on failure for the `annotate` operation.

5. Temperature:
   - `ANTHROPIC_TEMPERATURE_OP_ANNOTATE`, `OPENAI_TEMPERATURE_OP_ANNOTATE`, `OLLAMA_TEMPERATURE_OP_ANNOTATE`: Set the temperature for the LLM during annotation. This affects the randomness of the output.

6. Other LLM Parameters:
   - `TOP_K`, `TOP_P`, `SEED`, `REPEAT_PENALTY`, `FREQ_PENALTY`, `PRESENCE_PENALTY`: These parameters can be set specifically for the `annotate` operation by appending `_OP_ANNOTATE` to the variable name (e.g., `ANTHROPIC_TOP_K_OP_ANNOTATE`). Mostly useful for local Ollama provider, not needed to set with OpenAI or Anthropic models.

Example configuration in `.env` file:

```shell
LLM_PROVIDER="anthropic"
ANTHROPIC_MODEL_OP_ANNOTATE="claude-3-haiku-20240307"
ANTHROPIC_MAX_TOKENS_OP_ANNOTATE="512"
ANTHROPIC_TEMPERATURE_OP_ANNOTATE="0.5"
ANTHROPIC_ON_FAIL_RETRIES_OP_ANNOTATE="3"
```

This configuration uses the Anthropic provider with the Claude 3 Haiku model, sets a maximum of 512 tokens for annotations, uses a temperature of 0.5, and allows up to 3 retries on failure.

Note that if operation-specific variables (with `_OP_ANNOTATE` suffix) are not set, the `annotate` operation will fall back to the general variables for the chosen LLM provider. This allows for flexible configuration where you can set general defaults and override them specifically for the `annotate` operation if needed.

Warning: `annotate` will process all files, even source-code files marked with ###NOUPLOAD### comments. You may setup `Perpetual` to use local LLM with Ollama for `annotate` operation only in order to improve your privacy (use decent LLM such as deepseek-coder-33b-instruct or better).
