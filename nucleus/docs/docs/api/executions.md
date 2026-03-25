---
sidebar_position: 3
title: Executions
---

# Executions API

Execute shell commands and query execution history through the Nucleus API.

## Execute a Command

```
POST /api/v1/agent/execute
```

Execute a command with full risk evaluation, snapshotting, and DAG recording.

**Request body:**

```json
{
  "command": "npm test",
  "agent_id": "my-agent",
  "dry_run": false
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `command` | string | Yes | Shell command to execute |
| `agent_id` | string | No | Identifier for the calling agent |
| `dry_run` | boolean | No | If true, evaluate risk but do not execute |

**Response:**

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "command": {
      "raw": "npm test",
      "binary": "npm",
      "args": ["test"],
      "category": "build"
    },
    "exit_code": 0,
    "stdout": "Tests: 12 passed, 12 total\nTime: 3.2s",
    "stderr": "",
    "duration_ms": 3200,
    "risk_flags": [],
    "risk_level": "low",
    "rollback_available": false,
    "session_id": "uuid",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

**Dry run response** (when `dry_run: true`):

```json
{
  "data": {
    "command": {
      "raw": "rm -rf ./build",
      "binary": "rm",
      "args": ["-rf", "./build"],
      "category": "filesystem"
    },
    "risk_level": "high",
    "risk_score": 7,
    "risk_flags": ["recursive-delete"],
    "would_snapshot": true,
    "blocked": false
  }
}
```

**Error — command blocked:**

```json
{
  "error": {
    "code": "COMMAND_BLOCKED",
    "message": "Command blocked: rm -rf / is classified as critical risk",
    "risk_level": "critical",
    "risk_flags": ["root-filesystem-deletion"]
  }
}
```

## Generate an Execution Plan

```
POST /api/v1/agent/plan
```

Generate a multi-step plan for a goal. This uses the configured AI provider to break a goal into individual commands.

**Request:**

```json
{
  "goal": "Set up a new Node.js project with TypeScript and tests",
  "context": "Working in an empty directory",
  "agent_id": "claude"
}
```

**Response:**

```json
{
  "data": {
    "plan_id": "uuid",
    "goal": "Set up a new Node.js project with TypeScript and tests",
    "steps": [
      {
        "command": "npm init -y",
        "rationale": "Initialize package.json",
        "risk_level": "low",
        "reversible": true
      },
      {
        "command": "npm install -D typescript @types/node ts-node",
        "rationale": "Install TypeScript toolchain",
        "risk_level": "low",
        "reversible": true
      },
      {
        "command": "npx tsc --init",
        "rationale": "Create tsconfig.json",
        "risk_level": "low",
        "reversible": true
      },
      {
        "command": "npm install -D jest @types/jest ts-jest",
        "rationale": "Install test framework",
        "risk_level": "low",
        "reversible": true
      }
    ]
  }
}
```

## List Executions

```
GET /api/v1/executions
```

**Query parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | 20 | Number of results (max 100) |
| `offset` | int | 0 | Pagination offset |
| `session_id` | uuid | — | Filter by session |
| `search` | string | — | Full-text search |
| `risk_level` | string | — | Filter by risk level |

**Example:**

```bash
curl "http://localhost:8080/api/v1/executions?limit=5&risk_level=high" \
  -H "X-Nucleus-Key: dev-nucleus-key-local"
```

## Get Execution Detail

```
GET /api/v1/executions/:id
```

Returns the full execution record including stdout, stderr, risk assessment, snapshot info, and graph edges.

```bash
curl http://localhost:8080/api/v1/executions/550e8400-e29b-41d4-a716-446655440000 \
  -H "X-Nucleus-Key: dev-nucleus-key-local"
```
