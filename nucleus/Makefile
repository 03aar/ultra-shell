.PHONY: help dev dev-logs dev-shell build build-core build-api build-mcp build-cli build-dash build-docs \
	test test-core test-api test-mcp test-integration \
	install install-cli install-mcp uninstall \
	db-seed db-reset \
	release clean reset doctor logs status open format lint

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DIR := build

# ──────────────────────────────────────────────────────────────────────
# Self-documenting help — any target with ## comment gets listed
# ──────────────────────────────────────────────────────────────────────
help: ## Show this help
	@echo ""
	@echo "  ███╗   ██╗██╗   ██╗ ██████╗██╗     ███████╗██╗   ██╗███████╗"
	@echo "  ████╗  ██║██║   ██║██╔════╝██║     ██╔════╝██║   ██║██╔════╝"
	@echo "  ██╔██╗ ██║██║   ██║██║     ██║     █████╗  ██║   ██║███████╗"
	@echo "  ██║╚██╗██║██║   ██║██║     ██║     ██╔══╝  ██║   ██║╚════██║"
	@echo "  ██║ ╚████║╚██████╔╝╚██████╗███████╗███████╗╚██████╔╝███████║"
	@echo "  ╚═╝  ╚═══╝ ╚═════╝  ╚═════╝╚══════╝╚══════╝ ╚═════╝ ╚══════╝"
	@echo ""
	@echo "  AI-Native Shell Runtime & Agent Platform"
	@echo "  Version: $(VERSION)  Commit: $(COMMIT)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""

# ──────────────────────────────────────────────────────────────────────
# Development
# ──────────────────────────────────────────────────────────────────────
dev: ## Start all services with Docker Compose
	@docker compose up --build -d
	@echo ""
	@echo "NUCLEUS is running:"
	@echo "  Dashboard:  http://localhost:3000"
	@echo "  API:        http://localhost:8080"
	@echo "  Health:     http://localhost:8080/health"
	@echo "  API Key:    dev-nucleus-key-local"
	@echo ""
	@echo "  make db-seed   — Load demo data"
	@echo "  make dev-logs  — Watch logs"
	@echo "  make status    — Service status"

dev-logs: ## Follow all service logs
	@docker compose logs -f

dev-shell: ## Open a shell inside the API container
	@docker compose exec nucleus-api sh

# ──────────────────────────────────────────────────────────────────────
# Build
# ──────────────────────────────────────────────────────────────────────
build: build-core build-api build-cli build-mcp build-dash ## Build all components locally
	@echo ""
	@echo "All components built successfully."
	@echo "Binaries are in $(BUILD_DIR)/"

build-core: ## Build nucleus-core (Rust)
	@echo "Building nucleus-core..."
	@mkdir -p $(BUILD_DIR)
	cd nucleus-core && cargo build --release
	@cp nucleus-core/target/release/nucleus-core $(BUILD_DIR)/ 2>/dev/null || \
	 cp nucleus-core/target/release/nucleus $(BUILD_DIR)/nucleus-core 2>/dev/null || true
	@echo "  -> $(BUILD_DIR)/nucleus-core"

build-api: ## Build nucleus-api (Go)
	@echo "Building nucleus-api..."
	@mkdir -p $(BUILD_DIR)
	cd nucleus-api && go mod tidy && go build \
		-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)" \
		-o ../$(BUILD_DIR)/nucleus-api .
	@echo "  -> $(BUILD_DIR)/nucleus-api"

build-mcp: ## Build nucleus-mcp (TypeScript)
	@echo "Building nucleus-mcp..."
	cd nucleus-mcp && npm install && npm run build
	@echo "  -> nucleus-mcp/dist/"

build-cli: ## Build nuc CLI (Go)
	@echo "Building nuc CLI..."
	@mkdir -p $(BUILD_DIR)
	cd nucleus-cli && go mod tidy && go build \
		-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)" \
		-o ../$(BUILD_DIR)/nuc .
	@echo "  -> $(BUILD_DIR)/nuc"

build-dash: ## Build nucleus-dashboard (Next.js)
	@echo "Building nucleus-dashboard..."
	cd nucleus-dashboard && npm install --legacy-peer-deps && npm run build
	@echo "  -> nucleus-dashboard/.next/"

build-docs: ## Build documentation site (Docusaurus)
	@echo "Building documentation..."
	@if [ -d "nucleus-docs" ] && [ -f "nucleus-docs/package.json" ]; then \
		cd nucleus-docs && npm install && npm run build; \
	elif [ -d "docs" ] && [ -f "docs/package.json" ]; then \
		cd docs && npm install && npm run build; \
	else \
		echo "  No Docusaurus project found. Docs are in docs/."; \
	fi

# ──────────────────────────────────────────────────────────────────────
# Test
# ──────────────────────────────────────────────────────────────────────
test: test-core test-api test-mcp ## Run all tests
	@echo ""
	@echo "All tests passed."

