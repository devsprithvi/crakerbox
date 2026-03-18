#!/bin/bash
set -euo pipefail

# ─────────────────────────────────────────────────────────────────────────────
#  Crackerbox Full Installer
# ─────────────────────────────────────────────────────────────────────────────
#  Usage:
#    curl -sL https://<site>/install.sh | sudo bash
#    curl -sL https://<site>/install.sh | sudo bash -s -- --version 0.2.0
#
#  Installs:
#    crackerboxd         — Firecracker node daemon (manages microVMs)
#    crackerbox-manager  — Cluster orchestrator (manages multiple nodes)
#    firecracker         — VMM binary from upstream
#
#  Requires: Linux, root, KVM, curl, tar, sha256sum
# ─────────────────────────────────────────────────────────────────────────────

# ── Defaults ─────────────────────────────────────────────────────────────────
CRACKERBOX_VERSION="${CRACKERBOX_VERSION:-latest}"
CRACKERBOX_REPO="devsprithvi/crakerbox"
INSTALL_DIR="/usr/local/bin"
DATA_DIR="/var/lib/crackerbox"
CONFIG_DIR="/etc/crackerbox"
FIRECRACKER_VERSION="1.10.1"
SKIP_FIRECRACKER=false
COMPONENT="all"  # all | daemon | manager

# ── Parse CLI flags ──────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)          CRACKERBOX_VERSION="$2"; shift 2 ;;
    --install-dir)      INSTALL_DIR="$2";        shift 2 ;;
    --skip-firecracker) SKIP_FIRECRACKER=true;   shift   ;;
    --component)        COMPONENT="$2";          shift 2 ;;
    --help|-h)
      echo "Usage: install.sh [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  --version VERSION       Crackerbox version to install (default: latest)"
      echo "  --install-dir DIR       Binary install directory (default: /usr/local/bin)"
      echo "  --component COMPONENT   Install only: all, daemon, manager (default: all)"
      echo "  --skip-firecracker      Skip Firecracker VMM download"
      echo "  -h, --help              Show this help"
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
header "  │  🔥  Crackerbox Installer                     │"
header "  │      Daemon + Manager + Firecracker VMM       │"
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
require_cmd tar

# Architecture detection  (keep raw arch for Firecracker URLs)
RAW_ARCH=$(uname -m)
case "$RAW_ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64"  ;;
  *)       error "Unsupported architecture: $RAW_ARCH" ;;
esac

info "Platform: linux/${ARCH} (${RAW_ARCH})"

# KVM check (only required for daemon)
if [ "$COMPONENT" = "all" ] || [ "$COMPONENT" = "daemon" ]; then
  step "Checking hardware virtualisation..."
  if [ ! -e /dev/kvm ]; then
    error "KVM not available. Enable hardware virtualisation (VT-x/AMD-V) and ensure /dev/kvm exists."
  fi
  success "KVM available"
fi

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

# ── Helper: download a binary ───────────────────────────────────────────────
download_binary() {
  local name="$1"
  local url="$2"
  local dest="$3"

  step "Downloading ${name}..."
  HTTP_CODE=$(curl -fsSL -w "%{http_code}" -o "${dest}.tmp" "$url" 2>/dev/null || true)

  if [ "$HTTP_CODE" = "200" ] && [ -s "${dest}.tmp" ]; then
    mv "${dest}.tmp" "$dest"
    chmod +x "$dest"
    success "${name} installed → ${dest}"
    return 0
  else
    rm -f "${dest}.tmp"
    error "Failed to download ${name} from: ${url}
         HTTP status: ${HTTP_CODE:-unknown}
         
         Possible causes:
           • No release published for tag '${TAG_VERSION}'
           • No linux/${ARCH} binary in the release
           • Network connectivity issue
         
         Check releases: https://github.com/${CRACKERBOX_REPO}/releases"
    return 1
  fi
}

# ── Download crackerboxd ─────────────────────────────────────────────────────
if [ "$COMPONENT" = "all" ] || [ "$COMPONENT" = "daemon" ]; then
  echo ""
  download_binary "crackerboxd" \
    "${GITHUB_DOWNLOAD}/crackerboxd-linux-${ARCH}" \
    "${INSTALL_DIR}/crackerboxd"
fi

# ── Download crackerbox-manager ──────────────────────────────────────────────
if [ "$COMPONENT" = "all" ] || [ "$COMPONENT" = "manager" ]; then
  echo ""
  download_binary "crackerbox-manager" \
    "${GITHUB_DOWNLOAD}/crackerbox-manager-linux-${ARCH}" \
    "${INSTALL_DIR}/crackerbox-manager"
fi

