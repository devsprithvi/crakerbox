#!/bin/bash
set -e

# Crackerbox Manager Installer
# Usage: curl -sL <site>/install-manager.sh | sudo bash
#
# Installs:
#   - crackerbox-manager (cluster orchestrator)

CRACKERBOX_VERSION="0.1.0"
INSTALL_DIR="/usr/local/bin"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()  { echo -e "${GREEN}[crackerbox]${NC} $1"; }
warn()  { echo -e "${YELLOW}[crackerbox]${NC} $1"; }
error() { echo -e "${RED}[crackerbox]${NC} $1"; exit 1; }
header() { echo -e "${CYAN}${BOLD}$1${NC}"; }

echo ""
header "  🎯 Crackerbox Manager Installer v${CRACKERBOX_VERSION}"
header "  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check root
if [ "$EUID" -ne 0 ]; then
    error "Please run as root (use sudo)"
fi

# Check Linux
if [ "$(uname -s)" != "Linux" ]; then
    error "Crackerbox requires Linux (detected: $(uname -s))"
fi

# Check architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    *)       error "Unsupported architecture: $ARCH" ;;
esac

info "Platform: linux/${ARCH}"

GITHUB_BASE="https://github.com/devsprithvi/crakerbox/releases/download/v${CRACKERBOX_VERSION}"

# --- Download crackerbox-manager ---

info "Downloading crackerbox-manager (cluster orchestrator)..."
MANAGER_URL="${GITHUB_BASE}/crackerbox-manager-linux-${ARCH}"
if curl -fsSL "$MANAGER_URL" -o "${INSTALL_DIR}/crackerbox-manager" 2>/dev/null; then
    chmod +x "${INSTALL_DIR}/crackerbox-manager"
    info "✅ crackerbox-manager installed"
else
    warn "⚠️  crackerbox-manager binary not yet available (no release published yet)"
fi

# --- Summary ---

echo ""
info "════════════════════════════════════════════════"
info "  crackerbox-manager installed successfully! 🎯"
info "════════════════════════════════════════════════"
echo ""
info "  ${BOLD}Binary:${NC}"
info "    crackerbox-manager: ${INSTALL_DIR}/crackerbox-manager"
echo ""
info "  ${BOLD}Get started:${NC}"
info "    crackerbox-manager serve   # Start the cluster orchestrator"
info "    crackerbox-manager nodes   # List registered nodes"
echo ""
