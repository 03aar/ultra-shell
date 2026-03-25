---
sidebar_position: 1
title: Agent Orchestration
---

# Agent Orchestration

The `nucleus-orchestrator` is a Python engine that gives AI models structured access to your shell. Instead of fragile prompt-and-parse loops, agents interact with Nucleus through typed APIs that return structured results.

## Why Agent Orchestration?

Traditional approaches to giving AI agents shell access are brittle:

1. The agent generates a bash command as text.
2. A wrapper runs `subprocess.Popen()` and captures output.
3. The raw text output is sent back to the agent.
4. Errors, side effects, and context are lost.

Nucleus provides a better substrate:

1. The agent calls `POST /api/v1/agent/execute` with a command.
2. Nucleus parses, risk-evaluates, and snapshots before execution.
3. The agent receives structured data: exit code, parsed output, risk assessment, rollback availability.
4. The agent can observe context (`GET /api/v1/context`), plan multi-step workflows, and roll back mistakes.

## Agent Lifecycle

```
┌─────────────────────────────────────────┐
│              Agent Loop                  │
│                                          │
│  1. Observe    → GET /api/v1/context     │
│  2. Plan       → POST /api/v1/agent/plan │
│  3. Execute    → POST /api/v1/agent/execute │
│  4. Evaluate   → Check exit code, output │
│  5. Rollback?  → POST /api/v1/rollback/:id │
│  6. Repeat                               │
└─────────────────────────────────────────┘
```

### 1. Observe

The agent reads the current environment state:

```python
context = client.get_context()
# Returns: cwd, git branch, modified files, running processes, env vars
```

### 2. Plan

For complex goals, the agent generates a multi-step plan:

```python
plan = client.create_plan(
    goal="Deploy the application to staging",
    context=context
)
# Returns: ordered list of commands with rationale and risk assessment
```

### 3. Execute

The agent executes commands one at a time:

```python
result = client.execute("npm run build")
# Returns: exit_code, stdout, stderr, duration, risk_level, rollback_available
```

### 4. Evaluate

The agent checks the result and decides what to do next:

```python
if result.exit_code != 0:
    # Analyze the error
    error_analysis = agent.analyze_error(result.stderr)
    # Maybe try a fix
    fix_result = client.execute(error_analysis.suggested_fix)
```

### 5. Rollback

If something goes wrong, the agent can undo:

```python
if result.rollback_available:
    client.rollback(result.id)
```

## Running the Orchestrator

```bash
# Set up the orchestrator
cd nucleus-orchestrator
pip install -r requirements.txt

# Run an agent with a goal
nucleus-agent run "Set up CI/CD for this project" \
  --provider claude \
  --model claude-sonnet-4-5

# Run with a different provider
nucleus-agent run "Optimize the Dockerfile" \
  --provider openai \
  --model gpt-4o
```

## Agent Modes

### Autonomous Mode

The agent executes its plan without confirmation:

```bash
nucleus-agent run "goal" --mode autonomous
```

Risk guardrails still apply. Critical commands are blocked, high-risk commands produce warnings in the log.

### Interactive Mode (Default)

The agent presents each step and waits for approval:

```bash
nucleus-agent run "goal" --mode interactive
```

```
Agent: I'll set up CI/CD. Here's my plan:
  1. Create .github/workflows/ci.yml     [risk: low]
  2. Add test script to package.json     [risk: low]
  3. Push to trigger first run           [risk: low]

Execute step 1? [Y/n]
```

### Plan-Only Mode

The agent generates a plan but does not execute:

```bash
nucleus-agent run "goal" --mode plan-only
```

## Configuration

Configure the orchestrator in `.env` or environment variables:

```bash
# Required: at least one provider key
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=AI...

# Nucleus API connection
NUCLEUS_API_URL=http://localhost:8080
NUCLEUS_API_KEY=dev-nucleus-key-local

# Agent defaults
AGENT_DEFAULT_PROVIDER=claude
AGENT_DEFAULT_MODEL=claude-sonnet-4-5
AGENT_MODE=interactive
AGENT_MAX_STEPS=50
```

## See Also

- [Providers](/docs/agents/providers) — Configure Claude, GPT, Gemini, and Ollama.
- [API Reference](/docs/api/overview) — Full API documentation.
- [Python SDK](/docs/sdk/python) — Build custom agents.
