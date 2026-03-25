#!/usr/bin/env bash
#
# NUCLEUS Installer
#
# Usage:
#   curl -fsSL https://install.nucleusshell.dev | bash
#   OR
#   ./scripts/install.sh
#
# Options (env vars):
#   NUCLEUS_VERSION   — specific version to install (default: latest)
#   NUCLEUS_DIR       — install directory (default: ~/.nucleus)
#   NUCLEUS_NO_DOCKER — skip Docker service startup (default: false)
#   NUCLEUS_NO_MCP    — skip MCP server installation (default: false)
#
set -euo pipefail

# ── Colors ────────────────────────────────────────────
BOLD="\033[1m"
DIM="\033[2m"
GREEN="\033[32m"
YELLOW="\033[33m"
RED="\033[31m"
CYAN="\033[36m"
BLUE="\033[34m"
RESET="\033[0m"

info()    { echo -e "${GREEN}[nucleus]${RESET} $*"; }
warn()    { echo -e "${YELLOW}[nucleus]${RESET} $*"; }
error()   { echo -e "${RED}[nucleus]${RESET} $*"; exit 1; }
step()    { echo -e "${CYAN}  -->  ${RESET}$*"; }
success() { echo -e "${GREEN}  [OK]${RESET} $*"; }
skip()    { echo -e "${DIM}  [SKIP]${RESET} $*"; }

# ── Banner ────────────────────────────────────────────
banner() {
    echo -e "${BLUE}"
    echo '  ███╗   ██╗██╗   ██╗ ██████╗██╗     ███████╗██╗   ██╗███████╗'
    echo '  ████╗  ██║██║   ██║██╔════╝██║     ██╔════╝██║   ██║██╔════╝'
    echo '  ██╔██╗ ██║██║   ██║██║     ██║     █████╗  ██║   ██║███████╗'
    echo '  ██║╚██╗██║██║   ██║██║     ██║     ██╔══╝  ██║   ██║╚════██║'
    echo '  ██║ ╚████║╚██████╔╝╚██████╗███████╗███████╗╚██████╔╝███████║'
    echo '  ╚═╝  ╚═══╝ ╚═════╝  ╚═════╝╚══════╝╚══════╝ ╚═════╝ ╚══════╝'
    echo -e "${RESET}"
    echo -e "  ${BOLD}AI-Native Shell Runtime & Agent Platform${RESET}"
    echo ""
}

# ── Detect OS / Architecture ─────────────────────────
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Linux*)
            PLATFORM="linux"
            # Detect WSL
            if grep -qiE "(microsoft|wsl)" /proc/version 2>/dev/null; then
                IS_WSL=true
            else
                IS_WSL=false
            fi
            ;;
        Darwin*)
            PLATFORM="darwin"
            IS_WSL=false
            ;;
        CYGWIN*|MINGW*|MSYS*)
            PLATFORM="windows"
            IS_WSL=false
            ;;
        *)
            error "Unsupported operating system: $OS"
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)   ARCH="amd64" ;;
        aarch64|arm64)  ARCH="arm64" ;;
        armv7l)         ARCH="arm"   ;;
        *)              error "Unsupported architecture: $ARCH" ;;
    esac

    info "Platform: ${PLATFORM}/${ARCH}$([ "$IS_WSL" = true ] && echo ' (WSL)')"
}

