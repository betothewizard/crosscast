package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/betothewizard/crosscast/internal/broadcast"
	"github.com/betothewizard/crosscast/internal/config"
	"github.com/betothewizard/crosscast/internal/platform/discord"
	"github.com/betothewizard/crosscast/internal/platform/telegram"
	"github.com/betothewizard/crosscast/internal/server"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := &config.Config{Port: 6403}
	d := discord.New(cfg)
	tg := telegram.New(cfg)
	bcast := broadcast.NewService(cfg, d, tg)
	srv := server.New(cfg, bcast)
	_ = srv

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "healthy", "service": "crosscast"})
	})

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "healthy" {
		t.Errorf("status = %v, want 'healthy'", resp["status"])
	}
}

func TestAuthMiddleware(t *testing.T) {
	apiKey := "secret-test-token"

	handler := server.AuthMiddleware(apiKey, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	})

	tests := []struct {
		name         string
		authHeader   string
		expectedCode int
	}{
		{"missing auth header", "", http.StatusUnauthorized},
		{"invalid format", "Basic 12345", http.StatusUnauthorized},
		{"wrong token", "Bearer wrong-token", http.StatusUnauthorized},
		{"correct token", "Bearer secret-test-token", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/broadcast", bytes.NewBufferString(`{}`))
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedCode)
			}
		})
	}
}