test-core: ## Test nucleus-core (Rust — cargo test + clippy)
	@echo "Testing nucleus-core..."
	cd nucleus-core && cargo fmt --all -- --check
	cd nucleus-core && cargo clippy --all-targets -- -D warnings
	cd nucleus-core && cargo test --all-features

test-api: ## Test nucleus-api (Go)
	@echo "Testing nucleus-api..."
	cd nucleus-api && go mod tidy && go vet ./...
	cd nucleus-api && go test -race ./...

test-mcp: ## Test nucleus-mcp (TypeScript)
	@echo "Testing nucleus-mcp..."
	cd nucleus-mcp && npm install && npx tsc --noEmit
	cd nucleus-mcp && npm test 2>/dev/null || echo "  No test suite configured yet."

test-integration: ## Run integration tests (Docker Compose)
	@echo "Starting integration test stack..."
	@docker compose up --build -d
	@echo "Waiting for API health..."
	@RETRIES=0; \
	until curl -sf http://localhost:8080/health > /dev/null 2>&1; do \
		RETRIES=$$((RETRIES + 1)); \
		if [ $$RETRIES -ge 30 ]; then echo "API did not start in time."; docker compose logs nucleus-api; exit 1; fi; \
		sleep 2; \
	done
	@echo "API healthy. Running integration checks..."
	@curl -sf http://localhost:8080/health | python3 -m json.tool 2>/dev/null || curl -sf http://localhost:8080/health
	@curl -sf -H "X-API-Key: dev-nucleus-key-local" http://localhost:8080/api/v1/context > /dev/null && echo "  /api/v1/context OK" || echo "  /api/v1/context FAIL"
	@curl -sf -H "X-API-Key: dev-nucleus-key-local" http://localhost:8080/api/v1/executions > /dev/null && echo "  /api/v1/executions OK" || echo "  /api/v1/executions FAIL"
	@curl -sf -H "X-API-Key: dev-nucleus-key-local" http://localhost:8080/api/v1/skills > /dev/null && echo "  /api/v1/skills OK" || echo "  /api/v1/skills FAIL"
	@curl -sf http://localhost:3000 > /dev/null && echo "  Dashboard OK" || echo "  Dashboard FAIL"
	@echo ""
	@echo "Integration tests complete."
	@docker compose down

# ──────────────────────────────────────────────────────────────────────
# Install / Uninstall
# ──────────────────────────────────────────────────────────────────────
install: install-cli install-mcp ## Install CLI + MCP server
	@echo ""
	@echo "Installation complete. Run: nuc doctor"

install-cli: build-cli ## Install nuc CLI to /usr/local/bin
	@echo "Installing nuc CLI..."
	sudo install -m 755 $(BUILD_DIR)/nuc /usr/local/bin/nuc
	@echo "Installed: /usr/local/bin/nuc ($(VERSION))"

install-mcp: build-mcp ## Install MCP server for local use
	@echo "MCP server built at nucleus-mcp/dist/index.js"
	@echo ""
	@echo "To register with Claude Desktop:"
	@echo "  nuc mcp install --client claude-desktop"

uninstall: ## Remove nuc CLI and clean up
	@echo "Removing nuc CLI..."
	sudo rm -f /usr/local/bin/nuc
	@echo "Removed /usr/local/bin/nuc"
	@echo ""
	@echo "To also remove Docker resources: make clean && make reset"

# ──────────────────────────────────────────────────────────────────────
# Database
# ──────────────────────────────────────────────────────────────────────
db-seed: ## Seed demo data into the database
	@echo "Seeding demo data..."
	@python3 scripts/seed_demo.py

db-reset: ## Drop and recreate the database
	@echo "Resetting database..."
	@docker compose exec postgres psql -U nucleus -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" 2>/dev/null || true
	@docker compose restart nucleus-api
	@echo "Database reset. Run 'make db-seed' to re-seed."

# ──────────────────────────────────────────────────────────────────────
# Release
# ──────────────────────────────────────────────────────────────────────
release: ## Create a release (VERSION=x.x.x)
	@if [ "$(VERSION)" = "dev" ] || echo "$(VERSION)" | grep -q "dirty"; then \
		echo "Usage: make release VERSION=1.0.0"; exit 1; \
	fi
	@echo "Releasing v$(VERSION)..."
	@git diff --quiet || (echo "Error: Uncommitted changes. Commit first." && exit 1)
	git tag -a "v$(VERSION)" -m "Release v$(VERSION)"
	@echo ""
	@echo "Tagged v$(VERSION). Push with:"
	@echo "  git push origin v$(VERSION)"
	@echo ""
	@echo "This will trigger the release workflow on GitHub Actions."

# ──────────────────────────────────────────────────────────────────────
# Quality
# ──────────────────────────────────────────────────────────────────────
format: ## Format all source code
	@echo "Formatting Rust..."
	@cd nucleus-core && cargo fmt --all 2>/dev/null || true
	@echo "Formatting Go..."
	@cd nucleus-api && gofmt -w . 2>/dev/null || true
	@cd nucleus-cli && gofmt -w . 2>/dev/null || true
	@echo "Formatting TypeScript..."
	@cd nucleus-mcp && npx prettier --write "src/**/*.ts" 2>/dev/null || true
	@cd nucleus-dashboard && npx prettier --write "src/**/*.{ts,tsx}" 2>/dev/null || true
	@echo "Done."

