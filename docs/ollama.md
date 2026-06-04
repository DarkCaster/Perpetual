# Trying Local LLMs with Ollama

In this document, I collect my subjective tests of local LLMs with `Perpetual`. Your results may vary.

**My test machine specs:** AMD Ryzen 9 7945HX, 32GB of RAM, Nvidia RTX 4070 Laptop GPU with 8GB of VRAM. Consider this as a minimum for working with local LLMs for now.

## Ollama Setup

To maximize performance and minimize VRAM usage, set the following environment variables for the `ollama serve` process:

```sh
OLLAMA_FLASH_ATTENTION="1"
OLLAMA_KV_CACHE_TYPE="q8_0"
```

This configuration sets KV cache quantization to 8-bit, allowing larger context window sizes with some quality loss. It is recommended for GPUs with a low amount of VRAM.

For models that support Ollama thinking/reasoning controls, `Perpetual` can use `OLLAMA_THINK` or per-operation variables such as `OLLAMA_THINK_OP_ANNOTATE`. Supported values are `true`, `false`, and, with newer Ollama versions and supported models, `low`, `medium`, and `high`.

## Models

It is possible to use models from the Ollama repository to cover some `Perpetual` operations.

### gpt-oss:20b

Download with:

```sh
ollama pull gpt-oss:20b
```

The 20b gpt-oss model can be used for generating annotations, providing results good enough for use in other stages and operations. You need to set the environment variable `OLLAMA_THINK_OP_ANNOTATE` or `OLLAMA_THINK` to `low` in order to fit annotations in a smaller context. It may be used for stage 1 of other operations in `plain` non-JSON mode with small projects, but results may be unstable. Consider using the `-sp` flag for multi-step file selection.

### qwen3:8b / qwen3:14b

Download with:

```sh
ollama pull qwen3:8b
or
ollama pull qwen3:14b
```

Both models can be used for generating annotations. You should disable thinking by setting the `OLLAMA_THINK_OP_ANNOTATE` or `OLLAMA_THINK` environment variable to `false`. If using the smaller 8b model, expect lower-quality annotations and prefer using it for smaller projects. The 14b model can be used for stage 1 of other operations on smaller projects; however, results are unstable.

### qwen3:30b

This is a MoE Qwen 3 model. It works well enough for generating annotations, providing similar or slightly worse results than qwen3:14b.

**NOTE: There are multiple different models available in the Ollama library with the same name.** Consider using `qwen3:30b-a3b-q4_K_M`; this is an older but more universal variant that can be used with or without thinking. Other models are untested, but should work.

### snowflake-arctic-embed2

This is an embeddings model that can only be used for local similarity search. Enable it by setting the `OLLAMA_MODEL_OP_EMBED="snowflake-arctic-embed2"` and `LLM_PROVIDER_OP_EMBED="ollama"` environment variables. Generated embeddings are good enough for production use with projects of any size.

Download with:

```sh
ollama pull snowflake-arctic-embed2
```
