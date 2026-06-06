# Embed Operation

The `embed` operation generates vector embeddings for your project's source files, enabling local semantic search and similarity queries. By converting file contents into numerical vectors, `embed` allows Perpetual to find files related by meaning rather than just name or pattern, improving search relevance for other operations.

Generated embeddings are stored in `.perpetual/.embeddings.msgpack`.

## Usage

The `embed` operation is optional and will only function when an embedding model is configured via environment variables in your `.env` file. It is supported with OpenAI, Ollama, and Generic providers, depending on the specific provider/model capabilities. Anthropic does not support embeddings.

```sh
Perpetual embed [flags]
```

The `embed` operation has two primary modes:

1. **Embedding Generation Mode (default):** Processes your project files, detects changes, and regenerates embeddings as needed.

2. **Question/Search Mode:** Updates embeddings as needed and then performs semantic search to find files relevant to a specific question or query.

When invoked without question flags, `embed` processes your project files. It is also called internally by other operations, such as `doc`, `explain`, and `implement`, to keep embeddings updated and to complement LLM-driven file selection with local similarity search when annotation updates and embeddings are enabled. When properly set up, you generally do not need to run the `embed` operation manually except to force a rebuild, test configuration, or run standalone semantic search.

Available flags:

- `-f`  
  Force regeneration of all embeddings, even if up to date. Useful when you change embedding parameters in your `.env`.

- `-d`  
  Dry run: list files that **would** be processed without generating embeddings.

- `-r <file>`  
  Generate embeddings for a single file, even if its embedding already exists. The file path must point to a file inside the project and must pass the project whitelist/blacklist filters. If a user filter is supplied with `-x`, the requested file is still subject to that filter.

- `-q`  
  Read a question from stdin and find files relevant to it using semantic search.

- `-i <file>`  
  Read a question from a file, in plain text or markdown format, and find relevant files. Implies `-q`.

- `-s <limit>`  
  Limit the number of files returned that are relevant to the question (default: 5). Used with `-q` or `-i`. A value of `0` disables similarity search output.

- `-u`  
  In question/search mode, do not exclude files matching the project test-file blacklist. This flag has no additional effect in default embedding generation mode.

- `-x <file>`  
  Path to a JSON file containing an array of regex filters used to exclude files. In embedding generation mode, matching files are skipped for embedding. In question/search mode, matching files are skipped both from embedding updates and from search results.

- `-v`  
  Enable debug logging.

- `-vv`  
  Enable both debug and trace logging.

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

3. **Embed only a single file:**

   ```sh
   Perpetual embed -r cmd/main.go
   ```

4. **Search for files related to a question from stdin:**

   ```sh
   echo "How does authentication work?" | Perpetual embed -q
   ```

5. **Search for files related to a question in a file:**

   ```sh
   Perpetual embed -i question.txt -s 3
   ```

6. **Exclude files via a user-supplied filter:**

   ```sh
   Perpetual embed -x filters/skip_patterns.json
   ```

## LLM Configuration

To enable embeddings, set the appropriate model and parameters in your `.perpetual/.env` or global `.env` file. Embedding is supported for OpenAI, Ollama, and Generic providers, depending on the selected model and API endpoint. Anthropic does not support embeddings.

Standalone `embed` runs fail if no usable embedding provider/model is configured. Internal calls from other operations silently skip embedding updates when embeddings are unavailable.

### Key Environment Variables

- **Provider Selection:**  
  `LLM_PROVIDER_OP_EMBED` (fallback to `LLM_PROVIDER`)

  Numbered provider profiles are supported in the same way as other operations, for example `LLM_PROVIDER_OP_EMBED="ollama1"` with variables using the `OLLAMA1_` prefix.

- **Embedding Model:**  
  `<PROVIDER>_MODEL_OP_EMBED`

  For the `embed` operation, the operation-specific model variable must be set for OpenAI, Ollama, and Generic providers. The generic `<PROVIDER>_MODEL` fallback is not used for embedding mode.

- **Document Chunking:**  
  `<PROVIDER>_EMBED_DOC_CHUNK_SIZE` (default: 1024)  
  `<PROVIDER>_EMBED_DOC_CHUNK_OVERLAP` (default: 64)

- **Search Chunking:**  
  `<PROVIDER>_EMBED_SEARCH_CHUNK_SIZE` (default: 4096)  
  `<PROVIDER>_EMBED_SEARCH_CHUNK_OVERLAP` (default: 128)

- **Score Threshold:**  
  `<PROVIDER>_EMBED_SCORE_THRESHOLD` (default: 0.0)

- **Embedding Dimensions:**  
  `OPENAI_EMBED_DIMENSIONS`  
  `GENERIC_EMBED_DIMENSIONS`  
  `OLLAMA_EMBED_DIMENSIONS`

  Dimension overrides are model/provider dependent and may not be supported by all embedding models.

