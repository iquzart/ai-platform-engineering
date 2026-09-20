package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"chat-app/internal/config"
	"chat-app/internal/core/entities"
	"chat-app/internal/core/repositories"
	"chat-app/internal/meta"
)

type Chat struct {
	embedder  repositories.Embedder
	generator repositories.Generator
	cache     repositories.SemanticCache
	cfg       *config.SemanticCacheConfigs
	logger    *slog.Logger
	metrics   *meta.Metrics
	now       func() time.Time
}
type ChatResult struct {
	Answer     string
	CacheHit   bool
	Similarity float64
}

func NewChat(e repositories.Embedder, g repositories.Generator, c repositories.SemanticCache, cfg *config.SemanticCacheConfigs, logger *slog.Logger, metrics *meta.Metrics) *Chat {
	return &Chat{e, g, c, cfg, logger, metrics, time.Now}
}
func (c *Chat) Ask(ctx context.Context, query string) (ChatResult, error) {
	c.metrics.SemanticRequests.Inc()
	c.logger.Info("request received")
	var embedding []float32
	if c.cfg.Enabled {
		started := time.Now()
		var err error
		embedding, err = c.embedder.Embed(ctx, query)
		c.metrics.EmbeddingDuration.Observe(time.Since(started).Seconds())
		if err != nil {
			return ChatResult{}, fmt.Errorf("generate embedding: %w", err)
		}
		c.logger.Info("embedding generated", "dimensions", len(embedding))
		started = time.Now()
		entry, err := c.cache.Find(ctx, embedding)
		c.metrics.LookupDuration.Observe(time.Since(started).Seconds())
		if err != nil {
			c.logger.Warn("cache lookup unavailable; bypassing cache", "error", err)
		} else if entry != nil && c.valid(entry) {
			c.metrics.SemanticHits.Inc()
			c.logger.Info("cache hit", "similarity", entry.Similarity)
			return ChatResult{entry.Response, true, entry.Similarity}, nil
		} else {
			c.metrics.SemanticMisses.Inc()
			c.logger.Info("cache miss")
		}
		c.logger.Info("cache lookup")
	}
	started := time.Now()
	c.logger.Info("LLM request")
	answer, err := c.generator.Generate(ctx, query)
	c.metrics.LLMDuration.Observe(time.Since(started).Seconds())
	if err != nil {
		return ChatResult{}, fmt.Errorf("generate response: %w", err)
	}
	if c.cfg.Enabled {
		entry := entities.CacheEntry{Query: query, Embedding: embedding, Response: answer, LLMModel: c.cfg.LLMModel, EmbeddingModel: c.cfg.EmbeddingModel, PromptVersion: c.cfg.PromptVersion, CreatedAt: c.now(), ExpiresAt: c.now().Add(c.cfg.TTL)}
		if err := c.cache.Store(ctx, entry); err != nil {
			c.metrics.StoreErrors.Inc()
			c.logger.Warn("cache store failed", "error", err)
		} else {
			c.logger.Info("cache stored")
		}
	}
	return ChatResult{Answer: answer}, nil
}
func (c *Chat) valid(e *entities.CacheEntry) bool {
	return e.ExpiresAt.After(c.now()) && e.EmbeddingModel == c.cfg.EmbeddingModel && e.LLMModel == c.cfg.LLMModel && e.PromptVersion == c.cfg.PromptVersion && e.Similarity >= c.cfg.Threshold
}