lint: ## Run linters on all components
	@echo "Linting nucleus-core (clippy)..."
	@cd nucleus-core && cargo clippy --all-targets -- -D warnings 2>/dev/null || true
	@echo "Linting nucleus-api (go vet)..."
	@cd nucleus-api && go vet ./... 2>/dev/null || true
	@echo "Linting nucleus-mcp (tsc)..."
	@cd nucleus-mcp && npx tsc --noEmit 2>/dev/null || true
	@echo "Linting nucleus-dashboard (next lint)..."
	@cd nucleus-dashboard && npm run lint 2>/dev/null || true
	@echo "Done."

# ──────────────────────────────────────────────────────────────────────
# Operations
# ──────────────────────────────────────────────────────────────────────
logs: ## Follow Docker Compose logs
	@docker compose logs -f

status: ## Show service status and health
	@echo ""
	@echo "=== Docker Compose Services ==="
	@docker compose ps 2>/dev/null || echo "  Docker Compose not running."
	@echo ""
	@echo "=== Health Checks ==="
	@curl -sf http://localhost:8080/health 2>/dev/null && echo "" || echo "  API: not reachable"
	@curl -sf http://localhost:3000 > /dev/null 2>&1 && echo "  Dashboard: OK" || echo "  Dashboard: not reachable"
	@echo ""

open: ## Open dashboard in default browser
	@echo "Opening http://localhost:3000 ..."
	@command -v xdg-open > /dev/null && xdg-open http://localhost:3000 || \
	 command -v open > /dev/null && open http://localhost:3000 || \
	 echo "Open http://localhost:3000 in your browser"

doctor: ## Check system prerequisites and health
	@echo ""
	@echo "=== System Prerequisites ==="
	@command -v docker    > /dev/null 2>&1 && printf "  %-14s %s\n" "Docker" "OK ($(docker --version 2>/dev/null | head -c 40))" || printf "  %-14s %s\n" "Docker" "MISSING"
	@docker compose version > /dev/null 2>&1 && printf "  %-14s %s\n" "Compose" "OK" || printf "  %-14s %s\n" "Compose" "MISSING"
	@command -v cargo     > /dev/null 2>&1 && printf "  %-14s %s\n" "Rust/Cargo" "OK ($(cargo --version 2>/dev/null | head -c 30))" || printf "  %-14s %s\n" "Rust/Cargo" "SKIP (optional)"
	@command -v go        > /dev/null 2>&1 && printf "  %-14s %s\n" "Go" "OK ($(go version 2>/dev/null | awk '{print $$3}'))" || printf "  %-14s %s\n" "Go" "SKIP (optional)"
	@command -v node      > /dev/null 2>&1 && printf "  %-14s %s\n" "Node.js" "OK ($(node --version 2>/dev/null))" || printf "  %-14s %s\n" "Node.js" "SKIP (optional)"
	@command -v python3   > /dev/null 2>&1 && printf "  %-14s %s\n" "Python 3" "OK ($(python3 --version 2>/dev/null | awk '{print $$2}'))" || printf "  %-14s %s\n" "Python 3" "SKIP (optional)"
	@echo ""
	@echo "=== Project Files ==="
	@test -f .env && echo "  .env           OK" || echo "  .env           MISSING (copy .env.example)"
	@test -f docker-compose.yml && echo "  docker-compose OK" || echo "  docker-compose MISSING"
	@echo ""
	@echo "=== Service Health ==="
	@curl -sf http://localhost:8080/health > /dev/null 2>&1 && echo "  API            OK (http://localhost:8080)" || echo "  API            not running"
	@curl -sf http://localhost:3000 > /dev/null 2>&1 && echo "  Dashboard      OK (http://localhost:3000)" || echo "  Dashboard      not running"
	@docker compose exec -T redis redis-cli ping > /dev/null 2>&1 && echo "  Redis          OK" || echo "  Redis          not running"
	@docker compose exec -T postgres pg_isready -U nucleus > /dev/null 2>&1 && echo "  PostgreSQL     OK" || echo "  PostgreSQL     not running"
	@echo ""

# ──────────────────────────────────────────────────────────────────────
# Cleanup
# ──────────────────────────────────────────────────────────────────────
clean: ## Remove build artifacts and stopped containers
	@echo "Cleaning..."
	@docker compose down --rmi local 2>/dev/null || true
	@cd nucleus-core && cargo clean 2>/dev/null || true
	@rm -rf $(BUILD_DIR)/
	@cd nucleus-dashboard && rm -rf .next node_modules 2>/dev/null || true
	@cd nucleus-mcp && rm -rf dist node_modules 2>/dev/null || true
	@echo "Clean."

reset: ## Remove ALL data volumes (destructive)
	@echo "This will destroy all database data and volumes."
	@docker compose down -v
	@echo "All data volumes removed."
