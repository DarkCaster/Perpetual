# Trying local LLMs with Ollama

In this document I collect my subjective tests of local LLMs with `Perpetual`, your results may vary.

My test-machine specs: AMD Ryzen 9 7945HX, 32G of RAM, Nvidia RTX 4070 Laptop GPU with 8G of VRAM limited to 70W of power. I consider this as a bare minimum for working with local LLMs for now.

### Qwen2.5-Coder-7B-Instruct

Tests performed at Nov 2024 with Ollama `0.3.14` - `0.4.1`.

Works great for generating annotations with the `op_annotate` operation, tested with C#, Python and Golang. The quality and repeatability of the result is stable enough for everyday use (only with `op_annotate`). This model can actually replace commercial LLMs with this task.

Quality of generated annotations may be further improved by using multi-step/multi-variant generation and post-processing, see [`.env.example`](.perpetual/.env.example) for more info.

Have not tested this model for generating code with `op_implement`, probably it is not good enough for the task as other sub-10B models.

I used the following example Modelfile (based on official ollama template for qwen):

```sh
FROM Qwen2.5-Coder-7B-Instruct.i1-Q5_K_S.gguf
PARAMETER temperature 0.7
PARAMETER num_ctx 12288
PARAMETER num_predict 2048
PARAMETER repeat_penalty 1.0
PARAMETER penalize_newline false
SYSTEM You are helpful AI assistant. Given the following conversation, relevant context, and a follow up question, reply with an answer to the current question the user is asking. Return only your response to the question given the above information following the users instructions as needed.
TEMPLATE """{{- if .Suffix }}<|fim_prefix|>{{ .Prompt }}<|fim_suffix|>{{ .Suffix }}<|fim_middle|>
{{- else if .Messages }}
{{- if or .System .Tools }}<|im_start|>system
{{- if .System }}
{{ .System }}
{{- end }}
{{- if .Tools }}

# Tools

You may call one or more functions to assist with the user query.

You are provided with function signatures within <tools></tools> XML tags:
<tools>
{{- range .Tools }}
{"type": "function", "function": {{ .Function }}}
{{- end }}
</tools>

For each function call, return a json object with function name and arguments within <tool_call></tool_call> XML tags:
<tool_call>
{"name": <function-name>, "arguments": <args-json-object>}
</tool_call>
{{- end }}<|im_end|>
{{ end }}
{{- range $i, $_ := .Messages }}
{{- $last := eq (len (slice $.Messages $i)) 1 -}}
{{- if eq .Role "user" }}<|im_start|>user
{{ .Content }}<|im_end|>
{{ else if eq .Role "assistant" }}<|im_start|>assistant
{{ if .Content }}{{ .Content }}
{{- else if .ToolCalls }}<tool_call>
{{ range .ToolCalls }}{"name": "{{ .Function.Name }}", "arguments": {{ .Function.Arguments }}}
{{ end }}</tool_call>
{{- end }}{{ if not $last }}<|im_end|>
{{ end }}
{{- else if eq .Role "tool" }}<|im_start|>user
<tool_response>
{{ .Content }}
</tool_response><|im_end|>
{{ end }}
{{- if and (ne .Role "assistant") $last }}<|im_start|>assistant
{{ end }}
{{- end }}
{{- else }}
{{- if .System }}<|im_start|>system
{{ .System }}<|im_end|>
{{ end }}{{ if .Prompt }}<|im_start|>user
{{ .Prompt }}<|im_end|>
{{ end }}<|im_start|>assistant
{{ end }}{{ .Response }}{{ if .Response }}<|im_end|>{{ end }}"""
PARAMETER stop "<|endoftext|>"
PARAMETER stop "<|file_sep|>"
PARAMETER stop "<|fim_prefix|>"
PARAMETER stop "<|fim_middle|>"
```

The chosen context size of 12K tokens is the minimum for the typical code I work with. You can try reducing it to 8K to save some VRAM, a typical symptom of context window overflow is that LLM stops following the generation instructions and starts writing code instead of a short annotation.

I've used the model from here: <https://huggingface.co/mradermacher/Qwen2.5-Coder-7B-Instruct-i1-GGUF>, Q5_K_S variant.

### Yi-Coder-9B-Chat-i1

Tests performed at Oct 2024 with Ollama `0.3.12`.

It is possible to use this model for generating decent file annotations (`OLLAMA_MODEL_OP_ANNOTATE` param at `.env` file). Tested for GoLang only. It was not tested for other operations and languages.

