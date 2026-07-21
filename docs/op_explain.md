# Explain Operation

The `explain` operation is designed to provide insightful answers to questions and clarifications about your project based on analysis of its source code. This operation examines your project's structure, annotations, selected source files, and optional project description to generate comprehensive explanations, aiding developers in understanding complex codebases, identifying potential issues, planning changes, and more.

## Usage

To utilize the `explain` operation, execute the following command:

```sh
Perpetual explain [flags]
```

The `explain` operation offers a range of command-line flags to tailor its functionality to your specific needs:

- `-m <mode>`: Select the operation mode to perform (valid values: `normal`, `list`, `full`). This flag is required.
  - `normal`: Generate the final answer to the question.
  - `list`: Only list files that the LLM thinks are related to the question; do not generate the final answer. The output is one filename per line, with no formatting. In this mode, the operation stops before stage 2 and does not upload selected source files for final answer generation.
  - `full`: Include the question text and the list of relevant files in the generated answer. Files filtered out by the `no-upload` rule are marked using the configured filtered filename tags.
- `-i <file>`: Specify the input file containing the question to be answered (plain text or Markdown). If empty or `-`, the operation reads the question from standard input (stdin).
- `-o <file>`: Define the output file for writing the answer. If empty or `-`, the answer is output to standard output (stdout).
- `-df <file>`: Optional path to project description file for adding into LLM context. Valid values are `file-path` to specify a custom description file, or `disabled` to explicitly disable loading the project description. If omitted, the operation attempts to load the default `description.md` file from the `.perpetual` directory.
- `-e <file>`: Read instructions from a text or markdown file that will be used in stage 1 to select relevant files. Use this flag if the original question is too complex or not clear enough for the LLM to select relevant files, allowing you to provide separate instructions for the file selection process. The final answer is still generated for the original question from `-i` or stdin.
- `-c <mode>`: Set the context saving mode to reduce LLM context usage for large projects. Valid values are `auto`, `off`, `medium`, or `high` (default: `auto`).
- `-a`: Add project annotations in stage 2 in addition to the source files requested by the LLM. This can improve answer quality by providing the LLM with additional project-wide context, but it is disabled by default to save tokens and reduce context window requirements.
- `-n`: Enable "No Annotate" mode, which skips the automatic refresh of annotations and embeddings before processing. Existing annotations and embeddings are used if available. This can reduce API calls but may lower the quality of file selection and explanations.
- `-f`: Override the `no-upload` file filter to include files marked as `no-upload` for review. Use this flag with caution, as it may upload sensitive files to the LLM during the explanation process.
- `-u`: Include unit-test source files in the processing. By default, unit-test files are excluded using the project test-file blacklist.
- `-x <file>`: Provide a path to a user-supplied regex filter file. This file allows for the exclusion of specific files or patterns from processing based on custom criteria. See more info about using the filter [here](user_filter.md).
- `-s <n>`: Limit the number of additional files related to the question returned by local similarity search. Valid values are integer ≥ 0 (`0` disables local search; only use LLM-requested files). Default: `5`.
- `-sp <n>`: Set number of passes for related files selection at stage 1 (default: 1). Higher pass-count values may select more files, compensating for possible LLM errors when finding relevant files, but will cost more tokens and context use.
- `-h`: Display the help message, detailing all available flags and their descriptions.
- `-v`: Enable debug logging to receive detailed output about the operation's execution process.
- `-vv`: Activate both debug and trace logging for the highest level of verbosity, offering an in-depth view of the operation's internal workings.

### Examples

1. **Ask a question and receive an explanation:**

   ```sh
   echo "How does the authentication system work?" | Perpetual explain -m normal -o explanations/auth_system.md
   ```

   This command reads the question from standard input and writes the explanation to `explanations/auth_system.md`.

2. **List files relevant to a specific question without generating an explanation:**

   ```sh
   echo "What modules handle data processing?" | Perpetual explain -m list
   ```

   Instead of an answer, this command produces a list of files related to data processing.

