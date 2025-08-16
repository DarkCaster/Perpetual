# Report Operation

The `report` operation generates a comprehensive report of the project's source code that can be manually copied and pasted into an LLM user interface for further analysis or added to its internal knowledge base. The operation provides two types of reports: a detailed code report and a brief summary report.

The `report` operation relies on both the `op_report` and `project` configurations and heavily depends on the `annotate` operation internally for optimal results. In the case of the brief report type, the `annotate` operation is executed first to ensure that all file annotations are up-to-date.

## Usage

To run the `report` operation, use the following command:

```sh
Perpetual report [flags]
```

The `report` operation supports several command-line flags to customize its behavior:

- `-h`: Display the help message, showing all available flags and their descriptions.

- `-t <type>`: Select the report type. Valid values are:
  - `code` (default): Generates a detailed report containing the full source code of the project files.
  - `brief`: Generates a concise report from generated source code annotations, providing a summary of each file's contents and purpose.

- `-r <file>`: Specify the file path to write the report to. If not provided or empty, the report will be written to stdout (with all logging output sent to stderr).

- `-u`: Include unit test source files in the report. By default, unit test sources are excluded.

- `-x <file>`: Specify a path to a user-supplied regex filter file for excluding certain files from the report. See more info about using the filter [here](user_filter.md).

- `-c <mode>`: Set the context saving mode to reduce LLM context usage for large projects. Valid values are:
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

When executed, the `report` operation will process the project files and generate the requested report type. The report will include all files that match the project's whitelist and are not excluded by the blacklist, as defined in the project's configuration.

It is important to note that the `code` report type will include the contents of all project files, including those that might contain sensitive information. This is something to keep in mind before uploading the report to an external LLM provider.

For the `brief` report type, the operation will first run the `annotate` operation to ensure that all file annotations are current before generating the report. The `annotate` operation may also process files marked as no-upload, so it is possible to configure it to use a local LLM for privacy if needed.

The generated report can be used for various purposes, such as:

1. Providing a comprehensive overview of the project structure and contents.
2. Facilitating code reviews by presenting the entire codebase in a single document.
3. Enabling easy analysis of the project using an LLM by uploading the report into an LLM user interface for further manual analysis.
4. Creating documentation or summaries of the project's current state.
