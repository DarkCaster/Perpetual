# Annotate operation

The annotate operation is a crucial part of the `Perpetual`. It generates annotations for project source-code files, creating a summary of each file's contents and purpose. This operation is primarily used to maintain an up-to-date index of the project's structure and content, which is then utilized by other operations within the `Perpetual`. Project index is stored inside `.perpetual` directory and it only updated when neccecary saving your costs and time on LLM API access.

While the annotate operation is an essential component of the `Perpetual` workflow, it is not typically necessary to run it manually. Other operations, such as the `implement` operation, automatically trigger the annotate operation when needed to ensure that the project's annotations are current before proceeding with their tasks.

## Usage

To manually run the annotate operation, use the following command:

```shell
perpetual annotate [flags]
```

The annotate operation supports several command-line flags to customize its behavior:

- `-f`: Force annotation of all files, even for files which annotations are up to date. This flag is useful when you want to regenerate all annotations, regardless of whether the files have changed since the last annotation.

- `-d`: Perform a dry run without actually generating annotations. This flag will list the files that would be annotated, without making LLM requests and updating annotations.

- `-h`: Display the help message, showing all available flags and their descriptions.

- `-r <file>`: Only annotate a single specified file, even if its annotation is already up to date. This flag implies the `-f` flag. Use this when you want to update the annotation for a specific file. May be useful if annotating all changed project files in a batch hits LLM API limits.

- `-v`: Enable debug logging. This flag increases the verbosity of the operation's output, providing more detailed information about the annotation process.

- `-vv`: Enable both debug and trace logging. This flag provides the highest level of verbosity, useful for troubleshooting or understanding the internal workings of the annotation process.

Examples:

1. Annotate only new or changed files:

   ```shell
   perpetual annotate
   ```

2. Force (re)annotation of all files:

   ```shell
   perpetual annotate -f
   ```

3. Annotate a specific file:

   ```shell
   perpetual annotate -r path/to/file.go
   ```

When run, the annotate operation will process the specified files (or all changed files if no specific file is given) and generate or update their annotations. These annotations are then stored in the project's configuration directory for use by other `Perpetual` operations.
