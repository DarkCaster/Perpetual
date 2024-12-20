# Changelog

## v2.0.0

### Breaking changes

- Added initial support for json structured output mode. Disabled by default for now, currently only works with Ollama and OpenaAI. May improve results for supported operations/stages in future.

- Split extra reasoning mode for `op_implement` into dedicated stage, so implement operation now have 4 stages instead of 3. This allow not to mix reasonings and changes-detection instructions together within single query producing smaller and simplier instructions, which should provide better results with smaller models.

- User-customizable prompts at `.perpetual/prompts` dir moved to base `.perpetual` dir. Prompts now grouped together by opeartion-type and stored inside json config-files. Configs also include all needed text-tags and regexp definitions within corresponding json file.

**NOTE**: you will need to reinitialize your project by running `Perpetual init -l <lang>` to regenerate prompts. You should also update your `.env` file(s) for `op_implement` stages 2,3,4 config, if not using defaults.

## v1.9.0

### Breaking changes

- Improve `annotate` operation: Added multi-stage annotation generation with support for multiple annotation variants and different selection strategies. The operation now generates multiple variants of source file annotations and selects the optimal one using configurable strategies:
  - `short`: Selects the shortest variant (default)
  - `long`: Selects the longest variant
  - `combine`: Creates an optimal annotation by combining multiple variants through an additional LLM pass
  - Configuration is done via environment variables (see `.env.example`):
    - `*_VARIANT_COUNT_OP_ANNOTATE`: Number of variants to generate (default 1 - use old behavior)
    - `*_VARIANT_SELECTION_OP_ANNOTATE`: Selection strategy to use
    - Separate settings for different LLM providers (ANTHROPIC/OPENAI/OLLAMA)
    - Can be customized per operation using `*_OP_ANNOTATE` and `*_OP_ANNOTATE_POST` suffix (for post-processing stage)

**NOTE**: you will need to reinitialize your project by running `Perpetual init -l <lang>` to save new annotate-operation prompt-templates

### Improvements

- Added multiple profiles support (see `.env.example` for more info). You can now configure multiple profiles for each LLM provider using a numeric suffix in the provider name (e.g., `ANTHROPIC1_`, `OPENAI2_`, `OLLAMA3_`).

## v1.8.2

### Improvements

- Improve `annotate` operation prompts. Add separate prompts for golang unit-tests source files.

- Update example `.env` file, set more optimal default settings to the current time.

## v1.8.1

### Bug Fixes

- Various fixes to default regular expressions for selecting and filtering project files.

## v1.8.0

### Breaking changes

- Rework prompting for `annotate` operation: use separate prompts for different file-types when asking LLM to create a file summary. Now `annotate` operation is usable even with the small OSS models like `Yi-Coder` (9B) or `DeepSeek Coder V2 Lite` (16B) and similar.
  - For using local models with ollama [see comments, tips and tricks here](ollama.md)
  - The file annotations provided by LLM are now more specific and consistent, potentially improving overall results. The annotations are now larger, but because they are more specific, this can reduce the amount of data LLM will request in the final stages of code implementation.

- The `implement` operation now excludes unit test source files from processing by default, reducing LLM context pressure and your costs. If you need to work with unit tests, use the new `-u` flag to disable the unit test source file filter and include them in processing.

**NOTE**: you will need to reinitialize your project by running `Perpetual init -l <lang>`

### New Features

- Added support to the `implement`, `doc`, `report` operations to provide additional custom filters to exclude certain files from processing, exclude unit test source files from processing by default.

### Improvements

- Added more file type mappings for Markdown code block markup: bat/cmd, perl
- Added an additional safety check when generating a source file annotation: it must not contain any code blocks.
- Improve Python projects support: include shell scripts and bat files into the file-list by default, LLM prompts updated

### Bug Fixes

- Fixed checking for reaching the maximum number of tokens for Ollama

## v1.7.4

### Bug Fixes

- Check project root dir is not a symbolic link, as this behavior is not supported and will lead to no files to process

## v1.7.3

### Bug Fixes

- Fixed `implement` operation bug introduced with 1.7: when using the `-p` or `-pr` modes, new files are filtered-out from processing

## v1.7.2

### Bug Fixes

- Fixed `implement` operation bug introduced with 1.7: changes from some or all files can be skipped and not applied at the end when there are multiple files to modify
- Do not filter-out files marked as no-upload from the final results

## v1.7.1

- Fixed a typo in the draft document content for the `doc` operation
- Added CONTRIBUTORS.md

## v1.7

### New Features

- Added `doc` operation for creating, writing and refining documentation based on analysis of the project's source code.

### Improvements

- Allow for loading configuration from multiple `.env` files: from global config location and project directory
- Improved UTF encoding detection and conversion when reading text files
- Added documentation for each operation
- Split example from README.md into a separate document

### Bug Fixes

- Fixed minor issues with tag parsing logic in `implement` operation when parsing responses that hit token limit and LLM was asked to continue generation
- Addressed other minor bugs and made various code refinements across multiple files

### Other Changes

- README.md was updated
- Minor changes in logging (handling of fatal errors was updated)
- Updated GitHub workflow to use Go v1.23.1
- Added logic for creating an empty global .env config file if missing

## v1.6 and older versions

There was no changelog until this point.