3. **Generate an explanation with additional annotations and debug logging enabled:**

   ```sh
   Perpetual explain -m normal -i questions/query.txt -o explanations/data_flow.md -a -v
   ```

4. **Include unit-test files in the explanation process:**

   ```sh
   Perpetual explain -m normal -i questions/query.txt -o explanations/test_data_flow.md -u
   ```

5. **Override the `no-upload` filter to include all selected files for explanation:**

   ```sh
   Perpetual explain -m normal -i questions/query.txt -o explanations/full_data_flow.md -f
   ```

6. **Include the question and relevant files list in the answer:**

   ```sh
   Perpetual explain -m full -i questions/query.txt -o explanations/detailed_answer.md
   ```

7. **Use a separate instructions file for relevant files selection:**

   ```sh
   Perpetual explain -m normal -i questions/query.txt -e instructions/file_selection.txt -o explanations/targeted_answer.md
   ```

8. **Disable project description loading:**

   ```sh
   Perpetual explain -m normal -i questions/query.txt -o explanations/answer.md -df disabled
   ```

## LLM Configuration

The effectiveness of the `explain` operation relies on the configuration of the underlying LLM. Environment variables defined in `.env` files dictate the behavior and performance of the LLM during the explanation process. Proper configuration ensures accurate, relevant, and comprehensive explanations tailored to your project's specific needs.

Perpetual loads `.env` files from the project `.perpetual` directory and from the global Perpetual config directory. Environment variables already exported in the system environment take precedence over `.env` values.

1. **LLM Provider:**
   - `LLM_PROVIDER_OP_EXPLAIN_STAGE1`: Specifies the LLM provider for the first stage of the `explain` operation.
   - `LLM_PROVIDER_OP_EXPLAIN_STAGE2`: Specifies the LLM provider for the second stage of the operation.
   - If not set, both stages default to the general `LLM_PROVIDER`.

   Supported provider families include `anthropic`, `openai`, `ollama`, and `generic`.

2. **Model Selection:**
   - `ANTHROPIC_MODEL_OP_EXPLAIN_STAGE1`, `ANTHROPIC_MODEL_OP_EXPLAIN_STAGE2`: Define the Anthropic models used in each stage.
   - `OPENAI_MODEL_OP_EXPLAIN_STAGE1`, `OPENAI_MODEL_OP_EXPLAIN_STAGE2`: Specify the OpenAI models for each stage.
   - `OLLAMA_MODEL_OP_EXPLAIN_STAGE1`, `OLLAMA_MODEL_OP_EXPLAIN_STAGE2`: Specify the Ollama models for each stage.
   - `GENERIC_MODEL_OP_EXPLAIN_STAGE1`, `GENERIC_MODEL_OP_EXPLAIN_STAGE2`: Specify the models for the Generic OpenAI-compatible provider.

3. **Embeddings and Local Similarity Search:**
   - Local similarity search requires embeddings generated by the `embed` operation.
   - Configure the embedding model with provider-specific `*_MODEL_OP_EMBED` variables, for example:
     - `OPENAI_MODEL_OP_EMBED`
     - `OLLAMA_MODEL_OP_EMBED`
     - `GENERIC_MODEL_OP_EMBED`
   - Anthropic does not provide embedding support in this project. If you use Anthropic for `explain`, configure `LLM_PROVIDER_OP_EMBED` with another provider if you want local similarity search.
   - Embedding behavior can be tuned with provider-specific variables such as:
     - `*_EMBED_DOC_CHUNK_SIZE`
     - `*_EMBED_DOC_CHUNK_OVERLAP`
     - `*_EMBED_SEARCH_CHUNK_SIZE`
     - `*_EMBED_SEARCH_CHUNK_OVERLAP`
     - `*_EMBED_DIMENSIONS`
     - `*_EMBED_SCORE_THRESHOLD`
     - `*_EMBED_DOC_PREFIX`
     - `*_EMBED_SEARCH_PREFIX`

