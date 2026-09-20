package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"chat-app/internal/meta"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(body)
}

func Observe(metrics *meta.Metrics, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/system/metrics" {
				next.ServeHTTP(w, r)
				return
			}
			started := time.Now()
			recorder := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(recorder, r)
			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = "unmatched"
			}
			metrics.Requests.WithLabelValues(r.Method, route, strconv.Itoa(recorder.status)).Inc()
			metrics.Duration.WithLabelValues(r.Method, route).Observe(time.Since(started).Seconds())
			attributes := []any{"method", r.Method, "route", route, "status", recorder.status, "duration_ms", time.Since(started).Milliseconds()}
			if span := trace.SpanFromContext(r.Context()); span.SpanContext().IsValid() {
				attributes = append(attributes, "trace_id", span.SpanContext().TraceID().String())
			}
			logger.Info("http request", attributes...)
		})
	}
}