# ── Check Dependencies ───────────────────────────────
check_dependencies() {
    info "Checking dependencies..."
    MISSING_REQUIRED=()
    MISSING_OPTIONAL=()

    # Required
    if command -v curl &>/dev/null; then
        success "curl"
    elif command -v wget &>/dev/null; then
        success "wget (will use instead of curl)"
    else
        MISSING_REQUIRED+=("curl")
    fi

    if command -v tar &>/dev/null; then
        success "tar"
    else
        MISSING_REQUIRED+=("tar")
    fi

    if command -v git &>/dev/null; then
        success "git"
    else
        MISSING_REQUIRED+=("git")
    fi

    # Optional: Docker
    HAS_DOCKER=false
    if command -v docker &>/dev/null; then
        if docker compose version &>/dev/null; then
            success "Docker + Compose"
            HAS_DOCKER=true
        else
            success "Docker (Compose plugin missing)"
            MISSING_OPTIONAL+=("docker-compose-plugin")
        fi
    else
        skip "Docker (optional — needed for full stack)"
        MISSING_OPTIONAL+=("docker")
    fi

    # Optional: Build tools
    HAS_BUILD_TOOLS=false
    if command -v cargo &>/dev/null && command -v go &>/dev/null && command -v node &>/dev/null; then
        success "Rust + Go + Node.js (can build from source)"
        HAS_BUILD_TOOLS=true
    else
        skip "Build tools (optional — cargo, go, node)"
    fi

    # Optional: Node/npm for MCP
    HAS_NODE=false
    if command -v node &>/dev/null && command -v npm &>/dev/null; then
        HAS_NODE=true
    fi

    if [ ${#MISSING_REQUIRED[@]} -gt 0 ]; then
        echo ""
        warn "Missing required tools: ${MISSING_REQUIRED[*]}"
        echo ""
        echo "  Install them first:"
        case "$PLATFORM" in
            darwin)
                echo "    brew install ${MISSING_REQUIRED[*]}"
                ;;
            linux)
                echo "    sudo apt-get install -y ${MISSING_REQUIRED[*]}"
                echo "    # or: sudo yum install -y ${MISSING_REQUIRED[*]}"
                ;;
        esac
        echo ""
        error "Please install required dependencies and re-run."
    fi

    if [ ${#MISSING_OPTIONAL[@]} -gt 0 ] && [ "$HAS_DOCKER" = false ]; then
        echo ""
        warn "Docker is not installed. You can still install the CLI, but the full"
        warn "stack (API, dashboard, database) requires Docker."
        echo ""
        echo "  Install Docker:"
        case "$PLATFORM" in
            darwin)
                echo "    brew install --cask docker"
                ;;
            linux)
                echo "    curl -fsSL https://get.docker.com | sh"
                echo "    sudo usermod -aG docker \$USER && newgrp docker"
                ;;
            windows)
                echo "    Download: https://docs.docker.com/desktop/install/windows-install/"
                ;;
        esac
        echo ""
    fi
}

# ── Fetch / Download ──────────────────────────────────
fetch() {
    local url="$1"
    local dest="$2"
    if command -v curl &>/dev/null; then
        curl -fsSL "$url" -o "$dest"
    elif command -v wget &>/dev/null; then
        wget -qO "$dest" "$url"
    else
        error "No download tool available (curl or wget required)"
    fi
}

# ── Install CLI Binary ───────────────────────────────
install_cli() {
    local version="${NUCLEUS_VERSION:-latest}"
    local install_dir="${NUCLEUS_DIR:-$HOME/.nucleus}"
    local bin_dir="$install_dir/bin"

    mkdir -p "$bin_dir"

    local base_url="https://github.com/03aar/ultra-shell/releases"
    local target="${PLATFORM}-${ARCH}"
    local ext="tar.gz"
    [ "$PLATFORM" = "windows" ] && ext="zip"
    local filename="nuc-${target}.${ext}"

    if [ "$version" = "latest" ]; then
        local download_url="${base_url}/latest/download/${filename}"
    else
        local download_url="${base_url}/download/v${version}/${filename}"
    fi

    info "Downloading nuc CLI (${target})..."
    step "$download_url"

    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap 'rm -rf "$tmp_dir"' EXIT

    if fetch "$download_url" "$tmp_dir/$filename" 2>/dev/null; then
        step "Extracting..."
        if [ "$ext" = "zip" ]; then
            unzip -qo "$tmp_dir/$filename" -d "$tmp_dir"
        else
            tar xzf "$tmp_dir/$filename" -C "$tmp_dir"
        fi

        local binary_name="nuc-${target}"
        [ "$PLATFORM" = "windows" ] && binary_name="${binary_name}.exe"

        cp "$tmp_dir/$binary_name" "$bin_dir/nuc"
        chmod +x "$bin_dir/nuc"
        success "CLI installed to $bin_dir/nuc"
        INSTALL_METHOD="binary"
        return 0
    else
        warn "Could not download pre-built binary."
        return 1
    fi
}

