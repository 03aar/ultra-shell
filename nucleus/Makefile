.PHONY: dev build test install clean help stop logs setup

# Default target
help:
	@echo "NUCLEUS — AI-Native Shell Runtime"
	@echo ""
	@echo "Quick Start:"
	@echo "  make dev         Start all services with Docker Compose"
	@echo ""
	@echo "Development:"
	@echo "  make build       Build all production binaries locally"
	@echo "  make test        Run all tests"
	@echo "  make install     Install nucleus binary to /usr/local/bin"
	@echo ""
	@echo "Operations:"
	@echo "  make stop        Stop all services"
	@echo "  make logs        Follow all service logs"
	@echo "  make clean       Remove build artifacts and volumes"
	@echo "  make setup       Check prerequisites and show setup guide"
	@echo ""

# Check prerequisites
setup:
	@echo "Checking prerequisites..."
	@command -v docker >/dev/null 2>&1 && echo "  [OK] Docker" || echo "  [MISSING] Docker — install from https://docs.docker.com/get-docker/"
	@docker compose version >/dev/null 2>&1 && echo "  [OK] Docker Compose" || echo "  [MISSING] Docker Compose"
	@command -v cargo >/dev/null 2>&1 && echo "  [OK] Rust/Cargo (for local build)" || echo "  [SKIP] Rust/Cargo (optional, for local build)"
	@command -v go >/dev/null 2>&1 && echo "  [OK] Go (for local build)" || echo "  [SKIP] Go (optional, for local build)"
	@command -v node >/dev/null 2>&1 && echo "  [OK] Node.js (for local build)" || echo "  [SKIP] Node.js (optional, for local build)"
	@echo ""
	@echo "To start NUCLEUS with Docker: make dev"

# Start all services
dev:
	docker compose up --build -d
	@echo ""
	@echo "NUCLEUS is running:"
	@echo "  Dashboard:  http://localhost:3000"
	@echo "  API:        http://localhost:8080"
	@echo "  API Health: http://localhost:8080/health"
	@echo ""
	@echo "  API Key: dev-nucleus-key-local"
	@echo ""
	@echo "  Logs:  make logs"
	@echo "  Stop:  make stop"

# Build all production binaries (requires Rust, Go, Node.js locally)
build:
	@echo "Building nucleus-core (Rust)..."
	cd nucleus-core && cargo build --release
	@echo ""
	@echo "Building nucleus-api (Go)..."
	cd nucleus-api && go mod tidy && go build -o ../build/nucleus-api .
	@echo ""
	@echo "Building nucleus-dashboard (Next.js)..."
	cd nucleus-dashboard && npm install --legacy-peer-deps && npm run build
	@echo ""
	@echo "Build complete. Binaries in:"
	@echo "  nucleus-core/target/release/nucleus"
	@echo "  build/nucleus-api"

# Run all tests
test:
	@echo "Testing nucleus-core..."
	cd nucleus-core && cargo test
	@echo ""
	@echo "Testing nucleus-api..."
	cd nucleus-api && go mod tidy && go test ./...
	@echo ""
	@echo "All tests passed."

# Install nucleus binary (Unix only)
install:
	@echo "Building nucleus..."
	cd nucleus-core && cargo build --release
	@echo "Installing to /usr/local/bin/nucleus..."
	sudo cp nucleus-core/target/release/nucleus /usr/local/bin/nucleus
	sudo chmod +x /usr/local/bin/nucleus
	@echo ""
	@echo "Installed. Run 'nucleus' to start an AI-native shell session."

# Stop all services
stop:
	docker compose down

# View logs
logs:
	docker compose logs -f

# Clean everything
clean:
	docker compose down -v --rmi local 2>/dev/null || true
	cd nucleus-core && cargo clean 2>/dev/null || true
	rm -rf build/
	cd nucleus-dashboard && rm -rf .next node_modules 2>/dev/null || true
	@echo "Cleaned."
