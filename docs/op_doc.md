# Document Operation

The `doc` operation creates or reworks documentation files in Markdown or plain-text format. It streamlines generating and maintaining project documentation by producing context-aware documents based on your project's source code and any existing materials.

**Note:** Results may vary depending on the model used. The `doc` operation can use a large amount of tokens and context, especially for Stage 2 document generation. Use a capable model for best consistency and style. Reasoning models can improve style but may incur higher costs, and they are often better at creating initial documents with `-a write` rather than refining existing documents with `-a refine`. The `doc` operation is somewhat experimental for now.

## Usage

```sh
Perpetual doc [flags]
```

Available flags:

- `-r <file>`  
  Target documentation file for processing. If omitted, reads from stdin and writes to stdout.
- `-e <file>`  
  Example/reference document for style, structure, and format (not content).
- `-a <action>`  
  Action to perform (default: `write`):  
  - `draft`  Create an initial draft template.  
  - `write`  Write or complete an existing document.  
  - `refine` Refine and update an existing document.
- `-df <file>`  
  Optional path to project description file for adding into LLM context (valid values: file-path|disabled). If omitted, uses `.perpetual/description.md` if available.
- `-c <mode>`  
  Context saving mode: `auto` (default), `off`, `medium`, or `high`. Controls how aggressively LLM context usage is reduced on large projects.
- `-s <limit>`  
  Limit number of files for local similarity search via embeddings (default: 7; 0 disables local search and only uses LLM-requested files).
- `-sp <count>`  
  Set number of passes for related files selection at Stage 1 (default: 1). Higher pass-count values will select more files, compensating for possible LLM errors when finding relevant files, but it will cost you more tokens and context use.
- `-f`  
  Disable the `no-upload` file filter and include such files for review if requested.
- `-n`  
  No-annotate mode: skip re-annotating changed files and use current annotations and embeddings if available.
- `-u`  
  Include unit-test source files in processing (tests excluded by default).
- `-x <file>`  
  Path to a user-supplied regex filter file to exclude certain files. See more information about using the filter [here](user_filter.md).
- `-v`  
  Enable debug logging.
- `-vv`  
  Enable debug and trace logging.
- `-h`  
  Show help and exit.

### Examples

1. **Draft a new document template:**

   ```sh
   Perpetual doc -r docs/new_feature.md -a draft
   ```

   Then, edit the `docs/new_feature.md` draft by adding the most basic structure of the future document, your instructions, and notes about any aspect of the document starting with the words `Notes on implementation:`.

2. **Write or complete a draft:**

   ```sh
   Perpetual doc -r docs/new_feature.md -a write
   ```

3. **As an alternative, write using an example for style:**

   ```sh
   Perpetual doc -r docs/new_feature.md -e docs/old_feature.md -a write
   ```

4. **Refine an existing document:**

   ```sh
   Perpetual doc -r docs/installation_guide.md -e docs/user_guide.md -a refine
   ```

5. **Read from stdin, write to stdout:**

   ```sh
   cat draft.md | Perpetual doc -a write -e docs/user_guide.md > final_doc.md
   ```

6. **Exclude files via custom regex filter:**

   ```sh
   Perpetual doc -r docs/overview.md -a refine -x ../exclude_regexes.json
   ```

7. **Use custom project description file:**

   ```sh
   Perpetual doc -r docs/api_reference.md -df custom_description.md -a write
   ```

8. **Disable project description:**

   ```sh
   Perpetual doc -r docs/quick_start.md -df disabled -a write
   ```

## How It Works

When executed, the `doc` operation will analyze your project's structure, relevant source code, and existing documentation style (if provided) to generate or update the specified document. The operation uses a two-stage LLM process for `write` and `refine` actions:

1. **Stage 1:** Analyzes the project index, target document content, optional project description, optional example document, and any embedded instructions. It determines which project files are relevant for the documentation task.
2. **Stage 2:** Generates or refines the document content based on the selected source files, project annotations, the specified action, the target document, and any provided example document.

The `draft` action does not call the LLM. It writes a built-in Markdown draft template to the target file or stdout.

If embeddings are configured, the operation can also use local similarity search:

