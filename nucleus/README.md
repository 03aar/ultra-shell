# NUCLEUS

**AI-native shell runtime, agent orchestration platform, and MCP server.**

NUCLEUS is three things unified into one platform:

1. **Shell Runtime** — Drop-in replacement for bash/zsh that intercepts commands, evaluates risk, tracks file mutations, enables rollback, and persists memory.
2. **Agent Orchestration** — Structured execution substrate for Claude, GPT, Gemini, and Ollama. Agents get typed results, can plan/execute/observe/rollback through a single API.
3. **MCP Server** — Fully compliant Model Context Protocol server. Any MCP client (Claude Desktop, Cursor, Zed) can use Nucleus as its shell backend.

## Architecture

```
┌─────────────────────────────────────────────────┐
│              nucleus-dashboard                    │  :3000
│         Next.js 14 · ReactFlow · xterm.js        │
├─────────────────────────────────────────────────┤
│  nucleus-cli (nuc)  │  nucleus-orchestrator      │
│  Go CLI tool        │  Python agent engine        │
├─────────────────────┼───────────────────────────┤
│  nucleus-mcp        │     nucleus-api             │  :8080
│  MCP server (stdio) │  Go REST + WebSocket        │
├─────────────────────┴───────────────────────────┤
│                  nucleus-core                     │
│              Rust · portable-pty · sled           │
├──────────────┬──────────────────────────────────┤
│    Redis 7   │         PostgreSQL 15             │
└──────────────┴──────────────────────────────────┘
```

## Quickstart

```bash
git clone https://github.com/03aar/ultra-shell.git
cd ultra-shell/nucleus
cp .env.example .env
# Add ANTHROPIC_API_KEY and/or OPENAI_API_KEY to .env (optional)
docker compose up --build -d
```

- **Dashboard**: http://localhost:3000
- **API**: http://localhost:8080
- **Health**: http://localhost:8080/health
- **API Key**: `dev-nucleus-key-local`

```bash
make seed           # Load demo data
make logs           # Watch logs
```

## Install CLI

```bash
make install-cli
nuc doctor          # Verify installation
nuc shell           # Start Nucleus shell
```

## Connect to Claude Desktop (MCP)

```bash
nuc mcp install --client claude-desktop
# Restart Claude Desktop — Nucleus tools now available
```

## Run AI Agent

```bash
# Via CLI
nuc plan "set up a Node.js project with tests and Docker"

# Via Python orchestrator
pip install -e nucleus-orchestrator/
nucleus-agent run "set up CI/CD for this project" --provider claude
```

## Components

| Component | Language | Purpose |
|-----------|----------|---------|
| `nucleus-core` | Rust | PTY interception, command parsing, risk evaluation, execution DAG, rollback |
| `nucleus-api` | Go | REST + WebSocket API, agent endpoints, skills, session management |
| `nucleus-mcp` | TypeScript | MCP server (10 tools, 4 resources, 3 prompts) |
| `nucleus-cli` | Go | `nuc` CLI (20+ commands) |
| `nucleus-orchestrator` | Python | Agent engine (Claude, GPT, Gemini, Ollama) |
| `nucleus-dashboard` | Next.js 14 | Web UI (6 pages, real-time, dark theme) |
| `nucleus-sdk/python` | Python | Python client library |
| `nucleus-sdk/typescript` | TypeScript | TypeScript client library |
| `nucleus-sdk/go` | Go | Go client library |

## MCP Tools

| Tool | Description |
|------|-------------|
| `execute_command` | Run shell command with risk evaluation |
| `get_context` | Current environment state |
| `get_execution_history` | Recent command history |
| `rollback_execution` | Undo a command |
| `get_execution_graph` | Dependency DAG |
| `run_skill` | Execute pre-defined workflow |
| `search_history` | Search past commands |
| `get_risk_assessment` | Dry-run risk check |
| `create_session` | New named session |
| `watch_stream` | Live event subscription |

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/executions` | GET | List executions |
| `/api/v1/executions/:id` | GET | Execution detail |
| `/api/v1/executions/:id/context` | GET | Execution with dependency context |
| `/api/v1/context` | GET | Environment state |
| `/api/v1/context/graph` | GET | Session DAG |
| `/api/v1/rollback/:id` | POST | Rollback execution |
| `/api/v1/sessions` | GET/POST | Sessions CRUD |
| `/api/v1/sessions/:id/replay` | GET | Session replay |
| `/api/v1/agent/execute` | POST | Agent command execution |
| `/api/v1/agent/plan` | POST | Goal planning |
| `/api/v1/skills` | GET | List skills |
| `/api/v1/skills/:name/run` | POST | Run skill |
| `/api/v1/ws/stream` | WS | Real-time events |

## Built-in Skills

| Skill | Description |
|-------|-------------|
| `git_cleanup` | Clean merged branches |
| `docker_cleanup` | Remove unused Docker resources |
| `project_setup` | Scaffold new project (node/python/rust/go) |
| `deploy_check` | Pre-deployment validation |
| `env_audit` | Secret exposure audit |
| `process_debug` | Process diagnostics |

## Platform Support

| Platform | Docker | Local Build | CLI |
|----------|--------|-------------|-----|
| macOS (Intel/ARM) | Yes | Yes | Yes |
| Linux x86_64/ARM64 | Yes | Yes | Yes |
| Ubuntu 20.04+ | Yes | Yes | Yes |
| Windows 10/11 | Yes | Yes | Yes |
| WSL2 | Yes | Yes | Yes |

## Make Targets

```bash
make dev             # Start all services
make seed            # Load demo data
make build           # Build all binaries
make test            # Run all tests
make install-cli     # Install nuc CLI
make install-mcp     # Build MCP server
make stop            # Stop services
make logs            # Follow logs
make clean           # Remove artifacts
make reset           # Remove all data
make setup           # Check prerequisites
```

## Documentation

- [Architecture](docs/architecture.md)
- [MCP Integration](docs/mcp-integration.md)
- [CLI Reference](docs/cli-reference.md)
- [Agent API](docs/agent-api.md)
- [Skills](docs/skills.md)
- [AI Providers](docs/providers.md)

## License

MIT
