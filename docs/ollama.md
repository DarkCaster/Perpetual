# Trying local LLMs with Ollama

In this document I collect my subjective tests of local LLMs, your results may vary.

## CodeGemma-1.1-7B-it

Tests performed at May 2024 with Ollama `0.1.38`.

It is possible to use this model for generating file annotations (`OLLAMA_MODEL_OP_ANNOTATE` param at `.env` file).

For other operations it doesn't seem to work very well, it just doesn't follow the instructions well enough.

Model can also generate decent code (sometimes) at final stage 3 of implement operation (`OLLAMA_MODEL_OP_IMPLEMENT_STAGE3`).

Hovewer, it is not good enough to perform proper review, planning and reasonings at stages 1 and 2 of implement operation.

I've tested the model from here: <https://huggingface.co/bartowski/codegemma-1.1-7b-it-GGUF>

To run it with Ollama, I used the following example `Modelfile` (I don't guarantee it's correct or optimal, especially the prompt template). Models with less quantization should also work, but are less likely to succeed:

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

I'e tried the model from Ollama repo: `starcoder2:15b-instruct-v0.1-q6_K`