- Before Stage 1, context saving can preselect a subset of project files for large projects.
- After Stage 1, similarity search can add more relevant files based on the document content.

If embeddings are not configured or fail, the operation falls back to using the normal project file list and LLM-selected files.

## LLM Configuration

The `doc` operation can be configured using environment variables defined in `.env` files. Project-local `.env` files are loaded from `.perpetual`, and global `.env` files are loaded from the Perpetual configuration directory. These variables allow you to customize the behavior of the LLM used for generating documentation. Here are the key configuration options that affect the `doc` operation:

1. **LLM Provider:**
   - `LLM_PROVIDER_OP_DOC_STAGE1`: Specifies the LLM provider to use for the first stage of the `doc` operation.
   - `LLM_PROVIDER_OP_DOC_STAGE2`: Specifies the LLM provider to use for the second stage of the `doc` operation.
   - If not set, both stages fall back to the general `LLM_PROVIDER`.

2. **Model Selection:**
   - `ANTHROPIC_MODEL_OP_DOC_STAGE1`, `ANTHROPIC_MODEL_OP_DOC_STAGE2`: Specify the Anthropic models to use for each documentation stage.
   - `OPENAI_MODEL_OP_DOC_STAGE1`, `OPENAI_MODEL_OP_DOC_STAGE2`: Specify the OpenAI models to use for each documentation stage.
   - `OLLAMA_MODEL_OP_DOC_STAGE1`, `OLLAMA_MODEL_OP_DOC_STAGE2`: Specify the Ollama models to use for each documentation stage.
   - `GENERIC_MODEL_OP_DOC_STAGE1`, `GENERIC_MODEL_OP_DOC_STAGE2`: Specify the models to use for each stage when using the Generic provider (OpenAI-compatible API).

3. **Token Limits:**
   - `<PROVIDER>_MAX_TOKENS_OP_DOC_STAGE1`, `<PROVIDER>_MAX_TOKENS_OP_DOC_STAGE2`: Set the maximum number of output tokens for each documentation stage (replace `<PROVIDER>` with `ANTHROPIC`, `OPENAI`, `OLLAMA`, or `GENERIC`).
   - `<PROVIDER>_MAX_TOKENS_SEGMENTS`: Specify the maximum number of continuation segments allowed when the LLM token limit is reached.

   For comprehensive documentation, consider using higher token limits for Stage 2 to allow for detailed content generation. `Perpetual` will try to continue document generation if token limits are hit, but results may be suboptimal. If generating a smaller document, it is generally better to set a larger token limit and limit document size with embedded instructions starting with `Notes on implementation:` inside the document.

4. **JSON Structured Output Mode:**
   JSON-structured output can be enabled for Stage 1 for providers that support it. It can make file-list extraction more reliable for compatible models, but not all models support or work reliably with structured output.

   **Enable JSON-structured output mode:**

   ```sh
   ANTHROPIC_FORMAT_OP_DOC_STAGE1="json"
   OPENAI_FORMAT_OP_DOC_STAGE1="json"
   OLLAMA_FORMAT_OP_DOC_STAGE1="json"
   ```

   The Generic provider currently uses plain text output mode for documentation stages.

5. **Embeddings and Local Similarity Search:**
   The `doc` operation may call the `embed` operation internally to update embeddings unless `-n` is used. Embeddings are used for local similarity search and context-saving preselection.

   Relevant variables include:

   - `LLM_PROVIDER_OP_EMBED`: Provider used for embeddings.
   - `<PROVIDER>_MODEL_OP_EMBED`: Embedding model for the selected provider.
   - `<PROVIDER>_EMBED_DOC_CHUNK_SIZE`, `<PROVIDER>_EMBED_DOC_CHUNK_OVERLAP`: Chunking settings for project files.
   - `<PROVIDER>_EMBED_SEARCH_CHUNK_SIZE`, `<PROVIDER>_EMBED_SEARCH_CHUNK_OVERLAP`: Chunking settings for search queries.
   - `<PROVIDER>_EMBED_DIMENSIONS`: Optional embedding vector dimension count, if supported.
   - `<PROVIDER>_EMBED_SCORE_THRESHOLD`: Minimum cosine similarity score accepted by local search.
   - `<PROVIDER>_EMBED_DOC_PREFIX`, `<PROVIDER>_EMBED_SEARCH_PREFIX`: Optional model-specific text prefixes for embedding generation.

   Anthropic does not provide embedding support in Perpetual. OpenAI, Ollama, and Generic providers may support embeddings depending on the configured model and API.

