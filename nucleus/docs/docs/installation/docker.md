---
sidebar_position: 5
title: Docker
---

# Docker Installation

Docker Compose is the fastest way to run the full Nucleus stack. It bundles all services, databases, and networking with zero manual configuration.

## Prerequisites

- Docker Engine 24+ or Docker Desktop
- Docker Compose v2+

## Quick Start

```bash
git clone https://github.com/03aar/ultra-shell.git
cd ultra-shell/nucleus
cp .env.example .env
docker compose up --build -d
```

## Services

The `docker-compose.yml` starts these containers:

| Service | Port | Description |
|---------|------|-------------|
| `nucleus-core` | — | Shell runtime (Rust) |
| `nucleus-api` | 8080 | REST + WebSocket API (Go) |
| `nucleus-dashboard` | 3000 | Web UI (Next.js) |
| `nucleus-mcp` | — | MCP server (stdio, not network-exposed) |
| `postgres` | 5432 | PostgreSQL 15 |
| `redis` | 6379 | Redis 7 |

## Environment Variables

Key variables in `.env`:

```bash
# Required
DATABASE_URL=postgres://nucleus:nucleus@postgres:5432/nucleus?sslmode=disable
REDIS_URL=redis://redis:6379
API_PORT=8080

# Dev API key (always works in dev mode)
API_KEY=dev-nucleus-key-local

# Optional: AI provider keys
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=AI...
```

## Common Operations

```bash
# View logs for all services
docker compose logs -f

# View logs for a specific service
docker compose logs -f nucleus-api

# Restart a single service
docker compose restart nucleus-api

# Stop everything
docker compose down

# Stop and remove volumes (full reset)
docker compose down -v

# Rebuild after code changes
docker compose up --build -d
```

## Install the CLI

The `nuc` CLI runs on your host machine and talks to the API:

```bash
make install-cli
# or manually:
cd nucleus-cli && go build -o nuc . && sudo mv nuc /usr/local/bin/
```

Verify:

```bash
nuc doctor
```

## Load Demo Data

```bash
make seed
```

This populates the database with sample sessions, executions, and skills.

## Production Considerations

For production deployments:

1. **Change the API key.** Do not use `dev-nucleus-key-local` in production.
2. **Use a managed PostgreSQL** instance with proper credentials.
3. **Use a managed Redis** instance or Redis Cluster.
4. **Set up TLS** with a reverse proxy (nginx, Caddy, Traefik).
5. **Mount persistent volumes** for PostgreSQL and sled data.

Example production volume mounts:

```yaml
volumes:
  postgres-data:
    driver: local
  nucleus-data:
    driver: local

services:
  postgres:
    volumes:
      - postgres-data:/var/lib/postgresql/data
  nucleus-core:
    volumes:
      - nucleus-data:/data/sled
```

## Building Individual Images

```bash
# Build just the API
docker build -t nucleus-api -f nucleus-api/Dockerfile .

# Build just the dashboard
docker build -t nucleus-dashboard -f nucleus-dashboard/Dockerfile .
```

## Health Checks

```bash
# API health
curl http://localhost:8080/health

# PostgreSQL
docker compose exec postgres pg_isready

# Redis
docker compose exec redis redis-cli ping
```
