# Configuration

Perpetual employs a configuration system that allows you to tailor the application's behavior to suit your project's needs. The configuration is divided into two main types:

1. **Environment Configuration**  
   Machine- and instance-specific LLM settings such as LLM provider details, API keys, and model parameters. These can be set directly in the system environment or defined in `*.env` files (`.env` extension). They should **not** be committed to version control.

2. **Project Configuration**  
   Project-specific settings defined in JSON format. These control aspects such as file-selection filters, LLM prompts for different file types, response parsing tags, context-saving behavior, and operation-specific templates. These files reside in the `.perpetual` subdirectory and can usually be added to version control.

## LLM Configuration

Perpetual relies on Large Language Models (LLMs) for various operations. The LLM configuration includes provider selection, authentication, model parameters, embedding settings, and per-operation or per-stage overrides.

### Environment Variables and `.env` Files

LLM settings are read from environment variables and from `.env` files loaded by Perpetual. Loading occurs in this order:

1. **System Environment Variables**  
   Variables already set in your shell or operating system take highest priority.

2. **Project `*.env` Files**  
   All files ending in `.env` located in the project's `.perpetual` directory. They are loaded in alphabetical order; a variable already set by the system or by an earlier file is not overridden.

3. **Global `*.env` Files**  
   All files ending in `.env` in your global Perpetual config directory:
   - Unix/Linux: `$HOME/.config/Perpetual/`
   - Windows: `%AppData%\Roaming\Perpetual\`

When you run `perpetual init -l <lang>`, example files named `*.env.example` are created in `.perpetual` as a reference. **`*.env.example` files are not loaded** by Perpetual.

### Key Environment Variables

Use the generated `*.env.example` files as templates. Common settings include:

- **Directory Override**
  - `PERPETUAL_DIR`: Overrides the default `.perpetual` directory location. If not set, Perpetual searches for a `.perpetual` directory in the current directory and then parent directories. If `PERPETUAL_DIR` is set, the current working directory is treated as the project root.

- **Text Encoding**
  - `FALLBACK_TEXT_ENCODING`: Fallback encoding for files that cannot be read as UTF-8/16/32, for example `"windows-1252"`. Encoding names are resolved through `golang.org/x/text/encoding/ianaindex`.

- **Provider Selection**
  - `LLM_PROVIDER`: Default provider profile, e.g. `openai`, `anthropic`, `ollama`, or `generic`.
  - `LLM_PROVIDER_OP_<OPERATION>`: Operation- or stage-specific provider override.

  Operation names are uppercased internally. Examples:
  - `LLM_PROVIDER_OP_ANNOTATE`
  - `LLM_PROVIDER_OP_EMBED`
  - `LLM_PROVIDER_OP_IMPLEMENT_STAGE1`
  - `LLM_PROVIDER_OP_IMPLEMENT_STAGE2`
  - `LLM_PROVIDER_OP_IMPLEMENT_STAGE3`
  - `LLM_PROVIDER_OP_IMPLEMENT_STAGE4`
  - `LLM_PROVIDER_OP_DOC_STAGE1`
  - `LLM_PROVIDER_OP_DOC_STAGE2`
  - `LLM_PROVIDER_OP_EXPLAIN_STAGE1`
  - `LLM_PROVIDER_OP_EXPLAIN_STAGE2`

  You can append a numeric profile suffix to a provider name, for example `ollama1` or `generic2`. This makes Perpetual read settings from prefixes such as `OLLAMA1_` or `GENERIC2_`.

- **Authentication**  
  Options depend on the selected provider:
  - `<PROFILE>_API_KEY`: API key for providers that use this convention.
  - `<PROFILE>_AUTH`: API key, bearer token, or `login:password`.
  - `<PROFILE>_AUTH_TYPE`: `"Bearer"` or `"Basic"` where supported.

- **Model Selection**
  - `<PROFILE>_MODEL`: Default model name.
  - `<PROFILE>_MODEL_OP_<OPERATION>`: Model override for a specific operation or stage.
  - Embedding operations usually require an explicit `<PROFILE>_MODEL_OP_EMBED` value. Anthropic does not provide embedding support in Perpetual.

- **Common Generation Parameters**
  - `<PROFILE>_MAX_TOKENS`: Default maximum output tokens.
  - `<PROFILE>_MAX_TOKENS_OP_<OPERATION>`: Operation-specific maximum output tokens.
  - `<PROFILE>_MAX_TOKENS_SEGMENTS`: Maximum number of continuation segments when an operation supports continuing after token limit.
  - `<PROFILE>_ON_FAIL_RETRIES`: Default retry count for failed LLM calls.
  - `<PROFILE>_ON_FAIL_RETRIES_OP_<OPERATION>`: Operation-specific retry count.
  - `<PROFILE>_TEMPERATURE`: Default sampling temperature.
  - `<PROFILE>_TEMPERATURE_OP_<OPERATION>`: Operation-specific sampling temperature.
  - `<PROFILE>_TOP_P`, `<PROFILE>_TOP_K`, `<PROFILE>_SEED`, `<PROFILE>_REPEAT_PENALTY`, `<PROFILE>_FREQ_PENALTY`, `<PROFILE>_PRESENCE_PENALTY`: Provider-dependent tuning options where supported.

- **Streaming**
  - `<PROFILE>_ENABLE_STREAMING`
  - `<PROFILE>_ENABLE_STREAMING_OP_<OPERATION>`

  Streaming support is provider- and model-dependent. OpenAI and Generic providers expose streaming controls. Ollama uses streaming internally for response handling.

- **Incremental File-Change Mode**
  - `<PROFILE>_INCRMODE_SUPPORT`: Enables or disables incremental search-and-replace file modification mode.
  - `<PROFILE>_INCRMODE_SUPPORT_OP_<OPERATION>`: Operation-specific override.
  - `<PROFILE>_INCRMODE_RETRIES`: Number of retries for applying incremental changes before falling back to full-file generation.

- **Reasoning / Thinking Options**
  - Anthropic: `<PROFILE>_THINK_TOKENS` and `<PROFILE>_THINK_TOKENS_OP_<OPERATION>`.
  - Ollama: `<PROFILE>_THINK` and `<PROFILE>_THINK_OP_<OPERATION>`.
  - OpenAI/Generic: `<PROFILE>_REASONING_EFFORT` and `<PROFILE>_REASONING_EFFORT_OP_<OPERATION>` where supported by the provider and model.
  - Generic/Ollama response filtering: `<PROFILE>_THINK_RX_L`, `<PROFILE>_THINK_RX_R`, `<PROFILE>_OUT_RX_L`, and `<PROFILE>_OUT_RX_R`, with operation-specific variants, can be used to remove or extract parts of responses from models that include reasoning markup in plain text.

- **System Prompt Role**
  - `<PROFILE>_SYSPROMPT_ROLE`: Controls how the system prompt is sent for providers/models that need it.
  - Supported values are provider-dependent. Generic supports `system`, `developer`, and `user`; Ollama supports `system` and `user`.

- **Prompt Prefixes and Suffixes**
  - `<PROFILE>_SYSTEM_PFX`, `<PROFILE>_SYSTEM_SFX`
  - `<PROFILE>_USER_PFX`, `<PROFILE>_USER_SFX`
  - Operation-specific variants are also supported with `_OP_<OPERATION>` where implemented.

- **Embedding Settings**
  - `<PROFILE>_EMBED_DOC_CHUNK_SIZE`
  - `<PROFILE>_EMBED_DOC_CHUNK_OVERLAP`
  - `<PROFILE>_EMBED_SEARCH_CHUNK_SIZE`
  - `<PROFILE>_EMBED_SEARCH_CHUNK_OVERLAP`
  - `<PROFILE>_EMBED_DIMENSIONS`
  - `<PROFILE>_EMBED_SCORE_THRESHOLD`
  - `<PROFILE>_EMBED_DOC_PREFIX`
  - `<PROFILE>_EMBED_SEARCH_PREFIX`

- **Provider-Specific Request Options**
  - Generic provider:
    - `GENERIC_MAXTOKENS_FORMAT`: Selects old or new max-token request field format.
    - `GENERIC_API_VERSION`: Adds an `api-version` URL query parameter where needed.
    - `GENERIC_ADD_JSON`: Injects an additional JSON object into outgoing requests.
  - OpenAI provider:
    - `OPENAI_SERVICE_TIER`
    - `OPENAI_SERVICE_TIER_FALLBACK`
  - Ollama provider:
    - `OLLAMA_CONTEXT_SIZE`
    - `OLLAMA_CONTEXT_SIZE_OP_<OPERATION>`
    - `OLLAMA_CONTEXT_SIZE_LIMIT`
    - `OLLAMA_CONTEXT_MULT`
    - `OLLAMA_CONTEXT_ESTIMATE_MULT`
    - `OLLAMA_CONTEXT_MULT_MIN`

Refer to the comments within the generated `*.env.example` files for provider-specific defaults and additional options. You may create a single `.env` file or multiple `.env` files with settings for the providers you use.

**Security**: `.env` files may contain sensitive credentials. Do not commit them to version control.

## Project Configuration

Project configuration files allow you to customize Perpetual's behavior on a per-project basis. They are stored as JSON files under the `.perpetual` directory.

### Configuration Files

- **Global Project Settings**
  - `project.json`: Defines file-selection filters, Markdown code-block mappings, context-saving parameters, filename/code tags, no-upload markers, and incremental mode parsing settings.
  - `description.md`: Optional project description file that provides additional context to the LLM.

- **Operation-Specific Settings**
  - `op_annotate.json`: Prompts for source-file annotation and task annotation.
  - `op_implement.json`: Prompts and regexes for code implementation.
  - `op_doc.json`: Prompts for documentation generation and refinement.
  - `op_explain.json`: Prompts and output formatting for project explanation.
  - `op_report.json`: Prompts and filename formatting for report generation.

Perpetual validates these JSON files against built-in templates when loading them. Missing required keys or extra unknown keys cause configuration loading to fail.

### Generated and Ignored Runtime Files

`perpetual init` creates a `.gitignore` inside `.perpetual` that ignores runtime and sensitive files such as:

- `*.env`
- `.annotations.json`
- `.embeddings.msgpack`
- `.message_log.txt*`
- `.stash`

The operation JSON files and `project.json` are intended to be project configuration and are normally suitable for version control. Review `description.md` before committing it, because it may contain project-specific or sensitive context.

### `project.json` Parameters

`project.json` controls which files are included or excluded, how code is mapped to Markdown, and how context-saving behavior works.

File path regular expressions are matched against project-root-relative paths. For portability, default configs often use patterns that match both `/` and `\` separators.

- `project_files_whitelist`: Array of regex patterns for files to include.
- `project_files_blacklist`: Array of regex patterns for files to exclude.
- `project_test_files_blacklist`: Array of regex patterns used by operations that support excluding test files unless the operation is run with the option to include tests.
- `files_to_md_code_mappings`: A 2D array of `[pattern, language]` mappings for Markdown code blocks. If no mapping matches, Perpetual falls back to built-in extension-based mappings.
- `project_index_prompt`: Prompt used when presenting the project file index and annotations to the LLM.
- `project_index_response`: Simulated response paired with the project index prompt.
- `project_description_prompt`: Prompt used when adding `description.md` or another project description file to LLM context.
- `project_description_response`: Simulated response paired with the project description prompt.
- `filename_tags`: Tags used when embedding filenames in prompts.
- `filename_tags_rx`: Regex pairs used to parse filenames from LLM responses.
- `code_tags_rx`: Regex pairs used to parse code blocks from LLM responses.
- `noupload_comments_rx`: Regex patterns for comments that mark files as "no-upload".
- `medium_context_saving_file_count`: File count threshold for medium context-saving mode.
- `high_context_saving_file_count`: File count threshold for high context-saving mode.
- `medium_context_saving_select_percent`: Percentage of files to preselect in medium context-saving mode.
- `medium_context_saving_random_percent`: Percentage of randomized files in medium context-saving mode, calculated relative to the selected set.
- `high_context_saving_select_percent`: Percentage of files to preselect in high context-saving mode.
- `high_context_saving_random_percent`: Percentage of randomized files in high context-saving mode, calculated relative to the selected set.
- `files_incremental_mode_min_length`: 2D array of `[pattern, min_length]` records defining when incremental file-change mode may be used for a matching file.
- `files_incremental_mode_rx`: Regex patterns for parsing incremental search-and-replace blocks. The default format uses `SEARCH>>>`, `<<<REPLACE>>>`, and `<<<DONE`.

Partial example excerpt:

```json
{
  "project_files_whitelist": ["(?i)^.*\\.go$"],
  "project_files_blacklist": ["(?i)^vendor(\\\\|\\/).*"],
  "project_test_files_blacklist": ["(?i)^.*_test\\.go$"],
  "files_to_md_code_mappings": [
    [".*\\.go$", "go"],
    [".*\\.py$", "python"],
    [".*\\.md$", "markdown"]
  ],
  "medium_context_saving_file_count": 500,
  "high_context_saving_file_count": 1000,
  "medium_context_saving_select_percent": 60.0,
  "medium_context_saving_random_percent": 25.0,
  "high_context_saving_select_percent": 50.0,
  "high_context_saving_random_percent": 20.0,
  "files_incremental_mode_min_length": [
    [".*\\.go$", 1000],
    [".*\\.py$", 500]
  ],
  "files_incremental_mode_rx": [
    "(?m)(^|\\n)\\s*SEARCH>>>\\s*($|\\n)",
    "(?m)(^|\\n)\\s*<<<REPLACE>>>\\s*($|\\n)",
    "(?m)(^|\\n)\\s*<<<DONE\\s*($|\\n)"
  ]
}
```

This is only an excerpt. A real `project.json` must include all keys required by the template generated by `perpetual init`.

### Context-Saving Modes

Several operations support a context-saving flag:

```text
-c auto|off|medium|high
```

The mode affects how many project files are considered before asking the LLM to choose or process relevant files.

- `off`: disables context-saving preselection.
- `medium`: uses medium thresholds and percentages from `project.json`.
- `high`: uses high thresholds and percentages from `project.json`.
- `auto`: chooses context-saving behavior based on project file count.

For documentation, explanation, and implementation operations, context saving relies on local similarity search when embeddings are available. If embeddings are not configured or local search returns no results, Perpetual falls back to using the full available file list.

For annotation generation, context saving selects the shorter prompt variant from `annotate_file_prompts` when enabled.

### `op_*.json` Parameters

Operation configuration files define prompts, parsing tags, and operation-specific behavior.

#### Common Prompt Fields

Many operation configs include:

- `system_prompt`: The system context for the LLM.
- `system_prompt_ack`: A simulated acknowledgment used when a provider/model requires the system prompt to be represented as a user/assistant exchange.
- `code_prompt`: Prompt used before providing relevant source files to the LLM.
- `code_response`: Simulated response paired with `code_prompt`.

#### `op_annotate.json`

Annotation configuration includes:

- `system_prompt`
- `system_prompt_ack`
- `annotate_task_prompt`: Prompt used to summarize tasks marked with implementation comments.
- `annotate_task_response`: Simulated response for task annotation.
- `annotate_file_prompts`: Array of `[file_pattern, full_prompt, short_prompt]` records. The first matching file pattern selects the prompt. The short prompt is used when context saving is enabled for annotation.
- `annotate_file_response`: Simulated response before file contents are sent.

#### `op_implement.json`

Implementation configuration includes prompts and parsing settings for the implementation workflow:

- `implement_comments_rx`: Regex patterns used to find `###IMPLEMENT###` task comments.
- `filename_embed_rx`: Regex used to replace the filename placeholder in stage 4 prompts.

