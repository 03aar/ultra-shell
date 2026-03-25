---
sidebar_position: 1
title: Shell Runtime Overview
---

# Shell Runtime

`nucleus-core` is the heart of Nucleus. Written in Rust, it wraps your existing shell via PTY (pseudo-terminal) interception, making every command you run observable, reversible, and AI-accessible.

## How It Works

When you run `nuc shell`, Nucleus spawns your default shell (`$SHELL`, typically bash or zsh) inside a PTY. It sits between your terminal emulator and the child shell, intercepting all input and output.

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Terminal    │────>│ nucleus-core │────>│  Child Shell │
│  (your app)  │<────│  (Rust PTY)  │<────│  (bash/zsh)  │
└──────────────┘     └──────────────┘     └──────────────┘
                            │
                     ┌──────┴──────┐
                     │   sled DB   │  Local execution graph
                     └──────┬──────┘
                            │
                     ┌──────┴──────┐
                     │    Redis    │  Pub/sub to API layer
                     └─────────────┘
```

Your shell behaves exactly as before. Nucleus adds capabilities without changing the interface.

## Command Pipeline

Every command flows through this pipeline:

### 1. Parsing (`parser.rs`)

Raw input is parsed into a `ParsedCommand` structure:

```rust
pub struct ParsedCommand {
    pub raw: String,           // Original input
    pub binary: String,        // e.g., "git"
    pub args: Vec<String>,     // e.g., ["commit", "-m", "fix"]
    pub category: Category,    // e.g., VCS, FileSystem, Network
    pub pipes: Vec<PipeSegment>,
    pub redirects: Vec<Redirect>,
    pub env_overrides: HashMap<String, String>,
    pub background: bool,
}
```

Categories include: `FileSystem`, `Network`, `Process`, `VCS`, `Package`, `Build`, `Docker`, `System`, `Shell`, `Unknown`.

### 2. Risk Evaluation (`evaluator.rs`)

Each command receives a risk score and classification:

| Risk Level | Score | Example |
|-----------|-------|---------|
| `none` | 0 | `echo hello`, `pwd` |
| `low` | 1-3 | `ls`, `cat file.txt`, `git status` |
| `medium` | 4-6 | `npm install`, `docker build` |
| `high` | 7-8 | `rm -rf ./build`, `chmod 777` |
| `critical` | 9-10 | `rm -rf /`, `mkfs`, `dd if=/dev/zero` |

Commands scored `critical` are **blocked by default**. Commands scored `high` produce a warning annotation in the terminal.

### 3. File Snapshot (`snapshot.rs`)

Before executing commands that modify files, Nucleus captures the current state of affected files. Snapshots are stored in sled and limited to 100MB per file.

Snapshotted operations include:
- File writes and deletions (`rm`, `mv`, redirects with `>`)
- Package installations (`npm install`, `pip install`)
- Configuration changes

### 4. Execution

The command runs in the child shell. Nucleus captures:
- **stdout** and **stderr** (streaming)
- **Exit code**
- **Duration** (milliseconds)
- **Working directory** at time of execution

### 5. DAG Recording (`graph.rs`)

The execution is stored as a node in a directed acyclic graph:

```rust
pub struct ExecutionNode {
    pub id: Uuid,
    pub command: ParsedCommand,
    pub exit_code: i32,
    pub stdout: String,
    pub stderr: String,
    pub duration_ms: u64,
    pub timestamp: DateTime<Utc>,
    pub session_id: Uuid,
    pub risk: RiskAssessment,
    pub snapshot_id: Option<Uuid>,
    pub parent_ids: Vec<Uuid>,
}
```

Edges are created based on:
- Sequential execution order within a session
- Shared file dependencies (command A writes a file that command B reads)
- Pipe chains

### 6. Event Publishing (`redis.rs`)

The execution event is published to Redis on the `nucleus:executions` channel. The API server subscribes to this channel and broadcasts to WebSocket clients.

## Storage

`nucleus-core` uses [sled](https://github.com/spacejam/sled), an embedded database written in Rust. No external database is needed for the core runtime.

Sled stores:
- Execution nodes and their metadata
- Graph edges (parent/child relationships)
- File snapshots (pre-mutation state)
- Session metadata

Data directory: `~/.local/share/nucleus/sled/` (configurable via `NUCLEUS_DATA_DIR`).

## Next Steps

- [Terminal Annotations](/docs/shell/annotations) — Visual feedback in the shell.
- [Natural Language Commands](/docs/shell/natural-language) — The `?` prefix for NL queries.
- [Configuration](/docs/shell/configuration) — Customize the shell runtime.
