# Explain Operation

The `explain` operation is designed to provide insightful answers to questions and clarifications about your project based on thorough source code analysis. This operation examines your project's structure and code to generate comprehensive explanations, aiding developers in understanding complex codebases, identifying potential issues, implementing new features, and more.

## Usage

To utilize the `explain` operation, execute the following command:

```sh
Perpetual explain [flags]
```

The `explain` operation offers a range of command-line flags to tailor its functionality to your specific needs:

- `-i <file>`: Specify the input file containing the question to be answered. If omitted, the operation reads the question from standard input (stdin).

- `-r <file>`: Define the target file for writing the answer. If omitted, the answer is output to standard output (stdout) and all program logging is redirected to stderr.

- `-a`: Enable the addition of full project annotations in the request along with files requested by the LLM for analysis. This flag enhances the quality of answers by providing the LLM with additional context from annotated source files. Disabled by default to save tokens and lower the context window size requirement.

- `-l`: Activate "List Files Only" mode. Instead of generating a full answer, this flag lists the files relevant to the question based on project annotations (produced with the `annotate` operation). May be useful when performing simple semantic file-search queries. This task is always included at stage 1 of the operation in full mode.

- `-n`: Enable "No Annotate" mode, which skips the re-annotation of changed files and utilizes existing annotations if available. This can reduce API calls but may lower the quality of the answers.

- `-s`: Try to salvage incorrect filenames on stage 1. Experimental feature, use in projects with a large number of files where LLM tends to make more mistakes when generating list of files to analyze.

- `-f`: Override the 'no-upload' file filter to include files marked as 'no-upload' for review. As a result, it may upload sensitive files to the LLM when generating the final answer.

- `-u`: Include unit-test source files in the processing. By default, unit-test files are excluded to focus on primary source code.

- `-x <file>`: Provide a path to a user-supplied regex filter file. This file allows for the exclusion of specific files or patterns from processing based on custom criteria.

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

   Instead of an answer, this will list the files related to data processing based on project annotations.

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

## LLM Configuration

The `explain` operation's effectiveness relies on the configuration of the underlying LLM. Environment variables defined in the `.env` file dictate the behavior and performance of the LLM during the explanation process. Proper configuration ensures accurate, relevant, and comprehensive explanations tailored to your project's specific needs.

1. **LLM Provider:**
   - `LLM_PROVIDER_OP_EXPLAIN_STAGE1`: Specifies the LLM provider for the first stage of the `explain` operation.
   - `LLM_PROVIDER_OP_EXPLAIN_STAGE2`: Specifies the LLM provider for the second stage of the operation.
   - If not set, both stages default to the general `LLM_PROVIDER`.

2. **Model Selection:**
   - `ANTHROPIC_MODEL_OP_EXPLAIN_STAGE1`, `ANTHROPIC_MODEL_OP_EXPLAIN_STAGE2`: Define the Anthropic models used in each stage (e.g., "claude-3-sonnet-20240229" for stage 1 and "claude-3-opus-20240229" for stage 2).
   - `OPENAI_MODEL_OP_EXPLAIN_STAGE1`, `OPENAI_MODEL_OP_EXPLAIN_STAGE2`: Specify the OpenAI models for each stage.
   - `OLLAMA_MODEL_OP_EXPLAIN_STAGE1`, `OLLAMA_MODEL_OP_EXPLAIN_STAGE2`: Specify the Ollama models for each stage.
   - `GENERIC_MODEL_OP_EXPLAIN_STAGE1`, `GENERIC_MODEL_OP_EXPLAIN_STAGE2`: Specify the models for the Generic (OpenAI compatible) provider.

3. **Token Limits:**
   - `ANTHROPIC_MAX_TOKENS_OP_EXPLAIN_STAGE1`, `ANTHROPIC_MAX_TOKENS_OP_EXPLAIN_STAGE2`: Set the maximum number of tokens for each stage when using Anthropic.
   - `OPENAI_MAX_TOKENS_OP_EXPLAIN_STAGE1`, `OPENAI_MAX_TOKENS_OP_EXPLAIN_STAGE2`: Define token limits for each stage with OpenAI.
   - `OLLAMA_MAX_TOKENS_OP_EXPLAIN_STAGE1`, `OLLAMA_MAX_TOKENS_OP_EXPLAIN_STAGE2`: Define token limits for each stage with Ollama.
   - `GENERIC_MAX_TOKENS_OP_EXPLAIN_STAGE1`, `GENERIC_MAX_TOKENS_OP_EXPLAIN_STAGE2`: Define token limits for each stage with the Generic (OpenAI compatible) provider.
   - `ANTHROPIC_MAX_TOKENS_SEGMENTS`, `OPENAI_MAX_TOKENS_SEGMENTS`, `OLLAMA_MAX_TOKENS_SEGMENTS`, `GENERIC_MAX_TOKENS_SEGMENTS`: Limit the number of continuation segments if token limits are reached, preventing excessive API calls.

