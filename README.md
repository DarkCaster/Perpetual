# The Perpetual Project

LLM-driven software development helper.

This project is currently in an early stage of development, and users should expect to encounter various bugs and issues. The author(s) make no guarantees about the reliability, security, or suitability of this project for any particular use case.

## Description

`Perpetual` is an LLM assistant created to boost the productivity and efficiency of software developers. Its key feature is the ability to write and modify code based on textual descriptions provided by the programmer. `Perpetual` can automatically generate new code, make changes to existing code, and even create new files - all by analyzing the project's codebase and the programmer's instructions embedded as special comments. With `Perpetual`, developers can focus on high-level problem-solving while the AI assistant handles the tedious task of implementing the required functionality.

`Perpetual` is primarily focused on direct interaction with the project's codebase, without requiring any additional tools or server infrastructure (except the LLM access). This design allows `Perpetual` to remain simple and easily integrate into larger systems.

All of `Perpetual`'s operations are limited to the user's project directory, and it is not permitted to delete any files (at least for now).

**[TL;DR, go straight to Example](docs/example.md)**

## Limitations

Given `Perpetual`'s focus on direct codebase interaction and maintaining simplicity, the following limitations apply:

- It cannot install packages (npm, NuGet, etc.)
- It cannot set up the necessary development environment
- It cannot interact with version control systems (Git, SVN, etc.)
- It cannot execute arbitrary commands on the user's machine
- It provides only a command-line interface, which was chosen to preserve simplicity and enable integration
- Other user interface options are not currently planned

### Warning

While `Perpetual` tries to minimize the risk of destructive operations on the user's computer, there is still a small risk involved. The main danger lies in the unintentional modification of source files based on the LLM's responses. To reduce this risk, it automatically backs up the files it tries to change and creates a `stash` that can be (re)applied or reverted on command.

Since the LLM never provides a completely deterministic result, and the quality can vary from one run to the next, you may need to run the `Perpetual`'s `implement` operation multiple times to achieve a satisfactory result (see below how to use it).

It's important to remain vigilant and carefully review the changes made by `Perpetual` before integrating them into your codebase.

Note that `Perpetual` is a tool designed mainly to assist programmers, with the primary goal of writing routine code. `Perpetual` is not yet capable of designing the overall project architecture for you, unlike some other similar tools. Instead, `Perpetual` is focused on generating code based on your architectural vision. If you have created a poor architecture in which it is very difficult to create new code, `Perpetual` will likely produce a suboptimal result.

## Requirements

The key requirement for `Perpetual` is access to a Large Language Model (LLM) to perform the core tasks of code generation and project analysis. Access to LLM models requires API keys for the corresponding LLM provider.

Currently `Perpetual` supports working with OpenAI and Anthropic models. It also supports locally hosted models with Ollama (highly experimental). If using OpenAI - avoid using GPT-3.5-Turbo and other legacy models with small context windows, they simply cannot fit content of the request inside. For Anthropic - Claude 3 Haiku is the minimum suitable model. For Ollama - DeepSeek Coder (v1) 34b model can be used to offload some tasks locally.

`Perpetual` utilizes the LangChain library for Go, which can be found at the following GitHub project:

<https://github.com/tmc/langchaingo>

The quality of `Perpetual` results directly depends on the LLM used. `Perpetual` allows you to offload different tasks to different models and providers to save on your costs. For example, code annotation or change planning tasks can be performed on more affordable models like Claude 3 Haiku and Claude 3 Sonnet, while the actual code writing can be handled by a more advanced model like Claude 3 Opus or GPT-4o.

## Getting Started

### Obtain API Keys

To get started with `Perpetual`, you need to obtain the necessary API keys to access the LLM models that power the core functionality.

### Download or Compile Perpetual

Next, you need to download or compile the `Perpetual` executable file (you can download binaries from GitHub releases, or latest build from Actions)

### Command Line Usage

`Perpetual` is designed to be used from the command line. To see the available operations, you can simply run the `Perpetual` command without any parameters:

Supported operations:

