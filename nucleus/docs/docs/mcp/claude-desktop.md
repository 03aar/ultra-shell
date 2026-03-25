---
sidebar_position: 2
title: Claude Desktop Setup
---

# Connecting Claude Desktop

This guide walks through connecting Claude Desktop to Nucleus via MCP.

## Prerequisites

1. Nucleus is running (via Docker Compose or native install).
2. Claude Desktop is installed ([download](https://claude.ai/download)).
3. The `nuc` CLI is installed.

## Automatic Setup

The fastest way:

```bash
nuc mcp install --client claude-desktop
```

This writes the MCP configuration file and restarts the MCP server. **Restart Claude Desktop** for the change to take effect.

## Manual Setup

### macOS

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "nucleus": {
      "command": "node",
      "args": ["/path/to/ultra-shell/nucleus/nucleus-mcp/dist/index.js"],
      "env": {
        "NUCLEUS_API_URL": "http://localhost:8080",
        "NUCLEUS_API_KEY": "dev-nucleus-key-local"
      }
    }
  }
}
```

Replace `/path/to/ultra-shell` with your actual clone path.

### Linux

Edit `~/.config/Claude/claude_desktop_config.json` with the same content as above.

### Windows

Edit `%APPDATA%\Claude\claude_desktop_config.json` with the same content. If Nucleus is running in WSL2, use `http://localhost:8080` as the URL (WSL2 ports are forwarded).

## Verify Connection

After restarting Claude Desktop:

1. Open a new conversation.
2. You should see a hammer icon in the input area indicating MCP tools are available.
3. Ask Claude: *"What's my current working directory?"*

Claude will call the `get_context` tool and return your environment state.

## What You Can Ask Claude

Once connected, try these:

### Environment Awareness

> "What directory am I in? What git branch?"

Claude calls `get_context` and reads your environment.

### Execute Commands

> "Run the tests in this project."

Claude calls `execute_command` with `npm test` (or the appropriate test command based on your project).

### History and Search

> "What commands did I run in the last hour?"

Claude calls `get_execution_history` and summarizes your activity.

### Rollback

> "That last command broke something. Undo it."

Claude calls `rollback_execution` to restore snapshotted files.

### Planning

> "I need to set up a Docker deployment for this Node.js project."

Claude calls `get_context` to understand the project, then formulates a multi-step plan.

### Skills

> "Clean up my merged git branches."

Claude calls `run_skill` with `git_cleanup`.

## Troubleshooting

### No hammer icon in Claude Desktop

- Verify the config file path is correct for your OS.
- Check that the `node` command is available in your system PATH.
- Ensure `nucleus-mcp/dist/index.js` exists (run `npm run build` in `nucleus-mcp/` if not).
- Restart Claude Desktop completely (quit and reopen).

### "Connection refused" errors

- Verify Nucleus API is running: `curl http://localhost:8080/health`
- Check the API URL in your MCP config matches the running API.
- If using Docker, ensure port 8080 is mapped.

### Check MCP server status

```bash
nuc mcp status
```

This verifies the MCP server can connect to the API and reports any issues.

### View MCP server logs

```bash
# If running via Docker
docker compose logs -f nucleus-mcp

# If running manually
NUCLEUS_API_URL=http://localhost:8080 \
NUCLEUS_API_KEY=dev-nucleus-key-local \
node nucleus-mcp/dist/index.js 2>mcp.log
```
