# Onboard Operation

The `onboard` operation helps you set up and verify the **global LLM provider configuration** used by Perpetual. It validates your current environment configuration, reports the active settings, and can generate a default set of global environment files for a selected provider. This makes it easier to get a working LLM configuration before running operations that interact with a language model.

Unlike the `init` operation, `onboard` works with the **global configuration directory** rather than a project-local `.perpetual` directory. It does not require a project to be initialized and can be run from anywhere.

## Usage

To run the onboard operation, use the following command:

```sh
Perpetual onboard [flags]
```

The `onboard` operation supports several command-line flags to customize its behavior:

- `-c`: **Check the current global environment configuration**. Detects the active environment, prints it, and validates it. This flag cannot be combined with `-e`.
- `-e <provider>`: **Recreate the global env configuration** for the selected provider. Valid values are `anthropic`, `openai`, `ollama`, and `generic`.
- `-m <method>`: **Auth method to write with `-e`**, when applicable. Valid values are `Bearer` and `Basic`. This flag can only be used together with `-e`.
- `-k <value>`: **API key or `login:password` auth value to write with `-e`**. This flag can only be used together with `-e`.
- `-h`: **Display the help message**, showing all available flags and their descriptions.
- `-v`: **Enable debug logging**. This flag increases the verbosity of the operation's output.
- `-vv`: **Enable both debug and trace logging**. This flag provides the highest level of verbosity.

You must provide either `-c` (to check the current configuration) or `-e` (to recreate the configuration for a provider). Running the operation without one of these flags displays the usage message.

### Example Usage

Check the current global environment configuration:

```sh
Perpetual onboard -c
```

Recreate the global configuration for OpenAI and write the API key directly:

```sh
Perpetual onboard -e openai -k "sk-..."
```

Recreate the global configuration for a generic OpenAI-compatible provider using Bearer authentication:

```sh
Perpetual onboard -e generic -m Bearer -k "my-secret-token"
```

In all cases, after performing any requested configuration changes, the operation prints the detected active environment and validates it.

## Supported Providers

The `onboard` operation can generate configuration for the following LLM providers:

1. **Anthropic (`anthropic`)** – Uses an API key for authentication.
2. **OpenAI (`openai`)** – Uses an API key for authentication.
3. **Ollama (`ollama`)** – Supports optional authentication method and auth value.
4. **Generic (`generic`)** – OpenAI-compatible endpoints; supports authentication method and auth value, and requires a base URL.

The provider name is case-insensitive when passed with `-e`.

## Details

The `onboard` operation performs two main tasks, depending on the flags provided: recreating the global environment configuration (with `-e`), and detecting and validating the active environment (always performed).

### Recreating the Global Configuration

When run with `-e <provider>`, the operation rolls out a fresh set of global environment files:

1. **Validates the provider name** against the list of supported providers (`anthropic`, `openai`, `ollama`, `generic`).
2. **Validates the auth method** passed with `-m`, if any. Only `Bearer` and `Basic` are accepted (case-insensitive, normalized to `Bearer`/`Basic`).
3. **Creates the global configuration directory** if it doesn't exist.
4. **Removes existing configuration files** from the global configuration directory. A warning is printed before any files are removed.
5. **Writes example files** (`*.env.example`) for the providers that are not being generated as active configuration, so they remain available for reference.
6. **Generates a base env file** (e.g. `.env`) with the `LLM_PROVIDER` variable set to the selected provider.
7. **Generates a provider-specific env file** (e.g. `openai.env`) with the authentication details filled in based on `-m` and `-k`.

For the authentication value, the operation uses the value passed via `-k`. If `-k` is not provided, it attempts to detect a value from existing environment variables (`<PROVIDER>_AUTH` or `<PROVIDER>_API_KEY`). If no value can be determined, a warning is printed, instructing you to edit the global env config manually or rerun `onboard` with `-k`.

How authentication is written depends on the provider:

- For **Anthropic** and **OpenAI**, the auth value is written to `<PROVIDER>_API_KEY`.
- For **Ollama** and **Generic**, the auth method (if provided) is written to `<PROVIDER>_AUTH_TYPE`, and the auth value (if available) is written to `<PROVIDER>_AUTH`.

**Note:** Recreating the configuration removes all existing files in the global configuration directory. Back up any manual changes before running `onboard -e`.

### Detecting and Validating the Active Environment

After any configuration changes (or when run with `-c`), the operation detects the currently active environment and validates it:

1. **Loads the global env files** from the global configuration directory.
2. **Detects active provider selections** based on `LLM_PROVIDER` and per-operation `LLM_PROVIDER_OP_*` variables.
3. **Collects the active environment variables**, including provider-selection variables, the fallback text encoding, and provider-specific variables that belong to active providers.
4. **Prints the active environment** to standard output, masking sensitive values (such as API keys, auth values, and base URLs) as `<hidden>`.
5. **Validates the configuration** for each supported operation and prints the results.

If validation fails, the operation reports an error and exits with a non-zero status.

### Validated Operations

The validation step checks the configuration for each of the following operations:

- `ANNOTATE`
- `EMBED`
- `IMPLEMENT_STAGE1`, `IMPLEMENT_STAGE2`, `IMPLEMENT_STAGE3`, `IMPLEMENT_STAGE4`
- `DOC_STAGE1`, `DOC_STAGE2`
- `EXPLAIN_STAGE1`, `EXPLAIN_STAGE2`

For each operation, the provider is resolved from the per-operation `LLM_PROVIDER_OP_<OPERATION>` variable, falling back to the global `LLM_PROVIDER` variable. The validation then checks for the presence of the required variables for the resolved provider, such as authentication, model, base URL, and maximum token settings, depending on the provider type.

The `EMBED` operation is treated as **optional**. If embeddings are not configured (or the selected provider does not support them, as with Anthropic), validation does not fail. Instead, a notice is printed indicating that semantic search will be disabled.

### Validation Output

The validation output is organized into the following sections:

- **Missing required env variables**: Lists required variables that are not set. When multiple alternative variable names satisfy a requirement, they are shown joined with `or`.
- **Configuration errors**: Lists problems such as invalid or unsupported provider names.
- **Notices**: Informational messages, for example when embeddings are not configured and semantic search will be disabled.
- **Selected providers and models**: For each validated operation, shows the resolved provider and model.

If there are any missing variables or configuration errors, the operation reports a validation failure and exits with an error.

## Global Configuration Location

The `onboard` operation works with the global configuration directory, which uses the OS-specific user config directory, for example:

```text
~/.config/Perpetual/
├── .env
├── ollama.env.example
├── openai.env.example
├── anthropic.env.example
└── generic.env.example
```

on Linux, or:

```text
<User profile dir>\AppData\Roaming\Perpetual\
├── .env
├── ollama.env.example
├── openai.env.example
├── anthropic.env.example
└── generic.env.example
```

on Windows.

When recreating the configuration for a provider, the corresponding provider-specific env file (for example, `openai.env`) is also generated alongside the base `.env` file, and the example file for the active provider is omitted.
