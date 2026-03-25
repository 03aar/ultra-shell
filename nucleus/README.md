```
 ███╗   ██╗██╗   ██╗ ██████╗██╗     ███████╗██╗   ██╗███████╗
 ████╗  ██║██║   ██║██╔════╝██║     ██╔════╝██║   ██║██╔════╝
 ██╔██╗ ██║██║   ██║██║     ██║     █████╗  ██║   ██║███████╗
 ██║╚██╗██║██║   ██║██║     ██║     ██╔══╝  ██║   ██║╚════██║
 ██║ ╚████║╚██████╔╝╚██████╗███████╗███████╗╚██████╔╝███████║
 ╚═╝  ╚═══╝ ╚═════╝  ╚═════╝╚══════╝╚══════╝ ╚═════╝ ╚══════╝
```

<p align="center">
  <strong>The AI-native shell runtime. Execute, observe, rollback, orchestrate.</strong>
</p>

<p align="center">
  <a href="https://github.com/03aar/ultra-shell/actions/workflows/ci.yml"><img src="https://github.com/03aar/ultra-shell/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://www.npmjs.com/package/@nucleus/mcp-server"><img src="https://img.shields.io/npm/v/@nucleus/mcp-server?color=cb3837&label=npm" alt="npm"></a>
  <a href="https://pypi.org/project/nucleus-sdk/"><img src="https://img.shields.io/pypi/v/nucleus-sdk?color=3775a9&label=PyPI" alt="PyPI"></a>
  <a href="https://github.com/03aar/ultra-shell/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://discord.gg/nucleus"><img src="https://img.shields.io/discord/1234567890?color=5865F2&label=Discord&logo=discord&logoColor=white" alt="Discord"></a>
  <a href="https://github.com/03aar/ultra-shell"><img src="https://img.shields.io/github/stars/03aar/ultra-shell?style=social" alt="GitHub Stars"></a>
</p>

---

## What is Nucleus?

Nucleus is a **drop-in shell runtime** that intercepts every command, evaluates risk before execution, tracks file mutations, and enables instant rollback. It turns your terminal into a structured execution substrate that AI agents, human developers, and MCP clients can all operate through the same interface.

Think of it as **git for shell sessions, with an AI co-pilot built in**. Every command becomes a node in a dependency graph. Every side effect is recorded. Every mistake is reversible. Your shell finally has an undo button -- and an API.

---

## Install

```bash
curl -fsSL https://install.nucleusshell.dev | bash
```

Or clone and run locally:

```bash
git clone https://github.com/03aar/ultra-shell.git
cd ultra-shell/nucleus
cp .env.example .env
make dev
```

