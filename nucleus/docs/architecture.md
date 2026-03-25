# NUCLEUS Architecture

## System Overview

NUCLEUS is a three-layer system:

1. **nucleus-core** (Rust) — The shell runtime. Intercepts PTY I/O, parses commands, evaluates risk, builds execution DAG, manages snapshots for rollback.
2. **nucleus-api** (Go) — The API server. REST + WebSocket endpoints for executions, context, sessions, rollback, agent commands, and skills.
3. **nucleus-dashboard** (Next.js) — The web UI. Real-time execution feed, DAG visualization, session replay, agent console, API explorer.

Supporting components:
- **nucleus-mcp** (TypeScript) — MCP server for Claude Desktop/Cursor/Zed integration
- **nucleus-cli** (Go) — `nuc` CLI tool for terminal interaction
- **nucleus-orchestrator** (Python) — Agent orchestration engine for Claude/GPT/Gemini/Ollama
- **SDKs** — Python, TypeScript, Go client libraries

## Data Flow

```
User types command
    ↓
nucleus-core intercepts via PTY
    ↓
parser.rs → ParsedCommand
    ↓
evaluator.rs → RiskAssessment
    ↓ (if not blocked)
snapshot.rs → pre-execution file snapshot
    ↓
Command executes in child shell
    ↓
Output captured, execution node created
    ↓
graph.rs → DAG node + edges stored in sled
    ↓
redis.rs → Published to Redis pub/sub
    ↓
nucleus-api → Receives via Redis subscriber
    ↓
WebSocket hub → Broadcasts to all connected clients
    ↓
nucleus-dashboard → Real-time UI update
```

## Database Architecture

- **sled** (embedded in nucleus-core): Execution nodes, session metadata, file snapshots, graph edges. Zero-config, persists across restarts.
- **PostgreSQL**: Sessions, API keys, audit logs, rollback events, skills, agent sessions. Schema in `nucleus-api/database/init.sql`.
- **Redis**: Real-time pub/sub between nucleus-core and nucleus-api. Channels: `nucleus:executions`, `nucleus:rollbacks`, `nucleus:sessions`, `nucleus:warnings`.

## Security Model

- API keys stored hashed in PostgreSQL
- Dev key `dev-nucleus-key-local` always works in dev mode
- Sensitive env vars masked in API responses (SECRET, TOKEN, KEY, PASSWORD patterns)
- Risk evaluator blocks `rm -rf /` and similar destructive commands
- File snapshots limited to 100MB per file
