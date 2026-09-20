package config

import "testing"

func TestGetAppConfigsReadsRedisPassword(t *testing.T) {
	t.Setenv("REDIS_PASSWORD", "test-password")

	cfg, err := GetAppConfigs()
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.SemanticCache.RedisPassword; got != "test-password" {
		t.Fatalf("RedisPassword = %q, want %q", got, "test-password")
	}
}

func TestGetAppConfigsDefaultsRedisPasswordToEmpty(t *testing.T) {
	t.Setenv("REDIS_PASSWORD", "")

	cfg, err := GetAppConfigs()
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.SemanticCache.RedisPassword; got != "" {
		t.Fatalf("RedisPassword = %q, want empty", got)
	}
}

func TestGetAppConfigsReadsOpenAIConfiguration(t *testing.T) {
	t.Setenv("OPENAI_BASE_URL", "http://localhost:4000/v1")
	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("EMBEDDING_MODEL", "text-embedding-3-small")
	t.Setenv("LLM_MODEL", "gpt-4o-mini")

	cfg, err := GetAppConfigs()
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.SemanticCache.OpenAIBaseURL; got != "http://localhost:4000/v1" {
		t.Fatalf("OpenAIBaseURL = %q", got)
	}
	if got := cfg.SemanticCache.OpenAIAPIKey; got != "test-key" {
		t.Fatalf("OpenAIAPIKey = %q", got)
	}
	if got := cfg.SemanticCache.EmbeddingModel; got != "text-embedding-3-small" {
		t.Fatalf("EmbeddingModel = %q", got)
	}
	if got := cfg.SemanticCache.LLMModel; got != "gpt-4o-mini" {
		t.Fatalf("LLMModel = %q", got)
	}
}

func TestGetAppConfigsUsesOpenAIDefaults(t *testing.T) {
	t.Setenv("OPENAI_BASE_URL", "")
	t.Setenv("OPENAI_API_KEY", "")

	cfg, err := GetAppConfigs()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SemanticCache.OpenAIBaseURL != "https://api.openai.com/v1" || cfg.SemanticCache.OpenAIAPIKey != "" || cfg.SemanticCache.EmbeddingModel != "text-embedding-3-small" || cfg.SemanticCache.LLMModel != "gpt-4o-mini" {
		t.Fatalf("OpenAI configuration = base URL %q, API key %q, embedding model %q, chat model %q", cfg.SemanticCache.OpenAIBaseURL, cfg.SemanticCache.OpenAIAPIKey, cfg.SemanticCache.EmbeddingModel, cfg.SemanticCache.LLMModel)
	}
}
