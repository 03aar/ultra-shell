---
sidebar_position: 6
title: Sessions
---

# Sessions API

Sessions group related command executions together. They provide a way to organize work by task, project, or time period.

## List Sessions

```
GET /api/v1/sessions
```

**Query parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | 20 | Number of results |
| `offset` | int | 0 | Pagination offset |
| `active` | boolean | — | Filter active/inactive sessions |

**Response:**

```json
{
  "data": {
    "sessions": [
      {
        "id": "uuid",
        "name": "deploy-v2.1",
        "command_count": 23,
        "started_at": "2025-01-15T08:00:00Z",
        "last_activity": "2025-01-15T10:30:00Z",
        "active": true,
        "duration_seconds": 9000
      }
    ],
    "total": 45
  }
}
```

## Create a Session

```
POST /api/v1/sessions
```

**Request:**

```json
{
  "name": "deploy-v2.1"
}
```

**Response:**

```json
{
  "data": {
    "id": "uuid",
    "name": "deploy-v2.1",
    "created_at": "2025-01-15T10:30:00Z",
    "active": true
  }
}
```

## Get Session Detail

```
GET /api/v1/sessions/:id
```

Returns the session metadata and its execution history.

**Response:**

```json
{
  "data": {
    "id": "uuid",
    "name": "deploy-v2.1",
    "started_at": "2025-01-15T08:00:00Z",
    "last_activity": "2025-01-15T10:30:00Z",
    "active": true,
    "command_count": 23,
    "executions": [
      {
        "id": "uuid",
        "command": "git pull origin main",
        "exit_code": 0,
        "duration_ms": 2100,
        "risk_level": "low",
        "timestamp": "2025-01-15T08:01:00Z"
      }
    ]
  }
}
```

## Session Replay

```
GET /api/v1/sessions/:id/replay
```

Returns the full terminal recording for the session, suitable for playback with xterm.js.

**Response:**

```json
{
  "data": {
    "session_id": "uuid",
    "format": "asciicast",
    "events": [
      {"time": 0.0, "type": "o", "data": "$ "},
      {"time": 0.5, "type": "i", "data": "git status\r\n"},
      {"time": 0.8, "type": "o", "data": "On branch main\nnothing to commit\n"}
    ]
  }
}
```

The `events` array follows the [asciicast v2](https://github.com/asciinema/asciinema/blob/develop/doc/asciicast-v2.md) format:
- `time`: seconds since session start
- `type`: `"i"` for input, `"o"` for output
- `data`: terminal data

## Sessions via CLI

```bash
# List sessions
nuc session list

# Create a new session
nuc session new "deploy-v2.1"

# Replay a session in the terminal
nuc session replay <session-id>
```

## Auto-Naming

By default, Nucleus automatically names sessions based on the first few commands. A session that starts with `git checkout -b feature/auth` followed by `npm test` might be named `feature/auth-testing`.

Disable this in `config.toml`:

```toml
[session]
auto_name = false
```
