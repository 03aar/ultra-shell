---
sidebar_position: 2
title: Quickstart
---

# Quickstart

Get Nucleus running in 30 seconds.

## Prerequisites

- **Docker** and **Docker Compose** (v2+)
- **Git**

## Install

```bash
# Clone the repository
git clone https://github.com/03aar/ultra-shell.git
cd ultra-shell/nucleus

# Copy environment config
cp .env.example .env

# Start everything
docker compose up --build -d
```

That's it. Nucleus is running.

## Verify

```bash
# Check that all services are healthy
make install-cli   # Install the nuc CLI
nuc doctor
```

`nuc doctor` checks connectivity to every component:

```
Nucleus Doctor
==============
  API Server (localhost:8080)    ✓ healthy
  PostgreSQL                     ✓ connected
  Redis                          ✓ connected
  Shell Runtime                  ✓ active
  MCP Server                     ✓ ready
  Dashboard (localhost:3000)     ✓ serving

All systems operational.
```

## Launch the Shell

```bash
nuc shell
```

You are now inside a Nucleus-powered shell. Every command you run is:

- **Parsed** into structured data (binary, args, pipes, redirects)
- **Risk-evaluated** before execution
- **Snapshotted** so modified files can be rolled back
- **Recorded** as a node in the execution DAG
- **Streamed** in real time to the dashboard and any connected agents

Try it:

```bash
# Run a command — it executes normally but is fully tracked
ls -la

# Check your execution history
nuc history

# See the environment context Nucleus has built
nuc context

# Try a risk evaluation without executing
nuc exec "rm -rf /tmp/test" --dry-run
```

## Open the Dashboard

Visit [http://localhost:3000](http://localhost:3000) in your browser. You will see your commands appearing in real time on the execution feed.

## Connect Claude Desktop

If you use Claude Desktop, you can give it full shell access through Nucleus:

```bash
nuc mcp install --client claude-desktop
```

Restart Claude Desktop. You can now ask Claude to run commands, check your environment, search history, and roll back mistakes — all through the Nucleus MCP server.

See [MCP Integration](/docs/mcp/overview) for details.

## API Access

The REST API is available at `http://localhost:8080`. The development API key is `dev-nucleus-key-local`.

```bash
# Execute a command through the API
curl -X POST http://localhost:8080/api/v1/agent/execute \
  -H "X-Nucleus-Key: dev-nucleus-key-local" \
  -H "Content-Type: application/json" \
  -d '{"command": "echo hello from the API"}'

# Get environment context
curl http://localhost:8080/api/v1/context \
  -H "X-Nucleus-Key: dev-nucleus-key-local"
```

## Load Demo Data

```bash
make seed
```

This populates the database with sample sessions, executions, and skills so you can explore the dashboard and API with realistic data.

## What's Next

| Goal | Guide |
|------|-------|
| Understand the architecture | [What is Nucleus?](/docs/intro) |
| Install on bare metal | [Installation](/docs/installation/overview) |
| Use natural language commands | [Natural Language](/docs/shell/natural-language) |
| Build an AI agent on Nucleus | [Agent Orchestration](/docs/agents/overview) |
| Use the Python/TS/Go SDK | [SDKs](/docs/sdk/python) |
