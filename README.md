![CI/CD Status](https://github.com/NikitaMarakushev/uptime-mon/actions/workflows/ci.yaml/badge.svg?branch=master)

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

---

## Quick Start

### 1. Installation

```bash
# Clone the repository
git clone https://github.com/nmarakushev/uptime-monitor.git
cd uptime-monitor

# Build the binary
go build -o uptime-monitor main.go

# Or run directly
go run main.go
```

### 2. Telegram Setup
1) Create a bot via @BotFather → /newbot
2) Copy the API Token
3) Get your Chat ID from @userinfobot
4) Send /start to your bot (otherwise it won't be able to message you)

### 3. Configuration
Create a config.yaml file in the project root:
```
telegram:
  bot_token: "YOUR_BOT_TOKEN_HERE"
  chat_id: "YOUR_CHAT_ID_HERE"

check_interval_seconds: 15

targets:
  - name: "Google"
    url: "https://google.com"
  - name: "GitHub"
    url: "https://github.com"
  - name: "Yandex"
    url: "https://yandex.ru"
```

### 4. Build the Docker
1) Build image:
`docker build -t uptime-monitor:latest .`
2) Run:
```
docker run -d \
  --name uptime-monitor \
  -p 9116:9116 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  uptime-monitor:latest
```
3) Run full monitoring stack:
```
# Start all services
docker-compose up -d

# Access:
# - Uptime Monitor metrics: http://localhost:9116/metrics
# - Prometheus: http://localhost:9090
# - Grafana: http://localhost:3000 (admin/admin)

# Stop all services
docker-compose down
```

### 4. Kubernetes Deployment
The project includes a production-ready Helm chart for deployment to any Kubernetes cluster.
Prerequisites
- Kubernetes cluster (tested with k3s, minikube, kind)
- Helm 3.x installed
- Docker image built and imported into your cluster's container runtime
- Import Docker Image to k3s
Since k3s uses containerd instead of Docker, you need to import the image manually

```
docker save uptime-monitor:latest -o /tmp/uptime.tar

sudo k3s ctr images import /tmp/uptime.tar

sudo k3s ctr images ls | grep uptime-monitor

rm /tmp/uptime.tar
```

### 5. Deploy with Helm
```
helm install uptime-monitor ./helm/uptime-monitor

kubectl get pods -l app.kubernetes.io/name=uptime-monitor

kubectl logs -l app.kubernetes.io/name=uptime-monitor -f

kubectl port-forward svc/uptime-monitor 9116:9116

curl http://localhost:9116/metrics | grep uptime_target_up
```

### 6. Uninstall
`helm uninstall uptime-monitor`

### Development Workflow
This project follows the GitFlow branching model:
- develop: All development happens here. CI runs tests, lint, builds and scans the image, but does NOT publish it.
- master: Production branch. Merges from develop trigger full CI/CD pipeline including image publication to GHCR.
-Pull Requests: Required for merging into master. Branch protection rules enforce passing CI checks.

### Highlights
This project demonstrates a complete Cloud Native development lifecycle:
1) Go application with concurrent HTTP checks and Telegram alerting
2) Multi-stage Docker build with distroless base image (~20MB)
3) Security-first approach: non-root user, Trivy CVE scanning
4) Automated CI/CD via GitHub Actions (lint, test, build, scan, publish)
5) GitFlow branching model with branch protection
6) Helm chart for Kubernetes deployment with proper SecurityContext
7) Prometheus metrics with service discovery annotations


