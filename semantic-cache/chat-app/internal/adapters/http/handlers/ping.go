package handlers

import (
	"encoding/json"
	"net/http"

	"chat-app/internal/adapters/http/dto"
	"chat-app/internal/app/usecases"
)

func Ping(usecase usecases.Ping) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		result := usecase.Execute()
		writeJSON(w, http.StatusOK, dto.PingResponse{Message: "pong", Service: result.Service, Timestamp: result.Timestamp})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
