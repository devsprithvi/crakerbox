#!/bin/bash
set -e

# Crackerbox Daemon Installer
# Usage: curl -sL <site>/install-daemon.sh | sudo bash
#
# Installs:
#   - crackerboxd        (node daemon)
#   - firecracker        (VMM binary)

CRACKERBOX_VERSION="0.1.0"
INSTALL_DIR="/usr/local/bin"
DATA_DIR="/var/lib/crackerbox"
FIRECRACKER_VERSION="1.10.1"

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
header "  🔥 Crackerbox Daemon Installer v${CRACKERBOX_VERSION}"
header "  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
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

# Check KVM
info "Checking hardware virtualization..."
if [ ! -e /dev/kvm ]; then
    error "KVM not available. Ensure your server supports hardware virtualization and /dev/kvm exists."
fi
info "✅ KVM available"

GITHUB_BASE="https://github.com/devsprithvi/crakerbox/releases/download/v${CRACKERBOX_VERSION}"

# --- Download crackerboxd ---

info "Downloading crackerboxd (node daemon)..."
DAEMON_URL="${GITHUB_BASE}/crackerboxd-linux-${ARCH}"
if curl -fsSL "$DAEMON_URL" -o "${INSTALL_DIR}/crackerboxd" 2>/dev/null; then
    chmod +x "${INSTALL_DIR}/crackerboxd"
    info "✅ crackerboxd installed"
else
    warn "⚠️  crackerboxd binary not yet available (no release published yet)"
fi

# --- Download Firecracker ---

info "Downloading Firecracker v${FIRECRACKER_VERSION}..."
FC_URL="https://github.com/firecracker-microvm/firecracker/releases/download/v${FIRECRACKER_VERSION}/firecracker-v${FIRECRACKER_VERSION}-${ARCH}.tgz"
if curl -fsSL "$FC_URL" | tar xz -C /tmp 2>/dev/null; then
    mv "/tmp/release-v${FIRECRACKER_VERSION}-${ARCH}/firecracker-v${FIRECRACKER_VERSION}-${ARCH}" "${INSTALL_DIR}/firecracker"
    chmod +x "${INSTALL_DIR}/firecracker"
    info "✅ Firecracker installed"
else
    warn "⚠️  Firecracker download failed (check network)"
fi

# --- Create directory structure ---

info "Creating directory structure..."
mkdir -p "${DATA_DIR}"
mkdir -p "${DATA_DIR}/vms"
mkdir -p "${DATA_DIR}/images"
mkdir -p "${DATA_DIR}/kernels"
mkdir -p "${DATA_DIR}/config"
mkdir -p "${DATA_DIR}/logs"
info "✅ Created ${DATA_DIR}"

# --- Summary ---

echo ""
info "════════════════════════════════════════════"
info "  crackerboxd installed successfully! 🔥"
info "════════════════════════════════════════════"
echo ""
info "  ${BOLD}Binaries:${NC}"
info "    crackerboxd:  ${INSTALL_DIR}/crackerboxd"
info "    firecracker:  ${INSTALL_DIR}/firecracker"
info "    data dir:     ${DATA_DIR}"
echo ""
info "  ${BOLD}Get started:${NC}"
info "    crackerboxd serve    # Start the node daemon"
info "    crackerboxd status   # Check node status"
echo ""