- **Retries:**  
  `<PROVIDER>_ON_FAIL_RETRIES_OP_EMBED` (fallback to `<PROVIDER>_ON_FAIL_RETRIES`, default: 3)

- **Generic Provider API Version:**  
  `GENERIC_API_VERSION_OP_EMBED` (fallback to `GENERIC_API_VERSION`)

  This is appended as the `api-version` URL query parameter and is useful for some OpenAI-compatible providers, such as Azure endpoints.

- **OpenAI Service Tier:**  
  `OPENAI_SERVICE_TIER_OP_EMBED` (fallback to `OPENAI_SERVICE_TIER`)  
  `OPENAI_SERVICE_TIER_FALLBACK`

  If configured, service tier options are also applied to embedding requests. The fallback tier may be activated after eligible OpenAI rate-limit or server-side failures.

- **Generic Provider Prefixes:**  
  You can set `GENERIC_EMBED_DOC_PREFIX` and `GENERIC_EMBED_SEARCH_PREFIX` to prepend custom text to each document or search query before embedding. Some embedding models expect a specific prefix. Refer to the model's documentation or Hugging Face model card for recommended prefixes.

- **Ollama Prefixes:**  
  You can set `OLLAMA_EMBED_DOC_PREFIX` and `OLLAMA_EMBED_SEARCH_PREFIX` to prepend custom text to each document or search query before embedding. Some Ollama embedding models require this for good results. For example, `nomic-embed-text-v1.5` commonly uses `search_document:` and `search_query:` prefixes. `snowflake-arctic-embed2` does not require prefixes.

## Example Configurations in `.env` Files

```sh
# OpenAI embeddings
LLM_PROVIDER_OP_EMBED="openai"
OPENAI_API_KEY="<your api key goes here>"
OPENAI_BASE_URL="https://api.openai.com/v1"

OPENAI_MODEL_OP_EMBED="text-embedding-3-small"

# Document chunk size / overlap, in characters
OPENAI_EMBED_DOC_CHUNK_SIZE="1024"
OPENAI_EMBED_DOC_CHUNK_OVERLAP="64"

# Search query chunk size / overlap, in characters
OPENAI_EMBED_SEARCH_CHUNK_SIZE="4096"
OPENAI_EMBED_SEARCH_CHUNK_OVERLAP="128"

# Set dimension count of generated vectors, optional and model-dependent
OPENAI_EMBED_DIMENSIONS="1536"

# Cosine similarity threshold
OPENAI_EMBED_SCORE_THRESHOLD="0.0"
```

```sh
# Ollama embeddings
LLM_PROVIDER_OP_EMBED="ollama"
OLLAMA_BASE_URL="http://127.0.0.1:11434"

OLLAMA_MODEL_OP_EMBED="nomic-embed-text-v1.5"

OLLAMA_EMBED_DOC_CHUNK_SIZE="1024"
OLLAMA_EMBED_DOC_CHUNK_OVERLAP="64"
OLLAMA_EMBED_SEARCH_CHUNK_SIZE="1024"
OLLAMA_EMBED_SEARCH_CHUNK_OVERLAP="64"

# Optional and model-dependent
OLLAMA_EMBED_DIMENSIONS="1024"

OLLAMA_EMBED_SCORE_THRESHOLD="0.0"

# Optional Ollama prefixes for models that require them
OLLAMA_EMBED_DOC_PREFIX="search_document:\n"
OLLAMA_EMBED_SEARCH_PREFIX="search_query:\n"
```

```sh
# Generic OpenAI-compatible provider embeddings
LLM_PROVIDER_OP_EMBED="generic"

GENERIC_BASE_URL="https://your-provider.example.com/v1"
GENERIC_AUTH_TYPE="Bearer"
GENERIC_AUTH="<your api key or token goes here>"

GENERIC_MODEL_OP_EMBED="<your embedding model>"

GENERIC_EMBED_DOC_CHUNK_SIZE="1024"
GENERIC_EMBED_DOC_CHUNK_OVERLAP="64"
GENERIC_EMBED_SEARCH_CHUNK_SIZE="4096"
GENERIC_EMBED_SEARCH_CHUNK_OVERLAP="128"

# Optional and provider/model-dependent
GENERIC_EMBED_DIMENSIONS="1024"

GENERIC_EMBED_SCORE_THRESHOLD="0.0"

# Optional model-specific prefixes
GENERIC_EMBED_DOC_PREFIX="Process following document:\n"
GENERIC_EMBED_SEARCH_PREFIX="Process following search query:\n"
```

## Workflow

### Embedding Generation Mode

1. **Project Discovery**  
   Locate the project root and Perpetual directory (`.perpetual`).

