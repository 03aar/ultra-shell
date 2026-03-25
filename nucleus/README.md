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

## Installation

### Quick Start (Docker — all platforms)

**Prerequisites:** [Docker Desktop](https://docs.docker.com/get-docker/) (Mac, Windows, Linux)

```bash
git clone https://github.com/03aar/ultra-shell.git
cd ultra-shell/nucleus
docker compose up --build -d
```

Then open:
- **Dashboard**: http://localhost:3000
- **API**: http://localhost:8080
- **API Health**: http://localhost:8080/health
- **API Key**: `dev-nucleus-key-local`

### macOS

```bash
# Install Docker if not present
brew install --cask docker

# Clone and start
git clone https://github.com/03aar/ultra-shell.git
cd ultra-shell/nucleus
./scripts/install.sh
```

**Optional — install CLI locally:**
```bash
brew install rust
cd nucleus-core && cargo build --release
sudo cp target/release/nucleus /usr/local/bin/
```

### Linux / Ubuntu / Debian

```bash
# Install Docker if not present
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker

# Clone and start
git clone https://github.com/03aar/ultra-shell.git
cd ultra-shell/nucleus
./scripts/install.sh
```

**Optional — install CLI locally:**
```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
cd nucleus-core && cargo build --release
sudo cp target/release/nucleus /usr/local/bin/
```

### Windows

```powershell
# Install Docker Desktop from https://docs.docker.com/desktop/install/windows-install/

# Clone and start
git clone https://github.com/03aar/ultra-shell.git
cd ultra-shell\nucleus
.\scripts\install.ps1
```

**Optional — install CLI locally:**
```powershell
# Install Rust from https://rustup.rs
cd nucleus-core
cargo build --release
copy target\release\nucleus.exe C:\Windows\System32\nucleus.exe
```

### Using Make

```bash
make setup     # Check prerequisites
make dev       # Start all services (Docker)
make build     # Build all binaries locally (requires Rust, Go, Node.js)
make test      # Run all tests
make install   # Install nucleus CLI to /usr/local/bin
make stop      # Stop all services
make logs      # Follow service logs
make clean     # Remove build artifacts and volumes
```

## Configuration

Copy `.env.example` to `.env` to customize:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|----------|---------|-------------|
| `NUCLEUS_REDIS_URL` | `redis://localhost:6379` | Redis connection URL |
| `DATABASE_URL` | `postgres://nucleus:...@localhost:5432/nucleus` | PostgreSQL connection |
| `PORT` | `8080` | API server port |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080/api/v1` | Dashboard API URL |
| `NEXT_PUBLIC_WS_URL` | `ws://localhost:8080/api/v1` | Dashboard WebSocket URL |
| `NEXT_PUBLIC_API_KEY` | `dev-nucleus-key-local` | API key for dashboard |
| `NUCLEUS_DATA_DIR` | `~/.nucleus/data` | Sled database directory |
| `RUST_LOG` | `info` | Rust log level |

## Components

### nucleus-core (Rust)

The PTY interception layer and context engine:

- **PTY Multiplexer** — Forks a child shell (bash/zsh/fish on Unix, cmd.exe on Windows), intercepts all I/O
- **Command Parser** — Parses commands into structured data: binary, args, flags, pipes, redirections, env assignments. Classifies into 10 categories (filesystem, network, git, docker, build, etc.)
- **Risk Evaluator** — Pre-execution risk analysis. Blocks `rm -rf /`, warns on `chmod 777`, `git push --force`, `kill -9`, `dd`. Tracks repeated failures.
- **Execution Graph** — DAG of all executions stored in sled. Tracks file, env, and process dependencies.
- **Session Memory** — Persistent across restarts. Stores metadata, env snapshots, failure patterns.
- **Rollback Engine** — Pre-execution file snapshots. Restore files and env vars to pre-command state.
- **Redis IPC** — Publishes events to Redis pub/sub for real-time streaming.

### nucleus-api (Go)

REST + WebSocket API server (cross-platform: Linux, macOS, Windows):

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/executions` | GET | List executions (filterable by session, risk, category) |
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

Design system: `#000000` bg, `#0a0a0a` surface, `#00ff88` accent. JetBrains Mono / Syne / DM Sans.

### SDKs

**Python:**
```python
from nucleus import NucleusClient

client = NucleusClient("http://localhost:8080/api/v1", "dev-nucleus-key-local")
executions = client.get_executions(limit=10)
context = client.get_context()
result = client.rollback("execution-uuid")

# Real-time streaming
client.stream(lambda event_type, data: print(f"{event_type}: {data}"))
```

**TypeScript:**
```typescript
import { NucleusClient } from 'nucleus-sdk';

const client = new NucleusClient('http://localhost:8080/api/v1', 'dev-nucleus-key-local');
const executions = await client.getExecutions({ limit: 10 });
const context = await client.getContext();
const ws = client.stream((msg) => console.log(msg));
```

## Database

**PostgreSQL tables:**
- `sessions` — Shell session metadata
- `api_keys` — API key hashes for auth
- `audit_logs` — Action audit trail
- `rollback_events` — Rollback history

**Embedded sled DB** (nucleus-core): Execution nodes, session metadata, file snapshots, graph edges

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Shell Runtime | Rust, portable-pty, sled |
| API Server | Go, Gin, gorilla/websocket |
| Dashboard | Next.js 14, React, TypeScript, ReactFlow, xterm.js, Tailwind CSS |
| Database | PostgreSQL 15, sled (embedded), Redis 7 |
| DevOps | Docker Compose, Make |

## Platform Support

| Platform | Docker | Local Build | CLI |
|----------|--------|-------------|-----|
| macOS (Intel/ARM) | Yes | Yes | Yes |
| Linux (x86_64/ARM64) | Yes | Yes | Yes |
| Ubuntu 20.04+ | Yes | Yes | Yes |
| Debian 11+ | Yes | Yes | Yes |
| Windows 10/11 | Yes | Yes | Yes |
| WSL2 | Yes | Yes | Yes |

## License

MIT
