# Versioning Policy

Starting from **v3.0.0**, the following versioning policy is implemented:

- **Versions 3.0.x (Bugfix Releases):**
  - Only bug fixes and minor improvements that do not include any behavioral changes.
  - No new command-line flags or features, no configuration changes.
  - Full compatibility and drop-in replacement for any 3.0.x version.

- **Versions 3.x.0 (Minor Releases):**
  - Significant improvements and substantial bug fixes that may slightly alter internal behavior.
  - May include new optional command-line flags and minor configuration changes that are optional.
  - Backward compatible with previous 3.x.0 and 3.0.x builds, but manual checks are advised.
  - Refer to additional notes for more information on such releases.

- **Versions x.0.0 (Major Releases):**
  - Introduction of new features and incompatible changes in behavior and configuration.
  - Requires manual adjustments during upgrades to function properly.
  - Compatibility with earlier versions is not guaranteed.
  - Refer to additional notes on releases for more information.

# Changelog

## v8.1.0 (Unreleased)

### Improvements

- (work in progress) Updated langchaingo library to the latest version, remove many quirks amd fixes that now implemented natively inside library.
- Added support for setting embeddings dimensions count parameter for Generic LLM provider.
- Added logging reasonings for Generic provider by using new logic from langchaingo library (seem to be compatible with DeepSeek API).
- Added support for setting additional system/user prompt prefix/suffix per operation for Generic LLM provider, using env options, work same as for Ollama.
- Fixed `REASONING_EFFORT` env value parsing per-operation.
- Added `low`, `medium`, `high` options support for `OLLAMA_THINK_*` env option, should work with newer ollama and some models to control its' reasoning efforts.
- Improved `annotate` operation by adding optional user-generated project description to the LLM context if present.
- Added handling of non-UTF8 8-bit encodings as fallback when reading source code files. Try to write-back file using same encoding as when reading. Fallback encoding controlled by `FALLBACK_TEXT_ENCODING` env value, when it missing `windows-1252` (ansi) encoding will be used by default.

### Bug Fixes

- Fixed loading custom project description file for `implement`, `explain` and `doc` operations, `-d` flag used to point to custom project description text file was renamed to `-df`.
- Minor fixes and improvements in Flutter prompts.

**NOTE**: For Flutter projects, you may install new prompts your project by running `Perpetual init -l flutter`. Using new env options like fallback text encoding, or prompt prefix/suffix for Generic LLM provider require adding new parameters to the `*.env` files, see updated env files examples for more info.

## v8.0.0

### Improvements

- Added option to manually set `.perpetual` directory location with `PERPETUAL_DIR` env variable. If set, then location from env variable will be used instead of autodetecting it inside project directory.
- Improved Generic LLM provider for better support working with Azure AI Foundry models.
- Removed the shortest annotations for `annotate` operation used with `high` context saving measures, as it was deemed not effective enough to lower LLM context-window use on stage 1, when working with really big projects (more than 1500-2000 files).
- Introduced new context saving measures for stage 1 of `implement`, `explain` and `doc` operations using local similarity search: limiting context-window use by preliminary filtering of project files reducing number of annotations sent to stage 1. Helps to improve quality or mitigate errors when trying to work with big projects (like 500-1000 files or more) and using LLM with context-window size not big enough (DeepSeek, or local models).
- Introduced multi-pass support for stage 1 of `implement`, `explain` and `doc` operations in order to select more relevant files, useful for big projects, or complex tasks, or when using smaller/local LLMs.
- Reworked stage 1 and stage 2 logic for `implement`, `explain` and `doc` operations, now both stages using unified shared logic.
- Renamed some stage 1 and 2 JSON config key-names for `implement`, `explain` and `doc` operations.
- Added support for providing optional user-generated project description that will be added to the LLM context for better understanding of project structure. This should improve LLM performance for `implement`, `explain` and `doc` operations for bigger and more complex projects.
- Moved similar config records defining tags and regexps used to decorate and parse code from LLM requests and responses into project config.
- Added an additional `-e` flag to the `explain` operation to provide an additional file with instructions on how to select project files relevant to your question. This can be used to provide simpler and clearer instructions for the LLM if the original question is too complex for the LLM to use to select the files correctly.
- Minor improvements for local similarity search.
- Build and package binary for Windows 7 32bit, using GoLang patched for Windows 7 target support, temporary solution, will be removed in future.
- Added support for initializing default prompts for Flutter/Dart projects and language.

