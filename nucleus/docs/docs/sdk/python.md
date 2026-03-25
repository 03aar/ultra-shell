---
sidebar_position: 1
title: Python SDK
---

# Python SDK

The Nucleus Python SDK provides a typed client for interacting with the Nucleus API from Python applications.

## Installation

```bash
pip install nucleus-sdk
```

Or from the monorepo:

```bash
cd nucleus-sdk/python
pip install -e .
```

## Quick Start

```python
from nucleus import NucleusClient

client = NucleusClient(
    api_url="http://localhost:8080",
    api_key="dev-nucleus-key-local"
)

# Execute a command
result = client.execute("ls -la")
print(result.exit_code)    # 0
print(result.stdout)       # file listing
print(result.risk_level)   # "low"
print(result.duration_ms)  # 45

# Get environment context
ctx = client.get_context()
print(ctx.working_directory)  # "/home/user/project"
print(ctx.git.branch)        # "main"
```

## API Reference

### `NucleusClient`

```python
client = NucleusClient(
    api_url: str = "http://localhost:8080",
    api_key: str = "dev-nucleus-key-local",
    timeout: float = 30.0,
)
```

### `client.execute(command, dry_run=False, agent_id=None)`

Execute a shell command.

```python
result = client.execute("npm test")

# Result fields
result.id              # UUID
result.command         # ParsedCommand
result.exit_code       # int
result.stdout          # str
result.stderr          # str
result.duration_ms     # int
result.risk_level      # "none" | "low" | "medium" | "high" | "critical"
result.risk_flags      # list[str]
result.rollback_available  # bool

# Dry run
assessment = client.execute("rm -rf /tmp/data", dry_run=True)
print(assessment.risk_level)      # "high"
print(assessment.would_snapshot)  # True
```

### `client.get_context()`

Get the current environment state.

```python
ctx = client.get_context()
ctx.working_directory   # str
ctx.shell              # str
ctx.user               # str
ctx.hostname           # str
ctx.git.branch         # str
ctx.git.modified_files # list[str]
ctx.processes          # list[Process]
ctx.environment        # dict[str, str] (secrets masked)
```

### `client.get_history(limit=20, search=None)`

Get execution history.

```python
history = client.get_history(limit=10)
for execution in history:
    print(f"{execution.command} -> exit {execution.exit_code}")

# Search
results = client.get_history(search="docker build")
```

### `client.rollback(execution_id)`

Roll back an execution.

```python
result = client.rollback("550e8400-e29b-41d4-a716-446655440000")
for file in result.files_restored:
    print(f"{file.action}: {file.path}")
```

### `client.create_plan(goal, context="")`

Generate an execution plan.

```python
plan = client.create_plan("Set up a Python project with tests")
for step in plan.steps:
    print(f"  {step.command}  [{step.risk_level}]")
    print(f"    Reason: {step.rationale}")
```

### `client.run_skill(name, params=None)`

Run a skill.

```python
result = client.run_skill("git_cleanup")
result = client.run_skill("project_setup", {"project_name": "myapp", "type": "node"})
```

### `client.get_graph(format="json")`

Get the execution DAG.

```python
graph = client.get_graph(format="json")
for node in graph.nodes:
    print(f"{node.id}: {node.command}")
for edge in graph.edges:
    print(f"{edge.source} -> {edge.target}")
```

### `client.create_session(name)`

Create a new session.

```python
session = client.create_session("deploy-v2.1")
print(session.id)
```

## Async Client

For async applications:

```python
from nucleus import AsyncNucleusClient

client = AsyncNucleusClient(
    api_url="http://localhost:8080",
    api_key="dev-nucleus-key-local"
)

result = await client.execute("npm test")
ctx = await client.get_context()
```

## Building a Custom Agent

```python
from nucleus import NucleusClient

client = NucleusClient()

def deploy_agent(target: str):
    """Simple deployment agent."""
    # 1. Observe
    ctx = client.get_context()
    assert ctx.git.branch == "main", "Must be on main branch"

    # 2. Plan
    plan = client.create_plan(f"Deploy to {target}")

    # 3. Execute each step
    for step in plan.steps:
        print(f"Executing: {step.command}")
        result = client.execute(step.command)

        if result.exit_code != 0:
            print(f"Failed: {result.stderr}")
            if result.rollback_available:
                client.rollback(result.id)
                print("Rolled back.")
            return False

    return True

deploy_agent("staging")
```

## WebSocket Streaming

```python
from nucleus import NucleusClient

client = NucleusClient()

for event in client.watch(channels=["executions", "warnings"]):
    if event.type == "execution":
        print(f"[EXEC] {event.data.command} -> {event.data.exit_code}")
    elif event.type == "warning":
        print(f"[WARN] {event.data.command} risk:{event.data.risk_level}")
```
