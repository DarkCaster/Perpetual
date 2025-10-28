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

## Models

It is possible to use models from the Ollama repository to cover some `Perpetual` operations.

### gpt-oss:20b

Download with:

```sh
ollama pull gpt-oss:20b
```

20b gpt-oss model can be used for generating annotations providing results good enough for use on other stages and operations, you need to set env. variable `OLLAMA_THINK_OP_ANNOTATE` or `OLLAMA_THINK` to `low` in order to fit annotations in smaller context. May be used for stage 1 of other operations in `plain` non-json mode with small projects, results may be unstable - consider to use `-sp` flag for multi-step file selection.

### qwen3:8b / qwen3:14b

Download with:

```sh
ollama pull qwen3:8b
or
ollama pull qwen3:14b
```

Both models can be used for generating annotations, you should disable thinking by setting `OLLAMA_THINK_OP_ANNOTATE` or `OLLAMA_THINK` env. variable to `false`. If using smaller 8b model, consider using multi-stage annotation generation for better results. 14b model can be used for stage 1 of other operations on smaller projects, hovewer, result are unstable.

### qwen3:30b

This is MOE qwen 3 model. Works good enough for generating annotations, providing similar (or slightly worse) result as qwen3:14b.

**NOTE: There are multiple different models available in ollama library with the same name.** Consider using `qwen3:30b-a3b-q4_K_M`, this is an older but more universal - it can be used with or without thinking. Other models untested, but should work.

### snowflake-arctic-embed2

This is an embeddings model that can only be used for local similarity search. Enable by setting `OLLAMA_MODEL_OP_EMBED="snowflake-arctic-embed2"` and `LLM_PROVIDER_OP_EMBED="ollama"` env. variables. Generated embeddings are good enough for production use with projects of any size.

Download with:

```sh
ollama pull snowflake-arctic-embed2
```
