package handlers

import (
	"net/http"

	"chat-app/internal/adapters/http/dto"
)

func Liveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, dto.HealthResponse{Status: "live"})
}

func Readiness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, dto.HealthResponse{Status: "ready"})
}

func Version(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, dto.VersionResponse{Version: version})
	}
}
