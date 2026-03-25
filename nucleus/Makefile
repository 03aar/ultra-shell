.PHONY: dev build test install clean help

# Default target
help:
	@echo "NUCLEUS — AI-Native Shell Runtime"
	@echo ""
	@echo "Targets:"
	@echo "  make dev       Start all services with Docker Compose"
	@echo "  make build     Build all production binaries"
	@echo "  make test      Run all tests"
	@echo "  make install   Install nucleus binary to /usr/local/bin"
	@echo "  make clean     Remove build artifacts and volumes"
	@echo "  make stop      Stop all services"
	@echo "  make logs      Follow all service logs"
	@echo ""

# Start all services
dev:
	docker compose up --build -d
	@echo ""
	@echo "NUCLEUS is running:"
	@echo "  Dashboard:  http://localhost:3000"
	@echo "  API:        http://localhost:8080"
	@echo "  API Health: http://localhost:8080/health"
	@echo ""
	@echo "API Key: dev-nucleus-key-local"

# Build all production binaries
build:
	@echo "Building nucleus-core (Rust)..."
	cd nucleus-core && cargo build --release
	@echo "Building nucleus-api (Go)..."
	cd nucleus-api && go build -o ../build/nucleus-api .
	@echo "Building nucleus-dashboard (Next.js)..."
	cd nucleus-dashboard && npm install && npm run build
	@echo ""
	@echo "Build complete."

# Run all tests
test:
	@echo "Testing nucleus-core..."
	cd nucleus-core && cargo test
	@echo ""
	@echo "Testing nucleus-api..."
	cd nucleus-api && go test ./...
	@echo ""
	@echo "All tests passed."

# Install nucleus binary
install: build
	@echo "Installing nucleus to /usr/local/bin..."
	sudo cp nucleus-core/target/release/nucleus /usr/local/bin/nucleus
	sudo chmod +x /usr/local/bin/nucleus
	@echo "Installed. Run 'nucleus' to start."

# Stop all services
stop:
	docker compose down

# View logs
logs:
	docker compose logs -f

# Clean everything
clean:
	docker compose down -v --rmi local
	cd nucleus-core && cargo clean
	cd nucleus-api && rm -f nucleus-api
	cd nucleus-dashboard && rm -rf .next node_modules
	rm -rf build/
	@echo "Cleaned."
