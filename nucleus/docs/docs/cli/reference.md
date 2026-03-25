---
sidebar_position: 1
title: CLI Reference
---

# CLI Reference

`nuc` is the command-line interface for Nucleus. It communicates with the Nucleus API server to execute commands, query history, manage sessions, and more.

## Installation

```bash
# From the project root
make install-cli

# Or build manually
cd nucleus-cli && go build -o nuc . && sudo mv nuc /usr/local/bin/
```

Verify:

```bash
nuc --version
```

## Shell & Status

### `nuc shell`

Launch an interactive Nucleus shell session. This wraps your default shell with PTY interception.

```bash
nuc shell
```

### `nuc status`

Show the current session status including command count, session duration, and connection health.

```bash
nuc status
```

### `nuc doctor`

Run a health check on all Nucleus components.

```bash
nuc doctor
```

Checks connectivity to: API server, PostgreSQL, Redis, shell runtime, MCP server, and dashboard.

### `nuc dashboard`

Open the web dashboard in your default browser.

```bash
nuc dashboard
```

## History & Context

### `nuc history`

Show execution history for the current session.

```bash
nuc history                      # Default: last 10 commands
nuc history --limit 50           # Last 50 commands
nuc history --search "docker"    # Full-text search
nuc history --risk high          # Filter by risk level
nuc history --json               # JSON output
```

### `nuc context`

Show the current environment state.

```bash
nuc context                      # Human-readable output
nuc context --json               # Full JSON response
nuc context --section git        # Only git information
nuc context --section processes  # Only running processes
```

## Execution & Rollback

### `nuc exec`

Execute a command through Nucleus (without entering the interactive shell).

```bash
nuc exec "ls -la"                      # Execute and display result
nuc exec "rm -rf ./build" --dry-run    # Risk assessment only
nuc exec "npm test" --json             # JSON output
```

### `nuc rollback`

Roll back a command by restoring file snapshots.

```bash
nuc rollback <execution-id>      # Roll back specific execution
nuc rollback --last              # Roll back the last command
nuc rollback --last --dry-run    # Preview rollback without executing
```

## Graph

### `nuc graph`

Display the execution DAG for the current session.

```bash
nuc graph                        # ASCII graph in terminal
nuc graph --format mermaid       # Mermaid diagram syntax
nuc graph --format json          # Raw JSON edges and nodes
```

## Sessions

### `nuc session list`

List all sessions.

```bash
nuc session list
nuc session list --active        # Only active sessions
```

### `nuc session new`

Create a new named session.

```bash
nuc session new "deploy-v2.1"
```

### `nuc session replay`

Replay a session in the terminal.

```bash
nuc session replay <session-id>
nuc session replay <session-id> --speed 2x  # Double speed
```

## Skills

### `nuc skill list`

List all available skills.

```bash
nuc skill list
```

### `nuc skill run`

Run a skill.

```bash
nuc skill run git_cleanup
nuc skill run project_setup --params project_name=myapp type=node
nuc skill run docker_cleanup --dry-run
```

## Agent & Planning

### `nuc plan`

Generate an execution plan for a goal using AI.

```bash
nuc plan "set up CI/CD for this project"
nuc plan "optimize the Dockerfile" --provider openai
```

### `nuc watch`

Live event stream from the WebSocket.

```bash
nuc watch                        # All events
nuc watch --filter executions    # Only execution events
nuc watch --filter warnings      # Only risk warnings
```

## MCP

### `nuc mcp install`

Install MCP configuration for a client.

```bash
nuc mcp install --client claude-desktop
nuc mcp install --client cursor
```

### `nuc mcp status`

Check MCP server connectivity.

```bash
nuc mcp status
```

## Global Flags

These flags work with every command:

| Flag | Default | Description |
|------|---------|-------------|
| `--api-url` | `http://localhost:8080` | Nucleus API server URL |
| `--api-key` | `dev-nucleus-key-local` | API authentication key |
| `--output` | `text` | Output format: `text` or `json` |
| `--no-color` | `false` | Disable colored output |
| `--verbose` | `false` | Enable debug logging |
| `--version` | — | Print version and exit |
| `--help` | — | Print help and exit |

## Environment Variables

Flags can also be set via environment variables:

| Variable | Equivalent Flag |
|----------|----------------|
| `NUCLEUS_API_URL` | `--api-url` |
| `NUCLEUS_API_KEY` | `--api-key` |
| `NUCLEUS_OUTPUT` | `--output` |
| `NO_COLOR` | `--no-color` |
