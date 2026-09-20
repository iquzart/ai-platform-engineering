package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"chat-app/internal/adapters/http/dto"
	"chat-app/internal/app/usecases"
)

func Chat(chat *usecases.Chat, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request dto.ChatRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil || strings.TrimSpace(request.Query) == "" {
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "query must be a non-empty JSON string"})
			return
		}
		result, err := chat.Ask(r.Context(), request.Query)
		if err != nil {
			logger.Error("chat request failed", "error", err)
			writeJSON(w, http.StatusBadGateway, dto.ErrorResponse{Error: "chat service unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, dto.ChatResponse{Answer: result.Answer, CacheHit: result.CacheHit, Similarity: result.Similarity})
	}
}
