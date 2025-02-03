# Trying Local LLMs with Ollama

In this document, I collect my subjective tests of local LLMs with `Perpetual`. Your results may vary.

**My test machine specs:** AMD Ryzen 9 7945HX, 32GB of RAM, Nvidia RTX 4070 Laptop GPU with 8GB of VRAM limited to 70W of power. I consider this as a bare minimum for working with local LLMs for now.

## Ollama Setup

To maximize performance and minimize VRAM usage, set the following environment variables for the `ollama serve` process:

```sh
OLLAMA_FLASH_ATTENTION="1"
OLLAMA_KV_CACHE_TYPE=q8_0
```

This configuration sets KV cache quantization to 8-bit, allowing larger context window sizes with minimal quality loss. It is recommended for GPUs with a low amount of VRAM.

## Models

It is possible to use models from the Ollama repository to cover some `Perpetual` operations.

### Qwen2.5-Coder

The following models seem to work with `Perpetual`:

- **`qwen2.5-coder:7b-instruct`**: Can be used with the `annotate` operation more or less reliably (more reliable with multi-step/multi-try annotate mode), depending on the target programming language. It can fully fit in 8GB of VRAM, providing decent performance. This is a working alternative for commercial LLMs like Claude 3 Haiku or GPT-4-mini for the `annotate` operation only. Coding tasks with the `implement` operation are pretty unreliable.

- **`qwen2.5-coder:14b-instruct-q4_K_M`**: Can be used with the `annotate` operation more reliably than the 7b model and can handle small and most trivial coding tasks with the `implement` operation. It can only partially fit in 8GB of VRAM, leading to a significant performance drop. It is barely usable on a GPU with 8GB of VRAM.

- **`qwen2.5-coder:32b-instruct-q4_K_M`**: Can write more complicated code with the `implement` operation and is more reliable. However, it is practically unusable on a GPU with 8GB of VRAM.

If using these models with the `implement` operation, you can reliably use JSON mode for Stage 1 and Stage 3.