4. **JSON Structured Output Mode:**
   Enable JSON-structured output mode for faster and more cost-effective responses at stage 1 by setting the following variables:

   ```sh
   ANTHROPIC_FORMAT_OP_EXPLAIN_STAGE1="json"
   OPENAI_FORMAT_OP_EXPLAIN_STAGE1="json"
   OLLAMA_FORMAT_OP_EXPLAIN_STAGE1="json"
   ```

   Ensure that the selected models support JSON-structured output for reliable performance. This setting is somewhat experimental and may improve or worsen the ability to obtain a list of files related to the question, depending on the model used. The Generic (OpenAI compatible) provider profile does not support JSON-structured output mode for now.

5. **Retry Settings:**
   - `ANTHROPIC_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1`, `ANTHROPIC_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2`: Define retry attempts on failure for each stage with Anthropic.
   - `OPENAI_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1`, `OPENAI_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2`: Define retry attempts on failure for each stage with OpenAI.
   - `OLLAMA_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1`, `OLLAMA_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2`: Define retry attempts on failure for each stage with Ollama.
   - `GENERIC_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1`, `GENERIC_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2`: Define retry attempts on failure for each stage with the Generic provider.

6. **Temperature:**
   - `ANTHROPIC_TEMPERATURE_OP_EXPLAIN_STAGE1`, `ANTHROPIC_TEMPERATURE_OP_EXPLAIN_STAGE2`: Set the temperature for each stage when using Anthropic.
   - `OPENAI_TEMPERATURE_OP_EXPLAIN_STAGE1`, `OPENAI_TEMPERATURE_OP_EXPLAIN_STAGE2`: Set the temperature for each stage when using OpenAI.
   - `OLLAMA_TEMPERATURE_OP_EXPLAIN_STAGE1`, `OLLAMA_TEMPERATURE_OP_EXPLAIN_STAGE2`: Set the temperature for each stage when using Ollama.
   - `GENERIC_TEMPERATURE_OP_EXPLAIN_STAGE1`, `GENERIC_TEMPERATURE_OP_EXPLAIN_STAGE2`: Set the temperature for each stage when using the Generic provider.

   Lower values (e.g., 0.3-0.5) yield more focused and consistent explanations, while higher values (0.5-0.9) can produce more creative and varied outputs.

7. **Other LLM Parameters:**
   - `TOP_K`, `TOP_P`, `SEED`, `REPEAT_PENALTY`, `FREQ_PENALTY`, `PRESENCE_PENALTY`: Customize these parameters for each stage by appending `_OP_EXPLAIN_STAGE1` or `_OP_EXPLAIN_STAGE2` to the variable names (e.g., `ANTHROPIC_TOP_K_OP_EXPLAIN_STAGE1`). These are particularly useful for fine-tuning the output of different models.

### Example Configuration in `.env` File

```sh
LLM_PROVIDER="openai"

...

OPENAI_MODEL_OP_EXPLAIN_STAGE1="gpt-4"
OPENAI_MODEL_OP_EXPLAIN_STAGE2="gpt-4"
OPENAI_MAX_TOKENS_OP_EXPLAIN_STAGE1="2048"
OPENAI_MAX_TOKENS_OP_EXPLAIN_STAGE2="4096"
OPENAI_MAX_TOKENS_SEGMENTS="3"
OPENAI_TEMPERATURE_OP_EXPLAIN_STAGE1="0.2"
OPENAI_TEMPERATURE_OP_EXPLAIN_STAGE2="0.7"
OPENAI_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE1="3"
OPENAI_ON_FAIL_RETRIES_OP_EXPLAIN_STAGE2="2"

# Enable JSON-structured output mode
OPENAI_FORMAT_OP_EXPLAIN_STAGE1="json"
```

This configuration sets the `openai` provider with the `gpt-4` model for both stages, defines appropriate token limits, sets temperatures to balance creativity and consistency, enables JSON-structured output mode for stage 1, and allows for retries on failure.

