---
sidebar_position: 1
title: MCP Overview
---

# MCP Integration

Nucleus includes a fully compliant [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server. MCP is an open standard that lets AI clients — Claude Desktop, Cursor, Zed, Continue — connect to external tool servers. With the Nucleus MCP server, any MCP-compatible AI assistant gets full access to your shell.

## What This Means

When you connect Claude Desktop to Nucleus via MCP, Claude can:

- **Execute shell commands** with risk evaluation and rollback.
- **Read your environment** — working directory, git status, running processes, environment variables.
- **Search your command history** across sessions.
- **Roll back mistakes** by restoring file snapshots.
- **Run skills** — pre-built workflows like git cleanup or project setup.
- **Subscribe to live events** — watch commands execute in real time.

All of this happens through a structured protocol, not fragile text parsing.

## Architecture

```
┌──────────────────┐     stdio      ┌──────────────┐     HTTP      ┌──────────────┐
│  Claude Desktop  │ ◄────────────► │  nucleus-mcp │ ◄───────────► │  nucleus-api │
│  Cursor / Zed    │                │  (Node.js)   │               │  (Go :8080)  │
└──────────────────┘                └──────────────┘               └──────────────┘
```

The MCP server (`nucleus-mcp`) communicates with AI clients via stdio (standard input/output) and translates MCP requests into Nucleus API calls.

## MCP Capabilities

The Nucleus MCP server exposes three types of MCP primitives:

### Tools

Tools are actions the AI can invoke. See [Tools Reference](/docs/mcp/tools-reference) for complete documentation.

| Tool | Description |
|------|-------------|
| `execute_command` | Run a shell command with risk evaluation |
| `get_context` | Get current environment state |
| `get_execution_history` | Get recent command history |
| `rollback_execution` | Undo a previous command |
| `get_execution_graph` | Get the execution DAG |
| `run_skill` | Execute a pre-defined workflow |
| `search_history` | Full-text search past commands |
| `get_risk_assessment` | Dry-run risk evaluation |
| `create_session` | Create a named session |
| `watch_stream` | Subscribe to live execution events |

### Resources

Resources are data the AI can read:

| URI | Description |
|-----|-------------|
| `nucleus://context` | Current environment snapshot (cwd, git, env vars, processes) |
| `nucleus://graph` | Session execution graph in Mermaid format |
| `nucleus://history` | Last 50 executions with metadata |
| `nucleus://sessions` | All sessions list |

### Prompts

Prompts are pre-built conversation starters:

| Prompt | Description |
|--------|-------------|
| `debug_failure` | Analyze why a command failed |
| `plan_workflow` | Plan a multi-step workflow for a goal |
| `explain_session` | Summarize what happened in a session |

## Quick Setup

```bash
nuc mcp install --client claude-desktop
# Restart Claude Desktop
```

See [Claude Desktop Setup](/docs/mcp/claude-desktop) for detailed instructions.

## Supported Clients

| Client | Status | Setup Command |
|--------|--------|---------------|
| Claude Desktop | Fully supported | `nuc mcp install --client claude-desktop` |
| Cursor | Fully supported | `nuc mcp install --client cursor` |
| Zed | Supported | Manual configuration |
| Continue | Supported | Manual configuration |

## Security

The MCP server authenticates to the Nucleus API using the configured API key. The AI client does not have direct access to the shell — all commands pass through the Nucleus risk evaluator.

- Commands scored as `critical` risk are blocked.
- Sensitive environment variables (matching `SECRET`, `TOKEN`, `KEY`, `PASSWORD` patterns) are masked.
- All executions are logged and auditable.
