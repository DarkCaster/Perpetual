# Trying local LLMs with Ollama

In this document I collect my subjective tests of local LLMs with `Perpetual`, your results may vary.

## CodeGemma-1.1-7B-it

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
