# Semantic Cache Chat App

A small Go service that demonstrates semantic caching without RAG, agents, databases, or an AI framework. It uses the OpenAI embeddings and chat-completions APIs to embed an incoming query, searches Redis Stack for the closest cached query vector, and either returns that cached answer or generates and stores a new answer.

```
Request -> OpenAI embedding -> Redis KNN lookup -> cache hit: return cached answer
                                             -> cache miss: OpenAI chat completion -> Redis store -> answer
```

Redis stores cache hashes under `semantic_cache:<id>`. The RediSearch HNSW index uses cosine distance; the API reports `similarity = 1 - distance`. A cached entry is eligible only when it is unexpired, meets `SEMANTIC_CACHE_THRESHOLD`, and has matching embedding model, LLM model, and `PROMPT_VERSION`.

## Local services

| Service | Local URL / address | Purpose |
| --- | --- | --- |
| Chat API | `http://localhost:8080` | HTTP API and Prometheus metrics |
| Redis Stack | `localhost:6379` | Redis and RediSearch vector index; password authentication is required |
| Redis Insight | `http://localhost:8001` | Redis Stack's browser UI |
| LiteLLM | API: internal `http://litellm:4000/v1`; dashboard: `http://localhost:4000/ui` | OpenAI-compatible gateway for embedding and chat-completions requests |
| Ollama | host `http://localhost:11434` | Local model backend for LiteLLM |

Compose keeps the application on the OpenAI-compatible contract only: `OPENAI_BASE_URL=http://litellm:4000/v1` and `OPENAI_API_KEY` equal to `LITELLM_MASTER_KEY`. LiteLLM translates the application model names to Ollama; the application has no Ollama-specific configuration.

## Start the stack

