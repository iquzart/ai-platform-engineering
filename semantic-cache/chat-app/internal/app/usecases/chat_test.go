package usecases

import (
	"chat-app/internal/config"
	"chat-app/internal/core/entities"
	"chat-app/internal/meta"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

type fakeEmbedder struct{ err error }

func (f fakeEmbedder) Embed(context.Context, string) ([]float32, error) {
	return []float32{1, 2}, f.err
}

type fakeGenerator struct {
	calls int
	err   error
}

func (f *fakeGenerator) Generate(context.Context, string) (string, error) {
	f.calls++
	return "generated", f.err
}

type fakeCache struct {
	entry             *entities.CacheEntry
	findErr, storeErr error
	stored            bool
}

func (f *fakeCache) Find(context.Context, []float32) (*entities.CacheEntry, error) {
	return f.entry, f.findErr
}
func (f *fakeCache) Store(context.Context, entities.CacheEntry) error {
	f.stored = true
	return f.storeErr
}
func (f *fakeCache) Close() error { return nil }

type storedCache struct{ entry *entities.CacheEntry }

func (c *storedCache) Find(context.Context, []float32) (*entities.CacheEntry, error) {
	return c.entry, nil
}
func (c *storedCache) Store(_ context.Context, entry entities.CacheEntry) error {
	entry.Similarity = 1
	c.entry = &entry
	return nil
}
func (c *storedCache) Close() error { return nil }
func testChat(cache *fakeCache, generator *fakeGenerator) *Chat {
	cfg := &config.SemanticCacheConfigs{Enabled: true, Threshold: .92, TTL: time.Hour, EmbeddingModel: "embed", LLMModel: "llm", PromptVersion: "v1"}
	return NewChat(fakeEmbedder{}, generator, cache, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), meta.NewMetrics())
}
func entry(now time.Time) *entities.CacheEntry {
	return &entities.CacheEntry{Response: "cached", EmbeddingModel: "embed", LLMModel: "llm", PromptVersion: "v1", ExpiresAt: now.Add(time.Hour), Similarity: .96}
}
func TestChatCacheHit(t *testing.T) {
	now := time.Now()
	g := &fakeGenerator{}
	c := testChat(&fakeCache{entry: entry(now)}, g)
	c.now = func() time.Time { return now }
	got, err := c.Ask(context.Background(), "q")
	if err != nil || !got.CacheHit || got.Answer != "cached" || g.calls != 0 {
		t.Fatalf("got=%+v err=%v calls=%d", got, err, g.calls)
	}
}
func TestChatCacheMissAndStore(t *testing.T) {
	g := &fakeGenerator{}
	cache := &fakeCache{}
	got, err := testChat(cache, g).Ask(context.Background(), "q")
	if err != nil || got.CacheHit || g.calls != 1 || !cache.stored {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestChatSameRequestHitsAfterStore(t *testing.T) {
	g := &fakeGenerator{}
	cache := &storedCache{}
	cfg := &config.SemanticCacheConfigs{Enabled: true, Threshold: .92, TTL: time.Hour, EmbeddingModel: "embed", LLMModel: "llm", PromptVersion: "v1"}
	chat := NewChat(fakeEmbedder{}, g, cache, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), meta.NewMetrics())

	if _, err := chat.Ask(context.Background(), "same request"); err != nil {
		t.Fatal(err)
	}
	got, err := chat.Ask(context.Background(), "same request")
	if err != nil || !got.CacheHit || got.Answer != "generated" || g.calls != 1 {
		t.Fatalf("got=%+v err=%v generator calls=%d", got, err, g.calls)
	}
}
func TestChatExpiredEntryMisses(t *testing.T) {
	now := time.Now()
	g := &fakeGenerator{}
	e := entry(now)
	e.ExpiresAt = now.Add(-time.Second)
	c := testChat(&fakeCache{entry: e}, g)
	c.now = func() time.Time { return now }
	got, _ := c.Ask(context.Background(), "q")
	if got.CacheHit || g.calls != 1 {
		t.Fatal("expired entry was reused")
	}
}
func TestChatModelAndPromptMismatchMiss(t *testing.T) {
	now := time.Now()
	for _, mutate := range []func(*entities.CacheEntry){func(e *entities.CacheEntry) { e.LLMModel = "old" }, func(e *entities.CacheEntry) { e.PromptVersion = "v0" }} {
		g := &fakeGenerator{}
		e := entry(now)
		mutate(e)
		c := testChat(&fakeCache{entry: e}, g)
		c.now = func() time.Time { return now }
		got, _ := c.Ask(context.Background(), "q")
		if got.CacheHit || g.calls != 1 {
			t.Fatal("mismatched entry was reused")
		}
	}
}
func TestChatBelowThresholdMisses(t *testing.T) {
	now := time.Now()
	g := &fakeGenerator{}
	e := entry(now)
	e.Similarity = .91
	c := testChat(&fakeCache{entry: e}, g)
	c.now = func() time.Time { return now }
	got, _ := c.Ask(context.Background(), "q")
	if got.CacheHit {
		t.Fatal("below threshold hit")
	}
}
func TestChatRedisFailureFallsBack(t *testing.T) {
	g := &fakeGenerator{}
	got, err := testChat(&fakeCache{findErr: errors.New("redis down")}, g).Ask(context.Background(), "q")
	if err != nil || got.Answer != "generated" || g.calls != 1 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}
func TestChatStoreFailureStillReturns(t *testing.T) {
	g := &fakeGenerator{}
	got, err := testChat(&fakeCache{storeErr: errors.New("redis down")}, g).Ask(context.Background(), "q")
	if err != nil || got.Answer != "generated" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}
func TestChatEmbeddingFailure(t *testing.T) {
	g := &fakeGenerator{}
	cfg := &config.SemanticCacheConfigs{Enabled: true}
	c := NewChat(fakeEmbedder{err: errors.New("no embedding")}, g, &fakeCache{}, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), meta.NewMetrics())
	if _, err := c.Ask(context.Background(), "q"); err == nil || g.calls != 0 {
		t.Fatal("embedding failure should stop request")
	}
}
