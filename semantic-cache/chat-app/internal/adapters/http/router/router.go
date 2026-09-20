package router

import (
	"log/slog"

	"chat-app/internal/adapters/http/middleware"
	"chat-app/internal/adapters/http/routes"
	"chat-app/internal/app/usecases"
	"chat-app/internal/config"
	"chat-app/internal/meta"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func New(cfg *config.AppConfigs, logger *slog.Logger, metrics *meta.Metrics, chat *usecases.Chat) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID, chimiddleware.RealIP, chimiddleware.Recoverer)
	r.Use(otelhttp.NewMiddleware(cfg.Server.ServiceName))
	r.Use(middleware.Observe(metrics, logger))
	routes.AddSystem(r, cfg.Server.Version, metrics)
	routes.AddAPI(r, cfg, chat, logger)
	return r
}
