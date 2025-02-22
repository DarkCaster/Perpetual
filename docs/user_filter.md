# User Supplied Exclude Filter

The User Supplied Exclude Filter allows users to specify optional custom regular expression (regex) patterns to exclude certain files from being processed by the program's operations. The exclude filter is utilized in the following operations: `doc`, `explain`, `implement`, and `report`.

## Filter Structure

The exclude filter is defined using a JSON file containing an array of strings, where each entry is a regex pattern that matches file paths to be excluded from processing. This structure allows for straightforward customization and scalability, enabling users to manage large sets of exclusion rules efficiently.

## Example

```json
[
    "(?i)^vendor(\\\\|\\/).*",
    "other\\/.*\\.go$"
]
```

## Key Points

- **Regex Patterns**: Utilize standard Go regex syntax to define patterns.
- **Case Sensitivity**: Regex matching is case-sensitive by default. To make a regex case-insensitive, add `(?i)` to the beginning of the regex.
- **Directory Separator**: Directory separators are platform-specific. If you want your filter to work on any platform, you need to include both path separators in your regex.
- **Special Character Escaping**: Note that the `\` character has a special meaning in regexps. The `/` character does not need to be escaped in Go, but you may still want to use it for compatibility with other regex engines. To write a `\` character in a JSON string, you need to escape it with another `\`. For example, to pass `\\` to a Go regex, you need to specify it as `\\\\` inside a JSON string. Therefore, a regex group that matches path separators on any platform should be written in the JSON string as `(\\\\|\\/)`.
- **Relative Path**: File paths in `Perpetual` are passed as paths relative to the project root - they begin from the project root.
- **Anchors**: Use `^` and `$` to denote the start and end of a path for precise matching.
