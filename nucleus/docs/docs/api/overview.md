---
sidebar_position: 1
title: API Overview
---

# REST API Overview

The Nucleus API (`nucleus-api`) is a Go HTTP server running on port 8080 that exposes every Nucleus capability over REST and WebSocket.

## Base URL

```
http://localhost:8080
```

All API routes are prefixed with `/api/v1/`.

## Authentication

Every request requires an `X-Nucleus-Key` header:

```bash
curl http://localhost:8080/api/v1/context \
  -H "X-Nucleus-Key: dev-nucleus-key-local"
```

In development, the key `dev-nucleus-key-local` always works. See [Authentication](/docs/api/authentication) for production key management.

## Response Format

All responses follow a consistent JSON envelope:

```json
{
  "data": { ... },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

Error responses:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "command field is required",
    "request_id": "uuid"
  }
}
```

## Endpoints at a Glance

### Executions

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/agent/execute` | Execute a command |
| `GET` | `/api/v1/executions` | List recent executions |
| `GET` | `/api/v1/executions/:id` | Get execution detail |
| `POST` | `/api/v1/agent/plan` | Generate execution plan |

### Context

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/context` | Get environment state |

### Rollback

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/rollback/:id` | Rollback an execution |
| `GET` | `/api/v1/rollback/:id/preview` | Preview rollback without executing |

### Sessions

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/sessions` | List all sessions |
| `POST` | `/api/v1/sessions` | Create a new session |
| `GET` | `/api/v1/sessions/:id` | Get session detail |
| `GET` | `/api/v1/sessions/:id/replay` | Get session replay data |

### Skills

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/skills` | List all skills |
| `GET` | `/api/v1/skills/:name` | Get skill detail |
| `POST` | `/api/v1/skills/:name/run` | Run a skill |

### Graph

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/graph` | Get execution DAG |
| `GET` | `/api/v1/graph/mermaid` | Get DAG as Mermaid diagram |

### WebSocket

| Path | Description |
|------|-------------|
| `/ws` | Real-time event stream |

### Health

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check (no auth required) |

## Rate Limiting

The API does not enforce rate limits in development mode. In production, configure rate limiting through your reverse proxy (nginx, Caddy, etc.).

## CORS

CORS is enabled for `localhost` origins in development. Configure allowed origins via the `CORS_ORIGINS` environment variable:

```bash
CORS_ORIGINS=https://dashboard.example.com,https://admin.example.com
```

## Detailed Guides

- [Authentication](/docs/api/authentication) — API key management.
- [Executions](/docs/api/executions) — Execute commands and query history.
- [Context](/docs/api/context) — Read environment state.
- [Rollback](/docs/api/rollback) — Undo command side effects.
- [Sessions](/docs/api/sessions) — Manage execution sessions.
- [WebSocket](/docs/api/websocket) — Real-time event streaming.
