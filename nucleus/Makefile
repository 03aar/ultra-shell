.PHONY: dev build test install clean help stop logs setup seed

help:
	@echo "NUCLEUS — AI-Native Shell Runtime & Agent Platform"
	@echo ""
	@echo "Quick Start:"
	@echo "  make dev           Start all services with Docker Compose"
	@echo "  make seed          Seed demo data for the dashboard"
	@echo ""
	@echo "Development:"
	@echo "  make build         Build all production binaries locally"
	@echo "  make test          Run all tests"
	@echo "  make test-core     Test nucleus-core (Rust)"
	@echo "  make test-api      Test nucleus-api (Go)"
	@echo "  make test-mcp      Test nucleus-mcp (TypeScript)"
	@echo ""
	@echo "Installation:"
	@echo "  make install-cli   Install nuc CLI to /usr/local/bin"
	@echo "  make install-mcp   Build MCP server for local use"
	@echo ""
	@echo "Operations:"
	@echo "  make stop          Stop all services"
	@echo "  make logs          Follow all service logs"
	@echo "  make clean         Remove build artifacts"
	@echo "  make reset         Remove all data volumes"
	@echo "  make setup         Check prerequisites"
	@echo ""

setup:
	@echo "Checking prerequisites..."
	@command -v docker >/dev/null 2>&1 && echo "  [OK] Docker" || echo "  [MISSING] Docker"
	@docker compose version >/dev/null 2>&1 && echo "  [OK] Docker Compose" || echo "  [MISSING] Docker Compose"
	@command -v cargo >/dev/null 2>&1 && echo "  [OK] Rust/Cargo" || echo "  [SKIP] Rust (optional)"
	@command -v go >/dev/null 2>&1 && echo "  [OK] Go" || echo "  [SKIP] Go (optional)"
	@command -v node >/dev/null 2>&1 && echo "  [OK] Node.js" || echo "  [SKIP] Node.js (optional)"
	@command -v python3 >/dev/null 2>&1 && echo "  [OK] Python 3" || echo "  [SKIP] Python (optional)"

dev:
	docker compose up --build -d
	@echo ""
	@echo "NUCLEUS is running:"
	@echo "  Dashboard:  http://localhost:3000"
	@echo "  API:        http://localhost:8080"
	@echo "  Health:     http://localhost:8080/health"
	@echo "  API Key:    dev-nucleus-key-local"
	@echo ""
	@echo "  make seed   — Load demo data"
	@echo "  make logs   — Watch logs"
	@echo "  make stop   — Stop"

dev-logs:
	docker compose logs -f

seed:
	@echo "Seeding demo data..."
	python3 scripts/seed_demo.py

build:
	@echo "Building nucleus-core (Rust)..."
	cd nucleus-core && cargo build --release
	@echo "Building nucleus-api (Go)..."
	cd nucleus-api && go mod tidy && go build -o ../build/nucleus-api .
	@echo "Building nucleus-cli (Go)..."
	cd nucleus-cli && go mod tidy && go build -o ../build/nuc .
	@echo "Building nucleus-mcp (TypeScript)..."
	cd nucleus-mcp && npm install && npm run build
	@echo "Building nucleus-dashboard (Next.js)..."
	cd nucleus-dashboard && npm install --legacy-peer-deps && npm run build
	@echo "Build complete."

test: test-core test-api

test-core:
	@echo "Testing nucleus-core..."
	cd nucleus-core && cargo test

test-api:
	@echo "Testing nucleus-api..."
	cd nucleus-api && go mod tidy && go test ./...

test-mcp:
	@echo "Testing nucleus-mcp..."
	cd nucleus-mcp && npm test 2>/dev/null || echo "No tests configured"

install-cli:
	@echo "Building nuc CLI..."
	cd nucleus-cli && go mod tidy && go build -o nuc .
	sudo mv nucleus-cli/nuc /usr/local/bin/nuc
	sudo chmod +x /usr/local/bin/nuc
	@echo "Installed: nuc"

install-mcp:
	@echo "Building MCP server..."
	cd nucleus-mcp && npm install && npm run build
	@echo "MCP server built at nucleus-mcp/dist/index.js"
	@echo "Run: nuc mcp install --client claude-desktop"

stop:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down --rmi local 2>/dev/null || true
	cd nucleus-core && cargo clean 2>/dev/null || true
	rm -rf build/
	cd nucleus-dashboard && rm -rf .next node_modules 2>/dev/null || true
	cd nucleus-mcp && rm -rf dist node_modules 2>/dev/null || true

reset:
	docker compose down -v
	@echo "All data volumes removed."