- [`init`: Initialize new .perpetual directory to store the configuration](docs/op_init.md)
- [`annotate`: Generate annotations for project files](docs/op_annotate.md)
- [`implement`: Implement code according to instructions marked with ###IMPLEMENT### comments](docs/op_implement.md)
- [`stash`: Rollback or re-apply generated code](docs/op_stash.md)
- [`report`: Create report from project source code, that can be manually uploaded into the LLM for use as knowledge base or for manual analysys.](docs/op_report.md)
- [`doc`: Create or rework documentation files (in markdown or plain-text format)](docs/op_doc.md)

### Initialize a New Project

To initialize a new `Perpetual` project, navigate to the root directory of your project in the console, and run the following command:

```sh
Perpetual init -l <language>
```

The perpetual init command creates a .perpetual directory in the root of your project, which contains various system settings that you can customize as needed:

- Prompts for different operations and stages
- [`.env.example` file with example settings](.perpetual/.env.example)
- Regular expressions used for parsing responses from the LLM
- LLM chat logs

Additional files created when executing `Perpetual` operations. **DO NOT ADD THESE TO YOUR VCS** - these files are platform dependent:

- `.annotations.json` Current annotations generated for your project files.
- `.message_log.txt` Raw LLM interaction log, see below

You should be cautious when modifying these settings. You can always rewrite them by running the init operation in the project root directory again.

Next, you need to manually create a `.env` file by copying the [`.env.example`](.perpetual/.env.example) file. The `.env` file should be self-explanatory.

### Creating Project Annotations

After initializing a new `Perpetual` project and setting up `.env` file, the next step is to create your project source code annotations. These annotations will be used by the LLM to request relevant files for analysis, which is essential for generating accurate and relevant code, while not overloading LLM context window with unrelevant code. **NOTE**: it is not required now to run this command, it will be triggered automatically when needed. However, you still can do it. This operation may take pretty long time when run for the first time with a local LLM (so it may be convenient to start it manually and go for a nap)

To create source code annotations, you need to use perpetual annotate command:

```sh
Perpetual annotate
```

**Tip**: Use cheaper models like Claude 3 Haiku for generating annotations. This will be much more cost-effective and faster, because it needed to upload EVERY suitable source code file from your project to LLM in order to generate its summary. Next time, annotation will be run automatically before other operations, and **it will only re-annotate changed files** in order to minimize costs.

### Writing Code with Perpetual

Key function of `Perpetual` is to assist you in writing code for your project. `Perpetual` can generate code for tasks that are marked in your source code files using the special comment `###IMPLEMENT###` followed by instructions (also comments). It will automatically analyze the code of your project and write its own code in the context of your project. Depending on command line flags it may implement code for all files where `###IMPLEMENT###` comment found, or only for one specific file. It can also create new files to place the code it generate

Example:

```go
func ParseCustomer(jsonMessage string) (Customer,error) {
 //###IMPLEMENT###
 //parse jsonMessage into the Customer struct
 //check all fields in the same way as in the code for adding a new client
 //if reporting error, use helper methods from "reporting" subproject
 //write unit tests in a separate file
}
```

Then you need to run Perpetual with implement operation. [See this doc for more info](docs/op_implement.md)

### Generating project report for manual use with LLM

The `report` operation allows you to generate a report from your project's source code in Markdown format. You can then upload this file into your LLM chat-interface/knowledge base/Vector DB for manual analysis, bug searching, etc. Currently it only support Markdown formatting for code that seem to be optimal both for popular commercial and opensource LLMs.

[See this doc for more info](docs/op_report.md)

### Creating project documentation

The `doc` operation in `Perpetual` is designed to assist in creating and refining project documentation. For now it can only work with plain-text or markdown formatted files. The operation can be particularly useful for maintaining up-to-date documentation that accurately reflects the current state of your project. This is highly experimental feature, and it will provide good results only with big and smart enough models. It will also take much more tokens to generate or refine a document than writing code with `implement` operation.

[See this doc for more info](docs/op_doc.md)

## Conclusion

This project was an experimental endeavor, with a significant portion of the code written using an LLM (Claude 3), which inspired the name `Perpetual`. Despite many years of programming experience, Go was a new language for me, and this was my first program written in it. As a result, the quality of the project's code is, as expected, quite poor.

This project served as a valuable learning experience, allowing me to explore the capabilities and limitations of using an LLM to assist with code generation. While the LLM was able to help me write the code, the overall quality and architecture of the project suffered due to my own lack of familiarity with the language and best practices.

The process of developing `Perpetual` has provided insights into both the potential and the challenges of LLM-assisted programming. It highlighted the importance of human expertise in guiding the overall structure and design of a project, while demonstrating the power of AI in handling routine coding tasks and generating boilerplate code.

This README, like much of the project, was also written with the assistance of an LLM, showcasing the tool's potential for documentation as well as code generation.

Moving forward, `Perpetual` serves as a starting point for further exploration into AI-assisted software development. While it may not be production-ready in its current state, it provides a foundation for understanding how LLMs can be integrated into the software development process and what challenges need to be addressed for more robust AI-assisted coding tools in the future.

PS: I also noticed how LLM likes to praise himself. I deliberately did not delete several paragraphs above that he added completely without my request. As they say: "if you don't praise yourself, no one else will"

## Disclaimer

This project and its associated materials are provided "as is" and without warranty of any kind, either expressed or implied. The author(s) do not accept any liability for any issues, damages, or losses that may arise from the use of this project or its components. Users are responsible for their own use of the project and should exercise caution and due diligence when incorporating any of the provided code or functionality into their own projects.

The project is intended for educational and experimental purposes only. It should not be used in production environments or for any mission-critical applications without thorough testing and validation. The author(s) make no guarantees about the reliability, security, or suitability of this project for any particular use case.

Users are encouraged to review the project's documentation, logs, and source code carefully before relying on it. If you encounter any problems or have suggestions for improvement, please feel free to reach out to the project maintainers.