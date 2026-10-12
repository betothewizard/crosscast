package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/betothewizard/crosscast/internal/broadcast"
	"github.com/betothewizard/crosscast/internal/config"
)

// HTTPServer manages the inbound HTTP server and routing.
type HTTPServer struct {
	httpServer *http.Server
	port       int
}

// New initializes a new HTTPServer with routes and middlewares.
func New(cfg *config.Config, bcast *broadcast.Service) *HTTPServer {
	h := &Handler{
		broadcast: bcast,
		apiKey:    cfg.APIKey,
	}

	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("POST /broadcast", AuthMiddleware(h.apiKey, h.handleBroadcast))
	mux.HandleFunc("POST /api/v1/broadcast", AuthMiddleware(h.apiKey, h.handleBroadcast))
	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("GET /", h.handleRoot)

	// Apply middlewares (Recovery -> Logging -> Mux)
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
	log.Printf("[server] HTTP server listening on http://0.0.0.0:%d", s.port)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server failed: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	log.Printf("[server] stopping HTTP server...")
	return s.httpServer.Shutdown(ctx)
}
