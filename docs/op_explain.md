# Explain Operation

The `explain` operation is designed to provide insightful answers to questions and clarifications about your project based on a thorough analysis of its source code. This operation examines your project's structure and code to generate comprehensive explanations, aiding developers in understanding complex codebases, identifying potential issues, implementing new features, and more.

## Usage

To utilize the `explain` operation, execute the following command:

```sh
Perpetual explain [flags]
```

The `explain` operation offers a range of command-line flags to tailor its functionality to your specific needs:

- `-i <file>`: Specify the input file containing the question to be answered. If omitted, the operation reads the question from standard input (stdin).  
- `-r <file>`: Define the target file for writing the answer. If omitted, the answer is output to standard output (stdout) and all program logging is redirected to stderr.  
- `-df <file>`: Optional path to project description file for adding into LLM context. Valid values are `file-path` to specify a custom description file, or `disabled` to explicitly disable loading the project description. If omitted, the operation attempts to load the default `description.md` file from the `.perpetual` directory.  
- `-e <file>`: Read instructions from a text or markdown file that will be used in stage 1 to select relevant files. Use this flag if the original question is too complex or not clear enough for the LLM to select relevant files, allowing you to provide separate instructions for the file selection process.  
- `-c <mode>`: Set the context saving mode to reduce LLM context usage for large projects. Valid values are `auto`, `off`, `medium`, or `high` (default: `auto`).  
- `-a`: Enable the addition of full project annotations in the request along with the files requested by the LLM for analysis. This flag enhances the quality of answers by providing the LLM with additional context from annotated source files. It is disabled by default to save tokens and reduce the context window size requirement.  
- `-l`: Activate "List Files Only" mode. Instead of generating a full answer, this flag lists the files relevant to the question based on project annotations (produced with the `annotate` operation). This mode may be useful when performing simple semantic file-search queries. Note that the file-search task is always performed during stage 1 of a full explanation.  
- `-n`: Enable "No Annotate" mode, which skips the re-annotation of changed files and utilizes existing annotations if available. This can reduce API calls but may lower the quality of the explanations.  
- `-f`: Override the 'no-upload' file filter to include files marked as 'no-upload' for review. Use this flag with caution, as it may upload sensitive files to the LLM during the explanation process.  
- `-u`: Include unit-test source files in the processing. By default, unit-test files are excluded to focus on primary source code.  
- `-x <file>`: Provide a path to a user-supplied regex filter file. This file allows for the exclusion of specific files or patterns from processing based on custom criteria. See more info about using the filter [here](user_filter.md).  
- `-s <n>`: Limit number of files related to the question returned by local similarity search. Valid values are integer ≥ 0 (`0` disables local search; only use LLM-requested files). Default: `5`.  
- `-sp <n>`: Set number of passes for related files selection at stage 1 (default: 1). Higher pass-count values will select more files, compensating for possible LLM errors when finding relevant files, but it will cost you more tokens and context use.  
- `-q`: Include the question text and the list of relevant files in the generated answer. This provides additional context in the output, showing which files were considered relevant and displaying the original question.  
- `-h`: Display the help message, detailing all available flags and their descriptions.  
- `-v`: Enable debug logging to receive detailed output about the operation's execution process.  
- `-vv`: Activate both debug and trace logging for the highest level of verbosity, offering an in-depth view of the operation's internal workings.  

### Examples

1. **Ask a question and receive an explanation:**

   ```sh
   echo "How does the authentication system work?" | Perpetual explain -r explanations/auth_system.md
   ```

   This command reads the question from standard input and writes the explanation to `explanations/auth_system.md`.

2. **List files relevant to a specific question without generating an explanation:**

   ```sh
   echo "What modules handle data processing?" | Perpetual explain -l
   ```

   Instead of an answer, this command produces a list of files related to data processing based on project annotations.

3. **Generate an explanation with additional annotations and debug logging enabled:**

   ```sh
   Perpetual explain -i questions/query.txt -r explanations/data_flow.md -a -v
   ```

4. **Include unit-test files in the explanation process:**

   ```sh
   Perpetual explain -i questions/query.txt -r explanations/test_data_flow.md -u
   ```

5. **Override the 'no-upload' filter to include all files for explanation:**

   ```sh
   Perpetual explain -i questions/query.txt -r explanations/full_data_flow.md -f
   ```

