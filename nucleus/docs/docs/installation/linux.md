---
sidebar_position: 3
title: Linux
---

# Linux Installation

Native installation on Ubuntu/Debian, Fedora/RHEL, and Arch Linux.

## Prerequisites

### Ubuntu / Debian

```bash
# System packages
sudo apt update
sudo apt install -y build-essential pkg-config libssl-dev git curl

# Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.cargo/env

# Go (1.22+)
sudo snap install go --classic
# or download from https://go.dev/dl/

# Node.js (18+)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# PostgreSQL
sudo apt install -y postgresql postgresql-contrib
sudo systemctl enable --now postgresql

# Redis
sudo apt install -y redis-server
sudo systemctl enable --now redis-server
```

### Fedora / RHEL

```bash
sudo dnf install -y gcc openssl-devel git curl
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.cargo/env

sudo dnf install -y golang nodejs
sudo dnf install -y postgresql-server postgresql-contrib redis
sudo postgresql-setup --initdb
sudo systemctl enable --now postgresql redis
```

### Arch Linux

```bash
sudo pacman -S rust go nodejs npm postgresql redis git base-devel
sudo systemctl enable --now postgresql redis
```

## Clone and Build

```bash
git clone https://github.com/03aar/ultra-shell.git
cd ultra-shell/nucleus
cp .env.example .env
```

Build all components:

```bash
# Core (Rust)
cd nucleus-core && cargo build --release && cd ..

# API (Go)
cd nucleus-api && go build -o nucleus-api . && cd ..

# CLI (Go)
cd nucleus-cli && go build -o nuc . && sudo mv nuc /usr/local/bin/ && cd ..

# MCP server (Node.js)
cd nucleus-mcp && npm install && npm run build && cd ..

# Dashboard (Next.js)
cd nucleus-dashboard && npm install && npm run build && cd ..
```

## Database Setup

```bash
sudo -u postgres createuser $(whoami) --createdb
createdb nucleus
psql nucleus < nucleus-api/database/init.sql
```

## Environment Configuration

Edit `.env`:

```bash
DATABASE_URL=postgres://$(whoami)@localhost:5432/nucleus?sslmode=disable
REDIS_URL=redis://localhost:6379
API_PORT=8080
API_KEY=dev-nucleus-key-local
```

## Start Services

```bash
make run
```

Or start individually:

```bash
./nucleus-core/target/release/nucleus-core &
cd nucleus-api && ./nucleus-api &
cd nucleus-dashboard && npm start &
```

## Verify

```bash
nuc doctor
```

## systemd Service (Optional)

Create `/etc/systemd/system/nucleus.service` for production use:

```ini
[Unit]
Description=Nucleus Shell Runtime
After=postgresql.service redis.service

[Service]
Type=simple
User=nucleus
WorkingDirectory=/opt/nucleus
ExecStart=/opt/nucleus/nucleus-core
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable --now nucleus
```

## Troubleshooting

### Permission denied on PTY

Ensure your user is in the `tty` group:

```bash
sudo usermod -aG tty $(whoami)
# Log out and back in
```

### PostgreSQL peer authentication failed

Edit `pg_hba.conf` to use `md5` or `trust` for local connections:

```bash
sudo nano /etc/postgresql/15/main/pg_hba.conf
# Change "peer" to "trust" for local connections
sudo systemctl restart postgresql
```
