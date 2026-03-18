#!/bin/bash
set -euo pipefail

# ─────────────────────────────────────────────────────────────────────────────
#  Crackerbox Manager Installer
# ─────────────────────────────────────────────────────────────────────────────
#  Usage:
#    curl -sL https://<site>/install-manager.sh | sudo bash
#    curl -sL https://<site>/install-manager.sh | sudo bash -s -- --version 0.2.0
#
#  Installs:
#    crackerbox-manager — Cluster orchestrator (manages multiple crackerboxd nodes)
#
#  Requires: Linux, root, curl
# ─────────────────────────────────────────────────────────────────────────────

# ── Defaults ─────────────────────────────────────────────────────────────────
CRACKERBOX_VERSION="${CRACKERBOX_VERSION:-latest}"
CRACKERBOX_REPO="devsprithvi/crakerbox"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/crackerbox"

# ── Parse CLI flags ──────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)     CRACKERBOX_VERSION="$2"; shift 2 ;;
    --install-dir) INSTALL_DIR="$2";        shift 2 ;;
    --help|-h)
      echo "Usage: install-manager.sh [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  --version VERSION     Crackerbox version to install (default: latest)"
      echo "  --install-dir DIR     Binary install directory (default: /usr/local/bin)"
      echo "  -h, --help            Show this help"
      exit 0
      ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

# ── Colors & helpers ─────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

info()    { echo -e "${GREEN}[crackerbox]${NC} $1"; }
warn()    { echo -e "${YELLOW}[crackerbox]${NC} ⚠️  $1"; }
error()   { echo -e "${RED}[crackerbox]${NC} ✖  $1"; exit 1; }
success() { echo -e "${GREEN}[crackerbox]${NC} ✅ $1"; }
header()  { echo -e "${CYAN}${BOLD}$1${NC}"; }
step()    { echo -e "${DIM}[crackerbox]${NC} → $1"; }

# ── Dependency check ─────────────────────────────────────────────────────────
require_cmd() {
  if ! command -v "$1" &>/dev/null; then
    error "Required command '$1' not found. Please install it first."
  fi
}

# ── Banner ───────────────────────────────────────────────────────────────────
echo ""
header "  ┌──────────────────────────────────────────────┐"
header "  │  🎯  Crackerbox Manager Installer             │"
header "  │      Cluster Orchestrator                     │"
header "  └──────────────────────────────────────────────┘"
echo ""

# ── Pre-flight checks ───────────────────────────────────────────────────────
step "Running pre-flight checks..."

# Root check
if [ "$(id -u)" -ne 0 ]; then
  error "This installer must be run as root. Use: curl ... | sudo bash"
fi

# OS check
if [ "$(uname -s)" != "Linux" ]; then
  error "Crackerbox requires Linux (detected: $(uname -s))"
fi

# Required commands
require_cmd curl

# Architecture detection
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64"  ;;
  *)       error "Unsupported architecture: $ARCH" ;;
esac

info "Platform: linux/${ARCH}"

# ── Resolve version ─────────────────────────────────────────────────────────
GITHUB_API="https://api.github.com/repos/${CRACKERBOX_REPO}/releases"

