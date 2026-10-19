# Uptime Monitor

A simple and efficient HTTP/HTTPS service availability monitor with Telegram alerts and Prometheus metrics.

Written in Go, demonstrating goroutines, channels, HTTP clients, and external API integration.

## Features

- Parallel URL checking (goroutines)
- Smart Telegram alerts (only on state change: UP ↔ DOWN)
- Prometheus metrics export
- Flexible YAML configuration
- Configurable timeouts and check intervals
- Graceful shutdown (handles SIGTERM correctly)
- Compact binary with no external dependencies

## Quick Start

### 1. Installation

# Clone the repository
git clone https://github.com/nmarakushev/uptime-monitor.git
cd uptime-monitor`

# Build the binary
go build -o uptime-monitor main.go

# Or run directly
go run main.go

### 2.Telegram Setup
1. Create a bot via @BotFather → /newbot
2. Copy the API Token
3. Get your Chat ID from @userinfobot
4. Send /start to your bot (otherwise it won't be able to message you)

### 3. Setup config.yaml
1. telegram.bot_token - is a token from BotFather
2. telegram.chat_id - is your id, for example, use @userinfobot to get
3. You can use default targets as example and add your own

## Docker

### Build the image

```bash
docker build -t uptime-monitor:latest .