6. **Include the question and relevant files list in the answer:**

   ```sh
   Perpetual explain -i questions/query.txt -r explanations/detailed_answer.md -q
   ```

7. **Use separate instructions file for relevant files selection:**

   ```sh
   Perpetual explain -i questions/query.txt -e instructions/file_selection.txt -r explanations/targeted_answer.md
   ```

8. **Disable project description loading:**

   ```sh
   Perpetual explain -i questions/query.txt -r explanations/answer.md -df disabled
   ```

## LLM Configuration

The effectiveness of the `explain` operation relies on the configuration of the underlying LLM. Environment variables defined in the `.env` file dictate the behavior and performance of the LLM during the explanation process. Proper configuration ensures accurate, relevant, and comprehensive explanations tailored to your project's specific needs.

1. **LLM Provider:**
   - `LLM_PROVIDER_OP_EXPLAIN_STAGE1`: Specifies the LLM provider for the first stage of the `explain` operation.
   - `LLM_PROVIDER_OP_EXPLAIN_STAGE2`: Specifies the LLM provider for the second stage of the operation.
   - If not set, both stages default to the general `LLM_PROVIDER`.

2. **Model Selection:**
   - `ANTHROPIC_MODEL_OP_EXPLAIN_STAGE1`, `ANTHROPIC_MODEL_OP_EXPLAIN_STAGE2`: Define the Anthropic models used in each stage (for example, "claude-3-7-sonnet-latest" for stage 1 and "claude-sonnet-4-20250514" for stage 2).
   - `OPENAI_MODEL_OP_EXPLAIN_STAGE1`, `OPENAI_MODEL_OP_EXPLAIN_STAGE2`: Specify the OpenAI models for each stage (for example, "gpt-4.1" for stage 1 and "o4-mini" for stage 2).
   - `OLLAMA_MODEL_OP_EXPLAIN_STAGE1`, `OLLAMA_MODEL_OP_EXPLAIN_STAGE2`: Specify the Ollama models for each stage (for example, "qwen3:32b" for both stages).
   - `GENERIC_MODEL_OP_EXPLAIN_STAGE1`, `GENERIC_MODEL_OP_EXPLAIN_STAGE2`: Specify the models for the Generic (OpenAI compatible) provider.

3. **Token Limits:**
   - `ANTHROPIC_MAX_TOKENS_OP_EXPLAIN_STAGE1`, `ANTHROPIC_MAX_TOKENS_OP_EXPLAIN_STAGE2`: Set the maximum number of tokens for each stage when using Anthropic (typically 1024 for stage 1 and 8192 for stage 2).
   - `OPENAI_MAX_TOKENS_OP_EXPLAIN_STAGE1`, `OPENAI_MAX_TOKENS_OP_EXPLAIN_STAGE2`: Define token limits for each stage when using OpenAI (typically 1024 for stage 1 and 8192 for stage 2).
   - `OLLAMA_MAX_TOKENS_OP_EXPLAIN_STAGE1`, `OLLAMA_MAX_TOKENS_OP_EXPLAIN_STAGE2`: Define token limits for each stage when using Ollama (typically 1024 for stage 1 and 8192 for stage 2).
   - `GENERIC_MAX_TOKENS_OP_EXPLAIN_STAGE1`, `GENERIC_MAX_TOKENS_OP_EXPLAIN_STAGE2`: Define token limits for each stage with the Generic provider.
   - `ANTHROPIC_MAX_TOKENS_SEGMENTS`, `OPENAI_MAX_TOKENS_SEGMENTS`, `OLLAMA_MAX_TOKENS_SEGMENTS`, `GENERIC_MAX_TOKENS_SEGMENTS`: Limit the number of continuation segments if token limits are reached, preventing excessive API calls.

4. **JSON Structured Output Mode:**
   Enable JSON-structured output mode for faster and more cost-effective responses at stage 1 by setting the following variables:

   ```sh
   ANTHROPIC_FORMAT_OP_EXPLAIN_STAGE1="json"
   OPENAI_FORMAT_OP_EXPLAIN_STAGE1="json"
   OLLAMA_FORMAT_OP_EXPLAIN_STAGE1="json"
   ```

   Ensure that the selected models support JSON-structured output for reliable performance. This setting is somewhat experimental and may improve or worsen the ability to obtain a list of files related to the question, depending on the model used. The Generic (OpenAI compatible) provider profile does not support JSON-structured output mode at this time.

