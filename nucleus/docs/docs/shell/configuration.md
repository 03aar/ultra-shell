---
sidebar_position: 4
title: Configuration
---

# Shell Configuration

Nucleus is configured through a TOML file at `~/.config/nucleus/config.toml`. All settings have sensible defaults — you only need to create this file if you want to customize behavior.

## Default Configuration

```toml
[core]
shell = ""                    # Auto-detected from $SHELL
data_dir = "~/.local/share/nucleus"
log_level = "info"            # "trace", "debug", "info", "warn", "error"

[api]
url = "http://localhost:8080"
key = "dev-nucleus-key-local"

[redis]
url = "redis://localhost:6379"

[annotations]
enabled = true
show_risk = true
show_duration = true
show_files = true
show_snapshot = true
min_risk_level = "low"
position = "below"

[risk]
block_critical = true          # Block critical-risk commands
warn_high = true               # Show warnings for high-risk commands
custom_rules = []              # Additional risk rules (see below)

[snapshots]
enabled = true
max_file_size = "100MB"        # Skip files larger than this
exclude_patterns = [           # Glob patterns to never snapshot
  "node_modules/**",
  ".git/objects/**",
  "target/**",
  "*.log",
]

[natural_language]
enabled = true
provider = "claude"
model = "claude-sonnet-4-5"
auto_execute = false

[session]
auto_name = true               # Automatically name sessions based on activity
idle_timeout = "30m"           # Create new session after idle period
```

## Custom Risk Rules

Add patterns to increase or decrease risk scores for specific commands:

```toml
[[risk.custom_rules]]
pattern = "kubectl delete"
risk_level = "high"
reason = "Kubernetes resource deletion"

[[risk.custom_rules]]
pattern = "terraform destroy"
risk_level = "critical"
reason = "Infrastructure destruction"

[[risk.custom_rules]]
pattern = "echo *"
risk_level = "none"
reason = "Echo is always safe"
```

## Snapshot Exclusions

Prevent snapshotting of large or irrelevant directories:

```toml
[snapshots]
exclude_patterns = [
  "node_modules/**",
  ".git/objects/**",
  "target/**",
  "*.log",
  "*.sqlite",
  "/tmp/**",
  "vendor/**",
]
```

## Environment Variables

All configuration can also be set via environment variables. Environment variables take precedence over the config file.

| Variable | Config Equivalent | Example |
|----------|-------------------|---------|
| `NUCLEUS_SHELL` | `core.shell` | `/bin/zsh` |
| `NUCLEUS_DATA_DIR` | `core.data_dir` | `/data/nucleus` |
| `NUCLEUS_LOG_LEVEL` | `core.log_level` | `debug` |
| `NUCLEUS_API_URL` | `api.url` | `http://localhost:8080` |
| `NUCLEUS_API_KEY` | `api.key` | `my-api-key` |
| `NUCLEUS_REDIS_URL` | `redis.url` | `redis://localhost:6379` |
| `NUCLEUS_ANNOTATIONS` | `annotations.enabled` | `false` |
| `NUCLEUS_NL_PROVIDER` | `natural_language.provider` | `openai` |
| `NUCLEUS_NL_MODEL` | `natural_language.model` | `gpt-4o` |
| `ANTHROPIC_API_KEY` | — | `sk-ant-...` |
| `OPENAI_API_KEY` | — | `sk-...` |
| `GOOGLE_API_KEY` | — | `AI...` |

## Per-Directory Configuration

Create a `.nucleus.toml` in any directory to override settings for that project:

```toml
# ~/myproject/.nucleus.toml

[risk]
# Treat all docker commands as high risk in this project
[[risk.custom_rules]]
pattern = "docker *"
risk_level = "high"
reason = "Production Docker environment"

[snapshots]
exclude_patterns = ["dist/**", "coverage/**"]

[natural_language]
provider = "ollama"
model = "codellama"
```

Per-directory config is merged with the global config, with per-directory values taking precedence.

## Resetting Configuration

```bash
# Remove config file (revert to defaults)
rm ~/.config/nucleus/config.toml

# Reset sled database (clear all execution history)
rm -rf ~/.local/share/nucleus/sled/

# Full reset
rm -rf ~/.config/nucleus ~/.local/share/nucleus
```
