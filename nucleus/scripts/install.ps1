#
# NUCLEUS Installer — Windows (PowerShell)
#
# Usage:
#   .\scripts\install.ps1
#
# Requires: Docker Desktop for Windows
#

$ErrorActionPreference = "Stop"

function Write-Info($msg) { Write-Host "[NUCLEUS] $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "[NUCLEUS] $msg" -ForegroundColor Yellow }
function Write-Err($msg)  { Write-Host "[NUCLEUS] $msg" -ForegroundColor Red; exit 1 }

# ── Detect Architecture ──────────────────────────────
$arch = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else { "x86" }
Write-Info "Detected platform: windows/$arch"

# ── Check prerequisites ──────────────────────────────
$missing = @()

try {
    docker --version | Out-Null
} catch {
    $missing += "docker"
}

try {
    docker compose version | Out-Null
} catch {
    $missing += "docker-compose"
}

if ($missing.Count -gt 0) {
    Write-Warn "Missing required tools: $($missing -join ', ')"
    Write-Host ""
    Write-Host "Install Docker Desktop from:"
    Write-Host "  https://docs.docker.com/desktop/install/windows-install/"
    Write-Host ""
    Write-Host "After installation, ensure Docker Desktop is running."
    Write-Err "Please install the missing tools and re-run this script."
}

Write-Info "All prerequisites met."

# ── Find project directory ───────────────────────────
$nucleusDir = $PWD.Path
if (-not (Test-Path "$nucleusDir\docker-compose.yml")) {
    if (Test-Path "$nucleusDir\nucleus\docker-compose.yml") {
        $nucleusDir = "$nucleusDir\nucleus"
    } else {
        Write-Err "Could not find nucleus project. Run this script from the nucleus\ directory."
    }
}

Write-Info "Using project directory: $nucleusDir"

# ── Create .env file ────────────────────────────────
$envFile = "$nucleusDir\.env"
if (-not (Test-Path $envFile)) {
    @"
# NUCLEUS Environment Configuration
NUCLEUS_REDIS_URL=redis://redis:6379
DATABASE_URL=postgres://nucleus:nucleus_dev_password@postgres:5432/nucleus?sslmode=disable
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080/api/v1
NEXT_PUBLIC_API_KEY=dev-nucleus-key-local
"@ | Out-File -FilePath $envFile -Encoding UTF8

    Write-Info "Created .env file with default configuration."
}

# ── Build and start ──────────────────────────────────
Write-Info "Building and starting NUCLEUS..."
Set-Location $nucleusDir

docker compose build --parallel
docker compose up -d

Write-Host ""
Write-Info "NUCLEUS is running!"
Write-Host ""
Write-Host "  Dashboard:  http://localhost:3000"
Write-Host "  API:        http://localhost:8080"
Write-Host "  API Health: http://localhost:8080/health"
Write-Host ""
Write-Host "  API Key:    dev-nucleus-key-local"
Write-Host ""
Write-Host "  Manage:  docker compose logs -f   (watch logs)"
Write-Host "           docker compose down       (stop)"
Write-Host "           docker compose restart    (restart)"
Write-Host ""

# ── Optional: Check for local build tools ────────────
$hasRust = $null -ne (Get-Command cargo -ErrorAction SilentlyContinue)
$hasGo = $null -ne (Get-Command go -ErrorAction SilentlyContinue)
$hasNode = $null -ne (Get-Command node -ErrorAction SilentlyContinue)

if ($hasRust -and $hasGo -and $hasNode) {
    Write-Host ""
    Write-Host "To install the nucleus CLI locally:"
    Write-Host "  cd nucleus-core"
    Write-Host "  cargo build --release"
    Write-Host "  copy target\release\nucleus.exe C:\Windows\System32\nucleus.exe"
    Write-Host ""
}

Write-Info "Installation complete."
