# Document Operation

The `doc` operation is designed to create or rework documentation files in markdown or plain-text format. This operation streamlines the process of generating and maintaining project documentation by producing high-quality, context-aware documentation based on your project's source code and existing documentation.

**NOTE**: This operation may provide unstable results. It is challenging to create clear and universal LLM prompts that yield reproducible results across all documentation types. For optimal results, use large and intelligent models. Be aware that the `doc` operation requires significantly more tokens and a larger context size than the `implement` operation. The effectiveness of using reasoning models is uncertain; they may produce less consistent results, but the writing style is typically better. Additionally, the cost of using reasoning models can be extremely high.

## Usage

To use the `doc` operation, run the following command:

```sh
Perpetual doc [flags]
```

The `doc` operation supports several command-line flags to customize its behavior:

- `-r <file>`: Specify the target documentation file for processing. This flag is optional. If omitted, the operation will read from standard input and write the result to standard output.

- `-e <file>`: Optionally specify a documentation file to use as an example or reference for style, structure, and format (but not for content). This helps maintain consistency across your project's documentation.

- `-a <action>`: Select the action to perform. Valid values are:
  - `draft`: Create an initial draft of the document with template content.
  - `write`: Write or complete an existing document (default).
  - `refine`: Refine and update an existing document.

- `-f`: Disable the 'no-upload' file filter and upload such files for review if requested. Use this flag with caution, as it may include sensitive information in the LLM's context.

- `-n`: Enable "No annotate" mode, which skips re-annotating changed files and uses current annotations if available. This can save time and API calls if you're confident your annotations are up-to-date.

- `-s`: Try to salvage incorrect filenames on stage 1. Experimental feature; use in projects with a large number of files where the LLM tends to make more mistakes when generating the list of files to analyze.

- `-u`: Do not exclude unit-test source files from processing. By default, unit-test sources are excluded.

- `-x <file>`: Specify a path to a user-supplied regex filter file for filtering out certain files from processing. See more info about using the filter [here](user_filter.md).

- `-h`: Display the help message, showing all available flags and their descriptions.

- `-v`: Enable debug logging for more detailed output during the operation.

- `-vv`: Enable both debug and trace logging for the highest level of verbosity.

### Examples

1. **Create a new document:**

   ```sh
   Perpetual doc -r docs/new_feature.md -a draft
   ```

   Edit the `docs/new_feature.md` draft by adding the most basic structure of the future document, your instructions, and notes about any aspect of the document starting with the words `Notes on implementation:`. After editing the draft and adding basic notes, section drafts, and the basic document structure, run:

   ```sh
   Perpetual doc -r docs/new_feature.md -a write
   ```

   You can also use another document as a style and structure reference when writing the current document:

   ```sh
   Perpetual doc -r docs/new_feature.md -e docs/old_feature.md -a write
   ```

2. **Refine an existing document using an example for style:**

   ```sh
   Perpetual doc -r docs/installation_guide.md -e docs/user_guide.md -a refine
   ```

3. **Write a document with debug logging enabled:**

   ```sh
   Perpetual doc -r docs/troubleshooting.md -v
   ```

4. **Create a document by reading from standard input and writing to standard output, using an example for style:**

   ```sh
   cat draft.md | Perpetual doc -a write -e docs/user_guide.md > final_document.md
   ```

5. **Use a custom regex filter file to exclude specific files:**

   ```sh
   Perpetual doc -r docs/overview.md -a refine -x ../exclude_regexes.json
   ```

When executed, the `doc` operation will analyze your project's structure, relevant source code, and existing documentation style (if provided) to generate or update the specified document. The operation uses a two-stage process:

1. **Stage 1:** Analyzes the project and determines which files are relevant for the documentation task based on the target document's content and notes.

2. **Stage 2:** Generates or refines the document content based on the analyzed information, the specified action, and any provided example document.

## LLM Configuration

The `doc` operation can be configured using environment variables defined in the `.env` file. These variables allow you to customize the behavior of the LLM used for generating documentation. Here are the key configuration options that affect the `doc` operation:

1. **LLM Provider:**
   - `LLM_PROVIDER_OP_DOC_STAGE1`: Specifies the LLM provider to use for the first stage of the `doc` operation.
   - `LLM_PROVIDER_OP_DOC_STAGE2`: Specifies the LLM provider to use for the second stage of the `doc` operation.
   - If not set, both stages fall back to the general `LLM_PROVIDER`.