4. **Token Limits:**
   - `ANTHROPIC_MAX_TOKENS_OP_EXPLAIN_STAGE1`, `ANTHROPIC_MAX_TOKENS_OP_EXPLAIN_STAGE2`: Set the maximum number of output tokens for each stage when using Anthropic.
   - `OPENAI_MAX_TOKENS_OP_EXPLAIN_STAGE1`, `OPENAI_MAX_TOKENS_OP_EXPLAIN_STAGE2`: Define token limits for each stage when using OpenAI.
   - `OLLAMA_MAX_TOKENS_OP_EXPLAIN_STAGE1`, `OLLAMA_MAX_TOKENS_OP_EXPLAIN_STAGE2`: Define token limits for each stage when using Ollama.
   - `GENERIC_MAX_TOKENS_OP_EXPLAIN_STAGE1`, `GENERIC_MAX_TOKENS_OP_EXPLAIN_STAGE2`: Define token limits for each stage with the Generic provider.
   - `ANTHROPIC_MAX_TOKENS_SEGMENTS`, `OPENAI_MAX_TOKENS_SEGMENTS`, `OLLAMA_MAX_TOKENS_SEGMENTS`, `GENERIC_MAX_TOKENS_SEGMENTS`: Limit the number of continuation segments if token limits are reached, preventing excessive API calls.

   Stage 1 usually needs a smaller token limit because it only returns a file list. Stage 2 often needs a larger token limit because it generates the final answer.

5. **Retry Settings:**
   - `ANTHROPIC_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1`, `ANTHROPIC_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2`: Define the number of retry attempts on failure for each stage with Anthropic.
   - `OPENAI_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1`, `OPENAI_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2`: Define retry attempts on failure for each stage with OpenAI.
   - `OLLAMA_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1`, `OLLAMA_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2`: Define retry attempts on failure for each stage with Ollama.
   - `GENERIC_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1`, `GENERIC_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2`: Define retry attempts on failure for each stage with the Generic provider.

6. **Temperature:**
   - `ANTHROPIC_TEMPERATURE_OP_EXPLAIN_STAGE1`, `ANTHROPIC_TEMPERATURE_OP_EXPLAIN_STAGE2`: Set the temperature for each stage when using Anthropic.
   - `OPENAI_TEMPERATURE_OP_EXPLAIN_STAGE1`, `OPENAI_TEMPERATURE_OP_EXPLAIN_STAGE2`: Set the temperature for each stage when using OpenAI.
   - `OLLAMA_TEMPERATURE_OP_EXPLAIN_STAGE1`, `OLLAMA_TEMPERATURE_OP_EXPLAIN_STAGE2`: Set the temperature for each stage when using Ollama.
   - `GENERIC_TEMPERATURE_OP_EXPLAIN_STAGE1`, `GENERIC_TEMPERATURE_OP_EXPLAIN_STAGE2`: Set the temperature for each stage when using the Generic provider.

   Lower values (for example, 0.2–0.3) yield more focused and consistent file selection and explanations, while higher values (0.7–1.0) can produce more creative and varied outputs.

7. **Reasoning and Thinking Features:**
   - `ANTHROPIC_THINK_TOKENS_OP_EXPLAIN_STAGE2`: Set a thinking token budget for Anthropic models that support extended thinking.
   - `OPENAI_REASONING_EFFORT_OP_EXPLAIN_STAGE1`, `OPENAI_REASONING_EFFORT_OP_EXPLAIN_STAGE2`: Configure reasoning effort for OpenAI reasoning-capable models.
   - `OLLAMA_THINK_OP_EXPLAIN_STAGE1`, `OLLAMA_THINK_OP_EXPLAIN_STAGE2`: Enable or disable reasoning/thinking for Ollama models that support it. Supported values depend on the Ollama version and model; examples include `true`, `false`, `low`, `medium`, and `high`.
   - `GENERIC_REASONING_EFFORT_OP_EXPLAIN_STAGE1`, `GENERIC_REASONING_EFFORT_OP_EXPLAIN_STAGE2`: Configure reasoning effort for compatible Generic provider APIs.