if [ "$CRACKERBOX_VERSION" = "latest" ]; then
  step "Resolving latest release..."
  CRACKERBOX_VERSION=$(curl -fsSL "${GITHUB_API}/latest" 2>/dev/null | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//' || true)

  # Fallback to dev-latest pre-release if no stable release exists
  if [ -z "$CRACKERBOX_VERSION" ]; then
    step "No stable release found, trying dev-latest..."
    CRACKERBOX_VERSION=$(curl -fsSL "${GITHUB_API}/tags/dev-latest" 2>/dev/null | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//' || true)
  fi

  if [ -z "$CRACKERBOX_VERSION" ]; then
    error "Could not resolve version. No releases found at https://github.com/${CRACKERBOX_REPO}/releases"
  fi
fi

# Normalise version tag
DISPLAY_VERSION="${CRACKERBOX_VERSION#v}"
TAG_VERSION="${CRACKERBOX_VERSION}"
if [[ "$TAG_VERSION" =~ ^[0-9] ]]; then
  TAG_VERSION="v${DISPLAY_VERSION}"
fi

info "Version: ${DISPLAY_VERSION} (tag: ${TAG_VERSION})"

GITHUB_DOWNLOAD="https://github.com/${CRACKERBOX_REPO}/releases/download/${TAG_VERSION}"

# ── Download crackerbox-manager ──────────────────────────────────────────────
echo ""
step "Downloading crackerbox-manager..."
MANAGER_URL="${GITHUB_DOWNLOAD}/crackerbox-manager-linux-${ARCH}"
MANAGER_BIN="${INSTALL_DIR}/crackerbox-manager"

HTTP_CODE=$(curl -fsSL -w "%{http_code}" -o "${MANAGER_BIN}.tmp" "$MANAGER_URL" 2>/dev/null || true)

if [ "$HTTP_CODE" = "200" ] && [ -s "${MANAGER_BIN}.tmp" ]; then
  mv "${MANAGER_BIN}.tmp" "$MANAGER_BIN"
  chmod +x "$MANAGER_BIN"
  success "crackerbox-manager installed → ${MANAGER_BIN}"
else
  rm -f "${MANAGER_BIN}.tmp"
  error "Failed to download crackerbox-manager from: ${MANAGER_URL}
       HTTP status: ${HTTP_CODE:-unknown}
       
       Possible causes:
         • No release has been published yet for tag '${TAG_VERSION}'
         • The repository '${CRACKERBOX_REPO}' doesn't have linux/${ARCH} binaries
         • Network connectivity issue
       
       Check releases: https://github.com/${CRACKERBOX_REPO}/releases"
fi

# ── Create config directory ──────────────────────────────────────────────────
echo ""
step "Creating config directory..."
mkdir -p "${CONFIG_DIR}"

# Write default config if it doesn't already exist
if [ ! -f "${CONFIG_DIR}/manager.yaml" ]; then
  cat > "${CONFIG_DIR}/manager.yaml" <<EOF
# Crackerbox Manager Configuration
# Generated by installer on $(date -u +%Y-%m-%dT%H:%M:%SZ)

listen: ":9090"
log_level: "info"

# Registered crackerboxd nodes
nodes: []
EOF
  success "Default config created → ${CONFIG_DIR}/manager.yaml"
fi

success "Directory structure ready"

# ── Verify installation ─────────────────────────────────────────────────────
echo ""
step "Verifying installation..."

if [ -x "$MANAGER_BIN" ]; then
  SIZE=$(du -h "$MANAGER_BIN" | cut -f1)
  success "crackerbox-manager (${SIZE})"
else
  warn "crackerbox-manager not found at ${MANAGER_BIN}"
fi

# ── Summary ──────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}${BOLD}  ┌──────────────────────────────────────────────┐${NC}"
echo -e "${GREEN}${BOLD}  │  ✅  crackerbox-manager installed!           │${NC}"
echo -e "${GREEN}${BOLD}  └──────────────────────────────────────────────┘${NC}"
echo ""
info "  ${BOLD}Binary:${NC}"
info "    crackerbox-manager   ${MANAGER_BIN}"
info "    config               ${CONFIG_DIR}/manager.yaml"
echo ""
info "  ${BOLD}Quick start:${NC}"
info "    crackerbox-manager serve          Start the cluster orchestrator"
info "    crackerbox-manager status         Show cluster status"
info "    crackerbox-manager nodes          List registered nodes"
info "    crackerbox-manager version        Print version info"
info "    crackerbox-manager update         Self-update to latest release"
info "    crackerbox-manager help           Show all available commands"
echo ""
info "  ${BOLD}Update later:${NC}"
info "    crackerbox-manager update                    Update to latest"
info "    crackerbox-manager update --version 0.2.0    Update to specific version"
echo ""
