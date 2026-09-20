package redis

import (
	"chat-app/internal/core/entities"
	"context"
	"fmt"
	redisgo "github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"sync"
	"time"
)

const indexName = "semantic_cache_idx"

type Cache struct {
	client    *redisgo.Client
	dimension int
	mu        sync.Mutex
}

func New(address, password string, dimension int) *Cache {
	return &Cache{client: redisgo.NewClient(&redisgo.Options{Addr: address, Password: password, PoolSize: 10}), dimension: dimension}
}
func (c *Cache) Close() error { return c.client.Close() }
func (c *Cache) ensureIndex(ctx context.Context, dimension int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := validateDimension(c.dimension, dimension); err != nil {
		return err
	}
	if c.dimension == 0 {
		c.dimension = dimension
	}
	err := c.client.Do(ctx, "FT.CREATE", indexName, "ON", "HASH", "PREFIX", "1", "semantic_cache:", "SCHEMA", "query", "TEXT", "response", "TEXT", "embedding", "VECTOR", "HNSW", "6", "TYPE", "FLOAT32", "DIM", c.dimension, "DISTANCE_METRIC", "COSINE").Err()
	if err != nil && !strings.Contains(err.Error(), "Index already exists") {
		return fmt.Errorf("create vector index: %w", err)
	}
	return nil
}
func (c *Cache) Find(ctx context.Context, vector []float32) (*entities.CacheEntry, error) {
	if err := c.ensureIndex(ctx, len(vector)); err != nil {
		return nil, err
	}
	raw, err := c.client.Do(ctx, "FT.SEARCH", indexName, "*=>[KNN 1 @embedding $vector AS distance]", "PARAMS", "2", "vector", VectorBytes(vector), "SORTBY", "distance", "RETURN", "8", "response", "llm_model", "embedding_model", "prompt_version", "created_at", "expires_at", "distance", "query", "DIALECT", "2").Result()
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}
	m, found, err := searchFields(raw)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	expires, err := time.Parse(time.RFC3339Nano, m["expires_at"])
	if err != nil {
		return nil, fmt.Errorf("parse expiration: %w", err)
	}
	distance, _ := strconv.ParseFloat(m["distance"], 64)
	return &entities.CacheEntry{Query: m["query"], Response: m["response"], LLMModel: m["llm_model"], EmbeddingModel: m["embedding_model"], PromptVersion: m["prompt_version"], CreatedAt: parseTime(m["created_at"]), ExpiresAt: expires, Similarity: Similarity(distance)}, nil
}
func (c *Cache) Store(ctx context.Context, e entities.CacheEntry) error {
	if err := c.ensureIndex(ctx, len(e.Embedding)); err != nil {
		return err
	}
	key := "semantic_cache:" + strconv.FormatInt(e.CreatedAt.UnixNano(), 10)
	ttl := time.Until(e.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("cache entry is already expired")
	}
	err := c.client.HSet(ctx, key, map[string]any{"query": e.Query, "embedding": VectorBytes(e.Embedding), "response": e.Response, "llm_model": e.LLMModel, "embedding_model": e.EmbeddingModel, "prompt_version": e.PromptVersion, "created_at": e.CreatedAt.UTC().Format(time.RFC3339Nano), "expires_at": e.ExpiresAt.UTC().Format(time.RFC3339Nano)}).Err()
	if err != nil {
		return fmt.Errorf("store hash: %w", err)
	}
	if err := c.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("set cache ttl: %w", err)
	}
	return nil
}
func asInt(v interface{}) int {
	switch n := v.(type) {
	case int64:
		return int(n)
	case int:
		return n
	}
	return 0
}

// searchFields decodes FT.SEARCH replies returned in either RESP2 or RESP3.
// RediSearch returns an array in RESP2 and a results map in RESP3 (the
// go-redis default), with document fields under extra_attributes.
func searchFields(raw interface{}) (map[string]string, bool, error) {
	switch result := raw.(type) {
	case []interface{}:
		if len(result) < 3 || asInt(result[0]) == 0 {
			return nil, false, nil
		}
		fields, ok := result[2].([]interface{})
		if !ok {
			return nil, false, fmt.Errorf("unexpected Redis RESP2 search result")
		}
		return mapFields(fields), true, nil
	case map[interface{}]interface{}:
		return resp3SearchFields(result)
	case map[string]interface{}:
		generic := make(map[interface{}]interface{}, len(result))
		for key, value := range result {
			generic[key] = value
		}
		return resp3SearchFields(generic)
	default:
		return nil, false, fmt.Errorf("unexpected Redis search result type %T", raw)
	}
}

func resp3SearchFields(result map[interface{}]interface{}) (map[string]string, bool, error) {
	results, ok := result["results"].([]interface{})
	if !ok || len(results) == 0 {
		return nil, false, nil
	}
	document, ok := results[0].(map[interface{}]interface{})
	if !ok {
		return nil, false, fmt.Errorf("unexpected Redis RESP3 document result")
	}
	fields, ok := document["extra_attributes"]
	if !ok {
		return nil, false, fmt.Errorf("Redis RESP3 document has no fields")
	}
	return mapValues(fields), true, nil
}

func mapFields(v []interface{}) map[string]string {
	m := map[string]string{}
	for i := 0; i+1 < len(v); i += 2 {
		m[stringValue(v[i])] = stringValue(v[i+1])
	}
	return m
}

func mapValues(v interface{}) map[string]string {
	m := map[string]string{}
	switch values := v.(type) {
	case map[interface{}]interface{}:
		for key, value := range values {
			m[stringValue(key)] = stringValue(value)
		}
	case map[string]interface{}:
		for key, value := range values {
			m[key] = stringValue(value)
		}
	}
	return m
}

func stringValue(v interface{}) string {
	if bytes, ok := v.([]byte); ok {
		return string(bytes)
	}
	return fmt.Sprint(v)
}
func parseTime(v string) time.Time { t, _ := time.Parse(time.RFC3339Nano, v); return t }
