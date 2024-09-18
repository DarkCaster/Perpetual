# Report operation

The `report` operation generates a comprehensive report of the project's source code. This report can be manually uploaded into a 3rd-party LLM for further analysis, or added to its internal knowledge base. The operation provides two types of reports: a detailed code report and a brief summary report.

While the `report` operation doesn't have its own configuration, it heavily relies on the `annotate` operation internally. Therefore, it depends on the proper configuration of the `annotate` operation for optimal results.

## Usage

To run the `report` operation, use the following command:

```sh
./Perpetual report [flags]
```

The `report` operation supports several command-line flags to customize its behavior:

- `-h`: Display the help message, showing all available flags and their descriptions.

- `-t <type>`: Select the report type. Valid values are:
  - `code` (default): Generates a detailed report containing the full source code of the project files.
  - `brief`: Generates a concise report from generated source code annotations, providing a summary of each file's contents and purpose.

- `-r <file>`: Specify the file path to write the report to. If not provided or empty, the report will be written to stderr.

- `-v`: Enable debug logging. This flag increases the verbosity of the operation's output, providing more detailed information about the report generation process.

- `-vv`: Enable both debug and trace logging. This flag provides the highest level of verbosity, useful for troubleshooting or understanding the internal workings of the report generation process.

Examples:

1. Generate a detailed code report and display it in the console:

   ```sh
   ./Perpetual report
   ```

2. Generate a brief report and save it to a file:

   ```sh
   ./Perpetual report -t brief -r project_summary.txt
   ```

3. Generate a detailed code report with debug logging:

   ```sh
   ./Perpetual report -v
   ```

When executed, the `report` operation will process the project files and generate the requested report type. The report will include all files that match the project's whitelist and are not excluded by the blacklist, as defined in the project's configuration.

It's important to note that the `code` report type will include the contents of all project files, including those that might contain sensitive information. This is something to keep in mind before uploading the report to an external LLM provider.

For the `brief` report type, the operation will first run the `annotate` operation to ensure that all file annotations are up-to-date before generating the report. The `annotate` operation may also process files marked as no-upload, so you can configure it to use a local LLM for privacy, if needed.

The generated report can be used for various purposes, such as:

1. Providing a comprehensive overview of the project structure and contents.
2. Facilitating code reviews by presenting the entire codebase in a single document.
3. Enabling easy analysis of the project using an LLM by uploading the report into the 3rd-party LLM interface/provider.
4. Creating documentation or summaries of the project's current state.