Dashboard at [localhost:3000](http://localhost:3000) | API at [localhost:8080](http://localhost:8080) | API key: `dev-nucleus-key-local`

---

## For Developers

Nucleus evaluates risk **before** anything runs. Dangerous commands get flagged, explained, and require confirmation.

```bash
$ nuc exec "rm -rf /tmp/project"
  RISK: high (recursive delete, 847 files affected)
  Reversible: yes (snapshot created)
  Proceed? [y/N]
```

Made a mistake? Roll it back:

```bash
$ nuc rollback last
  Restored 847 files from snapshot s_a1b2c3d4
  Execution exec_5678 rolled back successfully.
```

Plan complex multi-step workflows and execute them safely:

```bash
$ nuc plan "set up a Node.js project with tests, Docker, and CI"
  Step 1/6: mkdir -p src tests          (risk: low)
  Step 2/6: npm init -y                 (risk: low)
  Step 3/6: npm install jest typescript (risk: low)
  Step 4/6: Generate Dockerfile         (risk: low)
  Step 5/6: Generate .github/workflows  (risk: low)
  Step 6/6: git init && git add .       (risk: low)
  Execute all? [y/N]
```

---

## For AI Agents

Any agent that speaks HTTP or MCP can drive Nucleus. Execute commands, read context, inspect the dependency graph, and roll back -- all through structured APIs.

**REST API:**

```bash
curl -X POST http://localhost:8080/api/v1/agent/execute \
  -H "X-API-Key: $NUCLEUS_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command": "ls -la", "require_approval": false}'
```

**Python SDK:**

```python
from nucleus import NucleusClient

nuc = NucleusClient(api_key="your-key")
result = nuc.execute("npm install express")
print(result.risk_level)   # "low"
print(result.exit_code)    # 0
nuc.rollback(result.id)    # undo
```

**TypeScript SDK:**

```typescript
import { Nucleus } from "@nucleus/sdk";

const nuc = new Nucleus({ apiKey: "your-key" });
const result = await nuc.execute("docker build -t app .");
console.log(result.riskLevel); // "medium"
await nuc.rollback(result.id);
```

---

## For Claude Desktop

Nucleus ships a fully compliant MCP server. One command connects it to Claude Desktop, Cursor, Zed, or any MCP client.

```bash
nuc mcp install --client claude-desktop
# Restart Claude Desktop — 10 Nucleus tools are now available
```

Claude can then execute shell commands, evaluate risk, inspect execution graphs, run skills, and roll back -- all through the MCP protocol with no custom integration needed.

---

## Architecture

```
                          ┌──────────────────────────────────────────┐
                          │          nucleus-dashboard               │  :3000
                          │     Next.js 14 · ReactFlow · xterm.js   │
                          └──────────────┬───────────────────────────┘
                                         │
         ┌───────────────┬───────────────┼───────────────┐
         │               │               │               │
  ┌──────┴──────┐ ┌──────┴──────┐ ┌──────┴──────┐ ┌─────┴────────┐
  │  nuc CLI    │ │  MCP Server │ │   REST API  │ │ Orchestrator │
  │  (Go)       │ │  (TS/stdio) │ │  (Go) :8080 │ │ (Python)     │
  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └─────┬────────┘
         │               │               │               │
         └───────────────┴───────┬───────┴───────────────┘
                                 │
                    ┌────────────┴────────────┐
                    │      nucleus-core       │
                    │  Rust · PTY · sled · DAG│
                    └─────┬──────────┬────────┘
                          │          │
                   ┌──────┴──┐ ┌────┴────────┐
                   │ Redis 7 │ │ PostgreSQL 15│
                   └─────────┘ └─────────────┘
```

---

## Components

| Component | Language | Purpose |
|-----------|----------|---------|
| `nucleus-core` | Rust | PTY interception, command parsing, risk evaluation, execution DAG, rollback engine |
| `nucleus-api` | Go | REST + WebSocket API, agent endpoints, skills, session management |
| `nucleus-mcp` | TypeScript | MCP server -- 10 tools, 4 resources, 3 prompts |
| `nucleus-cli` | Go | `nuc` CLI with 20+ commands |
| `nucleus-orchestrator` | Python | Multi-provider agent engine (Claude, GPT, Gemini, Ollama) |
| `nucleus-dashboard` | Next.js 14 | Web UI with 6 pages, real-time streaming, dark theme |
| `nucleus-sdk/python` | Python | Python client library |
| `nucleus-sdk/typescript` | TypeScript | TypeScript client library |
| `nucleus-sdk/go` | Go | Go client library |

---

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/executions` | GET | List executions with filtering and pagination |
| `/api/v1/executions/:id` | GET | Full execution detail with output |
| `/api/v1/executions/:id/context` | GET | Execution with dependency context |
| `/api/v1/context` | GET | Current environment state |
| `/api/v1/context/graph` | GET | Session dependency DAG |
| `/api/v1/rollback/:id` | POST | Rollback a specific execution |
| `/api/v1/sessions` | GET/POST | Session CRUD |
| `/api/v1/sessions/:id/replay` | GET | Replay a session step by step |
| `/api/v1/agent/execute` | POST | Execute command with risk evaluation |
| `/api/v1/agent/plan` | POST | Generate multi-step plan for a goal |
| `/api/v1/skills` | GET | List available skills |
| `/api/v1/skills/:name/run` | POST | Run a skill by name |
| `/api/v1/ws/stream` | WS | Real-time execution event stream |

---

## MCP Tools

| Tool | Description |
|------|-------------|
| `execute_command` | Run a shell command with risk evaluation and approval flow |
| `get_context` | Retrieve current environment state (cwd, env, git, etc.) |
| `get_execution_history` | List recent command executions with output |
| `rollback_execution` | Undo a specific execution by ID |
| `get_execution_graph` | Visualize the dependency DAG for a session |
| `run_skill` | Execute a pre-defined workflow (git_cleanup, deploy_check, etc.) |
| `search_history` | Search past commands by pattern, date, or risk level |
| `get_risk_assessment` | Dry-run risk evaluation without executing |
| `create_session` | Create a new named session |
| `watch_stream` | Subscribe to live execution events |

---

## Built-in Skills

| Skill | Description |
|-------|-------------|
| `git_cleanup` | Clean merged branches, prune remotes |
| `docker_cleanup` | Remove dangling images, stopped containers, unused volumes |
| `project_setup` | Scaffold a new project (node, python, rust, go) |
| `deploy_check` | Pre-deployment validation (tests, lint, env vars, ports) |
| `env_audit` | Scan for exposed secrets in environment and files |
| `process_debug` | Diagnose high CPU/memory processes, open ports, disk usage |

---

## Platform Support

| Platform | Docker | Local Build | CLI | MCP Server |
|----------|--------|-------------|-----|------------|
| macOS (Apple Silicon) | Yes | Yes | Yes | Yes |
| macOS (Intel) | Yes | Yes | Yes | Yes |
| Linux x86_64 | Yes | Yes | Yes | Yes |
| Linux ARM64 | Yes | Yes | Yes | Yes |
| Windows 10/11 (WSL2) | Yes | Yes | Yes | Yes |
| Windows (native) | Yes | -- | Yes | Yes |

---

## Make Targets

```bash
make help              # Show all targets
make dev               # Start all services
make dev-logs          # Follow logs
make build             # Build everything locally
make test              # Run all tests
make test-integration  # Docker-based integration tests
make install           # Install CLI + MCP server
make db-seed           # Load demo data
make doctor            # Check prerequisites and health
make format            # Format all source code
make lint              # Run all linters
make release VERSION=1.0.0  # Tag a release
make clean             # Remove build artifacts
make reset             # Destroy all data volumes
```

---

## Contributing

1. Fork the repo
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Run tests: `make test`
4. Format code: `make format`
5. Push and open a PR

All PRs run through CI automatically. Please include tests for new functionality and keep commits focused.

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## Documentation

- [Architecture](docs/architecture.md)
- [MCP Integration](docs/mcp-integration.md)
- [CLI Reference](docs/cli-reference.md)
- [Agent API](docs/agent-api.md)
- [Skills](docs/skills.md)
- [AI Providers](docs/providers.md)

---

## License

MIT -- see [LICENSE](LICENSE) for details.

Built by [@03aar](https://github.com/03aar) and contributors.
