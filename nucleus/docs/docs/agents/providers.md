---
sidebar_position: 2
title: AI Providers
---

# AI Providers

The Nucleus orchestrator supports multiple AI providers. Each provider can be used for agent orchestration and natural language command translation.

## Claude (Anthropic)

```bash
export ANTHROPIC_API_KEY=sk-ant-...
nucleus-agent run "goal" --provider claude --model claude-sonnet-4-5
```

**Available models:**

| Model | Best For |
|-------|----------|
| `claude-opus-4-5` | Complex multi-step reasoning, large codebases |
| `claude-sonnet-4-5` | Best balance of speed and capability (recommended) |
| `claude-haiku-4-5` | Fast, simple tasks, cost-sensitive workloads |

**Features:** Tool use, streaming, 200K context window.

Claude is the recommended provider for Nucleus. It has the strongest tool-use capabilities and handles multi-step shell workflows reliably.

## OpenAI

```bash
export OPENAI_API_KEY=sk-...
nucleus-agent run "goal" --provider openai --model gpt-4o
```

**Available models:**

| Model | Best For |
|-------|----------|
| `gpt-4o` | General purpose, strong coding |
| `gpt-4o-mini` | Fast, cost-effective |
| `o1` | Complex reasoning tasks |
| `o3-mini` | Reasoning with lower cost |

**Features:** Tool use (function calling), streaming, 128K context window.

## Gemini (Google)

```bash
export GOOGLE_API_KEY=AI...
nucleus-agent run "goal" --provider gemini --model gemini-2.0-flash
```

**Available models:**

| Model | Best For |
|-------|----------|
| `gemini-2.0-flash` | Fast, general purpose |
| `gemini-1.5-pro` | Complex tasks, large context |
| `gemini-1.5-flash` | Cost-effective |

**Features:** Tool use, streaming, up to 1M context window.

## Ollama (Local)

Run agents entirely locally with no API keys:

```bash
# Ensure Ollama is running
ollama serve

# Pull a model
ollama pull llama3

# Use with Nucleus
nucleus-agent run "goal" --provider ollama --model llama3
```

**Available models:** Any model installed in your local Ollama instance. Recommended:

| Model | Parameters | Best For |
|-------|-----------|----------|
| `llama3` | 8B | General purpose |
| `codellama` | 7B/13B/34B | Code-focused tasks |
| `mistral` | 7B | Fast, general purpose |
| `mixtral` | 8x7B | Complex reasoning |

**Configuration:**

```bash
# Custom Ollama URL (default: http://localhost:11434)
export OLLAMA_URL=http://my-gpu-server:11434
```

**Features:** Streaming, no API key required, runs on your hardware.

:::note
Ollama models have varying levels of tool-use support. For the most reliable agent experience, use a model with strong instruction-following like `llama3` or `mixtral`.
:::

## Provider Comparison

| Feature | Claude | OpenAI | Gemini | Ollama |
|---------|--------|--------|--------|--------|
| Tool use | Excellent | Excellent | Good | Varies |
| Streaming | Yes | Yes | Yes | Yes |
| Context window | 200K | 128K | 1M | Varies |
| Cost | Per-token | Per-token | Per-token | Free (local) |
| Privacy | Cloud | Cloud | Cloud | Local |
| Setup | API key | API key | API key | Install binary |

## Setting a Default Provider

In `~/.config/nucleus/config.toml`:

```toml
[natural_language]
provider = "claude"
model = "claude-sonnet-4-5"
```

Or via environment:

```bash
export AGENT_DEFAULT_PROVIDER=claude
export AGENT_DEFAULT_MODEL=claude-sonnet-4-5
```

## Provider Fallback

Configure fallback providers in case the primary is unavailable:

```toml
[agent]
providers = ["claude", "openai", "ollama"]
# If Claude fails, try OpenAI. If OpenAI fails, fall back to local Ollama.
```

## Using Multiple Providers

Different tasks can use different providers:

```bash
# Use Claude for complex planning
nucleus-agent run "refactor the auth module" --provider claude --model claude-sonnet-4-5

# Use Ollama for simple local tasks (private, no network)
nucleus-agent run "clean up temp files" --provider ollama --model llama3
```
