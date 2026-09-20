# Semantic Cache

This repository contains a local semantic-cache demonstration implemented by the [`chat-app`](chat-app/). The service accepts a chat query, creates an embedding through an OpenAI-compatible API, and uses Redis Stack/RediSearch to find the closest cached query vector. For an eligible match, it returns the stored answer; otherwise, it requests a chat completion and stores the query, embedding, and answer for later reuse.

```
Request -> embedding -> Redis KNN lookup -> eligible hit: cached answer
                                         -> miss: chat completion -> store answer
```

The cache uses a cosine-distance HNSW vector index. A candidate is reusable only when it has not expired, its similarity meets `SEMANTIC_CACHE_THRESHOLD`, and its embedding model, LLM model, and `PROMPT_VERSION` match the active configuration.

## Benefits

- An eligible cache hit returns a previously generated answer instead of making a chat-completions request.
- Similarity matching can reuse an answer for a sufficiently similar query, rather than requiring an exact text match.
- Entries expire according to `SEMANTIC_CACHE_TTL`, and model or prompt-version changes prevent reuse of entries created for a different configuration.
- The service exposes cache request, hit, miss, embedding, lookup, LLM-duration, and store-error metrics for observing the cache path.

An embedding request and Redis lookup are still performed before a cache hit is determined. Actual latency, cost, and hit-rate benefits depend on the configured models, threshold, query mix, and running environment.

## Suitable use cases

The implementation is best suited to repeated or semantically similar, stable questions where reusing a prior answer is acceptable. Examples include conceptual questions and repeated help-style prompts.

Do not rely on this simple cache for requests whose answers must reflect fresh or user-specific data, such as current infrastructure state or incident investigation. Those requests should bypass the cache or use additional application-specific controls.

## Chat app: integration and validation surface

The [`chat-app`](chat-app/) is the runnable integration surface for the cache. Its Compose stack connects:

- the Go HTTP API;
- Redis Stack with RediSearch vector search;
- LiteLLM as the OpenAI-compatible gateway; and
- a reachable Ollama backend for the configured embedding and chat models.

The app's `POST /chat` response reports `answer`, `cache_hit`, and `similarity`. With the local stack running, [`chat-app/validate-semantic-cache.sh`](chat-app/validate-semantic-cache.sh) validates the end-to-end path: liveness and readiness endpoints, an initial cache miss and store, an identical-request cache hit with the same answer, and the expected request/hit/miss metric deltas. Run it against a fresh local cache when validating the miss-and-store path, because a semantic nearest-neighbor match can cause even a newly worded request to hit an existing entry.

See the [chat-app README](chat-app/README.md) for setup, configuration, API details, and development checks.

## Limitations and caveats

- This is a local demonstration service, not a complete production cache design. The HTTP API has no authentication in the provided Compose setup.
- Redis returns only the nearest vector candidate. The application does not search for another candidate when that one is expired, below threshold, or has mismatched model or prompt metadata.
- Cache scope is not partitioned by user, tenant, conversation, authorization context, or source-data version. Reuse is controlled only by expiry, similarity, embedding model, LLM model, and prompt version.
- The cache has no explicit invalidation interface; entries are removed by their Redis TTL. Changing model or prompt configuration prevents reuse but does not itself remove old entries.
- A Redis lookup failure bypasses the cache and still tries to generate an answer; a cache-store failure still returns the generated answer. An embedding failure stops the request before generation.
- Similarity is model- and threshold-dependent. A semantic match is not a correctness guarantee, so thresholds and cache eligibility need evaluation against representative prompts.
