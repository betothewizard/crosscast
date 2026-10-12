package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/betothewizard/crosscast/internal/bot"
	"github.com/betothewizard/crosscast/internal/config"
	"github.com/betothewizard/crosscast/internal/gateway"
	"github.com/betothewizard/crosscast/internal/router"
	"github.com/betothewizard/crosscast/internal/telegram"
)

const banner = `
   ____                      ____           _   
  / ___|_ __ ___  ___ ___   / ___|__ _  ___| |_ 
 | |   | '__/ _ \/ __/ __| | |   / _' |/ __| __|
 | |___| | | (_) \__ \__ \ | |__| (_| |\__ \ |_ 
  \____|_|  \___/|___/___/  \____\__,_||___/\__|
  Crosscast v1.0.0 — Multi-Tenant Broadcast Gateway
`

func main() {
	log.Print(banner)

	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize Discord Bot Daemon
	b := bot.New(cfg)
	if err := b.Start(ctx); err != nil {
		log.Printf("[discord] connection error: %v", err)
	}
	defer b.Close()

	// 2. Initialize Telegram Bot Daemon
	tgBot := telegram.New(cfg)
	if err := tgBot.Start(ctx); err != nil {
		log.Printf("[telegram] connection error: %v", err)
	}
	defer tgBot.Close()

	// 3. Initialize Multi-Platform Routing Engine
	rtr := router.New(cfg)

	// 4. Initialize HTTP Gateway Server
	srv := gateway.NewServer(cfg, b, tgBot, rtr)

	// Start HTTP server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("[gateway] fatal: %v", err)
		}
	}()

	// 5. Wait for termination signals (SIGINT, SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	log.Printf("[crosscast] caught signal %v, shutting down gracefully...", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[gateway] error during shutdown: %v", err)
	}

	log.Println("[crosscast] shutdown complete. Goodbye!")
}
