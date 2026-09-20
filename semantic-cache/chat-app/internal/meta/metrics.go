package meta

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	Requests                                       *prometheus.CounterVec
	Duration                                       *prometheus.HistogramVec
	registry                                       *prometheus.Registry
	SemanticRequests, SemanticHits, SemanticMisses prometheus.Counter
	EmbeddingDuration, LLMDuration, LookupDuration prometheus.Histogram
	StoreErrors                                    prometheus.Counter
}

func NewMetrics() *Metrics {
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests."}, []string{"method", "route", "status"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "HTTP request duration in seconds."}, []string{"method", "route"})
	registry := prometheus.NewRegistry()
	semanticRequests := prometheus.NewCounter(prometheus.CounterOpts{Name: "semantic_cache_requests_total", Help: "Semantic cache chat requests."})
	hits := prometheus.NewCounter(prometheus.CounterOpts{Name: "semantic_cache_hits_total", Help: "Semantic cache hits."})
	misses := prometheus.NewCounter(prometheus.CounterOpts{Name: "semantic_cache_misses_total", Help: "Semantic cache misses."})
	embedding := prometheus.NewHistogram(prometheus.HistogramOpts{Name: "semantic_cache_embedding_duration_seconds", Help: "Embedding duration."})
	llm := prometheus.NewHistogram(prometheus.HistogramOpts{Name: "semantic_cache_llm_duration_seconds", Help: "LLM duration."})
	lookup := prometheus.NewHistogram(prometheus.HistogramOpts{Name: "semantic_cache_lookup_duration_seconds", Help: "Redis lookup duration."})
	storeErrors := prometheus.NewCounter(prometheus.CounterOpts{Name: "semantic_cache_store_errors_total", Help: "Cache store failures."})
	registry.MustRegister(requests, duration, semanticRequests, hits, misses, embedding, llm, lookup, storeErrors)
	return &Metrics{Requests: requests, Duration: duration, registry: registry, SemanticRequests: semanticRequests, SemanticHits: hits, SemanticMisses: misses, EmbeddingDuration: embedding, LLMDuration: llm, LookupDuration: lookup, StoreErrors: storeErrors}
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
