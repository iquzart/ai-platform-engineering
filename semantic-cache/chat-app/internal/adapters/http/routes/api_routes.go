package routes

import (
	"chat-app/internal/adapters/http/handlers"
	"chat-app/internal/app/usecases"
	"chat-app/internal/config"
	"github.com/go-chi/chi/v5"
	"log/slog"
)

func AddAPI(r chi.Router, cfg *config.AppConfigs, chat *usecases.Chat, logger *slog.Logger) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", handlers.Ping(usecases.NewPing(cfg.Server.ServiceName)))
	})
	r.Post("/chat", handlers.Chat(chat, logger))
	r.Get("/docs", handlers.SwaggerUI)
	r.Get("/docs/openapi.yaml", handlers.OpenAPISpec)
}
