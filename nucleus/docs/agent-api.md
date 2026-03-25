# Agent API Reference

## Execute Command

```
POST /api/v1/agent/execute
```

Execute a command through Nucleus with full context awareness.

**Request:**
```json
{
  "command": "npm test",
  "agent_id": "claude-agent-1",
  "dry_run": false
}
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "command": { "raw": "npm test", "binary": "npm", "category": "build" },
    "exit_code": 0,
    "duration_ms": 12000,
    "stdout": "Tests: 3 passed",
    "risk_flags": [],
    "rollback_available": false
  }
}
```

## Generate Plan

```
POST /api/v1/agent/plan
```

Generate an execution plan for a goal.

**Request:**
```json
{
  "goal": "Set up a new Node.js project with tests",
  "context": "",
  "agent_id": "claude"
}
```

**Response:**
```json
{
  "data": {
    "plan_id": "uuid",
    "goal": "Set up a new Node.js project with tests",
    "steps": [
      { "command": "mkdir project && cd project", "rationale": "Create directory", "risk_level": "low", "reversible": true },
      { "command": "npm init -y", "rationale": "Initialize package.json", "risk_level": "low", "reversible": true }
    ]
  }
}
```

## Skills

```
GET  /api/v1/skills              # List all skills
GET  /api/v1/skills/:name        # Get skill detail
POST /api/v1/skills/:name/run    # Run a skill
```

### Run Skill

**Request:**
```json
{
  "params": { "project_name": "myapp", "type": "node" }
}
```

## Authentication

All endpoints require `X-Nucleus-Key` header.

Dev key: `dev-nucleus-key-local`