I've used the model from here: <https://huggingface.co/mradermacher/Yi-Coder-9B-Chat-i1-GGUF>, Q4_K_M variant.

I used the following example `Modelfile`:

```sh
FROM Yi-Coder-9B-Chat.i1-Q4_K_M.gguf
PARAMETER temperature 0.5
PARAMETER num_ctx 8192
PARAMETER num_predict 4096
PARAMETER repeat_penalty 1.0
PARAMETER penalize_newline false
SYSTEM You are a highly skilled software developer. You always write concise and readable code. You do not overload the user with unnecessary details in your answers and answer only the question asked. You are not adding separate explanations after code-blocks, you adding comments within your code instead.
TEMPLATE """{{- if .Messages }}
{{- range $i, $_ := .Messages }}
{{- $last := eq (len (slice $.Messages $i)) 1 -}}
<|im_start|>{{ .Role }}
{{ .Content }}{{ if (or (ne .Role "assistant") (not $last)) }}<|im_end|>
{{ end }}
{{- if (and $last (ne .Role "assistant")) }}<|im_start|>assistant
{{ end }}
{{- end }}
{{- else }}
{{- if .System }}<|im_start|>system
{{ .System }}<|im_end|>
{{ end }}{{ if .Prompt }}<|im_start|>user
{{ .Prompt }}<|im_end|>
{{ end }}<|im_start|>assistant
{{ end }}{{ .Response }}{{ if .Response }}<|im_end|>{{ end }}
"""
PARAMETER stop "<|endoftext|>"
PARAMETER stop "<|im_end|>"
PARAMETER stop "<fim_prefix>"
PARAMETER stop "<fim_suffix>"
PARAMETER stop "<fim_middle>"
```

## Below are some outdated test results and models that did not perform well enough to be used with `Perpetual`

### DeepSeek-Coder-V2-Lite-Instruct-i1

Tests performed at Oct 2024 with Ollama `0.3.12`.

It is possible to use this model for generating decent file annotations (`OLLAMA_MODEL_OP_ANNOTATE` param at `.env` file). Tested for GoLang only. It was not tested for other operations and languages. Sometimes tends to answer in Chinese, often writes more verbosely than necessary. Sometimes partially ignores additional summarize instructions embedded in the source file.

I've used the model from here: <https://huggingface.co/mradermacher/DeepSeek-Coder-V2-Lite-Instruct-i1-GGUF>, IQ4_XS variant.

I used the following example `Modelfile`:

```sh
FROM DeepSeek-Coder-V2-Lite-Instruct.i1-IQ4_XS.gguf
PARAMETER temperature 0.5
PARAMETER num_ctx 8192
PARAMETER num_predict 4096
PARAMETER repeat_penalty 1.0
PARAMETER penalize_newline false
SYSTEM You are a highly skilled software developer. You always write concise and readable code. You do not overload the user with unnecessary details in your answers and answer only the question asked. You are not adding separate explanations after code-blocks, you adding comments within your code instead.
TEMPLATE """{{ if .System }}Answer in English. {{ .System }}

{{ end }}{{ if .Prompt }}User: Answer in English.

{{ .Prompt }}

{{ end }}Assistant: {{ .Response }}

"""
PARAMETER stop """
User:"""
PARAMETER stop """
Assistant:"""
```

### CodeGemma-1.1-7B-it

Tests performed at May 2024 with Ollama `0.1.38`.

It is possible to use this model for generating file annotations (`OLLAMA_MODEL_OP_ANNOTATE` param at `.env` file). For other operations it doesn't seem to work very well, it just doesn't follow the instructions well enough most of the times. Using low temperature like 0.3-0.5, `OLLAMA_REPEAT_PENALTY="1.0"` and `OLLAMA_TOP_K="20"` may help a bit to provide better results. Model can also generate decent code (sometimes) at final stage 3 of implement operation (`OLLAMA_MODEL_OP_IMPLEMENT_STAGE3`).

I've used the model from here: <https://huggingface.co/bartowski/codegemma-1.1-7b-it-GGUF>

To run it with Ollama, I used the following example `Modelfile` (I don't guarantee it is correct or optimal, especially the prompt template, I've manually increased context num_ctx, because original value is too small when working with multiple files at once). Models with less quantization should also work, but are less likely to succeed:

