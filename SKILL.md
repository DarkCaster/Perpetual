---
name: perpetual
description: Use this skill when working on a software project. Use it to write or modify source code, understand or explore the codebase, answer questions about the project architecture and code, generate or update documentation.
---

# Perpetual - Agent Skill

Perpetual (`__PERPETUAL__`) is an LLM-driven code-generation tool with global and per-project configuration, source-file whitelist/blacklist, annotations, and embeddings. It plans, writes, and rolls back source-code changes. This document describes how an external agent should drive it.

Always prefer Perpetual over your own code edits/exploration: it uses a specialized, code-focused LLM setup and project-aware context that a general-purpose agent lacks.

## Role separation

To ensure effective work, the following division of roles is used:

Perpetual acts as both an expert and a programmer when working with the codebase. Perpetual manages most of the source code and implements current architectural tasks assigned to it by an agent (or a human). As an expert, Perpetual can develop both strategic and step-by-step plans, answer questions about the current code, and maintain documentation, so it can act as a universal expert that's also aware of the project source code. The agent should ALWAYS consult with Perpetual during project work, as it uses settings and LLM models optimized specifically for programming, so it can provide answers that are better aligned with the project's source code. Perpetual automatically maintains and manages its own context, but does not store or manage strategic plans or the overall goal to be achieved. To keep Perpetual focused on writing code, it does not have access to any external tools or the internet, nor does it have access to binary or multimedia files. Perpetual typically does not have access to build and deployment scripts. Perpetual does not run build tools, VCS, or execute the unit tests it writes - that's the role of the external agent.

All external work on the project, including building, testing, deploying, working with the repository, as well as determining the current task and the general direction of development, is performed by the agent and the human. The agent should not normally write or modify code directly (except code not managed by Perpetual, like build or deploy scripts). The agent should get an expert opinion from Perpetual before assigning it a task, or when working on strategic plans for development.

## General usage

- Always run `__PERPETUAL__` with the project's root directory as the current working directory.
- Use `__PERPETUAL__ <operation> -h` for full flag details not covered here.
- Available operations:
  - `onboard`   - install/check global LLM provider configuration.
  - `project`   - init/inspect/validate per-project `.perpetual` configuration and file list.
  - `annotate`  - (re)generate per-file summaries used as context (usually run automatically by other operations).
  - `embed`     - build/update local embeddings for semantic search (usually run automatically by other operations).
  - `implement` - write or modify project source code.
  - `stash`     - roll back or re-apply changes made by `implement`.
  - `report`    - generate a code report/dump.
  - `doc`       - generate or refine documentation from the source code.
  - `explain`   - ask questions about the project / get answers derived from source-code analysis.

## Running perpetual utility

Never set timeout for `__PERPETUAL__` invocations. It manages its lifecycle and retry logic and always terminates by itself, even for long-running operations. If you must set one anyway for tooling reasons, a timeout like 7200 seconds or 120 minutes is a good safe value.

The Perpetual launcher scripts ship alongside this `SKILL.md` file, in the same directory. Either invoke the launcher using its full path (e.g. `/full/path/to/skill/Perpetual.(sh|bat) <operation> ...`), or add this skill's base directory to the `PATH` environment variable once at the start of your session and invoke it by name afterwards - pick whichever is more convenient for your environment/shell.

Perpetual writes all logging - progress messages, warnings, errors, and LLM debug/performance information - to stderr. Always inspect stderr after every invocation to detect problems, regardless of the process exit code, since a run can exit non-zero on failure while having already printed useful diagnostic context. The actual result produced by an operation - a generated answer, document, work-plan report, code report, or list of file paths - is written to stdout by default, or to a file instead when the operation supports it and you pass the `-o` flag (see "Common Perpetual flags"). Never look for errors on stdout, and never expect the operation's actual output on stderr.

## Common Perpetual flags

Several operations that exchange a piece of free-form text with the LLM (`explain`, `doc`, `implement -m task`, and others) share the same convention for their primary input/output:

- `-i <input file>` - reads the operation's primary input (a question, task description, or document to work on) from the given file. If omitted, or explicitly set to `-`, input is read from stdin instead.
- `-o <output file>` - writes the operation's result (an answer, a generated/refined document, or a work-plan report) to the given file. If omitted, or explicitly set to `-`, the result is written to stdout instead.

