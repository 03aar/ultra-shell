# NUCLEUS

**AI-native shell runtime that intercepts shell commands, builds a live context graph, and exposes a structured API for AI coding agents.**

NUCLEUS replaces raw bash/zsh as the interface layer for AI agents. It provides command interception, risk analysis, execution dependency tracking, file rollback, and a real-time dashboard — all out of the box.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    nucleus-dashboard                      │
│              Next.js 14 · React · TypeScript              │
│     ReactFlow DAG · xterm.js Terminal · Live WebSocket    │
│                     :3000                                 │
└──────────────────────┬──────────────────────────────────┘
                       │ REST + WebSocket
┌──────────────────────▼──────────────────────────────────┐
│                      nucleus-api                          │
│               Go · Gin · gorilla/websocket                │
│    Agent API · Sessions · Rollback · Auth                 │
│                     :8080                                 │
└──────────────────────┬──────────────────────────────────┘
                       │ Redis pub/sub
┌──────────────────────▼──────────────────────────────────┐
│                     nucleus-core                          │
│                    Rust · portable-pty                     │
│   PTY Interception · Command Parser · Risk Evaluator      │
│   Context Graph (sled) · Session Memory · Rollback        │
└─────────────────────────────────────────────────────────┘
          │                               │
     ┌────▼────┐                   ┌──────▼──────┐
     │  Redis   │                   │  PostgreSQL  │
     │  7       │                   │  15          │
     └─────────┘                   └──────────────┘
```

## Quickstart

```bash
git clone <repo-url> && cd nucleus
docker compose up --build
```

Then open:
- **Dashboard**: http://localhost:3000
- **API**: http://localhost:8080
- **API Health**: http://localhost:8080/health

Default API key: `dev-nucleus-key-local`

## Components

### nucleus-core (Rust)

The PTY interception layer and context engine:

- **PTY Multiplexer** — Forks a child shell (bash/zsh/fish), intercepts all I/O
- **Command Parser** — Parses commands into structured data: binary, args, flags, pipes, redirections, env assignments. Classifies by category (filesystem, network, git, docker, build, etc.)
- **Risk Evaluator** — Pre-execution risk analysis. Detects destructive commands (rm -rf /, chmod 777, kill -9), repeated failures, dangerous flags. Blocks critical operations.
- **Execution Graph** — DAG of all command executions stored in sled. Tracks file dependencies, env var dependencies, and process lifecycle edges.
- **Session Memory** — Persistent session data across restarts. Stores metadata, env snapshots, failure patterns.
- **Rollback Engine** — Pre-execution file snapshots. Restore files and env vars to pre-command state.
- **Redis IPC** — Publishes execution events, rollback events, and session events to Redis pub/sub channels.

### nucleus-api (Go)

REST + WebSocket API server:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/executions` | GET | List executions (filterable) |
| `/api/v1/executions/:id` | GET | Full execution details |
| `/api/v1/executions/:id/context` | GET | Execution + dependency context |
| `/api/v1/context` | GET | Current environment state |
| `/api/v1/context/graph` | GET | Full session DAG (reactflow compatible) |
| `/api/v1/rollback/:execution_id` | POST | Rollback an execution |
| `/api/v1/sessions` | GET/POST | List/create sessions |
| `/api/v1/sessions/:id/replay` | GET | Session replay data |
| `/api/v1/ws/stream` | WS | Real-time event stream |

Auth via `X-Nucleus-Key` header. Dev key: `dev-nucleus-key-local`

### nucleus-dashboard (Next.js 14)

Dark-theme dashboard with 5 views:

- **Dashboard** — Live command feed, stats bar, context panel, embedded terminal
- **Execution Graph** — ReactFlow DAG visualization with category-colored nodes
- **Sessions** — Session list, replay with timeline animation, JSON export
- **Environment** — Live env state, process table, file mutation tracker
- **API Explorer** — Interactive API tester with code snippets (curl/Python/TypeScript)

### SDKs

**Python:**
```python
from nucleus import NucleusClient

client = NucleusClient("http://localhost:8080/api/v1", "dev-nucleus-key-local")
executions = client.get_executions(limit=10)
context = client.get_context()
result = client.rollback("execution-uuid")
```

**TypeScript:**
```typescript
import { NucleusClient } from 'nucleus-sdk';

const client = new NucleusClient('http://localhost:8080/api/v1', 'dev-nucleus-key-local');
const executions = await client.getExecutions({ limit: 10 });
const context = await client.getContext();
const ws = client.stream((msg) => console.log(msg));
```

## Make Targets

```bash
make dev       # Start all services with Docker Compose
make build     # Build all production binaries
make test      # Run all tests
make install   # Install nucleus binary to /usr/local/bin
make stop      # Stop all services
make logs      # Follow service logs
make clean     # Remove build artifacts and volumes
```

## Database

PostgreSQL tables: `sessions`, `api_keys`, `audit_logs`, `rollback_events`

Embedded sled DB (in nucleus-core): execution nodes, session metadata, file snapshots, execution graph edges

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Shell Runtime | Rust, portable-pty, sled |
| API Server | Go, Gin, gorilla/websocket |
| Dashboard | Next.js 14, React, TypeScript, ReactFlow, xterm.js, Tailwind CSS |
| Database | PostgreSQL 15, sled (embedded), Redis 7 |
| DevOps | Docker Compose, Make |

## License

MIT
