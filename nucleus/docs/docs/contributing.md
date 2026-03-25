---
sidebar_position: 100
title: Contributing
---

# Contributing to Nucleus

Nucleus is open source under the MIT license. Contributions are welcome.

## Development Setup

```bash
git clone https://github.com/03aar/ultra-shell.git
cd ultra-shell/nucleus
cp .env.example .env
docker compose up --build -d
```

### Component Development

Each component can be developed independently:

| Component | Language | Dev Command |
|-----------|----------|-------------|
| `nucleus-core` | Rust | `cd nucleus-core && cargo watch -x run` |
| `nucleus-api` | Go | `cd nucleus-api && go run .` |
| `nucleus-cli` | Go | `cd nucleus-cli && go run . doctor` |
| `nucleus-mcp` | TypeScript | `cd nucleus-mcp && npm run dev` |
| `nucleus-dashboard` | Next.js | `cd nucleus-dashboard && npm run dev` |
| `nucleus-orchestrator` | Python | `cd nucleus-orchestrator && python -m nucleus_agent` |

### Running Tests

```bash
# All tests
make test

# Per component
cd nucleus-core && cargo test
cd nucleus-api && go test ./...
cd nucleus-cli && go test ./...
cd nucleus-mcp && npm test
cd nucleus-dashboard && npm test
cd nucleus-orchestrator && pytest
```

## Project Structure

```
nucleus/
├── nucleus-core/          # Rust shell runtime
│   ├── src/
│   │   ├── main.rs        # Entry point, PTY setup
│   │   ├── parser.rs      # Command parser
│   │   ├── evaluator.rs   # Risk evaluator
│   │   ├── snapshot.rs    # File snapshot engine
│   │   ├── graph.rs       # Execution DAG
│   │   └── redis.rs       # Event publisher
│   └── Cargo.toml
├── nucleus-api/           # Go API server
│   ├── main.go
│   ├── handlers/          # HTTP handlers
│   ├── middleware/         # Auth, CORS, logging
│   ├── database/          # PostgreSQL schema and queries
│   └── websocket/         # WebSocket hub
├── nucleus-cli/           # Go CLI
├── nucleus-mcp/           # TypeScript MCP server
├── nucleus-dashboard/     # Next.js web UI
├── nucleus-orchestrator/  # Python agent engine
├── nucleus-sdk/           # Client SDKs
│   ├── python/
│   ├── typescript/
│   └── go/
├── docs/                  # This documentation site
├── scripts/               # Build and deployment scripts
├── docker-compose.yml
└── Makefile
```

## Making Changes

1. **Fork** the repository.
2. **Create a branch** from `main`: `git checkout -b feature/my-feature`.
3. **Make your changes** with clear, focused commits.
4. **Add tests** for any new functionality.
5. **Run the test suite** to ensure nothing is broken.
6. **Open a pull request** against `main`.

## Code Style

| Language | Style | Tool |
|----------|-------|------|
| Rust | Standard Rust style | `cargo fmt`, `cargo clippy` |
| Go | Standard Go style | `gofmt`, `golangci-lint` |
| TypeScript | Prettier + ESLint | `npm run lint`, `npm run format` |
| Python | Black + Ruff | `black .`, `ruff check .` |

## Commit Messages

Use conventional commits:

```
feat: add session replay endpoint
fix: correct risk score for piped commands
docs: add MCP tools reference
refactor: extract snapshot logic into separate module
test: add integration tests for rollback API
```

## Areas for Contribution

- **New skills:** Add reusable workflows in YAML format.
- **Provider support:** Add new AI providers to the orchestrator.
- **Risk rules:** Improve the command risk evaluator with better heuristics.
- **Dashboard features:** Add visualizations, improve UX.
- **SDK improvements:** Better error handling, new language SDKs.
- **Documentation:** Fix errors, add examples, improve clarity.

## Reporting Issues

File issues at [github.com/03aar/ultra-shell/issues](https://github.com/03aar/ultra-shell/issues) with:

- A clear title describing the problem.
- Steps to reproduce.
- Expected vs actual behavior.
- Your OS, shell, and Nucleus version (`nuc --version`).
