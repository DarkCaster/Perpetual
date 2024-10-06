# Changelog

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

- Fixed a typo in the draft document content for the `doc` operation.
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