# ── Download Firecracker ─────────────────────────────────────────────────────
if [ "$SKIP_FIRECRACKER" = false ] && { [ "$COMPONENT" = "all" ] || [ "$COMPONENT" = "daemon" ]; }; then
  echo ""
  step "Downloading Firecracker v${FIRECRACKER_VERSION}..."

  # Firecracker releases use raw arch names (x86_64, aarch64), NOT amd64/arm64
  FC_URL="https://github.com/firecracker-microvm/firecracker/releases/download/v${FIRECRACKER_VERSION}/firecracker-v${FIRECRACKER_VERSION}-${RAW_ARCH}.tgz"
  FC_BIN="${INSTALL_DIR}/firecracker"

  FC_TMP=$(mktemp -d)
  trap "rm -rf ${FC_TMP}" EXIT

  if curl -fsSL "$FC_URL" -o "${FC_TMP}/firecracker.tgz"; then
    tar xzf "${FC_TMP}/firecracker.tgz" -C "${FC_TMP}"

    # Locate the firecracker binary inside the extracted archive
    FC_EXTRACTED=$(find "${FC_TMP}" -name "firecracker-v*" -type f ! -name "*.yaml" ! -name "*jailer*" | head -1)
    if [ -n "$FC_EXTRACTED" ]; then
      mv "$FC_EXTRACTED" "$FC_BIN"
      chmod +x "$FC_BIN"
      success "Firecracker v${FIRECRACKER_VERSION} installed → ${FC_BIN}"
    else
      warn "Firecracker archive extracted but binary not found in expected path"
    fi

    # Also install jailer if available
    JAILER_EXTRACTED=$(find "${FC_TMP}" -name "jailer-v*" -type f | head -1)
    if [ -n "$JAILER_EXTRACTED" ]; then
      mv "$JAILER_EXTRACTED" "${INSTALL_DIR}/jailer"
      chmod +x "${INSTALL_DIR}/jailer"
      success "jailer v${FIRECRACKER_VERSION} installed → ${INSTALL_DIR}/jailer"
    fi
  else
    error "Failed to download Firecracker from: ${FC_URL}
         
         This is an upstream dependency. Verify the version exists:
         https://github.com/firecracker-microvm/firecracker/releases/tag/v${FIRECRACKER_VERSION}"
  fi
fi

# ── Create directory structure ───────────────────────────────────────────────
echo ""
step "Creating directory structure..."
mkdir -p "${CONFIG_DIR}"

if [ "$COMPONENT" = "all" ] || [ "$COMPONENT" = "daemon" ]; then
  mkdir -p "${DATA_DIR}"/{vms,images,kernels,logs}

  if [ ! -f "${CONFIG_DIR}/crackerboxd.yaml" ]; then
    cat > "${CONFIG_DIR}/crackerboxd.yaml" <<EOF
# Crackerbox Daemon Configuration
# Generated by installer on $(date -u +%Y-%m-%dT%H:%M:%SZ)

listen: ":8090"
data_dir: "${DATA_DIR}"
log_level: "info"
EOF
    success "Daemon config created → ${CONFIG_DIR}/crackerboxd.yaml"
  fi
fi

if [ "$COMPONENT" = "all" ] || [ "$COMPONENT" = "manager" ]; then
  if [ ! -f "${CONFIG_DIR}/manager.yaml" ]; then
    cat > "${CONFIG_DIR}/manager.yaml" <<EOF
# Crackerbox Manager Configuration
# Generated by installer on $(date -u +%Y-%m-%dT%H:%M:%SZ)

listen: ":9090"
log_level: "info"

# Registered crackerboxd nodes
nodes: []
EOF
    success "Manager config created → ${CONFIG_DIR}/manager.yaml"
  fi
fi

success "Directory structure ready"

# ── Verify installation ─────────────────────────────────────────────────────
echo ""
step "Verifying installation..."

for bin in crackerboxd crackerbox-manager firecracker jailer; do
  BIN_PATH="${INSTALL_DIR}/${bin}"
  if [ -x "$BIN_PATH" ]; then
    SIZE=$(du -h "$BIN_PATH" | cut -f1)
    success "${bin} (${SIZE})"
  fi
done

# ── Summary ──────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}${BOLD}  ┌──────────────────────────────────────────────┐${NC}"
echo -e "${GREEN}${BOLD}  │  ✅  Crackerbox installed successfully!      │${NC}"
echo -e "${GREEN}${BOLD}  └──────────────────────────────────────────────┘${NC}"
echo ""
info "  ${BOLD}Binaries:${NC}"
[ -x "${INSTALL_DIR}/crackerboxd" ]        && info "    crackerboxd          ${INSTALL_DIR}/crackerboxd"
[ -x "${INSTALL_DIR}/crackerbox-manager" ] && info "    crackerbox-manager   ${INSTALL_DIR}/crackerbox-manager"
[ -x "${INSTALL_DIR}/firecracker" ]        && info "    firecracker          ${INSTALL_DIR}/firecracker"
[ -x "${INSTALL_DIR}/jailer" ]             && info "    jailer               ${INSTALL_DIR}/jailer"
info "    config dir           ${CONFIG_DIR}/"
info "    data dir             ${DATA_DIR}/"
echo ""
info "  ${BOLD}Quick start:${NC}"
if [ "$COMPONENT" = "all" ] || [ "$COMPONENT" = "daemon" ]; then
  info "    crackerboxd serve              Start the node daemon"
  info "    crackerboxd status             Show daemon & system status"
  info "    crackerboxd version            Print version info"
  info "    crackerboxd update             Self-update to latest release"
fi
if [ "$COMPONENT" = "all" ] || [ "$COMPONENT" = "manager" ]; then
  info "    crackerbox-manager serve       Start the cluster orchestrator"
  info "    crackerbox-manager nodes       List registered nodes"
  info "    crackerbox-manager version     Print version info"
  info "    crackerbox-manager update      Self-update to latest release"
fi
echo ""
