package gateway

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/betothewizard/crosscast/internal/bot"
	"github.com/betothewizard/crosscast/internal/config"
	"github.com/betothewizard/crosscast/internal/router"
	"github.com/betothewizard/crosscast/internal/telegram"
)

// HTTPServer manages the inbound HTTP gateway.
type HTTPServer struct {
	httpServer *http.Server
	port       int
}

// NewServer initializes a new HTTPServer with Discord and Telegram bot backends wired.
func NewServer(cfg *config.Config, b *bot.Bot, tg *telegram.Bot, rtr *router.Router) *HTTPServer {
	s := &Server{
		bot:    b,
		tgBot:  tg,
		router: rtr,
		apiKey: cfg.APIKey,
	}

	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("POST /broadcast", AuthMiddleware(s.apiKey, s.handleBroadcast))
	mux.HandleFunc("POST /api/v1/broadcast", AuthMiddleware(s.apiKey, s.handleBroadcast))
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /", s.handleRoot)

	// Apply middlewares (Recovery -> Logging -> Router)
	handler := RecoveryMiddleware(LoggerMiddleware(mux))

	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &HTTPServer{
		httpServer: httpSrv,
		port:       cfg.Port,
	}
}

// Start listens and serves inbound HTTP requests.
func (s *HTTPServer) Start() error {
	log.Printf("[gateway] HTTP server listening on http://0.0.0.0:%d", s.port)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server failed: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	log.Printf("[gateway] stopping HTTP server...")
	return s.httpServer.Shutdown(ctx)
}
