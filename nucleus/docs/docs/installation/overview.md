---
sidebar_position: 1
title: Installation Overview
---

# Installation Overview

Nucleus can be installed in several ways depending on your environment and needs.

## Methods

| Method | Best For | Time |
|--------|----------|------|
| [Docker Compose](/docs/installation/docker) | Quick evaluation, development | 2 minutes |
| [macOS](/docs/installation/macos) | Native macOS development | 10 minutes |
| [Linux](/docs/installation/linux) | Native Linux / servers | 10 minutes |
| [Windows](/docs/installation/windows) | WSL2 development | 15 minutes |

## Requirements

All installation methods require:

| Dependency | Minimum Version | Purpose |
|------------|----------------|---------|
| PostgreSQL | 15+ | Session storage, API keys, audit logs |
| Redis | 7+ | Real-time pub/sub between core and API |
| Rust toolchain | 1.75+ | Building `nucleus-core` |
| Go | 1.22+ | Building `nucleus-api` and `nucleus-cli` |
| Node.js | 18+ | Building `nucleus-mcp` and `nucleus-dashboard` |

**Docker Compose** bundles all dependencies automatically. For native installs, you need to provide PostgreSQL and Redis yourself.

## Optional Dependencies

| Dependency | Purpose |
|------------|---------|
| Python 3.11+ | Agent orchestrator (`nucleus-orchestrator`) |
| Ollama | Local LLM inference |
| `ANTHROPIC_API_KEY` | Claude agent provider |
| `OPENAI_API_KEY` | OpenAI agent provider |
| `GOOGLE_API_KEY` | Gemini agent provider |

## Architecture Recap

When fully running, these services are active:

```
nucleus-core      (Rust binary)          — Shell runtime
nucleus-api       (Go binary)      :8080 — REST + WebSocket API
nucleus-dashboard (Next.js)        :3000 — Web dashboard
nucleus-mcp       (Node.js)              — MCP server (stdio)
PostgreSQL                         :5432 — Persistent storage
Redis                              :6379 — Pub/sub messaging
```

## Recommended: Docker Compose

For most users, Docker Compose is the fastest path to a working Nucleus installation. It handles all dependencies, networking, and configuration:

```bash
git clone https://github.com/03aar/ultra-shell.git
cd ultra-shell/nucleus
cp .env.example .env
docker compose up --build -d
```

Then install just the CLI natively:

```bash
make install-cli
nuc doctor
```

See the [Docker installation guide](/docs/installation/docker) for details.