5. **Retry Settings:**
   - `ANTHROPIC_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1`, `ANTHROPIC_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2`: Define the number of retry attempts on failure for each stage with Anthropic (typically 2 for stage 1 and 10 for stage 2).
   - `OPENAI_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1`, `OPENAI_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2`: Define retry attempts on failure for each stage with OpenAI (typically 2 for stage 1 and 10 for stage 2).
   - `OLLAMA_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1`, `OLLAMA_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2`: Define retry attempts on failure for each stage with Ollama (typically 3 for both stages).
   - `GENERIC_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1`, `GENERIC_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2`: Define retry attempts on failure for each stage with the Generic provider.

6. **Temperature:**
   - `ANTHROPIC_TEMPERATURE_OP_EXPLAIN_STAGE1`, `ANTHROPIC_TEMPERATURE_OP_EXPLAIN_STAGE2`: Set the temperature for each stage when using Anthropic (typically 0.2 for stage 1 and 1.0 for stage 2 with thinking models).
   - `OPENAI_TEMPERATURE_OP_EXPLAIN_STAGE1`, `OPENAI_TEMPERATURE_OP_EXPLAIN_STAGE2`: Set the temperature for each stage when using OpenAI (typically 0.2 for stage 1 and 0.7 for stage 2).
   - `OLLAMA_TEMPERATURE_OP_EXPLAIN_STAGE1`, `OLLAMA_TEMPERATURE_OP_EXPLAIN_STAGE2`: Set the temperature for each stage when using Ollama (typically 0.2 for stage 1 and 0.7 for stage 2).
   - `GENERIC_TEMPERATURE_OP_EXPLAIN_STAGE1`, `GENERIC_TEMPERATURE_OP_EXPLAIN_STAGE2`: Set the temperature for each stage when using the Generic provider.

   Lower values (e.g., 0.2–0.3) yield more focused and consistent explanations, while higher values (0.7–1.0) can produce more creative and varied outputs.

7. **Special Features:**
   - `ANTHROPIC_THINK_TOKENS_OP_EXPLAIN_STAGE2`: Set thinking token budget for Anthropic models that support extended thinking (typically 4096 for answer generation).
   - `OLLAMA_THINK_OP_EXPLAIN_STAGE2`: Enable or disable reasoning/thinking for Ollama models that support it (e.g., Qwen3, DeepSeek R1).

8. **Other LLM Parameters:**
   - `TOP_K`, `TOP_P`, `SEED`, `REPEAT_PENALTY`, `FREQ_PENALTY`, `PRESENCE_PENALTY`: Customize these parameters for each stage by appending `_OP_EXPLAIN_STAGE1` or `_OP_EXPLAIN_STAGE2` to the variable names (for example, `ANTHROPIC_TOP_K_OP_EXPLAIN_STAGE1`). These are particularly useful for fine-tuning outputs from different models.

### Example Configuration in `.env` File

```sh
LLM_PROVIDER="openai"

...

OPENAI_MODEL_OP_EXPLAIN_STAGE1="gpt-4.1"
OPENAI_MODEL_OP_EXPLAIN_STAGE2="o4-mini"
OPENAI_MAX_TOKENS_OP_EXPLAIN_STAGE1="1024"
OPENAI_MAX_TOKENS_OP_EXPLAIN_STAGE2="8192"
OPENAI_MAX_TOKENS_SEGMENTS="3"
OPENAI_TEMPERATURE_OP_EXPLAIN_STAGE1="0.2"
OPENAI_TEMPERATURE_OP_EXPLAIN_STAGE2="0.7"
OPENAI_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1="2"
OPENAI_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2="10"

# Enable JSON-structured output mode
OPENAI_FORMAT_OP_EXPLAIN_STAGE1="json"
```

This configuration sets the `openai` provider with the `gpt-4.1` model for stage 1 and `o4-mini` for stage 2, defines appropriate token limits, sets temperatures to balance creativity and consistency, enables JSON-structured output mode for stage 1, and allows for retries on failure.

## Prompts Configuration

Customization of LLM prompts for the `explain` operation is managed through the `.perpetual/op_explain.json` configuration file. This file is initialized using the `init` operation, which sets up default prompts tailored to various aspects of your project. Adjusting these prompts can significantly influence the quality and relevance of the explanations generated by the LLM.