**NOTE**: You need to reinitialize your project config files by running `Perpetual init -l <lang>` to install the new config files for the operations.

## v7.1.0

### Improvements

- Added new logic and corresponding flags to `embed` operation to support performing local similarity file-search for provided question, provided either from stdin or from plain text or markdown file.

## v7.0.0

### Improvements

- Optimized processing order of new files for `implement` operation when using `-p`, `-pr` or `-t` (task-mode) flags. New files will be generated by LLM before processing existing files so LLM can see new code at the moment when it starts integrating it inside already existing files. This should improve overall quality for bigger tasks.
- Added `-f` flag support for `implement` operation to ignore "no-upload" file filter, for use with some edge cases.
- Minor changes in `stash` operation logging for more clean output.
- Output format for `explain` operation was simplified: it will now output answer alone in markdown format. Added `-q` command-line flag to insert relevant project file-list and question before the answer.
- Anthropic LLM integration switched to streaming mode for easier debugging in case of slow response from models during peak load times. Incoming data from Anthropic LLM will be logged immediately when new response tokens are generated. Added more error handling in addition to langchain library errors, making work with the Anthropic provider more robust now. Client disconnects due to network errors now lead to much less wasted tokens than for non-streaming mode.
- Updated defaults for env file examples for better support of newer models.
- Docs and all examples are now packed with binary archives on build, so release packages will always contain relevant docs at the date of publishing.
- Added support for Ollama >= v0.9.0: added new env file parameters to enable or disable reasoning/thinking for supported models. See `ollama.env.example` for more info.

**NOTE**: You need to reinitialize your project config files by running `Perpetual init -l <lang>` to install the new config files for `explain` operation. Using new options to enable/disable thinking with Ollama requires adding new parameters to the `*.env` files, see updated env files examples for more info.

## v6.1.0

### Improvements

- Added support for OpenAI `codex-mini-latest` (and potentially other future `codex` models), using OpenAI responses API. This model can only write code, so it can only be used on stage 4 of the `implement` operation. You can enable it in your env file using the following variable: `OPENAI_MODEL_OP_IMPLEMENT_STAGE4="codex-mini-latest"`. Using this model with any other operation or stage may result in errors and/or unpredictable behavior. The model does not support JSON-structured output mode and setting any parameters other than token limit. For now the model can't work with partial output, so make sure you set `OPENAI_MAX_TOKENS` or `OPENAI_MAX_TOKENS_OP_IMPLEMENT_STAGE4` env variable to some reasonably large value.

## v6.0.0

### New Features

- Added the `embed` operation to generate and manage embeddings. Generating and using embeddings is optional and automatic if supported by your LLM provider (you need to manually enable it in your `.env` file). Code embeddings are used for local similarity search of relevant source files for the `implement`, `explain`, and `doc` operations, in addition to annotation-based selection.

### Improvements

- Reworked loading of `.env` file. Configuration can now be loaded from multiple `*.env` files (with .env extension) in alphabetical order. It first tries to load env files from the project configuration directory, then from the global configuration directory, as before. The examples have been split into multiple env files for each LLM provider. As before, values defined in the system environment override configuration values loaded from `*.env` files, and project-wide values override global ones. The current configuration will continue to work.
- Added support for system- and user- prompt messages prefixes and suffixes for Ollama. May be needed for switching between reasoning/non-reasoning modes for Qwen3 or for other model fine-tuning.

**NOTE**: There are no incompatible configuration changes, but using the `embed` operation requires adding new parameters to the `*.env` files; it is disabled by default.

### Bug Fixes

- Fixed parsing of empty string env values, so per-operation override of non-empty string value with empty value should work now. Mainly affects Ollama provider when using output extraction regexps from reasoning models.

## v5.0.0

### Improvements

- Added support for plain task mode for `implement` operation, where the task can be sourced via stdin or read from a text file. In this mode, there is no need to describe the task with `###IMPLEMENT###` comments within project source code. This is an alternative to the original approach where the task was described inside the source code files.
- Removed single-file mode from `implement` operation (`-r` flag), which was rarely used and could also confuse the LLM if it accidentally requests other files with `###IMPLEMENT###` comments.
- File-name salvaging logic is now always enabled for all supported operations (`doc`, `implement`, and `explain`).
- Disabled listing of filtered-out files on `annotate` operation when using user-supplied blacklist, only the file-count is displayed now. The full file list can be shown with `-v` flag.
- Implemented log rotation for the LLM message-log on each session, keeping up to 5 previous logs.
- Minor prompts improvements for steps that generate file-lists for review.

