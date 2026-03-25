---
sidebar_position: 7
title: WebSocket
---

# WebSocket API

Nucleus provides a WebSocket endpoint for real-time streaming of execution events. The dashboard, agents, and custom clients can subscribe to live updates.

## Connecting

```
ws://localhost:8080/ws
```

Authentication is via query parameter:

```
ws://localhost:8080/ws?key=dev-nucleus-key-local
```

### JavaScript Example

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?key=dev-nucleus-key-local');

ws.onopen = () => {
  console.log('Connected to Nucleus');

  // Subscribe to specific event types
  ws.send(JSON.stringify({
    type: 'subscribe',
    channels: ['executions', 'rollbacks', 'warnings']
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Event:', data);
};

ws.onclose = () => {
  console.log('Disconnected');
};
```

## Event Types

### `execution`

Emitted when a command finishes executing.

```json
{
  "type": "execution",
  "data": {
    "id": "uuid",
    "command": "npm test",
    "exit_code": 0,
    "stdout": "Tests: 12 passed",
    "stderr": "",
    "duration_ms": 3200,
    "risk_level": "low",
    "session_id": "uuid",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

### `rollback`

Emitted when a rollback completes.

```json
{
  "type": "rollback",
  "data": {
    "execution_id": "uuid",
    "command": "rm -rf ./config",
    "files_restored": 3,
    "timestamp": "2025-01-15T10:35:00Z"
  }
}
```

### `session`

Emitted when a session is created or ended.

```json
{
  "type": "session",
  "data": {
    "action": "created",
    "session_id": "uuid",
    "name": "deploy-v2.1",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

### `warning`

Emitted when a command is flagged as risky.

```json
{
  "type": "warning",
  "data": {
    "command": "chmod 777 /etc/passwd",
    "risk_level": "high",
    "risk_flags": ["permission-escalation", "system-file"],
    "blocked": false,
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

## Channels

Subscribe to specific channels to filter events:

| Channel | Events |
|---------|--------|
| `executions` | All command executions |
| `rollbacks` | All rollback operations |
| `sessions` | Session create/end events |
| `warnings` | Risk warnings and blocked commands |

Subscribe to all channels by omitting the `channels` field in the subscribe message.

## Redis Pub/Sub Channels

The WebSocket hub subscribes to these Redis channels:

| Redis Channel | Description |
|---------------|-------------|
| `nucleus:executions` | Execution completion events from `nucleus-core` |
| `nucleus:rollbacks` | Rollback events |
| `nucleus:sessions` | Session lifecycle events |
| `nucleus:warnings` | Risk warnings |

You can also subscribe directly to Redis for custom consumers:

```bash
redis-cli SUBSCRIBE nucleus:executions
```

## Heartbeat

The server sends a ping frame every 30 seconds. Clients should respond with a pong. Connections that miss 3 consecutive pings are closed.

## CLI Watch Mode

```bash
nuc watch
```

This connects to the WebSocket and streams events to your terminal in real time:

```
[10:30:01] EXEC  npm test                    exit:0  risk:low   3200ms
[10:30:05] EXEC  git add .                   exit:0  risk:low     45ms
[10:30:08] EXEC  git commit -m "fix tests"   exit:0  risk:low    120ms
[10:30:15] WARN  rm -rf /tmp/data           risk:high  (allowed)
[10:30:20] ROLL  rm -rf ./config            files:3  restored
```
