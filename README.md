# The Perpetual Project

LLM-driven software development assistant.

## Description

`Perpetual` is an LLM-driven software development assistant designed to enhance the productivity and efficiency of software developers. Its primary function is to streamline the coding process by automating the generation and modification of code based on textual descriptions provided by programmers. It achieves this by analyzing the project's codebase and interpreting programmer instructions embedded as special comments. `Perpetual` can generate new code or entities in existing or new files, make changes to existing code, and write or refine documentation.

It focuses on direct interaction with the project's codebase, eliminating the need for additional tools, deployment, or server infrastructure (apart from LLM API access keys). This approach results in a simple and easily deployable tool that can be used directly by developers or integrated into larger AI software development ecosystems.

`Perpetual` operates strictly inside the user's project directory, ensuring a controlled and safe environment for code manipulation. Currently, it does not have the capability to delete files or run any external tools on the user's system, further safeguarding the project's integrity.

**TL;DR: Go straight to Examples**:

- [Conway's Game of Life](docs/example_2025_03/game_of_life.md)
- [Fractal Visualizer](docs/example_2025_02/fractal_visualizer.md)

## Limitations

[See current technical limitations](docs/limitations.md)

### Warning

While `Perpetual` tries to minimize the risk of destructive operations on the user's computer, there is still a small risk involved. The main danger lies in the unintentional modification of source files based on the LLM's responses. To reduce this risk, it automatically backs up the files it attempts to change and creates a `stash` that can be (re)applied or reverted on command.

Since the LLM almost never provides a completely deterministic result, and the quality can vary from one run to the next, you may need to run `Perpetual`'s `implement` operation multiple times to achieve a satisfactory result ([see below how to use it](#writing-code-with-perpetual)).

It's important to remain vigilant and carefully review the changes made by `Perpetual` before integrating them into your codebase.

Note that `Perpetual` is a tool designed mainly to assist programmers, with the primary goal of writing routine code. `Perpetual` is not yet capable of designing the overall project architecture for you, unlike some other similar tools. Instead, `Perpetual` is focused on generating code based on your architectural vision. If you have created a poor architecture in which it is very difficult to create new code, `Perpetual` will likely produce a suboptimal result.

## Requirements

The key requirement for `Perpetual` is access to a Large Language Model (LLM) to perform the core tasks of code generation and project analysis. Access to LLM models requires API keys for the corresponding LLM provider.

Currently, `Perpetual` supports working with OpenAI, Anthropic, Ollama, and generic OpenAI-compatible providers (with some limitations). If using OpenAI, avoid using GPT-3.5-Turbo and other legacy models with small context windows, as they simply cannot fit the content of the request. For Anthropic, Claude 3 Haiku is the minimum suitable model. For Ollama, the Qwen2.5-Coder-Instruct (7B and up) model can be used to offload some tasks locally. For other OpenAI-API compatible providers, [deepseek](https://www.deepseek.com) is known to work.

`Perpetual` utilizes the LangChain library for Go, which can be found at the following GitHub project:

<https://github.com/tmc/langchaingo>

The quality of `Perpetual`'s results directly depends on the LLM used. `Perpetual` allows you to offload different tasks to different models and providers to save on costs. For example, code annotation or change planning tasks can be performed on more affordable models like Claude 3 Haiku, while the actual code writing can be handled by a more advanced model like Claude 3 Opus, Claude 3.5 Sonnet, or GPT-4o.

## Getting Started

### Obtain API Keys

To get started with `Perpetual`, you need to obtain the necessary API keys to access the LLM models that power its core functionality.

### Download or Compile Perpetual

Download the latest `Perpetual` executable from GitHub Releases or GitHub Actions, or compile it yourself. You can extract the executable anywhere you like, but it is required to run the executable from your project root directory for which you want to use it.

### Configuration

In order to use `Perpetual`, you need to configure it first. See [this doc](docs/configuration.md) for configuration overview. Default project configuration may be installed by running `Perpetual init` (see below), and you will need to add your API keys and select the LLM provider to use as a minimum.

### Command Line Usage

`Perpetual` is designed to be used from the command line. To see the available operations, you can simply run the `Perpetual` command without any parameters:

Supported operations:

- [`init`: Initialize a new `.perpetual` directory to store the configuration](docs/op_init.md)
- [`annotate`: Generate annotations for project files](docs/op_annotate.md)
- [`implement`: Implement code according to instructions marked with `###IMPLEMENT###` comments](docs/op_implement.md)
- [`stash`: Rollback or re-apply generated code](docs/op_stash.md)
- [`report`: Create a report from project source code that can be manually uploaded into the LLM for use as a knowledge base or for manual analysis](docs/op_report.md)
- [`doc`: Create or rework documentation files (in markdown or plain-text format)](docs/op_doc.md)
- [`explain`: Answer questions about your project based on thorough source code analysis](docs/op_explain.md)

### Initialize a New Project

To initialize a new `Perpetual` project, navigate to the root directory of your project in the console and run the following command:

```sh
Perpetual init -l <language>
```

The `init` command creates a `.perpetual` directory in the root of your project, which contains various system settings (that you can customize as needed) and other service files:

- Config files with prompts and settings for different operations and regular expressions used for parsing responses from the LLM.
- [`.env.example` file with example settings](.perpetual/.env.example)
- Automatic backups for source code files it changes
- LLM chat logs

Additional files created when executing `Perpetual` operations. **DO NOT ADD THESE TO YOUR VCS** — these files are platform-dependent:

- `.annotations.json` — Current annotations generated for your project files.
- `.message_log.txt` — Raw LLM interaction log (see below).
- `.stash` subdirectory — Contains backups of source code files it changes.

You should be cautious when modifying these settings. You can always rewrite them by running the `init` command in the project root directory again.

Next, you need to manually create a `.env` file by copying the [`.env.example`](.perpetual/.env.example) file. The `.env` file should be self-explanatory.

### Creating Project Annotations

After initializing a new `Perpetual` project and setting up the `.env` file, the next step is to create your project source code annotations. These annotations will be used by the LLM to request relevant files for analysis, which is essential for generating accurate and relevant code while not overloading the LLM's context window with irrelevant code. **NOTE**: It is not required to run this command now; it will be triggered automatically when needed. However, you can still do it manually. This operation may take a considerable amount of time when run for the first time or with a local LLM (so it may be convenient to start it manually and take a break).

To create source code annotations, use the `annotate` command:

```sh
Perpetual annotate
```

**Tip**: Use more affordable models like Claude 3 Haiku for generating annotations. This will be much more cost-effective and faster because it needs to upload **ALL** suitable source code files from your project to the LLM to generate their summaries. Subsequent annotations will run automatically before other operations and **only re-annotate changed files** to minimize costs.

### Writing Code with Perpetual

The key function of `Perpetual` is to assist you in writing code for your project. `Perpetual` can generate code for tasks that are marked in your source code files using the special comment `###IMPLEMENT###` followed by instructions (also comments). It will automatically analyze the code of your project and write its own code in the context of your project. Depending on command-line flags, it may implement code for all files where the `###IMPLEMENT###` comment is found or only for one specific file. It can also create new files to place the code it generates.

#### Example

```go
func ParseCustomer(jsonMessage string) (Customer, error) {
	//###IMPLEMENT###
	// Parse jsonMessage into the Customer struct
	// Check all fields in the same way as in the code for adding a new client
	// If reporting an error, use helper methods from the "reporting" subproject
	// Write unit tests in a separate file
}
```

Then, run `Perpetual` with the `implement` operation:

```sh
Perpetual implement
```

[See this documentation for more info](docs/op_implement.md)

### Generating Project Report for Manual Use with LLM

The `report` operation allows you to generate a report from your project's source code in Markdown format. You can then upload this file into your LLM chat interface, knowledge base, or Vector DB for manual analysis, bug searching, etc. Currently, it only supports Markdown formatting for code that seems to be optimal for both popular commercial and open-source LLMs.

[See this documentation for more info](docs/op_report.md)

### Creating Project Documentation

The `doc` operation in `Perpetual` is designed to assist in creating and refining project documentation. Currently, it can only work with plain-text or markdown-formatted files. The operation can be particularly useful for maintaining up-to-date documentation that accurately reflects the current state of your project. This is a highly experimental feature, and it will provide good results only with large and sufficiently advanced models. It will also consume more tokens to generate or refine a document than writing code with the `implement` operation.

[See this documentation for more info](docs/op_doc.md)

### Explaining Project Code

The `explain` operation allows you to ask questions about your project based on thorough source code analysis. `Perpetual` analyzes the relevant parts of your codebase to provide accurate and insightful answers to your queries.

[See this documentation for more info](docs/op_explain.md)

## Disclaimer

This project and its associated materials are provided "as is" and without warranty of any kind, either expressed or implied. The author(s) do not accept any liability for any issues, damages, or losses that may arise from the use of this project or its components. Users are responsible for their own use of the project and should exercise caution and due diligence when incorporating any of the provided code or functionality into their own projects.

The project is intended for educational and experimental purposes only. It should not be used in production environments or for any mission-critical applications without thorough testing and validation. The author(s) make no guarantees about the reliability, security, or suitability of this project for any particular use case.

Users are encouraged to review the project's documentation, logs, and source code carefully before relying on it. If you encounter any problems or have suggestions for improvement, please feel free to reach out to the project maintainers.
