package entities

import "time"

type CacheEntry struct {
	Query, Response, LLMModel, EmbeddingModel, PromptVersion string
	Embedding                                                []float32
	CreatedAt, ExpiresAt                                     time.Time
	Similarity                                               float64
}
