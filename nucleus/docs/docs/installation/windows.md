---
sidebar_position: 4
title: Windows (WSL2)
---

# Windows Installation

Nucleus runs on Windows through WSL2 (Windows Subsystem for Linux). Native Windows support is not available because Nucleus relies on Unix PTY for shell interception.

## Prerequisites

### Install WSL2

Open PowerShell as Administrator:

```powershell
wsl --install -d Ubuntu-22.04
```

Restart your computer if prompted. After reboot, the Ubuntu terminal will open and ask you to create a user account.

### Install Docker Desktop (Recommended)

1. Download [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop/).
2. During installation, enable **WSL2 backend**.
3. In Docker Desktop settings, go to **Resources > WSL Integration** and enable your Ubuntu distro.

## Option A: Docker Compose (Recommended)

Inside your WSL2 Ubuntu terminal:

```bash
git clone https://github.com/03aar/ultra-shell.git
cd ultra-shell/nucleus
cp .env.example .env
docker compose up --build -d
```

Install the CLI:

```bash
make install-cli
nuc doctor
```

Access the dashboard at `http://localhost:3000` from your Windows browser.

## Option B: Native Build in WSL2

Follow the [Linux installation guide](/docs/installation/linux) inside your WSL2 terminal. The Ubuntu/Debian instructions apply directly.

## Accessing from Windows

Services running in WSL2 are accessible from Windows at `localhost`:

| Service | URL |
|---------|-----|
| Dashboard | http://localhost:3000 |
| API | http://localhost:8080 |
| Health | http://localhost:8080/health |

## VS Code Integration

Install the [WSL extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-wsl) for VS Code, then:

```bash
# From WSL2 terminal
code ~/ultra-shell/nucleus
```

This opens the project in VS Code with full WSL2 integration.

## Claude Desktop on Windows

To connect Claude Desktop (Windows) to Nucleus running in WSL2:

1. Install Node.js on Windows (not WSL2).
2. Build `nucleus-mcp` on Windows or copy the built files.
3. Configure Claude Desktop to point at the WSL2 API:

```json
{
  "mcpServers": {
    "nucleus": {
      "command": "node",
      "args": ["C:\\path\\to\\nucleus-mcp\\dist\\index.js"],
      "env": {
        "NUCLEUS_API_URL": "http://localhost:8080",
        "NUCLEUS_API_KEY": "dev-nucleus-key-local"
      }
    }
  }
}
```

## Troubleshooting

### WSL2 networking issues

If `localhost` does not resolve from Windows, find the WSL2 IP:

```bash
# In WSL2
hostname -I
```

Use that IP instead of `localhost` in your browser and configurations.

### Docker not available in WSL2

Ensure Docker Desktop WSL2 integration is enabled:
1. Open Docker Desktop.
2. Go to Settings > Resources > WSL Integration.
3. Enable your Ubuntu distribution.
4. Click Apply & Restart.