2. **Model Selection:**
   - `ANTHROPIC_MODEL_OP_DOC_STAGE1`, `ANTHROPIC_MODEL_OP_DOC_STAGE2`: Specify the Anthropic models to use for each stage of documentation (e.g., "claude-3-sonnet-20240229" for stage 1 and "claude-3-opus-20240229" for stage 2).
   - `OPENAI_MODEL_OP_DOC_STAGE1`, `OPENAI_MODEL_OP_DOC_STAGE2`: Specify the OpenAI models to use for each stage of documentation.
   - `OLLAMA_MODEL_OP_DOC_STAGE1`, `OLLAMA_MODEL_OP_DOC_STAGE2`: Specify the Ollama models to use for each stage of documentation.
   - `GENERIC_MODEL_OP_DOC_STAGE1`, `GENERIC_MODEL_OP_DOC_STAGE2`: Specify the models to use for each stage when using the Generic provider (OpenAI compatible).

3. **Token Limits:**
   - `*_MAX_TOKENS_OP_DOC_STAGE1`, `*_MAX_TOKENS_OP_DOC_STAGE2`: Set the maximum number of tokens for each stage of the documentation response (replace * with provider name).
   - `*_MAX_TOKENS_SEGMENTS`: Specify the maximum number of continuation segments allowed when the LLM token limit is reached.

   For comprehensive documentation, consider using higher token limits (e.g., 4096 or more, if possible by model) for stage 2 to allow for detailed content generation. `Perpetual` will try to continue document generation if token limits are hit, but results may be suboptimal. If needed to generate a small document, it is generally better to set a larger token limit and limit document size with embedded instructions (`Notes on implementation:`) inside the document.

4. **JSON Structured Output Mode:**
   To enable JSON-structured output mode for the `doc` operation, set the appropriate environment variables in your `.env` file. This mode can be enabled for Stage 1 for all providers except Generic, providing faster responses and slightly lower costs. Note that not all models may support or work reliably with JSON-structured output.

   **Enable JSON-structured output mode:**

   ```sh
   ANTHROPIC_FORMAT_OP_DOC_STAGE1="json"
   OPENAI_FORMAT_OP_DOC_STAGE1="json"
   OLLAMA_FORMAT_OP_DOC_STAGE1="json"
   ```

   Replace * with the provider name (ANTHROPIC, OPENAI, OLLAMA, or GENERIC).

5. **Authentication and API Settings:**
   - For Generic provider:
     - `GENERIC_BASE_URL`: Required. Base URL for the API endpoint.
     - `GENERIC_AUTH_TYPE`: Authentication type ("basic" or "bearer").
     - `GENERIC_AUTH`: Authentication credentials.
     - `GENERIC_MAXTOKENS_FORMAT`: Format for max tokens parameter ("old" or "new").
     - `GENERIC_ENABLE_STREAMING`: Enable streaming mode (0 or 1).
     - `GENERIC_SYSPROMPT_ROLE`: System prompt role ("system", "developer", or "user").

   Similar authentication options exist for the Ollama provider, for use with public instances wrapped with an HTTPS reverse proxy.

6. **Common Parameters for all Providers:**
   - `*_ON_FAIL_RETRIES_OP_DOC_STAGE*`: Number of retries on failure.
   - `*_TEMPERATURE_OP_DOC_STAGE*`: Temperature setting (0.0-1.0).
   - `*_TOP_K_OP_DOC_STAGE*`: Top-K sampling parameter.
   - `*_TOP_P_OP_DOC_STAGE*`: Top-P sampling parameter.
   - `*_SEED_OP_DOC_STAGE*`: Random seed for reproducibility.
   - `*_REPEAT_PENALTY_OP_DOC_STAGE*`: Penalty for repeated tokens.
   - `*_FREQ_PENALTY_OP_DOC_STAGE*`: Frequency penalty.
   - `*_PRESENCE_PENALTY_OP_DOC_STAGE*`: Presence penalty.
   - `*_REASONING_EFFORT_OP_DOC_STAGE*`: Reasoning effort ("low", "medium", "high").

   Replace * with the provider name. Not all parameters are supported by all providers.

### Example Configuration in `.env` File