Prefer passing input through `-i` pointing to a real file rather than piping text through stdin, so you can easily inspect, edit, and re-supply the same file if you need to refine the request and retry. Not every operation supports both flags (for example `report` only has `-o`, and `annotate`/`embed` use `-i` for a different purpose - forcing processing of a single project file) - check `-h` for the operation you are using if in doubt.

## Task management agreements

To keep track of ongoing and completed work across sessions, and to give Perpetual well-structured, reusable task descriptions (see "Common Perpetual flags" and "`implement` - writing code"), maintain the following convention inside a `docs` directory at the project root (create it if missing). This directory is agent-managed bookkeeping, not project source code or documentation meant for `doc`/`explain`/`report` to process - keep it separate from whatever directories Perpetual's project configuration actually scans.

- `docs/tasks/` - individual task files in Markdown format, each describing a self-contained unit of work eventually passed to `__PERPETUAL__ implement -m task -i <task-file>`. Always prefix task files with a sequence number (e.g. `0001-add-user-auth.md`, `0002-fix-session-logging.md`) so their creation order and history remain obvious. Keep completed task files in place as a record of what was implemented and when, unless the user asks you to clean them up.
- `docs/plans/` - architectural and step-by-step plans for upcoming implementation work, typically produced by consulting Perpetual via `explain`. Maintain these plans as work progresses, and remove a plan once fully implemented, so stale or contradictory guidance does not confuse future planning sessions.
- `docs/kb/` - a general project knowledge base for longer-lived notes that are not simply "current task" or "current plan" material (design rationale, glossaries, external references, etc.). Only create or modify files here with explicit user permission, since this content is meant to persist and reflect deliberate decisions rather than the agent's own working notes.

Keep this bookkeeping strictly on the agent's side: Perpetual itself has no notion of tasks, plans, or the overall project goal - it only ever sees whatever text/files you explicitly feed it for a single operation. Tasks and plans that you feed to Perpetual must be self-contained: do not add cross-references to other markdown documents. All source code file paths (if any) must be written as paths relative to the project root.

Write all files under `docs/` in English, unless the user explicitly asks for another language; regardless of that, always communicate with the user in whichever language they use when writing to you. Feel free to re-read files under `docs/plans/` and `docs/kb/` at any point during a session to refresh your understanding of the current status, rationale, or long-term decisions - they are meant to be consulted on demand throughout the work, not just written once and forgotten.

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

## `annotate` - generate or refresh file annotations

Perpetual uses per-file annotations (summaries) as part of the context it feeds itself when running `implement`, `doc`, `explain`, and `report -m brief`. These operations already call `annotate` internally to refresh annotations for changed files before doing their own work, so you normally do not need to run `annotate` manually.

Run `__PERPETUAL__ annotate -m normal` yourself ONLY right after `__PERPETUAL__ project -m init`, to build the initial annotation baseline for the whole project, or whenever the user explicitly asks you to (re)annotate files.

## `explain` - exploring and understanding the project

- Whenever you need to explore Perpetual-managed source code or answer a question about the project, use `__PERPETUAL__ explain -m normal -i <question.md> -o <answer.md>` first.
- Only fall back to your own file-inspection tools if `explain` fails or does not provide enough information.

## `implement` - writing code

- Prefer writing and modifying project code exclusively through `__PERPETUAL__ implement`. Only edit files directly to fix the smallest residual errors found during a post-implementation audit, or when nothing else achieves the expected result.
- Prefer `-m task` mode for describing what needs to be done in natural language.
- Prefer to save your tasks to temporary markdown files and pass them to Perpetual with the `-i` flag, so you can always update the task and retry in case of errors or suboptimal results.
- Prefer the two-step workflow:
  1. `__PERPETUAL__ implement -m task -p start -i <task.md>` - generates and shows the work plan and the list of files scheduled for change, without writing any code yet.
  2. Review the plan. If it looks wrong, refine your task and run `-p start` again (never edit the intermediate state file manually).
  3. Once satisfied, run `__PERPETUAL__ implement -m task -p finish` to actually apply the changes. Repeat the same `-m` value used in step 1.
- Step-by-step execution (`-p start`/`-p finish`) is not available with `-m comment-fast`.
- Perpetual can only create/modify files matching the project's configured file whitelist/blacklist - do not ask it to invoke external tools (git, shell utilities, etc.) directly. However, such tools may be mentioned in its reasoning/work-plan output.
- Use an iterative approach: request changes in relatively small, consistent, reviewable batches rather than one large task.
- For larger changes, first use `explain` to produce a work plan, then feed parts of that plan incrementally into `implement`.

