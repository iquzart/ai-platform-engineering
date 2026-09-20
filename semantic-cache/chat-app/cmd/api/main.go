package main

import (
	"chat-app/internal/adapters/http/server"
	"chat-app/internal/bootstrap"
	"chat-app/internal/config"
	"chat-app/internal/meta"
)

func main() {
	// Load configs.
	cfg, err := config.GetAppConfigs()
	if err != nil {
		meta.Fatal(meta.NewLogger("error"), "failed to load application configs", "error", err)
	}

	deps, err := bootstrap.Initialize(cfg)
	if err != nil {
		meta.Fatal(meta.NewLogger("error"), "failed to initialize dependencies", "error", err)
	}
	defer deps.Close()

	// Start server.
	server.New(cfg, deps).Run()
}