```sh
FROM codegemma-1.1-7b-it-Q6_K.gguf
PARAMETER temperature 0.5
PARAMETER num_ctx 32768
PARAMETER num_predict 4096
PARAMETER repeat_penalty 1.0
PARAMETER penalize_newline false
SYSTEM You are a highly skilled software developer. You always write concise and readable code. You do not overload the user with unnecessary details in your answers and answer only the question asked. You are not adding separate explanations after code-blocks, you adding comments within your code instead.
TEMPLATE """<start_of_turn>user
{{ if .System }}{{ .System }} {{ end }}{{ .Prompt }}<end_of_turn>
<start_of_turn>model
{{ .Response }}<end_of_turn>
"""
PARAMETER stop "<start_of_turn>"
PARAMETER stop "<end_of_turn>"

```

## StarCoder2

Tests performed at May 2024 with Ollama `0.1.38`.

Despite the fact that it generates code quite well, the model is almost completely unwilling to follow instructions from `Perpetual` needed to perform step-by-step planning and implementation split by multiple files. It might be related to this issue: <https://github.com/ollama/ollama/issues/3760>

Thus, as of May 2024, it cannot be used at all.

I've used the model from Ollama repo: `starcoder2:15b-instruct-v0.1-q6_K`

## Llama-3-11.5B-Instruct-Coder-v2

Tests performed at May 2024 with Ollama `0.1.38`.

Works slightly better for planning and reasonings, than `CodeGemma-1.1-7B-it`, hovewer, still not good enough most of the times. Using low temperature like 0.3-0.5, `OLLAMA_REPEAT_PENALTY="1.0"` and `OLLAMA_TOP_K="20"` may produce better results. Implementing code at multiple files at once most likely produce poor code.

Generating file annotations with `annotate` operation gives acceptable results.

I've used the model from here: <https://huggingface.co/bartowski/Llama-3-11.5B-Instruct-Coder-v2-GGUF>

I've manually increased context num_ctx, because original value is too small when working with multiple files at once.

```sh
FROM Llama-3-11.5B-Instruct-Coder-v2-Q6_K.gguf
PARAMETER temperature 0.5
PARAMETER num_ctx 32768
PARAMETER num_predict 4096
PARAMETER repeat_penalty 1.0
PARAMETER penalize_newline false
SYSTEM You are a highly skilled software developer. You always write concise and readable code. You do not overload the user with unnecessary details in your answers and answer only the question asked. You are not adding separate explanations after code-blocks, you adding comments within your code instead.
TEMPLATE """{{ if .System }}<|start_header_id|>system<|end_header_id|>

{{ .System }}<|eot_id|>{{ end }}{{ if .Prompt }}<|start_header_id|>user<|end_header_id|>

{{ .Prompt }}<|eot_id|>{{ end }}<|start_header_id|>assistant<|end_header_id|>

{{ .Response }}<|eot_id|>
"""
PARAMETER stop "<|start_header_id|>"
PARAMETER stop "<|end_header_id|>"
PARAMETER stop "<|eot_id|>"
```

## DeepDeek-coder-33b-instruct-iMat

Tests performed at May 2024 with Ollama `0.1.38`.

Works better than `Llama-3-11.5B`. It provides acceptable reasoning, planning, and decent coding when using the following parameters:

```sh
OLLAMA_TEMPERATURE="0.5"
OLLAMA_TOP_K="20"
OLLAMA_REPEAT_PENALTY_OP_ANNOTATE="1.2"
OLLAMA_REPEAT_PENALTY_OP_IMPLEMENT_STAGE1="1.0"
OLLAMA_REPEAT_PENALTY_OP_IMPLEMENT_STAGE2="1.0"
OLLAMA_REPEAT_PENALTY_OP_IMPLEMENT_STAGE3="1.0"
OLLAMA_REPEAT_PENALTY="1.0"
```

I've used the model from here: <https://huggingface.co/dranger003/deepseek-coder-33b-instruct-iMat.GGUF>. According to the discussion, it provides better quantization than other deepseek models out there. I used the following example `Modelfile`. I've manually increased context num_ctx, because original value is too small when working with multiple files at once (maybe this change is incorrect).

```sh
FROM ggml-deepseek-coder-33b-instruct-q4_k_m.gguf
PARAMETER temperature 0.5
PARAMETER num_ctx 32768
PARAMETER num_predict 4096
PARAMETER repeat_penalty 1.0
PARAMETER penalize_newline false
SYSTEM You are a highly skilled software developer. You always write concise and readable code. You do not overload the user with unnecessary details in your answers and answer only the question asked. You are not adding separate explanations after code-blocks, you adding comments within your code instead.
TEMPLATE """{{ .System }}
### Instruction:
{{ .Prompt }}
### Response:
"""
PARAMETER stop """
### Instruction:"""
```
