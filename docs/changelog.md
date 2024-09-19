# Changelog

## v1.7

### New Features

- Added new operation `op_doc` for creating, writing, and refining documentation based on project source code analysis

### Improvements

- Allow for loading configuration from multiple `.env` files: from global config location and project directory
- Improved UTF encoding detection and conversion when reading text files
- Added documentation for each operation
- Split example from README.md into a separate document

### Bug Fixes

- Fixed minor issues with tag parsing logic in `op_implement` when parsing responses that hit token limit and LLM was asked to continue generation
- Addressed other minor bugs and made various code refinements across multiple files

### Other Changes

- README.md was updated
- Minor changes in logging (handling of fatal errors was updated)
- Updated GitHub workflow to use Go v1.23.1
- Added logic for creating an empty global .env config file if missing

## v1.6 and older versions

There was no changelog until this point.
