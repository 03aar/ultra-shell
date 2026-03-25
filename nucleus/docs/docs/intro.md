---
sidebar_position: 1
slug: /intro
title: What is Nucleus?
---

# What is Nucleus?

Nucleus is an **AI-native shell runtime** that sits between you and your operating system. It wraps your existing shell (bash, zsh, fish) via PTY interception, turning every command you run into a tracked, reversible, AI-accessible operation.

## The Problem

Terminal sessions are **stateless, opaque, and irreversible**.

- You run `rm -rf` on the wrong directory and there is no undo.
- An AI agent executes shell commands with zero guardrails or context awareness.
- Your past commands vanish into `.bash_history` with no structure, no metadata, no relationships.
- There is no API to let an AI assistant observe, plan, and execute shell workflows safely.

Nucleus fixes all of this.

## Three Layers

Nucleus is composed of three layers that work together:

### Layer 1: Shell Runtime (`nucleus-core`)

**Language:** Rust

The foundation. `nucleus-core` intercepts your terminal via a PTY (pseudo-terminal) wrapper. When you type a command:

1. **Parser** (`parser.rs`) breaks the command into structured data: binary name, arguments, pipes, redirects, environment variables.
2. **Risk Evaluator** (`evaluator.rs`) scores the command for destructive potential. Commands like `rm -rf /` are blocked. Risky commands are flagged with warnings.
3. **Snapshot Engine** (`snapshot.rs`) captures the state of files that will be modified, enabling rollback.
4. **Execution Engine** runs the command in a child shell process, capturing stdout, stderr, exit code, and timing.
5. **DAG Builder** (`graph.rs`) records the execution as a node in a directed acyclic graph stored in [sled](https://github.com/spacejam/sled), with edges connecting related commands.
6. **Publisher** (`redis.rs`) pushes the execution event to Redis pub/sub for real-time consumers.

The result: every command you run is parsed, evaluated, recorded, and reversible.

### Layer 2: API Server (`nucleus-api`)

**Language:** Go

The bridge between the shell runtime and everything else. `nucleus-api` provides:

- **REST API** for executions, context, sessions, rollback, agent commands, and skills.
- **WebSocket hub** for real-time streaming of execution events to dashboards and agents.
- **PostgreSQL** for persistent storage of sessions, API keys, audit logs, and agent data.
- **Redis subscriber** to receive events from `nucleus-core` in real time.

Every capability of Nucleus is accessible through the API. If you can do it in the shell, you can do it through HTTP.

### Layer 3: Dashboard (`nucleus-dashboard`)

**Language:** Next.js 14 with React

A real-time web interface that visualizes your shell:

- **Execution Feed** showing commands as they happen, with risk badges and timing.
- **DAG Visualization** using ReactFlow to render the execution graph.
- **Session Replay** using xterm.js to play back terminal sessions.
- **Agent Console** to interact with AI agents that have shell access.
- **API Explorer** for testing API endpoints directly from the browser.

## Supporting Components

| Component | Language | Purpose |
|-----------|----------|---------|
| `nucleus-mcp` | TypeScript | MCP server for Claude Desktop, Cursor, Zed |
| `nucleus-cli` | Go | `nuc` command-line tool |
| `nucleus-orchestrator` | Python | Agent orchestration for Claude/GPT/Gemini/Ollama |
| `nucleus-sdk` | Python, TS, Go | Client libraries for building on Nucleus |

## Who Is Nucleus For?

- **Developers** who want rollback, risk warnings, and structured history for their shell sessions.
- **DevOps/SRE teams** who need auditable, replayable terminal sessions with real-time monitoring.
- **AI agent builders** who need a safe, structured execution substrate for LLM-driven workflows.
- **MCP users** who want Claude Desktop or Cursor to have full shell access with guardrails.

## How It Compares

| Feature | Traditional Shell | Nucleus |
|---------|-------------------|---------|
| Command history | Flat text file | Structured DAG with metadata |
| Rollback | Not possible | One-command file restoration |
| Risk evaluation | None | Automatic per-command scoring |
| AI access | Fragile wrapper scripts | Native API + MCP server |
| Session replay | Terminal recording tools | Built-in with xterm.js |
| Real-time monitoring | `tail -f` on logs | WebSocket + dashboard |

## Next Steps

- [Quickstart](/docs/quickstart) — Install and run Nucleus in 30 seconds.
- [Installation](/docs/installation/overview) — Detailed setup for your platform.
- [Shell Runtime](/docs/shell/overview) — Understand how the shell layer works.
- [MCP Integration](/docs/mcp/overview) — Connect Claude Desktop or Cursor.
