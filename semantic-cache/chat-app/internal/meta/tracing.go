package meta

import (
	"context"
	"log/slog"

	"chat-app/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func InitTracing(ctx context.Context, cfg *config.AppConfigs, logger *slog.Logger) (func(context.Context) error, error) {
	if !cfg.Server.TracingEnabled {
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(cfg.Server.OTLPEndpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes("", semconv.ServiceName(cfg.Server.ServiceName))),
	)
	otel.SetTracerProvider(provider)
	logger.Info("tracing enabled", "endpoint", cfg.Server.OTLPEndpoint)
	return provider.Shutdown, nil
}
