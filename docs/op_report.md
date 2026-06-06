# Report Operation

The `report` operation generates a comprehensive report of the project's source code that can be manually copied and pasted into an LLM user interface for further analysis or added to its internal knowledge base. The operation provides two types of reports: a detailed code report and a brief summary report.

The `report` operation relies on both the `op_report` and `project` configurations. For the brief report type, the `annotate` operation is executed first to ensure that file annotations are up to date before the report is generated.

## Usage

To run the `report` operation, use the following command:

```sh
Perpetual report [flags]
```

The `report` operation supports several command-line flags to customize its behavior:

- `-h`: Display the help message, showing all available flags and their descriptions.

- `-t <type>`: Select the report type. Valid values are:
  - `code` (default): Generates a detailed report containing the full source code of the selected project files.
  - `brief`: Generates a concise report from generated source code annotations, providing a summary of each selected file's contents and purpose.

- `-r <file>`: Specify the file path to write the report to. If not provided or empty, the report will be written to stdout, with human-readable logging output sent to stderr.

- `-df <file|disabled>`: Optional path to a project description file to forward into the `annotate` operation when generating a brief report. Use `disabled` to explicitly disable loading the project description during annotation.

- `-u`: Include unit test source files in the report. By default, unit test sources are excluded using the project's test-file blacklist.

- `-x <file>`: Specify a path to a user-supplied regex filter file for excluding certain files from the report. See more info about using the filter [here](user_filter.md).

- `-c <mode>`: Set the context saving mode used when the brief report type runs annotation generation. Valid values are:
  - `auto` (default)
  - `off`
  - `medium`
  - `high`

- `-v`: Enable debug logging. This flag increases the verbosity of the operation's output, providing more detailed information about the report generation process.

- `-vv`: Enable both debug and trace logging. This flag provides the highest level of verbosity, useful for troubleshooting or understanding the internal workings of the report generation process.

### Examples

1. **Generate a detailed code report and display it in the console:**

   ```sh
   Perpetual report
   ```

2. **Generate a brief report and save it to a file:**

   ```sh
   Perpetual report -t brief -r project_summary.txt
   ```

3. **Generate a detailed code report with debug logging:**

   ```sh
   Perpetual report -v
   ```

4. **Generate a report including unit test files:**

   ```sh
   Perpetual report -u
   ```

5. **Generate a report using a custom filter file:**

   ```sh
   Perpetual report -x custom_filter.json
   ```

6. **Generate a brief report while forwarding a custom project description to annotation generation:**

   ```sh
   Perpetual report -t brief -df ./docs/project_description.md
   ```

When executed, the `report` operation finds the project root, loads environment files, validates the project and report configurations, collects project files, and generates the requested report type. The report includes files that match the project's whitelist and are not excluded by the project blacklist, the unit-test blacklist unless `-u` is used, or the user-supplied blacklist.

For the `code` report type, the operation reads the selected source files and formats them into a single report using the report configuration's code prompt and filename tags. The output renderer also uses the project's Markdown code block mappings when formatting file contents.

For the `brief` report type, the operation first runs the `annotate` operation as an internal step, then loads the stored annotations and formats them into a summary report using the report configuration's brief prompt and filename tags. The `-c`, `-df`, and `-x` flags are forwarded to the internal annotation step where applicable. The `-u` flag controls which files are included in the final report; annotation generation itself follows the `annotate` operation's own file-selection behavior.

It is important to note that the `code` report type will include the contents of all selected project files, including files that might contain sensitive information. The `report` operation itself does not apply the `no-upload` comment filter, so use the project blacklist or a user filter file if particular files must never be included in generated reports.

For the `brief` report type, the operation will first run the `annotate` operation to ensure that annotations are current before generating the report. The `annotate` operation may also process files marked as no-upload, so it is possible to configure it to use a local LLM for privacy if needed.

The generated report can be used for various purposes, such as:

1. Providing a comprehensive overview of the project structure and contents.
2. Facilitating code reviews by presenting the codebase in a single document.
3. Enabling easy analysis of the project using an LLM by uploading the report into an LLM user interface for further manual analysis.
4. Creating documentation or summaries of the project's current state.
