# Embed Operation

The `embed` operation generates vector embeddings for your project’s source files, enabling local semantic search and similarity queries. By converting file contents into numerical vectors, `embed` allows Perpetual to find files related by meaning rather than just name or pattern, improving search of the relevant code for other operations.

## Usage

The `embed` operation is optional and will only function when an embedding model is configured via environment variables in your `.env` file. It is not supported with the Anthropic provider.

```sh
Perpetual embed [flags]
```

When invoked, `embed` processes your project files, detects changes, and regenerates embeddings as needed. It is also called internally by other operations (such as `doc`, `explain` and `implement`) to complement LLM-driven file selection with local similarity search using the project’s existing embeddings. When properly setup, it is generally not needed to run `embed` operation manually.

Available flags:

- `-f`: Force regeneration of all embeddings, even if up to date. Useful when you changed parameters for `embed` operation in your `.env`

- `-d`: Dry run: list files that **would** be processed without performing embedding.

- `-r <file>`: Generate embeddings for a single file, even if its embedding exists (implies `-f`).

- `-x <file>`: Path to a JSON file containing regex filters to exclude files from embedding.

- `-v`: Enable debug logging (Debug level).

- `-vv`: Enable both debug and trace logging (Debug + Trace levels).

- `-h`: Show help message and exit.

## Examples

1. **Dry run to see which files would be embedded:**

   ```sh
   Perpetual embed -d
   ```

2. **Force regenerate embeddings for the entire project:**

   ```sh
   Perpetual embed -f
   ```

3. **Embed only a single file (e.g., `main.go`):**

   ```sh
   Perpetual embed -r cmd/main.go
   ```

4. **Exclude tests and generated files via user-supplied filter:**

   ```sh
   Perpetual embed -x filters/skip_patterns.json
   ```

## LLM Configuration

To enable embeddings, set the appropriate model and parameters in your `.perpetual/.env` or global `.env` file. Embedding is supported for OpenAI, Ollama, and Generic providers; Anthropic does not support embeddings.

### Key Environment Variables

- **Provider Selection:**
  - `LLM_PROVIDER_OP_EMBED` (fallback to `LLM_PROVIDER`)

- **Model:**
  - `OPENAI_MODEL_OP_EMBED`
  - `OLLAMA_MODEL_OP_EMBED`
  - `GENERIC_MODEL_OP_EMBED`

- **Document Chunking:**
  - `<PROVIDER>_EMBED_DOC_CHUNK_SIZE` (default: 1024)
  - `<PROVIDER>_EMBED_DOC_CHUNK_OVERLAP` (default: 64)

- **Search Chunking:**
  - `<PROVIDER>_EMBED_SEARCH_CHUNK_SIZE` (default: 4096)
  - `<PROVIDER>_EMBED_SEARCH_CHUNK_OVERLAP` (default: 128)

- **Score Threshold:**
  - `<PROVIDER>_EMBED_SCORE_THRESHOLD` (default: 0.0)

- **Embedding Dimensions (OpenAI only):**
  - `OPENAI_EMBED_DIMENSIONS`

- **Retries & Segments:**
  - `<PROVIDER>_ON_FAIL_RETRIES_OP_EMBED` (default: 3)

Notes on implementation: write at this section about ollama-specific params OLLAMA_EMBED_DOC_PREFIX and OLLAMA_EMBED_SEARCH_PREFIX from `.env.example`. write that user should look at hugginface for prefixes for corresponding embedding models.

## Example Configuration in `.env` File

```sh
# Use OpenAI embeddings
LLM_PROVIDER_OP_EMBED="openai"
OPENAI_MODEL_OP_EMBED="text-embedding-3-small"

# Document chunk size / overlap (in characters)
OPENAI_EMBED_DOC_CHUNK_SIZE="1024"
OPENAI_EMBED_DOC_CHUNK_OVERLAP="64"

# Search query chunk size / overlap (in characters)
OPENAI_EMBED_SEARCH_CHUNK_SIZE="4096"
OPENAI_EMBED_SEARCH_CHUNK_OVERLAP="128"

# Cosine similarity threshold
OPENAI_EMBED_SCORE_THRESHOLD="0.0"

# Retry on failure
OPENAI_ON_FAIL_RETRIES_OP_EMBED="3"
OPENAI_MAX_TOKENS_SEGMENTS="3"

# Or with Ollama:
LLM_PROVIDER_OP_EMBED="ollama"
OLLAMA_MODEL_OP_EMBED="snowflake-arctic-embed2"
OLLAMA_EMBED_DOC_CHUNK_SIZE="1024"
OLLAMA_EMBED_DOC_CHUNK_OVERLAP="64"
OLLAMA_EMBED_SEARCH_CHUNK_SIZE="4096"
OLLAMA_EMBED_SEARCH_CHUNK_OVERLAP="128"
OLLAMA_EMBED_SCORE_THRESHOLD="0.0"
```

## Workflow

1. **Project Discovery**  
   Locate project root and Perpetual directory (`.perpetual`).

2. **File Listing & Filtering**  
   Gather all project files, apply whitelist/blacklist and user-supplied filters.

3. **Checksum Calculation**  
   Compute SHA-256 checksums for files to detect changes.

4. **Load Existing Embeddings**  
   Read embeddings and checksums from storage.

5. **Determine Files to Embed**  
   - If `-r` or `-f`, select specified or all files.  
   - Otherwise, detect files whose contents have changed.

6. **Apply Blacklist Filters**  
   Exclude files matching user-provided regex filters, preserving their old checksums for next run.

7. **Dry Run (optional)**  
   If `-d`, output the list of files to be embedded and exit.

8. **Generate Embeddings**  
   For each file:  
   - Read file content.  
   - Chunk and overlap based on configuration.  
   - Call LLM connector to create embeddings (with retry logic).  
   - Validate vector dimensions.

9. **Save Embeddings**  
   Update the embeddings storage file and checksums if any embeddings changed.

10. **Internal Use for Local Search**  
    Other operations call `embed` internally to perform local similarity search, combining these embeddings with LLM-driven file selection to improve and complement search results.

## Best Practices

- **Preferred Models:**  
  When using Ollama, it is generally recommended to use the `snowflake-arctic-embed2` model for high-quality embeddings. It is good enough for produiction use, recommended.  
  For OpenAI `text-embedding-3-small` is a minimum recommended model. You can try lowering `OPENAI_EMBED_DIMENSIONS` down to `1024` or less if RAM usage is too big.

- **Chunk Settings:**  
  Default optimal values are:

  ```sh
  <PROVIDER>_EMBED_DOC_CHUNK_SIZE=1024
  <PROVIDER>_EMBED_DOC_CHUNK_OVERLAP=64
  <PROVIDER>_EMBED_SEARCH_CHUNK_SIZE=4096
  <PROVIDER>_EMBED_SEARCH_CHUNK_OVERLAP=128
  ```

  Smaller overlaps reduce redundant vectors; larger overlaps can improve continuity at chunk boundaries. Current `DOC_CHUNK_SIZE` and `DOC_CHUNK_OVERLAP` defaults are optimal for general use and for most embedding models. You can lower `SEARCH_CHUNK_SIZE` if experiencing context window overflow and/or crashes with Ollama provider when using non `snowflake-arctic-embed2` embedding models.

- **Resource Optimization:**  
  Adjust chunk sizes and embedding thresholds for very large repositories to balance performance and accuracy: you can try increasing `DOC_CHUNK_SIZE` for a big projects (500+ files) in order to lower RAM requirements when working on embeddings.