6. **Authentication and API Settings:**
   - For the Generic provider:
     - `GENERIC_BASE_URL`: Required. Base URL for the API endpoint.
     - `GENERIC_AUTH_TYPE`: Authentication type (`basic` or `bearer`).
     - `GENERIC_AUTH`: Authentication credentials.
     - `GENERIC_MAXTOKENS_FORMAT`: Format for the max tokens parameter (`old` or `new`).
     - `GENERIC_ENABLE_STREAMING`: Enable streaming mode (`0` or `1`).
     - `GENERIC_SYSPROMPT_ROLE`: System prompt role (`system`, `developer`, or `user`).

   Similar authentication options exist for the Ollama provider, for use with public instances wrapped with an HTTPS reverse proxy. OpenAI also supports service tier configuration through `OPENAI_SERVICE_TIER_OP_DOC_STAGE1`, `OPENAI_SERVICE_TIER_OP_DOC_STAGE2`, and related fallback settings.

7. **Common Parameters for Providers:**
   - `<PROVIDER>_ON_FAIL_RETRIES_OP_DOC_STAGE1`, `<PROVIDER>_ON_FAIL_RETRIES_OP_DOC_STAGE2`: Number of retries on failure.
   - `<PROVIDER>_TEMPERATURE_OP_DOC_STAGE1`, `<PROVIDER>_TEMPERATURE_OP_DOC_STAGE2`: Temperature setting.
   - `<PROVIDER>_TOP_K_OP_DOC_STAGE1`, `<PROVIDER>_TOP_K_OP_DOC_STAGE2`: Top-K sampling parameter, where supported.
   - `<PROVIDER>_TOP_P_OP_DOC_STAGE1`, `<PROVIDER>_TOP_P_OP_DOC_STAGE2`: Top-P sampling parameter, where supported.
   - `<PROVIDER>_SEED_OP_DOC_STAGE1`, `<PROVIDER>_SEED_OP_DOC_STAGE2`: Random seed for reproducibility, where supported.
   - `<PROVIDER>_REPEAT_PENALTY_OP_DOC_STAGE1`, `<PROVIDER>_REPEAT_PENALTY_OP_DOC_STAGE2`: Penalty for repeated tokens, where supported.
   - `<PROVIDER>_FREQ_PENALTY_OP_DOC_STAGE1`, `<PROVIDER>_FREQ_PENALTY_OP_DOC_STAGE2`: Frequency penalty, where supported.
   - `<PROVIDER>_PRESENCE_PENALTY_OP_DOC_STAGE1`, `<PROVIDER>_PRESENCE_PENALTY_OP_DOC_STAGE2`: Presence penalty, where supported.

   Provider-specific reasoning options include:

   - `OPENAI_REASONING_EFFORT_OP_DOC_STAGE1`, `OPENAI_REASONING_EFFORT_OP_DOC_STAGE2`
   - `GENERIC_REASONING_EFFORT_OP_DOC_STAGE1`, `GENERIC_REASONING_EFFORT_OP_DOC_STAGE2`
   - `ANTHROPIC_THINK_TOKENS_OP_DOC_STAGE1`, `ANTHROPIC_THINK_TOKENS_OP_DOC_STAGE2`
   - `OLLAMA_THINK_OP_DOC_STAGE1`, `OLLAMA_THINK_OP_DOC_STAGE2`

   Replace `<PROVIDER>` with `ANTHROPIC`, `OPENAI`, `OLLAMA`, or `GENERIC`. Not all parameters are supported by all providers.

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
GENERIC_MAXTOKENS_FORMAT="old"
GENERIC_ENABLE_STREAMING="1"
GENERIC_SYSPROMPT_ROLE="system"
```

This configuration uses the Generic provider configured for DeepSeek LLM, sets appropriate token limits and continuation segments, uses slightly different temperatures, and allows for retries on failure.

Note that if stage-specific variables are not set, the `doc` operation will fall back to the general variables for the chosen LLM provider. This flexible configuration allows you to set general defaults while overriding them specifically for each stage of the `doc` operation if needed.

## Prompts Configuration

Customization of LLM prompts for the `doc` operation is handled through the `.perpetual/op_doc.json` configuration file. This file is populated using the `init` operation, which sets up default language-specific prompts tailored to your project's needs. Since the `doc` operation is experimental and requires a robust and intelligent LLM model to work effectively, you may want to alter some of the prompts to better suit your process.

Other important parameters (not recommended to change unless you are experiencing problems):

- **`stage1_output_key`**, **`stage1_output_schema`**, **`stage1_output_schema_desc`**, **`stage1_output_schema_name`**: Parameters used if JSON-structured output mode is enabled for Stage 1 of the operation.

## Workflow

The `doc` operation follows a structured workflow to ensure efficient and accurate documentation generation:

1. **Initialization:**
   - The operation begins by parsing command-line flags to determine its behavior.
   - It locates the project's root directory and the `.perpetual` configuration directory.
   - Environment variables are loaded from `.env` files to configure the core LLM parameters.
   - For `write` and `refine`, prompts and configuration are loaded from the `.perpetual/op_doc.json` and `.perpetual/project.json` files.

2. **File Discovery:**
   - The operation scans the project directory to locate source code files, applying whitelist and blacklist regular expressions.
   - Unit-test files are excluded by default unless `-u` is used.
   - A user-supplied regex blacklist can be appended with `-x`.
   - It automatically reannotates changed files unless the `-n` flag is used to skip this step.
   - Embeddings are generated or updated for similarity search unless the `-n` flag is used or embeddings are not configured.

3. **Documentation Generation or Refinement:**
   - **Preselection:** If context saving is enabled and embeddings are available, the operation may preselect a subset of project files before Stage 1.
   - **Stage 1:** Analyzes the project index, target document content, optional example document, and any provided project description to determine which files are relevant for the documentation task.
   - **Local Similarity Search:** If enabled, local similarity search may add files related to the target document.
   - **No-upload Filtering:** Unless `-f` is used, files marked with configured `no-upload` comments are excluded before their contents are sent to the LLM.
   - **Stage 2:** Generates or refines the document content based on the provided source files, annotations, analyzed information, and any example document provided.

## Best Practices

1. **Use Example Documents:** Use the `-e` flag to provide an example document. This helps maintain consistency in style and structure across your project's documentation. It is especially useful for `write` actions to copy the writing style and structure from the reference document.

2. **Iterative Refinement:** Start with a `draft` action, then use `write` to complete the document, and finally `refine` to polish the content. This iterative approach often yields the best results. Include instructions about the document topic, format, structure, and style inside the document draft (or the document you are about to rewrite or refine) in free form starting with the words `Notes on implementation:`. The LLM will follow these instructions when working on the document.

3. **Regular Updates:** As your project evolves, regularly use the `refine` action to keep your documentation up to date with the latest changes in your codebase.

4. **Review and Edit:** Always review and edit the generated documentation to ensure accuracy and add any project-specific nuances that the LLM might have missed.

5. **Version Control:** Keep your documentation files under version control along with your source code to precisely track changes made by the LLM.

6. **Use Filtering Options:** Utilize the `-u` flag to include unit test files in the documentation process when necessary. For more granular control, create a custom regex filter file and use it with the `-x` flag to exclude specific files or patterns from processing.

7. **Project Description:** Fill in the project description at `.perpetual/description.md` from the provided template, or use the `-df` flag to read it from a different file. This will populate LLM context with extra description about your project. It helps the LLM better understand the project's purpose and architecture, leading to more relevant and accurate documentation.

8. **Configure Embeddings for Large Projects:** For large projects, configure an embedding model and keep embeddings up to date. This improves local similarity search and context-saving preselection, reducing unnecessary context usage.

By leveraging the `doc` operation effectively, you can significantly streamline your documentation process, ensuring that your project's documentation remains comprehensive, up to date, and aligned with your codebase.
