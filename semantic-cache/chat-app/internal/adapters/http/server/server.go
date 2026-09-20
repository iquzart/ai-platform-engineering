package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chat-app/internal/adapters/http/router"
	"chat-app/internal/bootstrap"
	"chat-app/internal/config"
)

type Server struct {
	cfg    *config.ServerConfigs
	logger *slog.Logger
	http   *http.Server
}

func New(cfg *config.AppConfigs, deps *bootstrap.Dependencies) *Server {
	return &Server{
		cfg:    cfg.Server,
		logger: deps.Logger,
		http: &http.Server{
			Addr:              cfg.Server.Address(),
			Handler:           router.New(cfg, deps.Logger, deps.Metrics, deps.Chat),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      130 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (s *Server) Run() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	go func() {
		s.logger.Info("service started", "address", s.cfg.Address(), "service", s.cfg.ServiceName)
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("serve HTTP", "error", err)
		}
	}()

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	if err := s.http.Shutdown(ctx); err != nil {
		s.logger.Error("shutdown HTTP server", "error", err)
	}
}