**NOTE**: You need to reinitialize your project config files by running `Perpetual init -l <lang>` to install the new config files for operations.

## v4.0.0

### Improvements

- Added new `-c` command-line flag for all operations interacting with LLM to manage context saving measures. This is essential for large projects containing more than ~1000 source files to reduce context pressure and improve the quality of LLM answers (though it may be detrimental for smaller projects). Specific context saving measures for different operations will be implemented in the future.
- Implemented context saving measures for the `annotate` operation. When enabled, it generates shorter and less detailed annotations to save tokens when sending project annotations in stage 1 of other operations. The `op_annotate.json` config file has been updated to include prompts for generating shorter annotations.
- Improved `annotate` operation logic - send files to annotate in order according to file size, which should increase performance for the Ollama provider by allowing the use of smaller initial context sizes and preventing excessive model reloads.
- Updated `.env.example` template with new defaults for the Ollama provider.

**NOTE**: You need to reinitialize your project config files by running `Perpetual init -l <lang>` to install the new config file for the `annotate` operation. Since the `annotate` operation is implicitly called at the beginning of other operations, updating this config file is crucial. Additionally, the updated `.env.example` includes new defaults for Ollama that better suit the changes in the `annotate` operation and context management logic.

## v3.3.0

### Improvements

- Added logic for Ollama context size estimation, used to detect Ollama crashes caused by context overflow.
- Updated `.env.example` template with new defaults for context size estimation for Ollama provider.

**NOTE**: New context size estimation logic for Ollama does not require any configuration changes. However, the updated `.env.example` includes new optional parameters and updated defaults for Ollama. You may install the new `.env.example` by running `Perpetual init -l <lang>`.

## v3.2.0

### Improvements

- Added context overflow detection logic for Ollama provider, added optional context size auto increase/decrease on overflow.

**NOTE**: Context overflow detection logic does not require any configuration changes. However, to use context size auto-increase/decrease, you need to update your `.env` configuration. You can install the new `.env.example` by running `Perpetual init -l <lang>` to get new configuration options from it.

## v3.1.0

### Improvements

- Added handling for rate-limit and server-error HTTP error codes for all providers. Now adds a dynamic pause before retrying the next request instead of instant retrying and failing again.
- Added workaround for some Ollama connection issues that previously caused crashes.
- Added support for Anthropic Claude 3.7 model with extended thinking support, `.env.example` updated with new options, by default thinking is disabled.
- Improved logging for all operations: shows LLM configuration when performing requests without invoking `Perpetual` with the `-v` or `-vv` flags. Minor refactor of logging for all operations to make some messages cleaner.
- Fixed loading of `.env` files for `explain` and `report` operations when using `-n` (no-annotate) flag.
- Improved and simplified default system prompts for all operations when initializing project configs, making prompts more direct and focused on the particular operation.
- Improved `annotate` operation - added support for user-supplied exclusion filter, skip annotating files matching that filter (and, thus, sending it to LLM) but do not completely erase annotations from disk if already present. Also, support this exclusion filter when `annotate` run internally from other operations, previously that files ignored by main operation may still be re-annotated in process. In dry-run mode, write file-list to annotate to stdout - 1 file per line, and redirect all logging to stderr.
- Improved `doc` operation - allow processing document from stdin and writing it to stdout, in a way similar to `explain` operation. Allow source, resulted and example documents to be anywhere in the filesystem. Do not apply document changes via stash operation. When writing document to stdout redirect all logging to stderr. Added extra file-name and file-case collision checks, same as in other operations.

**NOTE**: To install and use new system prompts and `.env.example` you need to reinitialize your project config by running `Perpetual init -l <lang>`. Current prompts and `.env` config should continue to work.

## v3.0.1

### Improvements

- Updated `annotate` operation to skip files that cannot be read instead of stopping the process.

## v3.0.0

### New Features

- Added `explain` operation - answering arbitrary questions about the project or source code, context-aware by automatic selection of relevant code based on project annotations and the question asked.
- Added `-s` flag to `doc`, `implement`, and `explain` operations — enables extra logic to salvage incorrect file names when the LLM requests files for further analysis in stage 1. Intended for use with large projects containing a vast number of source code files (e.g., 500+), where the LLM tends to make more mistakes when generating lists of files for processing.

