---
sidebar_position: 4
title: Context
---

# Context API

The context endpoint provides a snapshot of the current environment state. This is the primary way agents and tools understand what is happening on the system.

## Get Context

```
GET /api/v1/context
```

**Response:**

```json
{
  "data": {
    "working_directory": "/home/user/myproject",
    "shell": "/bin/zsh",
    "user": "user",
    "hostname": "dev-machine",
    "os": {
      "name": "linux",
      "version": "Ubuntu 22.04",
      "arch": "x86_64"
    },
    "git": {
      "is_repo": true,
      "branch": "feature/auth",
      "repo_root": "/home/user/myproject",
      "remote_url": "git@github.com:user/myproject.git",
      "modified_files": ["src/auth.ts", "src/middleware.ts"],
      "staged_files": ["src/auth.ts"],
      "untracked_files": ["src/auth.test.ts"],
      "ahead": 2,
      "behind": 0,
      "last_commit": {
        "hash": "abc1234",
        "message": "Add auth middleware",
        "author": "user",
        "timestamp": "2025-01-15T09:00:00Z"
      }
    },
    "environment": {
      "NODE_ENV": "development",
      "PATH": "/usr/local/bin:/usr/bin:/bin",
      "HOME": "/home/user",
      "SHELL": "/bin/zsh",
      "ANTHROPIC_API_KEY": "***",
      "DATABASE_URL": "***"
    },
    "processes": [
      {
        "pid": 12345,
        "command": "node server.js",
        "cpu_percent": 2.1,
        "memory_mb": 120,
        "uptime_seconds": 3600
      },
      {
        "pid": 12350,
        "command": "postgres",
        "cpu_percent": 0.5,
        "memory_mb": 85,
        "uptime_seconds": 86400
      }
    ],
    "disk": {
      "total_gb": 500,
      "used_gb": 210,
      "available_gb": 290
    },
    "session": {
      "id": "uuid",
      "name": "dev-session",
      "command_count": 47,
      "started_at": "2025-01-15T08:00:00Z"
    }
  }
}
```

## Sensitive Variable Masking

Environment variables matching these patterns are masked with `***`:

- `*SECRET*`
- `*TOKEN*`
- `*KEY*` (except `TERM_KEY`, `KEYBOARD`)
- `*PASSWORD*`
- `*CREDENTIAL*`
- `*AUTH*` (as a standalone word boundary)

This prevents accidental exposure of secrets through the API. The masking applies to all API responses and MCP resource reads.

## JSON Output via CLI

```bash
nuc context --json
```

This calls the same API endpoint and outputs the raw JSON.

## Filtered Context

```bash
# Get only git information
nuc context --section git

# Get only running processes
nuc context --section processes
```

## Use Cases

### Agent Planning

AI agents call `GET /api/v1/context` before generating execution plans to understand:
- What project they are in (git repo, package.json, Makefile, etc.)
- What services are running (can I start a server or is one already running?)
- What branch is checked out (should I create a new branch?)
- What files have been modified (what should I commit?)

### Debugging

When a command fails, the context provides diagnostic information:
- Environment variables that may be misconfigured.
- Processes that may be consuming resources.
- Disk space that may be exhausted.
