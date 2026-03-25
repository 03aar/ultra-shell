---
sidebar_position: 101
title: Changelog
---

# Changelog

All notable changes to Nucleus are documented here.

## [Unreleased]

### Added
- Docusaurus documentation site with full API, CLI, MCP, and SDK reference.
- Natural language command support with `?` prefix in shell sessions.
- Agent orchestration engine with Claude, GPT, Gemini, and Ollama providers.
- File snapshot and rollback system in `nucleus-core`.
- WebSocket real-time event streaming.
- Session replay with asciicast format.
- Custom skills in YAML format.
- Python, TypeScript, and Go client SDKs.

### Changed
- Execution DAG storage migrated from in-memory to sled embedded database.
- Risk evaluator expanded with configurable custom rules.
- API response format standardized with consistent JSON envelope.

### Fixed
- PTY interception correctly handles multi-byte Unicode characters.
- Redis reconnection logic on connection drops.
- Snapshot exclusion patterns now support recursive globs.

## [0.1.0] — Initial Release

### Added
- `nucleus-core`: Rust shell runtime with PTY interception, command parsing, and risk evaluation.
- `nucleus-api`: Go REST API server with execution, context, session, and rollback endpoints.
- `nucleus-cli`: `nuc` CLI tool with shell, history, exec, rollback, graph, session, and skill commands.
- `nucleus-mcp`: MCP server with 10 tools, 4 resources, and 3 prompts.
- `nucleus-dashboard`: Next.js web dashboard with execution feed, DAG visualization, and session replay.
- `nucleus-orchestrator`: Python agent engine with multi-provider support.
- Docker Compose setup for full-stack deployment.
- PostgreSQL schema for sessions, API keys, audit logs, and skills.
- Redis pub/sub for real-time event broadcasting.
- Development API key (`dev-nucleus-key-local`) for local development.
- Built-in skills: `git_cleanup`, `docker_cleanup`, `project_setup`, `deploy_check`, `env_audit`, `process_debug`.
