package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/betothewizard/crosscast/internal/broadcast"
)

// Handler holds dependencies for HTTP endpoints.
type Handler struct {
	broadcast *broadcast.Service
	apiKey    string
}

// ErrorResponse is returned upon failure.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func (h *Handler) handleBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Limit request body to 1MB
	body, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: "failed to read request body"})
		return
	}

	var payload broadcast.BroadcastPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: fmt.Sprintf("invalid JSON payload: %v", err)})
		return
	}

	result, err := h.broadcast.Broadcast(r.Context(), &payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var discordStatus, telegramStatus any
	if h.broadcast != nil {
		discordStatus = h.broadcast.DiscordStatus()
		telegramStatus = h.broadcast.TelegramStatus()
	}

	data := map[string]any{
		"service":   "crosscast",
		"version":   "1.0.0",
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"discord":   discordStatus,
		"telegram":  telegramStatus,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var discordStatus, telegramStatus any
	if h.broadcast != nil {
		discordStatus = h.broadcast.DiscordStatus()
		telegramStatus = h.broadcast.TelegramStatus()
	}

	data := map[string]any{
		"name":     "Crosscast",
		"tagline":  "Ultra-lightweight multi-tenant social & community broadcasting gateway",
		"version":  "1.0.0",
		"discord":  discordStatus,
		"telegram": telegramStatus,
		"endpoints": map[string]string{
			"POST /broadcast": "Send rich messages to Discord channels and/or Telegram chats",
			"GET /health":     "Check service, Discord WebSocket, and Telegram bot health",
		},
		"documentation": "https://github.com/betothewizard/crosscast",
	}

	json.NewEncoder(w).Encode(data)
}
