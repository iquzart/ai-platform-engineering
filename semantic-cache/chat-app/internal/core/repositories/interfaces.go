package repositories

import (
	"chat-app/internal/core/entities"
	"context"
)

type Embedder interface {
	Embed(context.Context, string) ([]float32, error)
}
type Generator interface {
	Generate(context.Context, string) (string, error)
}
type SemanticCache interface {
	Find(context.Context, []float32) (*entities.CacheEntry, error)
	Store(context.Context, entities.CacheEntry) error
	Close() error
}