Stage 1 file-selection prompts:

- `stage1_analysis_prompt`
- `stage1_task_analysis_prompt`

Stage 2 review/planning prompts:

- `code_prompt`
- `code_response`
- `stage2_noplanning_prompt`
- `stage2_noplanning_response`
- `stage2_reasonings_prompt`
- `stage2_reasonings_prompt_final`
- `stage2_task_reasonings_prompt`
- `stage2_task_reasonings_prompt_final`

Stage 3 file-modification list prompts:

- `stage3_planning_prompt`
- `stage3_task_planning_prompt`
- `stage3_planning_lite_prompt`
- `stage3_task_extra_files_prompt`

Stage 4 code generation prompts:

- `stage4_changes_done_prompt`
- `stage4_changes_done_response`
- `stage4_process_prompt`
- `stage4_continue_prompt`
- `stage4_process_incremental_prompt`

#### `op_doc.json`

Documentation configuration includes:

- `system_prompt`
- `system_prompt_ack`
- `example_doc_prompt`
- `example_doc_response`
- `stage1_refine_prompt`
- `stage1_write_prompt`
- `code_prompt`
- `code_response`
- `stage2_refine_prompt`
- `stage2_write_prompt`
- `stage2_continue_prompt`

#### `op_explain.json`

Explanation configuration includes:

- `system_prompt`
- `system_prompt_ack`
- `output_files_header`
- `output_filename_tags`
- `output_filtered_filename_tags`
- `output_answer_header`
- `output_question_header`
- `stage1_question_prompt`
- `code_prompt`
- `code_response`
- `stage2_question_prompt`
- `stage2_continue_prompt`

The output formatting fields are used when Perpetual includes the original question and relevant file list in the generated answer.

#### `op_report.json`

Report configuration includes:

- `brief_prompt`: Prompt/header used when generating a brief report from annotations.
- `code_prompt`: Prompt/header used when generating a full source-code report.
- `filename_tags`: Tags used to render filenames in reports.

## Version Control and Customization

- Commit `project.json` and `op_*.json` files so all contributors use the same prompts, filters, and parsing rules.
- Do not commit real `.env` files.
- Review `description.md` before committing it, especially if it contains sensitive project details.
- Use the generated `*.env.example` files as references, but keep provider credentials and local model settings in `.env` files.
- You can customize prompts, regexes, and file filters to fit your project, but keep the required JSON keys intact because Perpetual rejects missing or extra keys during config loading.
