---
sidebar_position: 2
title: macOS
---

# macOS Installation

Native installation on macOS (Apple Silicon and Intel).

## Prerequisites

Install system dependencies with Homebrew:

```bash
brew install rust go node@18 postgresql@15 redis
```

Start the backing services:

```bash
brew services start postgresql@15
brew services start redis
```

## Clone and Build

```bash
git clone https://github.com/03aar/ultra-shell.git
cd ultra-shell/nucleus
cp .env.example .env
```

### Build nucleus-core (Rust)

```bash
cd nucleus-core
cargo build --release
# Binary at target/release/nucleus-core
cd ..
```

### Build nucleus-api (Go)

```bash
cd nucleus-api
go build -o nucleus-api .
cd ..
```

### Build nucleus-cli (Go)

```bash
cd nucleus-cli
go build -o nuc .
sudo mv nuc /usr/local/bin/
cd ..
```

### Build nucleus-mcp (Node.js)

```bash
cd nucleus-mcp
npm install
npm run build
cd ..
```

### Build nucleus-dashboard (Next.js)

```bash
cd nucleus-dashboard
npm install
npm run build
cd ..
```

## Database Setup

Create the PostgreSQL database and run the schema:

```bash
createdb nucleus
psql nucleus < nucleus-api/database/init.sql
```

## Environment Configuration

Edit `.env` with your local settings:

```bash
# Database
DATABASE_URL=postgres://$(whoami)@localhost:5432/nucleus?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# API
API_PORT=8080
API_KEY=dev-nucleus-key-local

# Optional: AI providers
# ANTHROPIC_API_KEY=sk-ant-...
# OPENAI_API_KEY=sk-...
```

## Start Services

Start each component (use separate terminal tabs or a process manager like `tmux`):

```bash
# Terminal 1: Shell runtime
./nucleus-core/target/release/nucleus-core

# Terminal 2: API server
cd nucleus-api && ./nucleus-api

# Terminal 3: Dashboard
cd nucleus-dashboard && npm start

# Terminal 4 (optional): MCP server for Claude Desktop
cd nucleus-mcp && node dist/index.js
```

Or use the Makefile:

```bash
make run  # Starts all components
```

## Verify

```bash
nuc doctor
```

## Claude Desktop Integration

```bash
nuc mcp install --client claude-desktop
# Restart Claude Desktop
```

This writes the MCP configuration to `~/Library/Application Support/Claude/claude_desktop_config.json`.

## Troubleshooting

### PostgreSQL connection refused

Ensure PostgreSQL is running:

```bash
brew services list | grep postgresql
pg_isready
```

### Redis connection refused

```bash
brew services list | grep redis
redis-cli ping  # Should return PONG
```

### Rust build fails on Apple Silicon

Ensure you have the latest Xcode command-line tools:

```bash
xcode-select --install
rustup update
```
