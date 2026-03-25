#!/usr/bin/env bash
#
# NUCLEUS Installer — Mac / Linux / Ubuntu / WSL
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/<repo>/main/scripts/install.sh | bash
#   OR
#   ./scripts/install.sh
#
set -euo pipefail

BOLD="\033[1m"
GREEN="\033[32m"
YELLOW="\033[33m"
RED="\033[31m"
RESET="\033[0m"

info()  { echo -e "${GREEN}[NUCLEUS]${RESET} $*"; }
warn()  { echo -e "${YELLOW}[NUCLEUS]${RESET} $*"; }
error() { echo -e "${RED}[NUCLEUS]${RESET} $*"; exit 1; }

# ── Detect OS ────────────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux*)   PLATFORM="linux"   ;;
  Darwin*)  PLATFORM="darwin"  ;;
  CYGWIN*|MINGW*|MSYS*) PLATFORM="windows" ;;
  *)        error "Unsupported OS: $OS" ;;
esac

case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)             error "Unsupported architecture: $ARCH" ;;
esac

info "Detected platform: ${PLATFORM}/${ARCH}"

# ── Check prerequisites ─────────────────────────────
check_cmd() {
  if ! command -v "$1" &>/dev/null; then
    return 1
  fi
  return 0
}

MISSING=()

# Docker is required for the full stack
if ! check_cmd docker; then
  MISSING+=("docker")
fi

if ! check_cmd docker && ! docker compose version &>/dev/null 2>&1; then
  MISSING+=("docker-compose")
fi

if [ ${#MISSING[@]} -gt 0 ]; then
  warn "Missing required tools: ${MISSING[*]}"
  echo ""
  echo "Install Docker:"
  case "$PLATFORM" in
    darwin)
      echo "  brew install --cask docker"
      echo "  OR download from https://docs.docker.com/desktop/install/mac-install/"
      ;;
    linux)
      echo "  curl -fsSL https://get.docker.com | sh"
      echo "  sudo usermod -aG docker \$USER"
      ;;
    windows)
      echo "  Download Docker Desktop from https://docs.docker.com/desktop/install/windows-install/"
      ;;
  esac
  echo ""
  error "Please install the missing tools and re-run this script."
fi

info "All prerequisites met."

# ── Optional: Build from source ──────────────────────
BUILD_FROM_SOURCE=false

# Check for Rust/Go/Node for local builds
if check_cmd cargo && check_cmd go && check_cmd node; then
  info "Rust, Go, and Node.js detected — local build available."
  BUILD_FROM_SOURCE=true
fi

# ── Clone or use existing repo ───────────────────────
NUCLEUS_DIR="$(pwd)"
if [ ! -f "$NUCLEUS_DIR/docker-compose.yml" ]; then
  if [ -f "$NUCLEUS_DIR/nucleus/docker-compose.yml" ]; then
    NUCLEUS_DIR="$NUCLEUS_DIR/nucleus"
  else
    error "Could not find nucleus project. Run this script from the nucleus/ directory."
  fi
fi

info "Using project directory: $NUCLEUS_DIR"

# ── Create .env file ────────────────────────────────
if [ ! -f "$NUCLEUS_DIR/.env" ]; then
  cat > "$NUCLEUS_DIR/.env" <<'ENVEOF'
# NUCLEUS Environment Configuration
NUCLEUS_REDIS_URL=redis://redis:6379
DATABASE_URL=postgres://nucleus:nucleus_dev_password@postgres:5432/nucleus?sslmode=disable
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080/api/v1
NEXT_PUBLIC_API_KEY=dev-nucleus-key-local
ENVEOF
  info "Created .env file with default configuration."
fi

# ── Build and start ─────────────────────────────────
info "Building and starting NUCLEUS..."
cd "$NUCLEUS_DIR"

docker compose build --parallel 2>&1 | tail -5
docker compose up -d

echo ""
info "NUCLEUS is running!"
echo ""
echo -e "  ${BOLD}Dashboard${RESET}:  http://localhost:3000"
echo -e "  ${BOLD}API${RESET}:        http://localhost:8080"
echo -e "  ${BOLD}API Health${RESET}: http://localhost:8080/health"
echo ""
echo -e "  ${BOLD}API Key${RESET}:    dev-nucleus-key-local"
echo ""
echo -e "  Manage:  ${GREEN}docker compose logs -f${RESET}  (watch logs)"
echo -e "           ${GREEN}docker compose down${RESET}     (stop)"
echo -e "           ${GREEN}docker compose restart${RESET}  (restart)"
echo ""

# ── Optional: Install CLI binary locally ─────────────
if [ "$BUILD_FROM_SOURCE" = true ]; then
  echo ""
  echo -e "To install the ${BOLD}nucleus${RESET} CLI locally:"
  echo -e "  cd nucleus-core && cargo build --release"
  echo -e "  sudo cp target/release/nucleus /usr/local/bin/nucleus"
  echo ""
fi

info "Installation complete."