8. **Other LLM Parameters:**
   Provider-specific parameters can be customized per stage by appending `_OP_EXPLAIN_STAGE1` or `_OP_EXPLAIN_STAGE2` to the variable names. Examples include:
   - `*_TOP_P`
   - `*_TOP_K` where supported
   - `*_SEED` where supported
   - `*_REPEAT_PENALTY` where supported
   - `*_FREQ_PENALTY`
   - `*_PRESENCE_PENALTY`
   - OpenAI-specific service tier settings such as `OPENAI_SERVICE_TIER_OP_EXPLAIN_STAGE1`, `OPENAI_SERVICE_TIER_OP_EXPLAIN_STAGE2`, and `OPENAI_SERVICE_TIER_FALLBACK`.

### Example Configuration in `.env` File

```sh
LLM_PROVIDER="openai"

OPENAI_API_KEY="<your api key goes here>"
OPENAI_BASE_URL="https://api.openai.com/v1"

OPENAI_MODEL_OP_EXPLAIN_STAGE1="<model-for-file-selection>"
OPENAI_MODEL_OP_EXPLAIN_STAGE2="<model-for-answer-generation>"
OPENAI_MODEL_OP_EMBED="text-embedding-3-small"

OPENAI_MAX_TOKENS_OP_EXPLAIN_STAGE1="4096"
OPENAI_MAX_TOKENS_OP_EXPLAIN_STAGE2="32768"
OPENAI_MAX_TOKENS_SEGMENTS="3"

OPENAI_TEMPERATURE_OP_EXPLAIN_STAGE1="0.2"
OPENAI_TEMPERATURE_OP_EXPLAIN_STAGE2="0.7"

OPENAI_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1="2"
OPENAI_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2="10"
```

This configuration sets the `openai` provider, configures separate models for file selection and answer generation, enables embeddings for local similarity search, defines token limits, sets temperatures to balance consistency and creativity, and allows retries on failure.

## Prompts Configuration

Customization of LLM prompts for the `explain` operation is managed through the `.perpetual/op_explain.json` configuration file. This file is initialized using the `init` operation, which sets up default prompts tailored to your project language. Adjusting these prompts can significantly influence the quality and relevance of the explanations generated by the LLM.

**Key Parameters:**

- **`system_prompt`**: Defines the initial prompt given to the LLM to set the context for the explanation task, tailored to the programming language of your project.

- **`system_prompt_ack`**: Provides an acknowledgment for the system prompt, used when the model does not support the system prompt and it needs to be converted to a regular user prompt.

- **`code_prompt`**: Prompt used in stage 2 when presenting the source code files selected for review to the LLM. This prompt instructs the LLM on how to process and analyze the provided source code files.

- **`code_response`**: Simulated acknowledgment response from the LLM after receiving the source code files, used to maintain conversation flow before the final question is asked.

- **`stage1_question_prompt`**: Frames the specific question or file-selection instructions in stage 1, prompting the LLM to generate a list of project files related to the question. The prompt should instruct the LLM to place filenames between the filename tags configured for the project.

- **`stage2_question_prompt`**: Formulates the main question for stage 2, building upon stage 1 findings and requesting the LLM to generate a comprehensive explanation based on the selected files.

- **`stage2_continue_prompt`**: Provides instructions for the LLM to continue generating responses if token limits are reached.

- **`output_question_header`**, **`output_files_header`**, **`output_answer_header`**: Headers used in `full` mode to structure the output with question, relevant files list, and answer sections.

- **`output_filename_tags`**, **`output_filtered_filename_tags`**: Tags used to format filenames in the output in `full` mode. `output_filtered_filename_tags` are used for files selected as relevant but filtered out from upload by the `no-upload` rule.

## Workflow

