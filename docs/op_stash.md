# Stash Operation

The `stash` operation is designed to manage and manipulate code changes generated by other operations, particularly the `implement` or `doc` operations. It allows users to rollback or re-apply generated code, providing a safety net for code modifications and enabling easy management of different versions of implemented changes.

The `stash` operation is primarily used internally with operations that modify files as an extra safety measure. Manual invocation is mainly needed when you want to quickly rollback changes. It is particularly useful when you are using slower and less convenient version control systems (VCS) like TFS or Perforce. In other cases, it is generally better to use file-change tracking within your preferred VCS.

## Usage

To use the `stash` operation, use the following command:

```sh
Perpetual stash [flags]
```

The `stash` operation supports several command-line flags to customize its behavior:

- `-h`: Display the help message, showing all available flags and their descriptions.

- `-l`: List all current stashes. This flag will display the names of all available stashes.

- `-a`: Apply changes from a specified stash. This flag is used to re-apply previously stashed changes.

- `-r`: Rollback changes from a specified stash. This flag is used to revert the changes made by a particular stash.

- `-lf`: List files in a specified stash. This flag shows the files that are affected by a particular stash.

- `-n <name>`: Set the stash name to apply or revert. If not specified, it defaults to the latest stash.

- `-f <filename>`: Select a single file to apply or revert from the stash. This is useful when you want to manipulate changes for a specific file.

- `-t <target_file>`: Specify a target file where the selected single file from the stash will be saved, relative to the project root. This is used in conjunction with the `-f` flag.

- `-v`: Enable debug logging. This flag increases the verbosity of the operation's output, providing more detailed information about the stash process.

- `-vv`: Enable both debug and trace logging. This flag provides the highest level of verbosity, useful for troubleshooting or understanding the internal workings of the stash process.

### Examples

1. **List all current stashes:**

   ```sh
   Perpetual stash -l
   ```

2. **Apply changes from the latest stash:**

   ```sh
   Perpetual stash -a
   ```

3. **Rollback changes from a specific stash:**

   ```sh
   Perpetual stash -r -n 2023-05-15_14-30-00
   ```

4. **List files in a specific stash:**

   ```sh
   Perpetual stash -lf -n 2023-05-15_14-30-00
   ```

5. **Apply changes for a single file from a stash:**

   ```sh
   Perpetual stash -a -n 2023-05-15_14-30-00 -f path/to/file.go
   ```

6. **Apply changes for a single file from a stash to a different target file:**

   ```sh
   Perpetual stash -a -n 2023-05-15_14-30-00 -f path/to/source_file.go -t path/to/target_file.go
   ```

When executed, the `stash` operation will perform the specified action (list, apply, or rollback) on the stashes stored in the project's `.perpetual/.stash` directory. Each stash is a JSON file containing the original and modified versions of the affected files.

## Stash Creation

While the `stash` operation itself doesn't create new stashes, stashes are automatically created by other operations, such as the `implement` operation. When code changes are generated, a new stash is created to store both the original and modified versions of the affected files. The stash is named using the current timestamp (format: `YYYY-MM-DD_HH-MM-SS.json`).

## Notes

- The `stash` operation ensures that changes can be safely applied or reverted, especially in environments where version control systems may not provide sufficient tracking.

- Using the `-f` and `-t` flags allows for granular control over individual file changes, providing flexibility in managing specific modifications without affecting the entire stash.
