package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/betothewizard/crosscast/internal/bot"
	"github.com/betothewizard/crosscast/internal/router"
)

// Server holds dependencies for HTTP handlers.
type Server struct {
	bot    *bot.Bot
	router *router.Router
	apiKey string
}

// BroadcastResponse is returned upon successful broadcast.
type BroadcastResponse struct {
	Success   bool      `json:"success"`
	Target    string    `json:"target"`
	ChannelID string    `json:"channel_id"`
	MessageID string    `json:"message_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorResponse is returned upon failure.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func (s *Server) handleBroadcast(w http.ResponseWriter, r *http.Request) {
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

	var payload router.BroadcastPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: fmt.Sprintf("invalid JSON payload: %v", err)})
		return
	}

	channelID, msg, err := s.router.PrepareMessage(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	sentMsg, err := s.bot.SendMessage(channelID, msg)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: fmt.Sprintf("failed to deliver message to Discord: %v", err)})
		return
	}

	resp := BroadcastResponse{
		Success:   true,
		Target:    payload.Target,
		ChannelID: channelID,
		MessageID: sentMsg.ID,
		Timestamp: time.Now().UTC(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := s.bot.GetStatus()

	data := map[string]any{
		"service":   "crosscast",
		"version":   "1.0.0",
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"bot":       status,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	status := s.bot.GetStatus()

	data := map[string]any{
		"name":        "Crosscast",
		"tagline":     "Ultra-lightweight multi-tenant social & community broadcasting gateway",
		"version":     "1.0.0",
		"bot_status":  status,
		"endpoints": map[string]string{
			"POST /broadcast": "Send rich messages to Discord channels or configured target",
			"GET /health":     "Check service and Discord WebSocket health",
		},
		"documentation": "https://github.com/betothewizard/crosscast",
	}

	json.NewEncoder(w).Encode(data)
}