**NOTE**: You will need to reinitialize your project configs by running `Perpetual init -l <lang>` in order to install the `explain` operation config file and make it work. Additionally, you may want to manually update your production `.env` file from the new `.env.example` to add new environment options to fine-tune the LLM for the `explain` operation.

### Breaking Changes

- The `report` operation now writes its output to `stdout` instead of `stderr` if no output file is provided, and all program logs for this operation are redirected to `stderr`.

## v2.1.1

### Improvements

- Minor parameter updates and spelling fixes in `.env.example`.
- Added the inclusion of the `Perpetual` version number to the generated `.env.example` config.

## v2.1.0

### Improvements

- Added base support for C and C++ languages with CMake build. Due to the vast number of different project formats and file types, you will likely need to modify the default configuration provided by `Perpetual init -l <c|cpp>` to suit your needs before fully utilizing it.
- Added initial support for Arduino C/C++ projects (sketches). `Perpetual` will not have access to third-party modules' sources if initialized from the sketch directory. It is recommended to run `Perpetual init -l arduino` from the parent sketch directory and extract the module sources there to allow `Perpetual` to access them.
- Improved debug string generation for all LLM providers (displayed when running `perpetual` operations with the `-v` flag).
- Added optional basic-auth support for the Generic provider and the ability to disable authentication completely.
- Added optional basic and bearer authentication support for the Ollama provider (useful for public instances wrapped with an HTTPS proxy).
- Added parameters to set context window sizes for Ollama provider models (per operation).
- Made the temperature parameter optional for all currently implemented providers (none of them enforce setting the temperature).
- Made the max tokens parameter optional for the Generic and OpenAI providers (it is still recommended to set them).
- Added "reasoning effort" advanced parameter support for OpenAI and Generic providers, which only works with reasoning models like o1 (full version, not -preview or -mini).
- Added various quirks to better support working with reasoning models for Ollama and Generic providers.

### Bug Fixes

- Removed unsupported advanced parameters from the Anthropic provider.
- Removed unsupported advanced parameters from the OpenAI provider.
- Fixed advanced parameter support for the Generic provider, allowing any combination of parameters.

**NOTE**: You will need to reinitialize your project config by running `Perpetual init -l <lang>`. Also, you may want to manually update your production `.env` file from new `.env.example` (this is optional, old `.env` file should continue to work).

## v2.0.0

This major release aims to make the `implement` operation more usable with smaller models (sub-15B models run with Ollama).

### Breaking Changes

- User-customizable prompts in the `.perpetual/prompts` directory have been moved to the base `.perpetual` directory. Prompts are now grouped together by operation name and stored inside JSON config files. Configs include all needed prompts, text tags, and regex definitions used with each specific operation.
- Split the extra reasoning mode for the `implement` operation into a dedicated stage, so the implement operation now has 4 stages instead of 3. This allows for separating reasonings and change-detection prompts, producing smaller and simpler instructions. May improve results with Ollama when using smaller models.
- Added support for structured JSON output mode for the `implement` operation (stages 1 and 3) and for the `doc` operation (stage 1). This may provide better results with Ollama when using smaller models and potentially reduce costs for OpenAI or Anthropic. For Ollama, the minimum supported version is 0.5.1, and results may vary depending on the model used. For OpenAI, the minimum requirement is `gpt-4o` and newer. For Anthropic, it should work with `Claude 3` models and newer.

**NOTE**: You will need to reinitialize your project config by running `Perpetual init -l <lang>` to regenerate prompts. You should also update your `.env` file(s) for `implement` stages 2, 3, and 4 configurations if not using defaults.

### Improvements