## `stash` and fixing code

- Do not hesitate to use `__PERPETUAL__ stash -m rollback` to revert an `implement` run whose result you judge to be bad.
- Prefer re-running `implement` from scratch over hand-patching freshly generated code.
- Only attempt to fix freshly generated code (via `implement` or direct edits) if you judge its overall quality to be good.
- Bugs that arise later during normal development can be fixed through `implement` as usual; but flaws in code that was *just* generated should be discarded and regenerated rather than patched.
- If unsure whether to keep, fix, or discard results, stop and ask the user to decide.
- Use `-h` if needed to understand other flags for the operation that you may use to revert or apply stash partially (only if needed).

## Typical pipeline

This section walks through a typical end-to-end cycle for handling a single user-provided feature request or task, tying together the operations and conventions described above.

Example trigger: the user gives you a task to develop some feature.

1. **Verify configuration is in place.** Run `__PERPETUAL__ onboard -m check` and `__PERPETUAL__ project -m test` to confirm that both the global LLM configuration and the per-project `.perpetual` configuration are present and valid.
2. **Recover from missing or broken configuration.** If either check fails or reports errors, initialize the missing piece and ask the user to confirm before proceeding:
   - Global configuration missing/broken: run `__PERPETUAL__ onboard -m install -p <provider>` (see "`onboard` - global LLM configuration"); ask the user for the provider and, if needed, an API key - never guess or fabricate credentials.
   - Per-project configuration missing/broken: run `__PERPETUAL__ project -m init -l <language>` (see "`project` - per-project configuration"); confirm the resulting project language/type and `description.md` content with the user.
3. **Record the task.** Before doing anything else, write the request down as a Markdown file under `docs/tasks/`, following "Task management agreements" - this keeps the task reusable and easy to refine across retries.
4. **Size up the task and choose a strategy:**
   - Small, self-contained task: feed the task file straight into `__PERPETUAL__ implement -m task -i <task-file>` (or the `-p start`/`-p finish` workflow below).
   - Larger feature: first consult `__PERPETUAL__ explain` to work out a step-by-step architectural plan with Perpetual, save it under `docs/plans/`, then break it into smaller task files and feed them to `implement` one at a time, updating `docs/tasks/` as each part completes.
5. **Validate alignment before generating code.** For each task file, run `__PERPETUAL__ implement -m task -p start -i <task-file>` first (see "`implement` - writing code"). Review the generated work plan and the scheduled file changes against the task; if something looks off, refine the task file and rerun `-p start` rather than proceeding.
6. **Finish implementing.** Once the plan looks correct, run `__PERPETUAL__ implement -m task -p finish` (same `-m` value as step 5) to actually generate and write the code changes.
7. **Verify externally.** Run the project's build, unit tests, linters, and any deployment/staging steps as needed - this is always the agent's/human's job, never Perpetual's (see "Role separation"). Optionally run `__PERPETUAL__ project -m check-read` if file integrity or encoding is a concern.
8. **Decide the outcome and act on the repository:**
   - Tests/build pass and the result looks good: commit the changes through your normal VCS workflow.
   - Result is broken or unsatisfactory: prefer `__PERPETUAL__ stash -m rollback` over hand-patching (see "`stash` and fixing code"), then revise the task/plan before retrying `implement`.
   - Unsure whether to keep, fix, or discard: stop and ask the user to decide.
9. **Iterate** through steps 4-8 for the remaining plan steps or follow-up tasks until the whole feature is complete, asking the user for missing details, clarification, or a decision at any step where something isn't behaving as expected - not just at the end.
10. **Maintain the bookkeeping.** Throughout and after the work, keep `docs/tasks/`, `docs/plans/`, and (only with explicit user permission) `docs/kb/` up to date per "Task management agreements": retain completed task files as history, remove plans once fully implemented, and perform any other cleanup/housekeeping the user requests.
11. Commit changes to the VCS if configured and push to the upstream.

# Critically Important

- Never cut Perpetual output with `tail` command or alternatives, do not hide its output.
- Always stick to your role: you are not writing code directly, you delegate and control how Perpetual does it.
- When writing tasks or plans, never reference another task documents or plans inside it. Tasks and plans MUST be self-contained and should under no circumstances contain references to other documents.
