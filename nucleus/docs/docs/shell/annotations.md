---
sidebar_position: 2
title: Terminal Annotations
---

# Terminal Annotations

Nucleus adds inline annotations to your terminal output. These are non-intrusive visual hints that appear alongside command results, giving you real-time feedback on risk, timing, and status.

## What Annotations Look Like

```bash
$ rm -rf ./build
  ⚠ risk:high  files:47  snapshot:taken  duration:120ms

$ git push origin main
  ✓ risk:low  duration:3200ms

$ curl https://api.example.com/data
  ✓ risk:low  network:external  duration:890ms

$ rm -rf /
  ✗ BLOCKED  risk:critical  reason:root-filesystem-deletion
```

## Annotation Fields

Each annotation can include:

| Field | Description | Example |
|-------|-------------|---------|
| `risk` | Risk level from evaluator | `risk:low`, `risk:high` |
| `duration` | Execution time | `duration:120ms` |
| `files` | Number of files affected | `files:47` |
| `snapshot` | Whether a snapshot was taken | `snapshot:taken` |
| `network` | Network access type | `network:external` |
| `exit` | Non-zero exit code | `exit:1` |
| `rollback` | Rollback availability | `rollback:available` |
| `reason` | Block/warning reason | `reason:root-filesystem-deletion` |

## Risk Indicators

| Symbol | Meaning |
|--------|---------|
| `✓` | Command completed successfully, low/no risk |
| `⚠` | Command completed but was flagged as risky |
| `✗` | Command was blocked or failed |
| `⏳` | Command is currently executing (streaming) |

## Configuration

Annotations are controlled in your Nucleus configuration file (`~/.config/nucleus/config.toml`):

```toml
[annotations]
enabled = true          # Master toggle
show_risk = true        # Show risk level
show_duration = true    # Show execution time
show_files = true       # Show affected file count
show_snapshot = true    # Show snapshot status
show_network = true     # Show network access
min_risk_level = "low"  # Minimum risk level to annotate ("none", "low", "medium", "high")
position = "below"      # "below" (default) or "inline"
```

### Disable Annotations

```toml
[annotations]
enabled = false
```

### Show Only High-Risk Warnings

```toml
[annotations]
enabled = true
min_risk_level = "high"
```

## Annotations in Non-Interactive Mode

When commands are executed through the API or MCP (non-interactive), annotations are included in the response payload as structured data rather than terminal escape sequences:

```json
{
  "annotations": {
    "risk": "medium",
    "duration_ms": 450,
    "files_affected": 3,
    "snapshot_taken": true,
    "rollback_available": true
  }
}
```

This allows AI agents and dashboards to consume annotation data programmatically.
