package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type ServerConfigs struct {
	Port            string
	ServiceName     string
	Version         string
	LogLevel        string
	TracingEnabled  bool
	OTLPEndpoint    string
	ShutdownTimeout time.Duration
}

type AppConfigs struct {
	Server        *ServerConfigs
	SemanticCache *SemanticCacheConfigs
}

type SemanticCacheConfigs struct {
	Enabled            bool
	Threshold          float64
	TTL                time.Duration
	RedisAddress       string
	RedisPassword      string
	OpenAIBaseURL      string
	OpenAIAPIKey       string
	EmbeddingModel     string
	LLMModel           string
	PromptVersion      string
	EmbeddingDimension int
}

func GetAppConfigs() (*AppConfigs, error) {
	return &AppConfigs{Server: &ServerConfigs{
		Port:            env("PORT", "8080"),
		ServiceName:     env("SERVICE_NAME", "chat-app"),
		Version:         env("VERSION", "dev"),
		LogLevel:        env("LOG_LEVEL", "info"),
		TracingEnabled:  env("TRACING_ENABLED", "false") == "true",
		OTLPEndpoint:    env("OTLP_ENDPOINT", "otel-collector:4317"),
		ShutdownTimeout: duration("SHUTDOWN_TIMEOUT", 5*time.Second),
	}, SemanticCache: &SemanticCacheConfigs{
		Enabled: env("SEMANTIC_CACHE_ENABLED", "true") == "true", Threshold: floatEnv("SEMANTIC_CACHE_THRESHOLD", 0.92),
		TTL: duration("SEMANTIC_CACHE_TTL", time.Hour), RedisAddress: env("REDIS_ADDR", "redis:6379"), RedisPassword: env("REDIS_PASSWORD", ""),
		OpenAIBaseURL: env("OPENAI_BASE_URL", "https://api.openai.com/v1"), OpenAIAPIKey: env("OPENAI_API_KEY", ""),
		EmbeddingModel: env("EMBEDDING_MODEL", "text-embedding-3-small"), LLMModel: env("LLM_MODEL", "gpt-4o-mini"),
		PromptVersion: env("PROMPT_VERSION", "v1"), EmbeddingDimension: intEnv("EMBEDDING_DIMENSION", 0),
	}}, nil
}

func (c ServerConfigs) Address() string { return fmt.Sprintf(":%s", c.Port) }

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func duration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(env(key, fallback.String()))
	if err != nil {
		return fallback
	}
	return value
}

func floatEnv(key string, fallback float64) float64 {
	value, err := strconv.ParseFloat(env(key, ""), 64)
	if err != nil {
		return fallback
	}
	return value
}
func intEnv(key string, fallback int) int {
	value, err := strconv.Atoi(env(key, ""))
	if err != nil {
		return fallback
	}
	return value
}
