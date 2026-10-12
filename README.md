# Crosscast

Lightweight broadcast gateway for Discord and Telegram written in Go.

Accepts inbound HTTP requests and dispatches formatted messages to Discord channels and Telegram chats.

## Features

- Multi-platform: Broadcast to Discord and Telegram simultaneously or independently.
- Minimal footprint: Single static binary, under 15 MB RAM at idle.
- Inbound HTTP API (`POST /broadcast`) with Bearer token authentication.
- Persistent bots: WebSocket presence for Discord and update polling for Telegram.
- JSON-based channel mapping configured via `.env`.

## Quickstart

1. Configure environment:
```bash
cp .env.example .env
```

2. Set credentials in `.env`:
```env
PORT=6403
CROSSCAST_API_KEY=secret_key

# Discord
DISCORD_BOT_TOKEN=bot_token
DISCORD_APP_ID=app_id

# Telegram
TELEGRAM_BOT_TOKEN=telegram_bot_token

# Channel routing map (JSON object)
CROSSCAST_CHANNELS={"alerts":{"discord":"123456789012345678","telegram":"-1001234567890"},"deployments":"123456789012345678"}
```

3. Run:
```bash
make run
```

## Usage

### Broadcast a message

```bash
curl -X POST http://localhost:6403/broadcast \
  -H "Authorization: Bearer secret_key" \
  -H "Content-Type: application/json" \
  -d '{
    "target": "alerts",
    "title": "Deployment Successful",
    "description": "Production release finished without errors.",
    "url": "https://example.com",
    "fields": [
      {"name": "Environment", "value": "Production", "inline": true},
      {"name": "Status", "value": "Healthy", "inline": true}
    ]
  }'
```

To limit delivery to a specific platform, include `"platform": "discord"` or `"platform": "telegram"` (default: `"all"`).

Direct channel IDs are also supported in `"target"`:
- Discord snowflake ID: `"123456789012345678"`
- Telegram chat ID: `"-1001234567890"` or `"@channel_username"`

### Health check

```bash
curl http://localhost:6403/health
```

## Docker

```bash
make docker-build
docker run -d -p 6403:6403 --env-file .env crosscast:latest
```

## License

MIT