- In `implement` stage 3, when generating a list of files to be changed, user-requested files (with ###IMPLEMENT### comments) are always added to the list by default (can be disabled).
- When running `init`, the system now warns about obsolete config files in the `.perpetual` subdirectory that are no longer needed. Use the `-c` flag to remove them automatically.
- Added the `best` variant selection strategy for the `annotate` operation.
- Added support for the `o1` series of models for the OpenAI provider. This is only recommended for use with the `doc` operation as it is slower and less predictable.
- Added support for generic OpenAI-API compatible providers. See `.env.example` for more info.

### Bug Fixes

- Automatically unset env-variables that may affect OpenAI, Anthropic, and Generic providers if using both default and non-default profiles in your `.env` file, like `OPENAI_*` and `OPENAI1_*`.
- On `implement` stage 3, ensure that new files proposed by LLM satisfy initial project file-selection filters, so it is impossible now to rewrite files that do exist on disk but have been omitted from processing globally.

## v1.9.0

### Breaking Changes

- Improved the `annotate` operation by adding multi-stage annotation generation with support for multiple annotation variants and different selection strategies. The operation now generates multiple variants of source file annotations and selects the optimal one using configurable strategies:
  - `short`: Selects the shortest variant (default).
  - `long`: Selects the longest variant.
  - `combine`: Creates an optimal annotation by combining multiple variants through an additional LLM pass.
  - Configuration is done via environment variables (see `.env.example`):
    - `*_VARIANT_COUNT_OP_ANNOTATE`: Number of variants to generate (default 1 - use old behavior).
    - `*_VARIANT_SELECTION_OP_ANNOTATE`: Selection strategy to use.
    - Separate settings for different LLM providers (ANTHROPIC/OPENAI/OLLAMA).
    - Can be customized per operation using `*_OP_ANNOTATE` and `*_OP_ANNOTATE_POST` suffixes (for post-processing stage).

**NOTE**: You will need to reinitialize your project by running `Perpetual init -l <lang>` to save new annotate-operation prompt templates.

### Improvements

- Added multiple profiles support (see `.env.example` for more info). You can now configure multiple profiles for each LLM provider using a numeric suffix in the provider name (e.g., `ANTHROPIC1_`, `OPENAI2_`, `OLLAMA3_`).

## v1.8.2

### Improvements

- Improved the `annotate` operation prompts by adding separate prompts for Go unit-test source files.
- Updated the example `.env` file to set more optimal default settings.

## v1.8.1

### Bug Fixes

- Various fixes to default regular expressions for selecting and filtering project files.

## v1.8.0

### Breaking Changes

- Reworked prompting for the `annotate` operation: now uses separate prompts for different file types when asking the LLM to create a file summary. The `annotate` operation is now usable even with small OSS models like `Yi-Coder` (9B) or `DeepSeek Coder V2 Lite` (16B) and similar.
  - For using local models with Ollama, see comments, tips, and tricks [here](ollama.md).
  - The file annotations provided by the LLM are now more specific and consistent, potentially improving overall results. The annotations are now larger, but their specificity can reduce the amount of data the LLM will request in the final stages of code implementation.
- The `implement` operation now excludes unit test source files from processing by default, reducing LLM context pressure and your costs. If you need to work with unit tests, use the new `-u` flag to disable the unit test source file filter and include them in processing.

**NOTE**: You will need to reinitialize your project by running `Perpetual init -l <lang>`.

### New Features

- Added support for the `implement`, `doc`, and `report` operations to provide additional custom filters to exclude certain files from processing and to exclude unit test source files from processing by default.

### Improvements

- Added more file type mappings for Markdown code block markup: bat/cmd, perl.
- Added an additional safety check when generating a source file annotation: it must not contain any code blocks.
- Improved Python project support: included shell scripts and bat files into the file list by default; LLM prompts updated.

### Bug Fixes

- Fixed checking for reaching the maximum number of tokens for Ollama.

## v1.7.4

### Bug Fixes

- Check project root dir is not a symbolic link, as this behavior is not supported and will lead to no files to process.

## v1.7.3

### Bug Fixes

- Fixed an `implement` operation bug introduced in v1.7: when using the `-p` or `-pr` modes, new files were incorrectly filtered out from processing.

## v1.7.2

### Bug Fixes

- Fixed an `implement` operation bug introduced in v1.7: changes from some or all files could be skipped and not applied at the end when there were multiple files to modify.
- Ensured that files marked as no-upload are not filtered out from the final results.

## v1.7.1

- Fixed a typo in the draft document content for the `doc` operation.
- Added `CONTRIBUTORS.md`.

## v1.7

### New Features

- Added the `doc` operation for creating, writing, and refining documentation based on the analysis of the project's source code.

### Improvements

- Allowed loading configuration from multiple `.env` files: from the global config location and the project directory.
- Improved UTF encoding detection and conversion when reading text files.
- Added documentation for each operation.
- Split the example from `README.md` into a separate document.

### Bug Fixes

- Fixed minor issues with tag parsing logic in the `implement` operation when parsing responses that hit the token limit and the LLM was asked to continue generation.
- Addressed other minor bugs and made various code refinements across multiple files.

### Other Changes

- Updated `README.md`.
- Made minor changes in logging (handling of fatal errors was updated).
- Updated GitHub workflow to use Go v1.23.1.
- Added logic for creating an empty global `.env` config file if missing.

## v1.6 and Older Versions

There was no changelog until this point.