The `explain` operation is divided into preparation and two main processing stages:

1. **Preparation:**
   - **Project Setup:** Finds the project root and `.perpetual` directory, loads project and operation configuration, and loads the optional project description.
   - **File List Preparation:** Builds the project file list using project whitelist/blacklist rules, user-supplied filters, and the test-file blacklist unless `-u` is specified.
   - **Question Loading:** Reads the question from `-i` or stdin. If `-e` is supplied, reads separate file-selection instructions for stage 1.
   - **Annotation and Embedding Refresh:** Unless `-n` is specified, runs `annotate` and `embed` internally to refresh annotations and embeddings used for file selection and local similarity search.
   - **Context Saving Preselection:** If context saving is enabled and embeddings are available, preselects a subset of project files for stage 1 to reduce context usage on large projects.

2. **Stage 1: Relevant File Selection:**
   - **Project Index Request:** Sends the project index, annotations for the preselected files, optional project description, and the question or separate stage 1 instructions to the LLM.
   - **Response Handling:** Parses the LLM's response to extract a list of files that require further examination. The resulting paths are validated against the project file list, duplicates are removed, filename case is normalized where possible, and invalid paths are rejected.
   - **Local Similarity Search:** If embeddings are available and local search is enabled with `-s`, performs cosine similarity search to find additional relevant files based on semantic similarity to the question.
   - **Multiple Passes:** If `-sp` is greater than 1, stage 1 runs multiple times and merges the selected file lists.

   If the `list` mode is selected with `-m`, execution stops after this stage and outputs only the selected file list.

3. **Stage 2: Detailed Explanation:**
   - **No-Upload Filtering:** Files marked with `no-upload` comments are filtered out before source content is sent to the LLM unless the `-f` flag is used.
   - **Content Compilation:** Aggregates the contents of the selected files and prepares them for detailed analysis by the LLM.
   - **Optional Annotation Context:** If `-a` is enabled, adds project annotations to the stage 2 context.
   - **Source Code Review:** If files were selected in stage 1, presents them to the LLM using the `code_prompt` to establish context about the relevant source code.
   - **Question Processing:** Sends the main question to the LLM to generate a comprehensive explanation using the provided project context and selected source files.
   - **Response Handling:** Receives and compiles the LLM's response, handling scenarios where token limits are reached by utilizing continuation segments as configured.
   - **Output Formatting:** In `full` mode, formats the output to include the original question, list of relevant files, indicators for files filtered out by the `no-upload` rule, and the generated answer.

## Best Practices

1. **Craft Clear and Specific Questions:**
   - Formulate precise and well-defined questions to guide the LLM in generating accurate and relevant explanations. Ambiguous queries may lead to generalized or off-target responses. You can control answer verbosity by adding instructions at the end of your question, such as "Answer with a detail level of 5 out of 10" or "Provide a brief explanation."

2. **Use Separate File-Selection Instructions When Needed:**
   - If the final question is broad, abstract, or phrased in a way that makes file selection difficult, use `-e` to provide a more direct description of which part of the codebase should be inspected.

3. **Protect Sensitive Information:**
   - Use the `-f` flag cautiously to include files marked as `no-upload` only when necessary. Avoid exposing sensitive or proprietary code to the LLM provider unless absolutely essential for the explanation task.

4. **Enable Verbose Logging During Development:**
   - When setting up or troubleshooting the `explain` operation, enable debug or trace logging to gain insights into the operation's internal processes and to identify potential issues.

5. **Optimize Local Search:**
   - Use the `-s` parameter to control the number of additional files returned by local similarity search. Setting it to `0` disables local search entirely, while higher values may include more context but can also introduce noise.

6. **Use Context Saving Mode for Large Projects:**
   - For large codebases, use the `-c` parameter to control context usage. The `auto` mode uses project file-count thresholds from `project.json` to decide when to preselect a smaller subset of files before stage 1. Use `medium` or `high` to force stronger context-saving behavior, or `off` to disable preselection.
