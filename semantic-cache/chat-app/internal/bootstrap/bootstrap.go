package bootstrap

import (
	"context"
	"log/slog"

	redisadapter "chat-app/internal/adapters/cache/redis"
	"chat-app/internal/adapters/openai"
	"chat-app/internal/app/usecases"
	"chat-app/internal/config"
	"chat-app/internal/meta"
)

type Dependencies struct {
	Logger          *slog.Logger
	Metrics         *meta.Metrics
	Chat            *usecases.Chat
	cacheClose      func() error
	shutdownTracing func(context.Context) error
}

func Initialize(cfg *config.AppConfigs) (*Dependencies, error) {
	logger := meta.NewLogger(cfg.Server.LogLevel)
	metrics := meta.NewMetrics()
	shutdownTracing, err := meta.InitTracing(context.Background(), cfg, logger)
	if err != nil {
		return nil, err
	}

	cache := redisadapter.New(cfg.SemanticCache.RedisAddress, cfg.SemanticCache.RedisPassword, cfg.SemanticCache.EmbeddingDimension)
	return &Dependencies{
		Logger:          logger,
		Metrics:         metrics,
		Chat:            usecases.NewChat(openai.NewEmbedder(cfg.SemanticCache.OpenAIBaseURL, cfg.SemanticCache.OpenAIAPIKey, cfg.SemanticCache.EmbeddingModel), openai.NewGenerator(cfg.SemanticCache.OpenAIBaseURL, cfg.SemanticCache.OpenAIAPIKey, cfg.SemanticCache.LLMModel), cache, cfg.SemanticCache, logger, metrics),
		cacheClose:      cache.Close,
		shutdownTracing: shutdownTracing,
	}, nil
}

func (d *Dependencies) Close() {
	if err := d.cacheClose(); err != nil {
		d.Logger.Error("close Redis", "error", err)
	}
	if err := d.shutdownTracing(context.Background()); err != nil {
		d.Logger.Error("shutdown tracing", "error", err)
	}
}