1. Install and start [Ollama](https://ollama.com/), then make both required models available before starting Compose:

   ```bash
   ollama pull nomic-embed-text
   ollama pull llama3.1:8b
   ```

   On Docker Desktop for macOS, the default `OLLAMA_API_BASE=http://host.docker.internal:11434` lets the LiteLLM container reach host-run Ollama. Verify that Ollama is running and listening on port `11434`. If that hostname is unavailable or Ollama is remote, set `OLLAMA_API_BASE` to a URL reachable from the container (for example, `http://<reachable-host>:11434`); ensure the Ollama listener and network policy permit that connection. Do not use `localhost` here: from LiteLLM it means the LiteLLM container.

2. Create a local configuration file and replace the Redis password, LiteLLM master key, and LiteLLM dashboard credential placeholders. `.env` is gitignored; do not commit it or paste credentials into documentation, commands, or logs.

   ```bash
   cp .env.example .env
   ```

     In `.env`, set the following values (the values below are deliberately placeholders, not usable credentials):

    ```dotenv
     REDIS_PASSWORD=<unique-local-redis-password>
      OLLAMA_API_BASE=http://host.docker.internal:11434
      LITELLM_MASTER_KEY=<unique-local-litellm-master-key>
      OPENAI_BASE_URL=http://localhost:4000/v1
      OPENAI_API_KEY=<same-value-as-LITELLM_MASTER_KEY>
      LITELLM_UI_USERNAME=<local-dashboard-username>
      LITELLM_UI_PASSWORD=<unique-local-dashboard-password>
    ```

       Compose passes the Redis password both to Redis (`--requirepass`) and to the chat application (`REDIS_PASSWORD`). `litellm-config.yaml` maps the application model names `nomic-embed-text` and `llama3.1:8b` to Ollama. Update that file and `OLLAMA_API_BASE` if using another reachable Ollama backend or model.

   ```dotenv
    REDIS_ADDR=redis:6379
     EMBEDDING_MODEL=nomic-embed-text
     LLM_MODEL=llama3.1:8b
   ```

       Compose overrides the OpenAI-compatible settings for the containerized API with `OPENAI_BASE_URL=http://litellm:4000/v1` and `OPENAI_API_KEY=$LITELLM_MASTER_KEY`. For a host-run API, use `REDIS_ADDR=localhost:6379`, `OPENAI_BASE_URL=http://localhost:4000/v1`, and the same `OPENAI_API_KEY`/`LITELLM_MASTER_KEY` value.

3. Build and run the local stack:

   ```bash
   docker compose up --build
   ```

    Stop it with `docker compose down`.

### LiteLLM dashboard authentication

Open [LiteLLM dashboard](http://localhost:4000/ui) and sign in with `LITELLM_UI_USERNAME` and `LITELLM_UI_PASSWORD` from your private `.env`. The dashboard is bound to `127.0.0.1` only and is not LAN-accessible by this Compose configuration. `LITELLM_MASTER_KEY` protects the proxy API and is intentionally separate from the dashboard password. Use unique, high-entropy values for both secrets; never use placeholders in a running stack or expose port 4000 beyond loopback without adding appropriate network and TLS controls.

### Redis authentication and Insight

Redis is intentionally not available without authentication. To check the instance locally, supply the password from your private `.env` only in your own terminal/session:

```bash
redis-cli --no-auth-warning -h localhost -p 6379 -a "$REDIS_PASSWORD" ping
# PONG
```

An unauthenticated `PING` is expected to be rejected with `NOAUTH Authentication required.`

Open [Redis Insight](http://localhost:8001) after Compose starts. Add the local database using host `localhost`, port `6379`, and the password from your local `.env`; do not share that credential. Use Insight to inspect the `semantic_cache:*` keys and `semantic_cache_idx` index.

## API reference

Set a base URL for the examples:

```bash
API_URL=http://localhost:8080
```

All JSON endpoints return `Content-Type: application/json`. There is no API authentication in this local demonstration service.

### `POST /chat`

Generates an answer or returns an eligible semantically similar cached answer.

```bash
curl --request POST "$API_URL/chat" \
  --header 'Content-Type: application/json' \
  --data '{"query":"How do I restart a Kubernetes pod?"}'
```

Request body:

```json
{"query":"How do I restart a Kubernetes pod?"}
```

Successful response — `200 OK`:

```json
{
  "answer": "...",
  "cache_hit": false,
  "similarity": 0
}
```

`answer` is generated text, so its value varies. A repeated identical query is expected to return `200 OK`, `cache_hit: true`, the same cached `answer`, and a numeric similarity (typically `1` for an exact repeated query). A paraphrase may hit or miss depending on its similarity and the configured threshold.

The endpoint returns the following errors:

| Status | Response | When |
| --- | --- | --- |
| `400 Bad Request` | `{"error":"query must be a non-empty JSON string"}` | The body is invalid JSON, has unknown fields, omits `query`, or `query` is empty/whitespace. Request bodies are limited to 1 MiB. |
| `502 Bad Gateway` | `{"error":"chat service unavailable"}` | The configured OpenAI endpoint embedding or generation request fails. The application logs include the endpoint and upstream HTTP status. |

### `GET /api/v1/ping`

Returns a service identity and a UTC timestamp.

```bash
curl --silent "$API_URL/api/v1/ping"
```

`200 OK` response (timestamp varies):

```json
{"message":"pong","service":"chat-app","timestamp":"2026-09-20T12:34:56.789Z"}
```

### `GET /system/version`

Returns the configured `VERSION`, which defaults to `dev`.

```bash
curl --silent "$API_URL/system/version"
```

`200 OK` response:

```json
{"version":"dev"}
```

### Health endpoints

Both health endpoints currently report the HTTP service state and return `200 OK` with a JSON status value.

```bash
curl --silent "$API_URL/system/health/live"
# {"status":"live"}

curl --silent "$API_URL/system/health/ready"
# {"status":"ready"}
```

### `GET /system/metrics`

Returns Prometheus text exposition format with `200 OK`.

```bash
curl --silent "$API_URL/system/metrics"
```

The response includes HTTP request counters/durations and semantic-cache metrics:

- `semantic_cache_requests_total`
- `semantic_cache_hits_total`
- `semantic_cache_misses_total`
- `semantic_cache_embedding_duration_seconds`
- `semantic_cache_lookup_duration_seconds`
- `semantic_cache_llm_duration_seconds`
- `semantic_cache_store_errors_total`

### API documentation endpoints

```bash
# Swagger UI (HTML, 200 OK)
open "$API_URL/docs"

# OpenAPI document (YAML, 200 OK)
curl --silent "$API_URL/docs/openapi.yaml"
```

The OpenAPI source in the repository is [`docs/openapi.yaml`](docs/openapi.yaml). Swagger UI loads that document from `/docs/openapi.yaml`.

## Configuration

`.env.example` lists all supported environment variables. Key settings are:

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | API listener port |
| `REDIS_ADDR` | `redis:6379` | Redis address |
| `REDIS_PASSWORD` | empty | Redis password; required by the Compose setup |
| `OPENAI_BASE_URL` | `https://api.openai.com/v1` | OpenAI-compatible API base URL; Compose sets `http://litellm:4000/v1` |
| `OPENAI_API_KEY` | empty | OpenAI-compatible API bearer token; Compose supplies `LITELLM_MASTER_KEY` |
| `EMBEDDING_MODEL` | `nomic-embed-text` in Compose | LiteLLM embedding model identifier |
| `LLM_MODEL` | `llama3.1:8b` in Compose | LiteLLM chat model identifier |
| `SEMANTIC_CACHE_ENABLED` | `true` | Enables embedding, lookup, and cache storage |
| `SEMANTIC_CACHE_THRESHOLD` | `0.92` | Minimum similarity required for a hit |
| `SEMANTIC_CACHE_TTL` | `1h` | Cache-entry lifetime; Compose configures `3600s` |
| `PROMPT_VERSION` | `v1` | Version that must match for a cache hit |
| `EMBEDDING_DIMENSION` | `0` | `0` discovers the first embedding dimension; a positive value validates the model dimension |

The application supports OpenAI-compatible APIs through `OPENAI_BASE_URL`, `OPENAI_API_KEY`, and model identifiers. In the Compose stack it uses LiteLLM, which routes those identifiers to Ollama. Tune the threshold using representative prompts and models. Dynamic or fresh-data requests (for example, current infrastructure state or incident investigation) should bypass this simple cache; stable conceptual questions are better candidates.

## Validate the semantic-cache path

With the stack running, Ollama reachable, and the required models pulled, run:

```bash
./validate-semantic-cache.sh "Explain the purpose of a cache hit."
```

The script requires Bash, `curl`, `jq`, `awk`, and standard local utilities (`cat`, `date`, `mktemp`, and `rm`); it does not require Python. Its only input is one non-empty chat message. It always calls the local API at `http://localhost:8080` with a 120-second request timeout. A unique textual identifier cannot guarantee a miss in a semantic nearest-neighbor cache; run it against a fresh local cache to validate the miss-and-store path. It emits newline-delimited JSON logs, including the JSON request and each JSON response. It performs the following checks:

1. `GET /system/health/live` returns `200` with `{"status":"live"}`.
2. `GET /system/health/ready` returns `200` with `{"status":"ready"}`.
3. A unique `POST /chat` request misses and is stored; the identical repeat hits and returns the same answer.
4. Semantic-cache metric deltas are at least `+2` requests, `+1` hit, and `+1` miss.

## Development checks

```bash
go fmt ./...
go vet ./...
go test ./...
```