2. **Configuration Loading**  
   Load `.env` files, `project.json`, and create the embedding connector. If the selected provider/model does not support embeddings, a standalone run fails.

3. **File Listing & Filtering**  
   Gather project files and apply project whitelist/blacklist rules from `project.json`.

4. **Checksum Calculation**  
   Compute SHA-256 checksums for files to detect content changes.

5. **Load Existing Embeddings**  
   Read stored embeddings and checksums from `.perpetual/.embeddings.msgpack`.

6. **Determine Files to Embed**  
   - With `-r`, select the specified file.  
   - With `-f`, remove the old embeddings storage and select all project files.  
   - Otherwise, select files whose checksums have changed or whose embeddings are missing.

7. **Apply User Filters**  
   Exclude files matching user-provided regex patterns from `-x`. For skipped files, old checksums are preserved where possible so they can be reconsidered on later runs.

8. **Dry Run (optional)**  
   If `-d` is specified, output the list of files to be embedded and exit.

9. **Generate Embeddings**  
   For each selected file:  
   - Read file content.  
   - Split it into chunks with overlap based on configuration.  
   - Call the configured embedding provider.  
   - Retry transient failures according to provider retry settings.  
   - Validate vector dimension consistency.

10. **Save Embeddings**  
    Update `.perpetual/.embeddings.msgpack` if any embeddings changed.

### Question/Search Mode

When using `-q` or `-i`, Perpetual still performs the embedding generation workflow first, so changed or missing embeddings are updated before searching.

1. **Read Question**  
   Load the question from stdin with `-q` or from a file with `-i`.

2. **Apply Search Filters**  
   Apply user filters from `-x`. If `-u` is not specified, also exclude files matching the project test-file blacklist.

3. **Update Needed Embeddings**  
   Generate or refresh embeddings for changed files that are still eligible after filtering.

4. **Generate Question Embeddings**  
   Create embeddings for the input question using the configured embedding provider. Search query embeddings are cached in memory during the current process to avoid recomputation.

5. **Load Project Embeddings**  
   Read project file embeddings from `.perpetual/.embeddings.msgpack`.

6. **Perform Similarity Search**  
   Calculate cosine similarity between the question embedding and stored project file embeddings. If a file has multiple vectors, the best score for that file is used.

7. **Return Results**  
   Print selected matching files, one per line, limited by `-s` and filtered by the configured similarity threshold.

### Internal Use for Local Search

Other operations use embeddings for local similarity search in addition to LLM-based file selection. This is used to improve relevance and reduce context pressure, especially when context saving is enabled.

Internal searches may use more than just a direct query. Depending on the operation, Perpetual can also compose search queries from task text, target-file annotations, or generated task summaries. These searches use the same embedding storage and provider configuration as the standalone `embed` operation.

## Best Practices

- **Preferred Models:**  
  - **Ollama:**  
    `snowflake-arctic-embed2` is recommended for high-quality local embeddings.  
  - **OpenAI:**  
    `text-embedding-3-small` is a practical minimum recommended model. Adjust `OPENAI_EMBED_DIMENSIONS` if RAM or disk usage is a concern.

- **Chunk Settings:**  
  Default values are:

  ```sh
  <PROVIDER>_EMBED_DOC_CHUNK_SIZE=1024
  <PROVIDER>_EMBED_DOC_CHUNK_OVERLAP=64
  <PROVIDER>_EMBED_SEARCH_CHUNK_SIZE=4096
  <PROVIDER>_EMBED_SEARCH_CHUNK_OVERLAP=128
  ```

  Smaller overlaps reduce redundant vectors; larger overlaps can improve continuity at chunk boundaries. These defaults work well for general use with models such as `text-embedding-3-small` and `snowflake-arctic-embed2`. For some Ollama models, you may need to lower `SEARCH_CHUNK_SIZE` to avoid context-window overflow or model errors.

- **Resource Optimization:**  
  For very large projects, consider increasing `DOC_CHUNK_SIZE` to reduce the number of generated vectors, trading off granularity for performance. You can also experiment with higher `SEARCH_CHUNK_SIZE` for large queries if your provider supports it without errors.

- **Model-Specific Prefixes:**  
  Some embedding models require specific prefixes for optimal performance. Check your model's documentation for recommended prefixes. Examples:

  - `nomic-embed-text-v1.5`: use `search_document:` and `search_query:` prefixes.
  - `mxbai-embed-large-v1`: no document prefix is usually needed, but search queries can use `Represent this sentence for searching relevant passages:\n`.
  - `qwen3-embedding-8b`: search queries can use `Instruct: retrieve code fragments relevant to the query\nQuery:\n`.

- **Sensitive Files:**  
  The `embed` operation sends file contents to the configured embedding provider. It does not use `no-upload` source comments as a filter. Use project whitelist/blacklist rules or the `-x` user filter file to exclude files that must not be embedded.