## Prompts Configuration

Customization of LLM prompts for the `explain` operation is managed through the `.perpetual/op_explain.json` configuration file. This file is initialized using the `init` operation, which sets up default prompts tailored to various aspects of your project. Adjusting these prompts can significantly influence the quality and relevance of the explanations generated by the LLM.

**Key Parameters:**

- **`system_prompt`**: Defines the initial prompt given to the LLM to set the context for the explanation task, tailored to the programming language of your project.

- **`system_prompt_ack`**: Provides an acknowledgment for the system prompt, used when the model does not support the system prompt and it needs to be converted to a regular user prompt.

- **`project_index_prompt`**: Instructs the LLM to analyze the project's annotations/index (used to determine relevant files for explanation).

- **`project_index_response`**: Simulated response from the LLM regarding the project annotations analysis.

- **`stage1_question_prompt`**: Frames the specific question to be addressed in stage 1 of the explanation process, prompting the LLM to generate a list of files from annotations related to the question.

- **`stage1_question_json_mode_prompt`**: Alternative prompt for JSON-structured output mode in stage 1.

- **`stage2_files_prompt`**: Guides the LLM in processing specific files selected for detailed explanation.

- **`stage2_files_response`**: Simulated response from the LLM regarding file processing.

- **`stage2_question_prompt`**: Formulates the main question for stage 2, building upon stage 1 findings.

- **`stage2_continue_prompt`**: Provides instructions for the LLM to continue generating responses if token limits are reached.

- **`filename_tags_rx`**: Regular expressions to identify and parse filenames in the LLM response.

- **`filename_tags`**: Tagging conventions that help the LLM recognize filenames within annotations.

- **`noupload_comments_rx`**: Regular expressions to detect `no-upload` comments, marking files that should not be processed for privacy or other reasons.

- **`stage1_output_key`**, **`stage1_output_schema`**, **`stage1_output_schema_desc`**, **`stage1_output_schema_name`**: Parameters that define the schema and structure of stage 1 outputs when JSON-structured output mode is enabled.

**Customization Tips:**

- **Adapt Prompts to Project Specifics**: Tailor the `system_prompt` and `stage<#>_question_prompt` to reflect the programming language, core technologies, and terminology used by your project for more accurate explanations.

- **Structured Responses**: Utilize JSON-structured output mode by configuring the relevant prompts and environment variables. This enables easier parsing and integration of LLM responses with other tools or documentation systems.

- **Handling Large Projects**: For extensive codebases, refine the prompts to focus on specific modules or components to avoid overwhelming the LLM and ensure detailed explanations.

- **Use Bigger and Smarter Models**: Utilizing more advanced models can enhance the quality and depth of explanations, providing more accurate and insightful answers to complex questions.

## Workflow

The `explain` operation is divided into two main stages.

1. **Stage 1: Index Analysis:**
   - **Project Index Request:** Sends a request to the LLM to analyze the project annotations/index and determine which files are related to the question asked.
   - **Response Handling:** Parses the LLM's response to extract a list of files that require further examination, ensuring they reside within the project's scope and adhere to filtering rules.

   If using the `-l` flag, execution stops here, outputting the list of files.

2. **Stage 2: Detailed Explanation:**
   - **Content Compilation:** Aggregates the contents of the identified files, preparing them for detailed analysis by the LLM.
   - **Question Processing:** Sends the main question along with the relevant file contents to the LLM to generate a comprehensive explanation.
   - **Response Handling:** Receives and compiles the LLM's response, handling scenarios where token limits are reached by utilizing continuation segments as configured.

## Best Practices

1. **Craft Clear and Specific Questions:**
   - Formulate precise and well-defined questions to guide the LLM in generating accurate and relevant explanations. Ambiguous queries may lead to generalized or off-target responses. Control answer verbosity by adding prompts at the end of your question, such as "Answer at a detail level of 5 out of 10," or "Answer briefly/in detail."

2. **Protect Sensitive Information:**
   - Use the `-f` flag cautiously to include files marked as 'no-upload' only when necessary. Avoid exposing sensitive or proprietary code to the LLM provider unless it's essential for the explanation task.

3. **Enable Verbose Logging During Development:**
    - When setting up or troubleshooting the `explain` operation, enable debug or trace logging to gain insights into the operation's internal processes and identify potential issues.
