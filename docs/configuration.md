# Configuration

Perpetual employs a configuration system that allows you to tailor the application's behavior to suit your project's needs. The configuration is divided into two main types:

1. **Environment Configuration**  
   Machine- and instance-specific LLM settings such as LLM provider details, API keys, and model parameters. These can be set directly in the system environment or defined in `*.env` files (`.env` extension). They should **not** be committed to version control.

2. **Project Configuration**  
   Project-specific settings defined in JSON format. These control aspects such as file-selection filters, LLM prompts for different file types, and operation-specific templates. These files reside in the `.perpetual` subdirectory and can be safely added to version control.

## LLM Configuration

Perpetual relies on Large Language Models (LLMs) for various operations. The LLM configuration includes provider selection, authentication, model parameters, and per-operation overrides.

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

Use `*.env.example` as templates. Common settings include:

- **Directory Override**  
  - `PERPETUAL_DIR`: Override the default `.perpetual` directory location. If not set, Perpetual attempts to automatically find `.perpetual` configuration directory in the current directory and parent directories.

- **Text Encoding**  
  - `FALLBACK_TEXT_ENCODING`: Fallback encoding for files that cannot be read as UTF-8/16/32 (e.g., `"windows-1252"`).

- **Provider Selection**  
  - `LLM_PROVIDER`: Default provider profile, e.g., `openai`, `anthropic`, `ollama`, or `generic`. You can append a profile number (e.g., `openai1`) to maintain multiple configurations.
  - `LLM_PROVIDER_OP_<OPERATION>`: Operation-specific provider override (e.g., `LLM_PROVIDER_OP_ANNOTATE`).

- **Authentication** (options depend on selected provider)  
  - `<PROFILE>_API_KEY` or `<PROFILE>_AUTH`: API key, token or `login:password` for the provider.
  - `<PROFILE>_AUTH_TYPE`: `"Bearer"` (API key/token) or `"Basic"` (login:password).

- **Model and Parameters**  
  - `<PROFILE>_MODEL`: Default model name (e.g., `OPENAI_MODEL="gpt-5"`).
  - `<PROFILE>_MODEL_OP_<OPERATION>`: Model override for a specific operation.
  - `<PROFILE>_TEMPERATURE`: Sampling temperature (`0.0`–`1.0`).
  - `<PROFILE>_MAX_TOKENS`: Maximum tokens per response.
  - `<PROFILE>_FORMAT_OP_<OPERATION>`: Response format (`plain` or `json`).
  - `<PROFILE>_VARIANT_COUNT`: Number of response variants to generate.
  - `<PROFILE>_VARIANT_SELECTION`: Strategy for selecting final variant (`short`, `long`, `combine`, or `best`).
  - `<PROFILE>_INCRMODE_SUPPORT`: Enable incremental file-change mode (`true`/`false`).
  - `<PROFILE>_THINK_TOKENS` (Anthropic) or `<PROFILE>_THINK` (Ollama): Enable reasoning/thinking for supported models.
  - `<PROFILE>_SYSPROMPT_ROLE`: System prompt role (`system`, `developer`, or `user`) for models without system prompt support.

Refer to the comments within `*.env.example` files for detailed defaults. You may create single or multiple `*.env` files with options for provider(s) you are using.

**Security**: `.env` files may contain sensitive credentials. Do not commit them to version control.

## Project Configuration

Project configuration files allow you to customize Perpetual's behavior on a per-project basis. They are stored in JSON files under the `.perpetual` directory.

### Configuration Files

- **Global Project Settings**  
  - `project.json`: Defines file-selection filters, Markdown code-block mappings, and context saving parameters.
  - `description.md`: Optional project description file that provides context to the LLM about your project.

- **Operation-Specific Settings**  
  - `op_annotate.json`: Prompts and templates for file annotation.
  - `op_implement.json`: Prompts, tags, and regexes for code implementation.
  - `op_doc.json`: Prompts and templates for documentation generation.
  - `op_explain.json`: Prompts and templates for project explanation.
  - `op_report.json`: Prompts for report generation.

### `project.json` Parameters

Controls which files are included or excluded, how code is mapped to Markdown, and context saving behavior:

- `project_files_whitelist`: Array of regex patterns for files to include.
- `project_files_blacklist`: Array of regex patterns for files to exclude.
- `project_test_files_blacklist`: Array of regex patterns to exclude test files.
- `files_to_md_code_mappings`: A 2D array of `[pattern, language]` mappings for Markdown code blocks.
- `project_index_prompt`: Prompt for generating project index.
- `project_index_response`: Response template for project index.
- `project_description_prompt`: Prompt for generating project description.
- `project_description_response`: Response template for project description.
- `filename_tags`: Tags used when embedding filenames in prompts.
- `filename_tags_rx`: Regex arrays to recognize tagged filenames in responses.
- `code_tags_rx`: Regex arrays to identify code blocks in responses.
- `noupload_comments_rx`: Regex arrays for comments that mark files as "no-upload".
- `medium_context_saving_file_count`: File count threshold for medium context saving mode.
- `high_context_saving_file_count`: File count threshold for high context saving mode.
- `medium_context_saving_select_percent`: Percentage of files to select in medium context saving.
- `medium_context_saving_random_percent`: Percentage of random files in medium context saving (calculated relative to the previous value).
- `high_context_saving_select_percent`: Percentage of files to select in high context saving.
- `high_context_saving_random_percent`: Percentage of random files in high context saving (calculated relative to the previous value).
- `files_incremental_mode_min_length`: Minimum file length for incremental mode.
- `files_incremental_mode_rx`: Regex patterns for incremental mode tags.

Example:

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

### `op_*.json` Parameters

#### Prompts and System Messages

- `system_prompt`: The initial system context for the LLM.
- `system_prompt_ack`: Acknowledgment message after the system prompt.
- Stage-specific prompts (e.g., `stage1_prompts`, `stage2_prompt_variant`).

#### Response Schemas (JSON Mode)

- `stage1_output_schema`, `stage3_output_schema`, etc.: Define expected JSON structure when using the JSON output mode.
- `stage1_output_key`, `stage3_output_key`: Key names for extracting data from JSON responses.
- `stage1_output_schema_name`, `stage3_output_schema_name`: Schema names for JSON mode.
- `stage1_output_schema_desc`, `stage3_output_schema_desc`: Schema descriptions for JSON mode.

**Version Control**: Operation configs in `.perpetual` should be committed to ensure consistency across environments.  
**Customization**: Feel free to adjust prompts, regexes, and schemas to fit your project's needs.
