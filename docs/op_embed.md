# Embed Operation

The `embed` operation generates vector embeddings for your project’s source files, enabling local semantic search and similarity queries. By converting file contents into numerical vectors, `embed` allows Perpetual to find files related by meaning rather than just name or pattern, improving search relevance for other operations.

## Usage

The `embed` operation is optional and will only function when an embedding model is configured via environment variables in your `.env` file. It is not supported with the Anthropic provider.

```sh
Perpetual embed [flags]
```

When invoked, `embed` processes your project files, detects changes, and regenerates embeddings as needed. It is also called internally by other operations (such as `doc`, `explain`, and `implement`) to complement LLM-driven file selection with local similarity search using the project’s existing embeddings. When properly set up, you generally do not need to run the `embed` operation manually.

Available flags:

- `-f`  
  Force regeneration of all embeddings, even if up to date. Useful when you change embedding parameters in your `.env`.

- `-d`  
  Dry run: list files that **would** be processed without performing embeddings.

- `-r <file>`  
  Generate embeddings for a single file, even if its embedding already exists (implies `-f`).

- `-x <file>`  
  Path to a JSON file containing regex filters to exclude files from embedding.

- `-v`  
  Enable debug logging (Debug level).

- `-vv`  
  Enable both debug and trace logging (Debug + Trace levels).

- `-h`  
  Show help message and exit.

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
  `LLM_PROVIDER_OP_EMBED` (fallback to `LLM_PROVIDER`)

- **Model:**  
  `<PROVIDER>_MODEL_OP_EMBED`

- **Document Chunking:**  
  `<PROVIDER>_EMBED_DOC_CHUNK_SIZE` (default: 1024)  
  `<PROVIDER>_EMBED_DOC_CHUNK_OVERLAP` (default: 64)

- **Search Chunking:**  
  `<PROVIDER>_EMBED_SEARCH_CHUNK_SIZE` (default: 4096)  
  `<PROVIDER>_EMBED_SEARCH_CHUNK_OVERLAP` (default: 128)

- **Score Threshold:**  
  `<PROVIDER>_EMBED_SCORE_THRESHOLD` (default: 0.0)

- **Embedding Dimensions (OpenAI only):**  
  `OPENAI_EMBED_DIMENSIONS`

- **Retries:**  
  `<PROVIDER>_ON_FAIL_RETRIES_OP_EMBED` (default: 3)

- **Ollama Prefixes:**  
  You can optionally set `OLLAMA_EMBED_DOC_PREFIX` and `OLLAMA_EMBED_SEARCH_PREFIX` to prepend custom text to each document or search query before embedding. Some Ollama embedding models (e.g., `nomic-embed-text-v1.5`) may expect a specific prompt prefix. Refer to the model’s documentation or its Hugging Face model card for recommended prefixes. **NOTE:** `snowflake-arctic-embed2` does not require any prefixes to be set.

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
OLLAMA_MODEL_OP_EMBED="nomic-embed-text-v1.5"
OLLAMA_EMBED_DOC_CHUNK_SIZE="1024"
OLLAMA_EMBED_DOC_CHUNK_OVERLAP="64"
OLLAMA_EMBED_SEARCH_CHUNK_SIZE="1024"
OLLAMA_EMBED_SEARCH_CHUNK_OVERLAP="64"
OLLAMA_EMBED_SCORE_THRESHOLD="0.0"

# Optional Ollama prefixes
OLLAMA_EMBED_DOC_PREFIX="search_document: \n"
OLLAMA_EMBED_SEARCH_PREFIX="search_query: \n"
```

## Workflow

1. **Project Discovery**  
   Locate the project root and Perpetual directory (`.perpetual`).

2. **File Listing & Filtering**  
   Gather all project files, apply built-in whitelist/blacklist rules and any user-supplied filters.

3. **Checksum Calculation**  
   Compute SHA-256 checksums for files to detect content changes.

4. **Load Existing Embeddings**  
   Read stored embeddings and checksums from the `.embeddings.msgpack` file.

5. **Determine Files to Embed**  
   - With `-r`, select the specified file.  
   - With `-f`, select all project files.  
   - Otherwise, detect files whose checksums have changed.

6. **Apply Blacklist Filters**  
   Exclude files matching user-provided regex patterns, but preserve their old checksums for future runs.

7. **Dry Run (optional)**  
   If `-d`, output the list of files to be embedded and exit.

8. **Generate Embeddings**  
   For each file:  
   - Read file content.  
   - Split into chunks with overlap based on configuration.  
   - Call the LLM provider to create embeddings (with retry logic).  
   - Validate vector dimensions consistency.

9. **Save Embeddings**  
   Update the embeddings storage file and checksums if any embeddings changed.

10. **Internal Use for Local Search**  
    Other operations invoke `embed` internally to perform local similarity search, combining these embeddings with LLM-driven file selection for improved relevance.

## Best Practices

- **Preferred Models:**  
  - **Ollama:**  
    `snowflake-arctic-embed2` is recommended for high-quality local embeddings.  
  - **OpenAI:**  
    `text-embedding-3-small` is a minimum recommended model. Adjust `OPENAI_EMBED_DIMENSIONS` (e.g., to 1024) if RAM usage is a concern.

- **Chunk Settings:**  
  Default optimal values are:

  ```sh
  <PROVIDER>_EMBED_DOC_CHUNK_SIZE=1024
  <PROVIDER>_EMBED_DOC_CHUNK_OVERLAP=64
  <PROVIDER>_EMBED_SEARCH_CHUNK_SIZE=4096
  <PROVIDER>_EMBED_SEARCH_CHUNK_OVERLAP=128
  ```

  Smaller overlaps reduce redundant vectors; larger overlaps can improve continuity at chunk boundaries. These defaults work well for general use (for `text-embedding-3-small` and `snowflake-arctic-embed2`). For some Ollama models, you may need to lower `SEARCH_CHUNK_SIZE` to avoid context-window overflow/crashes.

- **Resource Optimization:**  
  For very large projects (500+ files), consider increasing `DOC_CHUNK_SIZE` to reduce the number of generated vectors (trading off granularity for performance). You can also experiment with higher `SEARCH_CHUNK_SIZE` for large queries if your provider supports it without errors.
