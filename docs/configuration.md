# Configuration

Perpetual employs a flexible and robust configuration system that allows you to tailor the application's behavior to suit your project's needs. The configuration is divided into two main types:

1. **Environment Configuration**: This includes machine-specific settings such as LLM provider details, API keys, and model configurations. These settings can be injected directly into the environment before running Perpetual or specified in `.env` files. They are specific to the machine or instance and should not be added to version control systems (VCS).

2. **Project Configuration**: These are project-specific settings defined in JSON format. They control aspects such as file selection filters, LLM prompts for different file types, and configurations for various tasks performed on the project's source files. These files reside in the `.perpetual` subdirectory and can be safely added to VCS.

## LLM Configuration

Perpetual relies on Large Language Models (LLMs) for various operations, and the LLM configuration is crucial for connecting to these services. This configuration includes provider details, API keys, and model parameters.

### Environment Variables and `.env` Files

LLM configurations are specified using environment variables, which can be set directly in your system environment or defined in `.env` files. Perpetual supports loading environment variables from different `.env` files, which are processed in the following order:

1. **System Environment Variables**: Perpetual first uses environment variables present in the current system environment.

2. **Project-Specific `.env` File**: Next, it tries to load the `.env` file located in the project's `.perpetual` directory, typically at `<project_root>/.perpetual/.env`.

3. **Global Configuration Directory**: Finally, it attempts to load the `.env` file from the global configuration directory:
   - **Unix/Linux**: `$HOME/.config/Perpetual/.env`
   - **Windows**: `%AppData%\Perpetual\.env`

Variables loaded earlier override those loaded later, allowing you to customize configurations at different levels.

When performing `Perpetual init -l <lang>`, an example `.env` file is placed at `<project_root>/.perpetual/.env.example`. Use this as a reference when creating your configuration. Note that `.env.example` **will not be loaded** by Perpetual.

### Key Environment Variables

The following environment variables are commonly used for LLM configuration:

- **LLM Provider Settings**:
  - `LLM_PROVIDER`: Specifies the LLM provider profile to use. Supported values include `openai`, `anthropic`, `ollama`, or `generic`. It can also include a profile number in formats like `openai1`, `openai2`, `generic3`, etc., allowing multiple distinct profiles for a single provider and enabling different configurations for different operations.

- **Authentication**:
  - `<PROFILE_NAME>_API_KEY`: Provider-specific API key for the LLM provider, such as `OPENAI_API_KEY` or `ANTHROPIC_API_KEY`. This is typically required for authentication.

Use `.env.example` as a reference; it contains sane defaults for different providers. Not all options are strictly required for each operation. Refer to the comments within the `.env.example` file for more detailed information. You can also remove provider-specific sections if not using a particular LLM provider.

### Important Notes

- **Security**: `.env` files may contain sensitive information (e.g., API keys) and should **not** be committed to version control.

- **Overriding Variables**: Environment variables set directly in the system environment have higher precedence over those defined in `.env` files.

## Project Configuration

Project configuration files allow for extensive customization of Perpetual's operations on a per-project basis. These configurations are stored as JSON files within the `.perpetual` subdirectory of your project.

### Configuration Files

The primary configuration files include:

- **Project Configuration**: Defines global project settings, such as which files can be selected for processing, which files are related to unit tests (and may be omitted when not needed), and how to map particular file types to Markdown code blocks.
  - `project.json`

- **Operation-Specific Configurations**: Customize behavior for specific operations.
  - `op_annotate.json`
  - `op_implement.json`
  - `op_doc.json`
  - `op_report.json`

### Configurable Parameters

#### `project.json`

Controls which files are included or excluded during processing using regular expressions:

- `project_files_whitelist`: An array of regex patterns specifying files to include.
- `project_files_blacklist`: An array of regex patterns specifying files to exclude.
- `project_test_files_blacklist`: An array of regex patterns to exclude test files.
- `files_to_md_code_mappings`: A two-dimensional array representing mappings from file types to Markdown code block languages. Each sub-array contains two elements: the first is a regex pattern matching the file type, and the second is the corresponding Markdown language identifier. You can skip filling up this field and provide an empty array - most popular source-file types will be detected automatically by their extension, this field is particularly useful when using non-standard file types.

**Example**:

```json
{
  "project_files_whitelist": ["(?i)^.*\\.go$"],
  "project_files_blacklist": ["(?i)^vendor(\\\\|\\/).*"],
  "project_test_files_blacklist": ["(?i)^.*_test\\.go$", "(?i)^.*(\\\\|\\/)test(\\\\|\\/).*\\.go$", "(?i)^test(\\\\|\\/).*\\.go$"],
  "files_to_md_code_mappings": [
    [".*\\.go$", "go"],
    [".*\\.py$", "python"],
    [".*\\.md$", "markdown"]
  ]
}
```

In this example:

- The `project_files_whitelist` includes all `.go` files (case insensitive).
- The `project_files_blacklist` excludes anything in the `vendor/` directory (case insensitive).
- The `project_test_files_blacklist` specifically excludes test-files across the project.
- The `files_to_md_code_mappings` maps `.go` files to `go` code blocks, `.py` files to `python` code blocks, and `.md` files to `markdown` code blocks in Markdown documentation.

#### `op_*.json` Config Files: LLM Prompts and Templates

Customize the prompts sent to the LLM for different operations and stages:

- **System Prompt**: A general prompt that sets the context for LLM interactions.
  - `system_prompt`

- **Operation-Specific Prompts**: Different stages of an operation can have unique prompts.
  - Examples include `stage1_prompts`, `stage2_prompt_variant`, `stage2_prompt_combine`, etc.

- **Response Templates**: Define expected response formats from the LLM when using the structured JSON output format. This feature is experimental and may not be supported by all LLM providers or models. If JSON output mode is disabled in LLM `.env` configuration, these templates are unused.
  - Examples include `stage1_output_schema`, `stage3_output_schema`, etc.

#### `op_*.json` Config Files: Mappings and Tags

Define how files are represented and parsed within LLM interactions:

- `filename_tags`: Tags used to wrap filenames when sending them to the LLM to identify and process them correctly.

- **Regular Expressions for Parsing LLM Responses**:
  - `filename_tags_rx`: Regex patterns to recognize tagged filenames in LLM responses.
  - `code_tags_rx`: Regex patterns to identify code blocks in responses.
  - `no_upload_comments_rx`: Regex patterns to detect comments indicating files should not be uploaded to LLM on implement and doc operations if requested by LLM.
  - Additional regex-based configurations as needed.

### Important Notes

- **Version Control**: Project configuration files are intended to be added to version control, ensuring consistency across different environments and team members.
- **Customization**: While default configurations are provided, you are encouraged to customize the prompts and settings to align with your project's requirements.
