# Changelog

## v2.0.0

This major release aims to make the `implement` operation more usable with smaller models (sub-15B models run with Ollama).

### Breaking Changes

- User-customizable prompts in the `.perpetual/prompts` directory have been moved to the base `.perpetual` directory. Prompts are now grouped together by operation name and stored inside JSON config files. Configs include all needed prompts, text tags, and regex definitions used with each specific operation.

- Split the extra reasoning mode for the `implement` operation into a dedicated stage, so the implement operation now has 4 stages instead of 3. This allows for separating reasonings and change-detection prompts, producing smaller and simpler instructions. May improve results with Ollama when using smaller models.

- Added support for structured JSON output mode for the `implement` operation (stages 1 and 3) and for the `doc` operation (stage 1). This may provide better results with Ollama when using smaller models and potentially reduce costs for OpenAI or Anthropic. For Ollama, the minimum supported version is 0.5.1, and results may vary depending on the model used. For OpenAI, the minimum requirement is `gpt-4o` and newer. For Anthropic, it should work with `Claude 3` models and newer.

**NOTE**: You will need to reinitialize your project by running `Perpetual init -l <lang>` to regenerate prompts. You should also update your `.env` file(s) for `implement` stages 2, 3, and 4 configurations if not using defaults.

### Improvements

- In `implement` stage 3, when generating a list of files to be changed, user-requested files (with ###IMPLEMENT### comments) are always added to the list by default (can be disabled).

- When running `init`, the system now warns about obsolete config files in the `.perpetual` subdirectory that are no longer needed. Use the `-c` flag to remove them automatically.

- Added the `best` variant selection strategy for the `annotate` operation.

- Added support for the `o1` series of models for the OpenAI provider. This is only recommended for use with the `doc` operation as it is slower and less predictable.

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

- Check project root dir is not a symbolic link, as this behavior is not supported and will lead to no files to process

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
