# Perpetual - Agent Skill

Perpetual (`__PERPETUAL__`) is a code-generation tool driven by an LLM. It manages its own project configuration, source-file whitelist/blacklist, annotations and embeddings, and can plan, write, and roll back changes to a project's source code. This document describes how an external coding agent should drive Perpetual as a tool.

## General usage

- Always run `__PERPETUAL__` with the project's root directory as the current working directory.
- Detailed flags for any operation can be obtained with `__PERPETUAL__ <operation> -h`. This document intentionally does not enumerate every flag - use `-h` when you need specifics not covered here.
- Available operations:
  - `onboard`   - install/check global LLM provider configuration.
  - `project`   - init/inspect/validate per-project `.perpetual` configuration and file list.
  - `annotate`  - (re)generate per-file summaries used as context (usually run automatically by other operations).
  - `embed`     - build/update local embeddings for semantic search (usually run automatically by other operations).
  - `implement` - write or modify project source code.
  - `stash`     - roll back or re-apply changes made by `implement`.
  - `report`    - generate a code report/dump.
  - `doc`       - generate documentation from the source code.
  - `explain`   - ask questions about the project / get answers derived from source-code analysis.

## `onboard` - global LLM configuration

- Run `__PERPETUAL__ onboard -m check` to print the currently active environment configuration and validate it (provider selection, API keys/auth, models, etc.).
- Run `__PERPETUAL__ onboard -m install -p <anthropic|openai|ollama|generic>` to write a fresh default global configuration for the chosen provider. Optional `-k` supplies an API key / `login:password`, `-km` sets the auth method (`Bearer`|`Basic`).
- If `onboard -m check` reports missing variables or configuration errors, ask the user to manually edit the global env config files (you may assist them). Use the bundled `*.env.example` files (e.g. `anthropic.env.example`, `openai.env.example`, `ollama.env.example`, `generic.env.example`, `.env.example`) shipped alongside the Perpetual distribution as a reference for available settings - they document every supported option with comments.
- Never guess or fabricate API keys/secrets; only the user should provide them.

## `project` - per-project configuration

- Run `__PERPETUAL__ project -m init -l <language>` once per project to create the `.perpetual` directory with default prompts/config for the given language/project type.
- You may run `__PERPETUAL__ project -m <test|list|check-read|check-ascii>` at any time to query the current state: `test` validates configuration, `list` shows files Perpetual can see, `check-read`/`check-ascii` verify file readability/encoding.
- Before committing Perpetual-generated changes to VCS, run `__PERPETUAL__ project -m check-read` to verify consistency of the files Perpetual wrote; use `__PERPETUAL__ project -m save-utf` to fix files with encoding issues if needed.
- As with `onboard`, consult the bundled `*.env.example` files for available global settings when helping the user configure the project's environment.

## Per-project `description.md`

- `.perpetual/description.md` gives Perpetual extra context about the project. Keep it concise and focused on the most important architectural concepts and decisions (purpose, key technologies, architecture, coding standards, project structure).
- Because Perpetual only works with text/source files, it cannot see binary, resource, or media files directly. Use `description.md` to also list such files together with their purpose and relevant details, so that code Perpetual generates can correctly reference them. List them with paths relative to the project root.

## `explain` - exploring and understanding the project

- Whenever you need to explore Perpetual-managed source code or answer a question about the project, use `__PERPETUAL__ explain -m normal` (or `-m full`) first.
- Only fall back to your own file-inspection tools if `explain` fails or does not provide enough information.

## `implement` - writing code

- Prefer writing and modifying project code exclusively through `__PERPETUAL__ implement`. Only edit files directly to fix the smallest residual errors found during a post-implementation audit, or when nothing else achieves the expected result.
- Prefer `-m task` mode for describing what needs to be done in natural language.
- Prefer to save your tasks to the temporary markdown files and source them to Perpetual with `-i` flag. So you can always update the task and retry in case of errors or suboptimal results.
- Prefer the two-step workflow:
  1. `__PERPETUAL__ implement -m task -p start -i <path to the task file>` - generates and shows the work plan and the list of files scheduled for change, without writing any code yet.
  2. Review the plan. If it looks wrong, refine your task and run `-p start` again (never edit the intermediate state file manually).
  3. Once satisfied, run `__PERPETUAL__ implement -p finish` to actually apply the changes.
- Perpetual can only create/modify files matching the project's configured file whitelist/blacklist - do not ask it to invoke external tools (git, shell utilities, etc.) directly; However such tools may be mentioned in its reasoning/work-plan output.
- Use an iterative approach: request changes in relatively small, consistent, reviewable batches rather than one large task.
- For larger changes, first use `explain` to produce a work plan, then feed parts of that plan incrementally into `implement`.

## `stash` and fixing code

- Do not hesitate to use `__PERPETUAL__ stash -m rollback` to revert an `implement` run whose result you judge to be bad.
- Prefer re-running `implement` from scratch over hand-patching freshly generated code.
- Only attempt to fix freshly generated code (via `implement` or direct edits) if you judge its overall quality to be good.
- Bugs that arise later during normal development can be fixed through `implement` as usual; but flaws in code that was *just* generated should be discarded and regenerated rather than patched.
- If unsure whether to keep, fix, or discard results, stop and ask the user to decide.
