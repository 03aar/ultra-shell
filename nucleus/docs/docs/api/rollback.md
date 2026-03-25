---
sidebar_position: 5
title: Rollback
---

# Rollback API

Nucleus can undo command side effects by restoring file snapshots taken before execution. This is one of the most powerful features of the platform — every destructive file operation can be reversed.

## How Rollback Works

1. Before a command executes, `nucleus-core` identifies files that will be modified.
2. The current contents of those files are saved as a snapshot in sled.
3. After execution, if you want to undo the changes, the rollback endpoint restores the snapshotted files.

## Roll Back an Execution

```
POST /api/v1/rollback/:id
```

**Path parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | uuid | Execution ID to roll back |

**Response:**

```json
{
  "data": {
    "execution_id": "550e8400-e29b-41d4-a716-446655440000",
    "command": "rm -rf ./config",
    "files_restored": [
      {
        "path": "/home/user/project/config/database.yml",
        "action": "recreated",
        "size_bytes": 1024
      },
      {
        "path": "/home/user/project/config/app.yml",
        "action": "recreated",
        "size_bytes": 512
      }
    ],
    "rolled_back_at": "2025-01-15T10:35:00Z"
  }
}
```

**File action types:**

| Action | Meaning |
|--------|---------|
| `restored` | File existed before and was modified; contents restored to pre-execution state |
| `recreated` | File existed before and was deleted; file recreated with original contents |
| `removed` | File did not exist before and was created by the command; file removed |

## Preview Rollback

See what would happen without actually restoring files:

```
GET /api/v1/rollback/:id/preview
```

**Response:**

```json
{
  "data": {
    "execution_id": "uuid",
    "command": "rm -rf ./config",
    "rollback_available": true,
    "files": [
      {
        "path": "/home/user/project/config/database.yml",
        "action": "recreated",
        "size_bytes": 1024,
        "snapshot_age_seconds": 300
      }
    ]
  }
}
```

## Rollback via CLI

```bash
# Roll back a specific execution
nuc rollback 550e8400-e29b-41d4-a716-446655440000

# Roll back the last command
nuc rollback --last

# Preview without executing
nuc rollback --last --dry-run
```

## Limitations

- **Snapshot size limit:** Files larger than 100MB are not snapshotted (configurable in `config.toml`).
- **Excluded patterns:** Files matching exclusion patterns (e.g., `node_modules/**`) are not snapshotted.
- **Non-file side effects:** Rollback restores files but cannot undo network requests, database writes, or process state changes.
- **Time window:** Snapshots are retained for the lifetime of the sled database. Old snapshots can be pruned with `nuc prune --older-than 30d`.

## Rollback Events

When a rollback occurs, a `rollback` event is published to the `nucleus:rollbacks` Redis channel. This triggers:

- A real-time notification on the dashboard.
- A WebSocket event to all connected clients.
- An audit log entry in PostgreSQL.
