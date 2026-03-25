---
sidebar_position: 2
title: Authentication
---

# API Authentication

Every API request must include a valid API key in the `X-Nucleus-Key` header.

## Header Format

```
X-Nucleus-Key: <your-api-key>
```

Example:

```bash
curl http://localhost:8080/api/v1/context \
  -H "X-Nucleus-Key: dev-nucleus-key-local"
```

## Development Key

In development mode, the key `dev-nucleus-key-local` is always accepted. This key is configured in `.env` and works out of the box with Docker Compose.

:::caution
Do not use the development key in production. It provides full access to all API endpoints with no restrictions.
:::

## Creating API Keys

API keys are stored as bcrypt hashes in PostgreSQL. To create a new key:

```bash
# Via the CLI
nuc api-key create --name "my-service" --scope "execute,context,history"
# Output: nuc-key-a1b2c3d4e5f6...

# Via the API
curl -X POST http://localhost:8080/api/v1/keys \
  -H "X-Nucleus-Key: dev-nucleus-key-local" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "CI Pipeline",
    "scopes": ["execute", "context", "history"]
  }'
```

## Key Scopes

Keys can be scoped to limit access:

| Scope | Access |
|-------|--------|
| `execute` | Execute commands, run skills |
| `context` | Read environment state |
| `history` | Query execution history |
| `rollback` | Roll back executions |
| `sessions` | Manage sessions |
| `admin` | Full access including key management |
| `*` | All scopes (development key default) |

Example: a monitoring service might only need `context` and `history` scopes.

## Key Management

```bash
# List all keys
nuc api-key list

# Revoke a key
nuc api-key revoke --name "my-service"
```

## Security Details

- Keys are hashed with bcrypt (cost factor 12) before storage.
- The raw key is shown only at creation time and never stored.
- Keys are prefixed with `nuc-key-` for easy identification in logs and configs.
- Failed authentication attempts are logged with the source IP.
- The `/health` endpoint does not require authentication.

## Using Keys with SDKs

### Python

```python
from nucleus import NucleusClient

client = NucleusClient(
    api_url="http://localhost:8080",
    api_key="nuc-key-a1b2c3d4e5f6"
)
```

### TypeScript

```typescript
import { NucleusClient } from '@nucleus/sdk';

const client = new NucleusClient({
  apiUrl: 'http://localhost:8080',
  apiKey: 'nuc-key-a1b2c3d4e5f6',
});
```

### Go

```go
client := nucleus.NewClient(
    nucleus.WithAPIURL("http://localhost:8080"),
    nucleus.WithAPIKey("nuc-key-a1b2c3d4e5f6"),
)
```