# ── Build from Source ─────────────────────────────────
build_from_source() {
    local install_dir="${NUCLEUS_DIR:-$HOME/.nucleus}"
    local bin_dir="$install_dir/bin"
    mkdir -p "$bin_dir"

    info "Building from source..."

    local repo_dir
    if [ -f "$(pwd)/docker-compose.yml" ] && [ -d "$(pwd)/nucleus-cli" ]; then
        repo_dir="$(pwd)"
    elif [ -f "$(pwd)/nucleus/docker-compose.yml" ]; then
        repo_dir="$(pwd)/nucleus"
    else
        step "Cloning repository..."
        local tmp_clone
        tmp_clone=$(mktemp -d)
        git clone --depth 1 https://github.com/03aar/ultra-shell.git "$tmp_clone/ultra-shell"
        repo_dir="$tmp_clone/ultra-shell/nucleus"
    fi

    step "Building nuc CLI (Go)..."
    cd "$repo_dir/nucleus-cli"
    go mod tidy
    go build -ldflags "-s -w" -o "$bin_dir/nuc" .
    chmod +x "$bin_dir/nuc"
    success "CLI built at $bin_dir/nuc"

    INSTALL_METHOD="source"
    REPO_DIR="$repo_dir"
}

# ── Install MCP Server ───────────────────────────────
install_mcp() {
    if [ "${NUCLEUS_NO_MCP:-false}" = "true" ]; then
        skip "MCP server (NUCLEUS_NO_MCP=true)"
        return
    fi

    if [ "$HAS_NODE" = false ]; then
        skip "MCP server (Node.js/npm not found)"
        return
    fi

    info "Installing MCP server..."

    # Try npm global install first
    if npm install -g @nucleus/mcp-server 2>/dev/null; then
        success "MCP server installed globally via npm"
        return
    fi

    # Fall back to building from source if we have the repo
    if [ -n "${REPO_DIR:-}" ] && [ -d "${REPO_DIR}/nucleus-mcp" ]; then
        step "Building MCP server from source..."
        cd "${REPO_DIR}/nucleus-mcp"
        npm install
        npm run build
        success "MCP server built at ${REPO_DIR}/nucleus-mcp/dist/index.js"
        echo ""
        step "Register with Claude Desktop:"
        echo "    nuc mcp install --client claude-desktop"
    else
        skip "MCP server (npm package not published yet, no local source)"
    fi
}

# ── Start Docker Services ────────────────────────────
start_docker_services() {
    if [ "${NUCLEUS_NO_DOCKER:-false}" = "true" ]; then
        skip "Docker services (NUCLEUS_NO_DOCKER=true)"
        return
    fi

    if [ "$HAS_DOCKER" = false ]; then
        skip "Docker services (Docker not installed)"
        return
    fi

    local project_dir="${REPO_DIR:-}"

    if [ -z "$project_dir" ]; then
        if [ -f "$(pwd)/docker-compose.yml" ]; then
            project_dir="$(pwd)"
        elif [ -f "$(pwd)/nucleus/docker-compose.yml" ]; then
            project_dir="$(pwd)/nucleus"
        else
            skip "Docker services (no project directory found)"
            return
        fi
    fi

    echo ""
    info "Starting Docker services..."
    cd "$project_dir"

    # Create .env if it doesn't exist
    if [ ! -f ".env" ]; then
        if [ -f ".env.example" ]; then
            cp .env.example .env
            step "Created .env from .env.example"
        else
            cat > .env <<'ENVEOF'
NUCLEUS_REDIS_URL=redis://redis:6379
DATABASE_URL=postgres://nucleus:nucleus_dev_password@postgres:5432/nucleus?sslmode=disable
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080/api/v1
NEXT_PUBLIC_API_KEY=dev-nucleus-key-local
ENVEOF
            step "Created default .env"
        fi
    fi

    docker compose build --parallel 2>&1 | tail -3
    docker compose up -d

    # Wait for API
    step "Waiting for API to be ready..."
    local retries=0
    until curl -sf http://localhost:8080/health > /dev/null 2>&1; do
        retries=$((retries + 1))
        if [ $retries -ge 30 ]; then
            warn "API did not become healthy in time. Check: docker compose logs nucleus-api"
            return
        fi
        sleep 2
    done
    success "API is healthy"

    DOCKER_RUNNING=true
}

