# MCP Integration Guide

## What is MCP?

The Model Context Protocol (MCP) allows AI clients like Claude Desktop, Cursor, Continue, and Zed to connect to external tool servers. NUCLEUS provides a fully compliant MCP server exposing all shell capabilities.

## Quick Setup

### Claude Desktop

```bash
nuc mcp install --client claude-desktop
# Restart Claude Desktop
```

Or manually add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS):

```json
{
  "mcpServers": {
    "nucleus": {
      "command": "node",
      "args": ["./nucleus-mcp/dist/index.js"],
      "env": {
        "NUCLEUS_API_URL": "http://localhost:8080",
        "NUCLEUS_API_KEY": "dev-nucleus-key-local"
      }
    }
  }
}
```

### Cursor

```bash
nuc mcp install --client cursor
```

## Available MCP Tools

| Tool | Description |
|------|-------------|
| `execute_command` | Run shell command with risk evaluation |
| `get_context` | Get environment state (cwd, git, processes) |
| `get_execution_history` | Get recent command history |
| `rollback_execution` | Undo a previous command |
| `get_execution_graph` | Get the dependency DAG |
| `run_skill` | Run a pre-defined workflow |
| `search_history` | Full-text search past commands |
| `get_risk_assessment` | Dry-run risk check |
| `create_session` | Create a named session |
| `watch_stream` | Subscribe to live events |

## Available MCP Resources

| Resource URI | Description |
|-------------|-------------|
| `nucleus://context` | Current environment snapshot |
| `nucleus://graph` | Session execution graph (Mermaid) |
| `nucleus://history` | Last 50 executions |
| `nucleus://sessions` | All sessions list |

## Available MCP Prompts

| Prompt | Description |
|--------|-------------|
| `debug_failure` | Analyze a failed command |
| `plan_workflow` | Plan a multi-step workflow |
| `explain_session` | Summarize session activity |