```sh
LLM_PROVIDER="generic"
GENERIC_BASE_URL="https://api.deepseek.com/v1"
GENERIC_AUTH_TYPE="Bearer"
GENERIC_AUTH="<deep seek api key>"
GENERIC_MODEL_OP_DOC_STAGE1="deepseek-chat"
GENERIC_MODEL_OP_DOC_STAGE2="deepseek-chat"
GENERIC_MAX_TOKENS_OP_DOC_STAGE1="1024"
GENERIC_MAX_TOKENS_OP_DOC_STAGE2="4096"
GENERIC_MAX_TOKENS_SEGMENTS="3"
GENERIC_TEMPERATURE_OP_DOC_STAGE1="0.5"
GENERIC_TEMPERATURE_OP_DOC_STAGE2="0.7"
GENERIC_ON_FAIL_RETRIES_OP_DOC_STAGE1="3"
GENERIC_ON_FAIL_RETRIES_OP_DOC_STAGE2="2"
GENERIC_FORMAT_OP_DOC_STAGE1="json"
GENERIC_MAXTOKENS_FORMAT="old"
GENERIC_ENABLE_STREAMING="1"
GENERIC_SYSPROMPT_ROLE="system"
```

This configuration uses the Generic provider configured for DeepSeek LLM, sets appropriate token limits and continuation segments, uses slightly different temperatures, and allows for retries on failure.

Note that if stage-specific variables are not set, the `doc` operation will fall back to the general variables for the chosen LLM provider. This allows for flexible configuration where you can set general defaults and override them specifically for each stage of the `doc` operation if needed.

## Prompts Configuration

Customization of LLM prompts for the `doc` operation is handled through the `.perpetual/op_doc.json` configuration file. This file is populated using the `init` operation, which sets up default language-specific prompts tailored to your project's needs. Since the `doc` operation is somewhat experimental and requires a robust and intelligent LLM model to work effectively, you may want to alter some of the prompts to better suit your process.

Other important parameters (not recommended to change unless having problems):

- **`filename_tags_rx`**: Regular expressions used to detect and parse the list of files for stage 1 LLM response if not using JSON structured output mode.

- **`filename_tags`**: Tagging conventions used to identify filenames within the annotations. This allows the LLM to recognize and process filenames accurately, facilitating better integration with the project's file structure.

- **`noupload_comments_rx`**: Regular expressions used to detect `no-upload` comments that mark source files forbidden to upload for processing due to privacy or other concerns.

- **`stage1_output_key`**, **`stage1_output_schema`**, **`stage1_output_schema_desc`**, **`stage1_output_schema_name`**: Parameters used if JSON structured output mode is enabled for stage 1 of the operation.

## Workflow

The `doc` operation follows a structured workflow to ensure efficient and accurate documentation generation:

1. **Initialization:**
   - The `doc` operation begins by parsing command-line flags to determine its behavior.
   - It locates the project's root directory and the `.perpetual` configuration directory.
   - Environment variables are loaded from `.env` files to configure the core LLM parameters.
   - Prompts and configuration are loaded from the `.perpetual/op_doc.json` file.

2. **File Discovery:**
   - The operation scans the project directory to locate project source code files, applying whitelist and blacklist regex patterns.
   - It automatically reannotates changed files unless the `-n` flag is used to skip this step.

3. **Documentation Generation or Refinement:**
   - **Stage 1:** Analyzes the project-index, target document content, and determines which files are relevant for the documentation task.
   - **Stage 2:** Generates or refines the document content based on the provided source files, analyzed information, and any example document provided.

## Best Practices

1. **Use Example Documents**: Use the `-e` flag to provide an example document. This helps maintain consistency in style and structure across your project's documentation. This is mostly useful for `write` actions to copy the writing style and structure from the reference document.

2. **Iterative Refinement**: Start with a `draft` action, then use `write` to complete the document, and finally `refine` to polish the content. This iterative approach often yields the best results. You should add instructions about the document topic, format, structure, and style inside the document draft (or the document you are about to rewrite or refine) in free form starting with the words `Notes on implementation:`. The LLM will follow these instructions when working on the document.

3. **Regular Updates**: As your project evolves, regularly use the `refine` action to keep your documentation up-to-date with the latest changes in your codebase.

4. **Review and Edit**: Always review and edit the generated documentation to ensure accuracy and add any project-specific nuances that the LLM might have missed.

5. **Version Control**: Keep your documentation files under version control along with your source code to precisely track changes made by the LLM.

6. **Use Filtering Options**: Utilize the `-u` flag to include unit test files in the documentation process when necessary. For more granular control, create a custom regex filter file and use it with the `-x` flag to exclude specific files or patterns from processing.

By leveraging the `doc` operation effectively, you can significantly streamline your documentation process, ensuring that your project's documentation remains comprehensive, up-to-date, and aligned with your codebase.
