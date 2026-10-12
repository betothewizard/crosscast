package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/betothewizard/crosscast/internal/bot"
	"github.com/betothewizard/crosscast/internal/router"
	"github.com/betothewizard/crosscast/internal/telegram"
)

// Server holds dependencies for HTTP handlers.
type Server struct {
	bot    *bot.Bot
	tgBot  *telegram.Bot
	router *router.Router
	apiKey string
}

// DeliveryResult records delivery metadata for a specific platform.
type DeliveryResult struct {
	Platform  string `json:"platform"`
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id,omitempty"`
}

// BroadcastResponse is returned upon successful broadcast.
type BroadcastResponse struct {
	Success    bool                      `json:"success"`
	Target     string                    `json:"target"`
	Deliveries map[string]DeliveryResult `json:"deliveries"`
	Timestamp  time.Time                 `json:"timestamp"`
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

	res, err := s.router.Resolve(payload.Target, payload.Platform)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	deliveries := make(map[string]DeliveryResult)
	var errs []string

	// 1. Deliver to Discord if target resolved
	if res.DiscordChannelID != "" {
		if s.bot != nil && s.bot.IsReady() {
			discordMsg := s.router.PrepareDiscordMessage(&payload, res.DefaultColor)
			sentMsg, err := s.bot.SendMessage(res.DiscordChannelID, discordMsg)
			if err != nil {
				errs = append(errs, fmt.Sprintf("discord: %v", err))
			} else {
				deliveries["discord"] = DeliveryResult{
					Platform:  "discord",
					ChannelID: res.DiscordChannelID,
					MessageID: sentMsg.ID,
				}
			}
		} else {
			errs = append(errs, "discord: bot session not active")
		}
	}

	// 2. Deliver to Telegram if target resolved
	if res.TelegramChatID != "" {
		if s.tgBot != nil && s.tgBot.IsReady() {
			sentMsg, err := s.tgBot.SendMessage(res.TelegramChatID, &payload)
			if err != nil {
				errs = append(errs, fmt.Sprintf("telegram: %v", err))
			} else {
				deliveries["telegram"] = DeliveryResult{
					Platform:  "telegram",
					ChannelID: res.TelegramChatID,
					MessageID: strconv.Itoa(sentMsg.MessageID),
				}
			}
		} else {
			errs = append(errs, "telegram: bot session not active")
		}
	}

	if len(deliveries) == 0 {
		w.WriteHeader(http.StatusBadGateway)
		errMsg := "failed to deliver to any platform"
		if len(errs) > 0 {
			errMsg = strings.Join(errs, "; ")
		}
		json.NewEncoder(w).Encode(ErrorResponse{Success: false, Error: errMsg})
		return
	}

	resp := BroadcastResponse{
		Success:    true,
		Target:     payload.Target,
		Deliveries: deliveries,
		Timestamp:  time.Now().UTC(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	discordStatus := any(nil)
	if s.bot != nil {
		discordStatus = s.bot.GetStatus()
	}

	telegramStatus := any(nil)
	if s.tgBot != nil {
		telegramStatus = s.tgBot.GetStatus()
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

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	discordStatus := any(nil)
	if s.bot != nil {
		discordStatus = s.bot.GetStatus()
	}

	telegramStatus := any(nil)
	if s.tgBot != nil {
		telegramStatus = s.tgBot.GetStatus()
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