# ── Configure Shell Integration ──────────────────────
configure_shell() {
    local install_dir="${NUCLEUS_DIR:-$HOME/.nucleus}"
    local bin_dir="$install_dir/bin"

    # Check if already in PATH
    if echo "$PATH" | tr ':' '\n' | grep -q "^${bin_dir}$"; then
        success "Already in PATH"
        return
    fi

    local shell_rc=""
    local current_shell="${SHELL:-/bin/bash}"
    case "$current_shell" in
        */zsh)  shell_rc="$HOME/.zshrc"  ;;
        */bash) shell_rc="$HOME/.bashrc" ;;
        */fish) shell_rc="$HOME/.config/fish/config.fish" ;;
        *)      shell_rc="$HOME/.bashrc" ;;
    esac

    local path_line="export PATH=\"${bin_dir}:\$PATH\""

    if [ -n "$shell_rc" ] && [ -f "$shell_rc" ]; then
        if ! grep -q "nucleus" "$shell_rc" 2>/dev/null; then
            echo "" >> "$shell_rc"
            echo "# Nucleus CLI" >> "$shell_rc"
            echo "$path_line" >> "$shell_rc"
            success "Added to PATH in $shell_rc"
        else
            success "Already configured in $shell_rc"
        fi
    elif [ -n "$shell_rc" ]; then
        echo "# Nucleus CLI" > "$shell_rc"
        echo "$path_line" >> "$shell_rc"
        success "Created $shell_rc with PATH"
    fi

    # Export for current session
    export PATH="${bin_dir}:$PATH"
}

# ── Print Success ────────────────────────────────────
print_success() {
    local install_dir="${NUCLEUS_DIR:-$HOME/.nucleus}"

    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    echo -e "${GREEN}${BOLD}  NUCLEUS installed successfully.${RESET}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    echo ""
    echo -e "  ${BOLD}Get started:${RESET}"
    echo ""
    echo -e "    ${CYAN}nuc doctor${RESET}          Check installation health"
    echo -e "    ${CYAN}nuc shell${RESET}           Start a Nucleus shell session"
    echo -e "    ${CYAN}nuc exec \"ls -la\"${RESET}   Execute a command with risk evaluation"
    echo -e "    ${CYAN}nuc plan \"...\"${RESET}      Plan a multi-step workflow"
    echo ""

    if [ "${DOCKER_RUNNING:-false}" = true ]; then
        echo -e "  ${BOLD}Services running:${RESET}"
        echo ""
        echo -e "    Dashboard:  ${CYAN}http://localhost:3000${RESET}"
        echo -e "    API:        ${CYAN}http://localhost:8080${RESET}"
        echo -e "    API Key:    ${DIM}dev-nucleus-key-local${RESET}"
        echo ""
    fi

    echo -e "  ${BOLD}Connect to Claude Desktop:${RESET}"
    echo ""
    echo -e "    ${CYAN}nuc mcp install --client claude-desktop${RESET}"
    echo ""
    echo -e "  ${BOLD}Links:${RESET}"
    echo ""
    echo -e "    Docs:     https://nucleusshell.dev/docs"
    echo -e "    GitHub:   https://github.com/03aar/ultra-shell"
    echo -e "    Discord:  https://discord.gg/nucleus"
    echo ""

    if ! echo "$PATH" | tr ':' '\n' | grep -q "${install_dir}/bin"; then
        echo -e "  ${YELLOW}Restart your shell or run:${RESET}"
        echo -e "    export PATH=\"${install_dir}/bin:\$PATH\""
        echo ""
    fi
}

# ── Main ──────────────────────────────────────────────
main() {
    banner
    detect_platform
    check_dependencies
    echo ""

    INSTALL_METHOD=""
    REPO_DIR=""
    DOCKER_RUNNING=false

    # Strategy 1: Download pre-built binary
    # Strategy 2: Build from source
    if ! install_cli; then
        if [ "$HAS_BUILD_TOOLS" = true ]; then
            build_from_source
        else
            error "Could not install CLI. Install Go (https://go.dev) to build from source, or check https://github.com/03aar/ultra-shell/releases"
        fi
    fi

    configure_shell
    install_mcp
    start_docker_services
    print_success
}

main "$@"
