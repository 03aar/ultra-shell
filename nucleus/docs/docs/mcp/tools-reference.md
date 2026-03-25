---
sidebar_position: 3
title: MCP Tools Reference
---

# MCP Tools Reference

Complete documentation for every tool exposed by the Nucleus MCP server. These tools are available to any MCP-compatible AI client (Claude Desktop, Cursor, Zed).

## execute_command

Run a shell command through Nucleus with full risk evaluation, snapshotting, and recording.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `command` | string | Yes | The shell command to execute |
| `working_directory` | string | No | Directory to execute in (defaults to current) |

**Example call:**

```json
{
  "name": "execute_command",
  "arguments": {
    "command": "npm test",
    "working_directory": "/home/user/myproject"
  }
}
```

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "command": "npm test",
  "exit_code": 0,
  "stdout": "Tests: 12 passed, 12 total\nTime: 3.2s",
  "stderr": "",
  "duration_ms": 3200,
  "risk_level": "low",
  "rollback_available": false
}
```

**Notes:**
- Commands scored as `critical` risk are blocked and return an error.
- File mutations are automatically snapshotted.
- The execution is recorded in the DAG.

---

## get_context

Get the current environment state including working directory, git info, running processes, and environment variables.

**Parameters:** None.

**Response:**

```json
{
  "working_directory": "/home/user/myproject",
  "shell": "/bin/zsh",
  "user": "user",
  "hostname": "dev-machine",
  "os": "linux",
  "git": {
    "branch": "main",
    "repo_root": "/home/user/myproject",
    "modified_files": ["src/index.ts"],
    "staged_files": [],
    "ahead": 0,
    "behind": 0
  },
  "environment": {
    "NODE_ENV": "development",
    "PATH": "/usr/local/bin:/usr/bin"
  },
  "running_processes": [
    {"pid": 1234, "command": "node server.js", "cpu": 2.1, "memory_mb": 120}
  ]
}
```

**Notes:** Sensitive environment variables (matching `SECRET`, `TOKEN`, `KEY`, `PASSWORD`) are masked as `***`.

---

## get_execution_history

Get recent command execution history.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | number | No | Number of entries to return (default: 20, max: 100) |

**Response:**

```json
{
  "executions": [
    {
      "id": "uuid",
      "command": "git status",
      "exit_code": 0,
      "duration_ms": 45,
      "risk_level": "low",
      "timestamp": "2025-01-15T10:30:00Z"
    }
  ]
}
```

---

## rollback_execution

Roll back a previously executed command by restoring snapshotted files to their pre-execution state.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `execution_id` | string | Yes | UUID of the execution to roll back |

**Response:**

```json
{
  "success": true,
  "files_restored": [
    {"path": "/home/user/myproject/config.json", "action": "restored"},
    {"path": "/home/user/myproject/temp.txt", "action": "recreated"}
  ]
}
```

**Notes:** Only executions with `rollback_available: true` can be rolled back. Not all commands produce snapshots (e.g., read-only commands like `ls` or `cat`).

---

## get_execution_graph

Get the execution dependency graph for the current session.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `format` | string | No | Output format: `mermaid` (default) or `json` |

**Response (Mermaid):**

```
graph TD
  A[git status] --> B[npm install]
  B --> C[npm test]
  C --> D[git commit]
```

---

## run_skill

Execute a pre-defined skill workflow.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `skill_name` | string | Yes | Name of the skill to run |
| `params` | object | No | Key-value parameters for the skill |

**Available skills:**

| Skill | Parameters | Description |
|-------|-----------|-------------|
| `git_cleanup` | — | Delete merged git branches |
| `docker_cleanup` | — | Remove unused Docker resources |
| `project_setup` | `project_name`, `type` | Scaffold a new project |
| `deploy_check` | — | Pre-deployment validation |
| `env_audit` | — | Audit for exposed secrets |
| `process_debug` | `pid` or `port` | Debug a running process |

**Example:**

```json
{
  "name": "run_skill",
  "arguments": {
    "skill_name": "project_setup",
    "params": {
      "project_name": "myapp",
      "type": "node"
    }
  }
}
```

---

## search_history

Full-text search across all past command executions.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `query` | string | Yes | Search query |
| `limit` | number | No | Max results (default: 20) |

**Example:**

```json
{
  "name": "search_history",
  "arguments": {
    "query": "docker build",
    "limit": 10
  }
}
```

---

## get_risk_assessment

Perform a dry-run risk evaluation on a command without executing it.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `command` | string | Yes | Command to evaluate |

**Response:**

```json
{
  "command": "rm -rf ./node_modules",
  "risk_level": "medium",
  "risk_score": 5,
  "risk_flags": ["recursive-delete", "large-directory"],
  "would_snapshot": true,
  "estimated_files_affected": 12000
}
```

---

## create_session

Create a new named session for organizing related commands.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Human-readable session name |

**Response:**

```json
{
  "session_id": "uuid",
  "name": "deploy-v2.1",
  "created_at": "2025-01-15T10:30:00Z"
}
```

---

## watch_stream

Subscribe to a live stream of execution events. This is a long-lived connection that emits events as commands are executed.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `events` | string[] | No | Event types to filter: `execution`, `rollback`, `session`, `warning` |

**Event format:**

```json
{
  "type": "execution",
  "data": {
    "id": "uuid",
    "command": "npm test",
    "exit_code": 0,
    "duration_ms": 3200,
    "risk_level": "low",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```
