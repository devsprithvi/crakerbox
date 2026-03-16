#!/bin/bash
set -e

# Matchbox Installer
# Usage: curl -sL https://get.matchbox.dev/install.sh | sudo bash

MATCHBOX_VERSION="0.1.0"
INSTALL_DIR="/usr/local/bin"
DATA_DIR="/var/lib/matchbox"
FIRECRACKER_VERSION="1.10.1"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[matchbox]${NC} $1"; }
warn()  { echo -e "${YELLOW}[matchbox]${NC} $1"; }
error() { echo -e "${RED}[matchbox]${NC} $1"; exit 1; }

# --- Pre-flight checks ---

info "Matchbox Installer v${MATCHBOX_VERSION}"
echo ""

# Check root
if [ "$EUID" -ne 0 ]; then
    error "Please run as root (use sudo)"
fi

# Check Linux
if [ "$(uname -s)" != "Linux" ]; then
    error "Matchbox requires Linux (detected: $(uname -s))"
fi

# Check architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    *)       error "Unsupported architecture: $ARCH" ;;
esac

# Check KVM
info "Checking hardware virtualization..."
if [ ! -e /dev/kvm ]; then
    error "KVM not available. Ensure your server supports hardware virtualization and /dev/kvm exists."
fi
info "✅ KVM available"

# --- Download matchboxd ---

info "Downloading matchboxd binary..."
MATCHBOX_URL="https://github.com/matchbox-platform/matchbox/releases/download/v${MATCHBOX_VERSION}/matchboxd-linux-${ARCH}"
# TODO: Replace with actual download URL when releases are published
# curl -sL "$MATCHBOX_URL" -o "${INSTALL_DIR}/matchboxd"
# chmod +x "${INSTALL_DIR}/matchboxd"
warn "⚠️  Binary download not yet available (placeholder)"

# --- Download Firecracker ---

info "Downloading Firecracker v${FIRECRACKER_VERSION}..."
FC_URL="https://github.com/firecracker-microvm/firecracker/releases/download/v${FIRECRACKER_VERSION}/firecracker-v${FIRECRACKER_VERSION}-${ARCH}.tgz"
# TODO: Uncomment when ready
# curl -sL "$FC_URL" | tar xz -C /tmp
# mv /tmp/release-v${FIRECRACKER_VERSION}-${ARCH}/firecracker-v${FIRECRACKER_VERSION}-${ARCH} "${INSTALL_DIR}/firecracker"
# chmod +x "${INSTALL_DIR}/firecracker"
warn "⚠️  Firecracker download not yet available (placeholder)"

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
info "════════════════════════════════════════"
info "  Matchbox installed successfully! 🔥"
info "════════════════════════════════════════"
echo ""
info "  Binary:      ${INSTALL_DIR}/matchboxd"
info "  Firecracker: ${INSTALL_DIR}/firecracker"
info "  Data dir:    ${DATA_DIR}"
echo ""
info "  Get started:"
info "    matchboxd serve    # Start the daemon"
info "    matchboxd status   # Check system status"
info "    matchboxd ui       # Open the dashboard"
echo ""
