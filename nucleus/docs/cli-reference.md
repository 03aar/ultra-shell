# CLI Reference — `nuc`

## Installation

```bash
make install-cli
# or
cd nucleus-cli && go build -o nuc . && sudo mv nuc /usr/local/bin/
```

## Commands

### Shell & Status
```bash
nuc shell                    # Launch Nucleus shell
nuc status                   # Show session status
nuc doctor                   # Check installation health
nuc dashboard                # Open web dashboard
```

### History & Context
```bash
nuc history                  # Show execution history
nuc history --limit 20       # Last 20 commands
nuc history --search "docker" # Search history
nuc context                  # Show environment state
nuc context --json           # JSON output
```

### Execution & Rollback
```bash
nuc exec "ls -la"            # Execute through Nucleus
nuc exec "rm -rf tmp" --dry-run  # Risk assessment only
nuc rollback <execution-id>  # Rollback specific execution
nuc rollback --last          # Rollback last command
```

### Graph
```bash
nuc graph                    # Show execution DAG
nuc graph --format mermaid   # Mermaid diagram output
nuc graph --format json      # Raw JSON
```

### Sessions
```bash
nuc session list             # List all sessions
nuc session new "deploy-v2"  # Create new session
nuc session replay <id>      # Replay session
```

### Skills
```bash
nuc skill list               # List available skills
nuc skill run git_cleanup    # Run a skill
```

### Agent & Planning
```bash
nuc plan "set up CI/CD"      # Generate execution plan
nuc watch                    # Live event stream
```

### MCP
```bash
nuc mcp install --client claude-desktop  # Install MCP config
nuc mcp status               # Check MCP server
```

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--api-url` | `http://localhost:8080` | API server URL |
| `--api-key` | `dev-nucleus-key-local` | API key |
| `--output` | `text` | Output format (text/json) |
| `--no-color` | `false` | Disable colors |