**Key Parameters:**

- **`system_prompt`**: Defines the initial prompt given to the LLM to set the context for the explanation task, tailored to the programming language of your project.

- **`system_prompt_ack`**: Provides an acknowledgment for the system prompt, used when the model does not support the system prompt and it needs to be converted to a regular user prompt.

- **`code_prompt`**: Prompt used in stage 2 when presenting the source code files selected for review to the LLM. This prompt instructs the LLM on how to process and analyze the provided source code files.

- **`code_response`**: Simulated acknowledgment response from the LLM after receiving the source code files, used to maintain conversation flow in the multi-stage dialogue.

- **`stage1_question_prompt`**: Frames the specific question to be addressed in stage 1 of the explanation process, prompting the LLM to generate a list of files from annotations related to the question.

- **`stage1_question_json_mode_prompt`**: Alternative prompt for JSON-structured output mode in stage 1.

- **`stage2_question_prompt`**: Formulates the main question for stage 2, building upon stage 1 findings and requesting the LLM to generate a comprehensive explanation based on the selected files.

- **`stage2_continue_prompt`**: Provides instructions for the LLM to continue generating responses if token limits are reached.

- **`output_question_header`**, **`output_files_header`**, **`output_answer_header`**: Headers used when the `-q` flag is enabled to structure the output with question, relevant files list, and answer sections.

- **`output_filename_tags`**, **`output_filtered_filename_tags`**: Tags used to format filenames in the output, with special formatting for filtered files when using the `-q` flag.

- **`stage1_output_key`**, **`stage1_output_schema`**, **`stage1_output_schema_desc`**, **`stage1_output_schema_name`**: Parameters that define the schema and structure of stage 1 outputs when JSON-structured output mode is enabled.

## Workflow

The `explain` operation is divided into two main stages:

1. **Stage 1: Index Analysis:**
   - **Project Index Request:** Sends a request to the LLM to analyze the project annotations/index and determine which files are related to the question asked.
   - **Response Handling:** Parses the LLM's response to extract a list of files that require further examination, ensuring they reside within the project's scope and adhere to filtering rules.
   - **Local Similarity Search:** If embeddings are available and local search is enabled, performs cosine similarity search to find additional relevant files based on semantic similarity to the question.

   If the `-l` flag is specified, execution stops at this stage, outputting only the list of files.

2. **Stage 2: Detailed Explanation:**
   - **Content Compilation:** Aggregates the contents of the identified files, preparing them for detailed analysis by the LLM. Files marked with 'no-upload' comments are filtered out unless the `-f` flag is used.
   - **Source Code Review:** If files were selected in stage 1, presents them to the LLM using the `code_prompt` to establish context about the relevant source code.
   - **Question Processing:** Sends the main question along with the relevant file contents to the LLM to generate a comprehensive explanation.
   - **Response Handling:** Receives and compiles the LLM's response, handling scenarios where token limits are reached by utilizing continuation segments as configured.
   - **Output Formatting:** If the `-q` flag is enabled, formats the output to include the original question, list of relevant files (with indicators for filtered files), and the generated answer.

## Best Practices

1. **Craft Clear and Specific Questions:**
   - Formulate precise and well-defined questions to guide the LLM in generating accurate and relevant explanations. Ambiguous queries may lead to generalized or off-target responses. You can control answer verbosity by adding instructions at the end of your question, such as "Answer with a detail level of 5 out of 10" or "Provide a brief explanation."

2. **Protect Sensitive Information:**
   - Use the `-f` flag cautiously to include files marked as 'no-upload' only when necessary. Avoid exposing sensitive or proprietary code to the LLM provider unless absolutely essential for the explanation task.

3. **Enable Verbose Logging During Development:**
   - When setting up or troubleshooting the `explain` operation, enable debug or trace logging to gain insights into the operation's internal processes and to identify potential issues.

4. **Optimize Local Search:**
   - Use the `-s` parameter to control the number of files returned by local similarity search. Setting it to 0 disables local search entirely, while higher values may include more context but could also introduce noise.

5. **Use Context Saving Mode for Large Projects:**
   - For large codebases, use the `-c` parameter to control context usage. The `auto` mode automatically balances between aggressive and conservative file selection based on the number of files found in stage 1.
