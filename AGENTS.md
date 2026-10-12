# Crosscast — Agent Guidelines & Architecture Protocol

`crosscast` is the ultra-lightweight multi-tenant social publishing & community broadcasting gateway written in Go.

---

## 1. Primary Responsibilities

1. **Inbound HTTP Broadcast Gateway (`POST /broadcast`)**:
   - Exposes port **`6403`**.
   - Authenticates requests via Bearer token (`CROSSCAST_API_KEY`).
   - Routes payloads to configured target channels (e.g. `alerts`, `announcements`, `deployments`, or raw Discord channel snowflake IDs).
   - Generates beautifully formatted Discord Rich Embeds with smart color coding.

2. **24/7 Persistent Discord Bot Daemon (`internal/bot`)**:
   - Maintains continuous WebSocket presence with the Discord Gateway.
   - Registers and handles interactive Slash Commands (`/ping`, `/status`).
   - Sits in guild member lists with customized presence activities.

---

## 2. Architecture Map

```text
crosscast/
├── cmd/
│   └── crosscast/
│       └── main.go           ← Entrypoint, signal handler & lifecycle orchestrator
├── internal/
│   ├── bot/                  ← Discordgo session, WebSocket connection & slash commands
│   │   ├── bot.go
│   │   └── handlers.go
│   ├── config/               ← Environment variable loader & defaults (.env support)
│   │   └── config.go
│   ├── gateway/              ← HTTP REST API server (Go 1.22+ net/http)
│   │   ├── handlers.go
│   │   ├── middleware.go     ← Bearer auth, logging & panic recovery
│   │   └── server.go
│   └── router/               ← Dynamic channel routing & Rich Embed transformer
│       ├── message.go
│       └── router.go
├── .env.example
├── Dockerfile
├── Makefile
├── LICENSE
└── README.md
```

---

## 3. Engineering Constraints & Rules

1. **Port Allocation**: Crosscast listens on port **`6403`** by default.
2. **Standard Library First**: Use `net/http` for HTTP routing; avoid heavy external web frameworks (no Gin, Fiber, or Echo).
3. **No CGO**: Keep Crosscast 100% pure Go to enable effortless cross-compilation and scratch Docker images.
4. **Memory Footprint**: Target idle memory consumption under **20 MB RAM**.
5. **Open Source Purity**: Crosscast is an open-source, general-purpose project. Do not commit private project names, credentials, or custom endpoints into the repository.

---

## 4. Useful Commands

```bash
# Run local development server
make run

# Run tests
make test

# Build production binary
make build

# Build minimal Docker image
make docker-build
```
