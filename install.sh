#!/bin/bash

set -e

# --- Configuration ---
GITHUB_REPO="zarigata/budgie"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/budgie"
DATA_DIR="/var/lib/budgie"
LOG_DIR="/var/log/budgie"

BINARY_NAME="budgie"
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# --- Colors ---
C_RESET='\033[0m'
C_RED='\033[0;31m'
C_GREEN='\033[0;32m'
C_BLUE='\033[0;34m'
C_BOLD='\033[1m'

# --- Helper Functions ---
info() {
    echo -e "${C_BLUE}INFO: $1${C_RESET}"
}

success() {
    echo -e "${C_GREEN}✅ SUCCESS: $1${C_RESET}"
}

error() {
    echo -e "${C_RED}❌ ERROR: $1${C_RESET}"
    exit 1
}

# --- Main Functions ---

check_root() {
    if [ "$EUID" -ne 0 ]; then
        error "This script must be run as root. Use sudo."
    fi
}

detect_platform() {
    local uname_s
    uname_s=$(uname -s)
    case "$uname_s" in
        Linux)
            PLATFORM="linux"
            ;;
        Darwin)
            PLATFORM="darwin"
            ;;
        *)
            error "Unsupported platform: $uname_s"
            ;;
    esac
}

detect_arch() {
    local uname_m
    uname_m=$(uname -m)
    case "$uname_m" in
        x86_64)
            ARCH="amd64"
            ;;
        arm64 | aarch64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $uname_m"
            ;;
    esac
}

download_binary() {
    local binary_url="https://github.com/$GITHUB_REPO/releases/download/$LATEST_RELEASE/budgie-$PLATFORM-$ARCH"
    info "Downloading Budgie from $binary_url..."
    curl -L -o "$INSTALL_DIR/$BINARY_NAME" "$binary_url"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
}

create_directories() {
    info "Creating directories..."
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$DATA_DIR/containers"
    mkdir -p "$LOG_DIR"
}

create_config_file() {
    info "Creating configuration file..."
    cat > "$CONFIG_DIR/config.yaml" << EOF
# Budgie Configuration
data_dir: "$DATA_DIR"

runtime:
  address: "/run/containerd/containerd.sock"

discovery:
  enabled: true
  port: 5353

proxy:
  type: "round-robin"
  health_check_interval: 30s

logging:
  level: "info"
  file: "$LOG_DIR/budgie.log"
EOF
}

check_containerd() {
    info "Checking for containerd..."
    if ! command -v containerd &> /dev/null; then
        error "containerd is not installed. Please install it before running Budgie."
    fi
    if ! systemctl is-active --quiet containerd; then
        error "containerd is not running. Please start it before running Budgie."
    fi
}

install() {
    check_root
    detect_platform
    detect_arch
    check_containerd

    info "Installing Budgie $LATEST_RELEASE for $PLATFORM-$ARCH..."
    
    download_binary
    create_directories
    create_config_file
    
    success "Budgie has been installed successfully!"
    info "Run 'budgie --help' to get started."
}

uninstall() {
    check_root
    info "Uninstalling Budgie..."
    rm -f "$INSTALL_DIR/$BINARY_NAME"
    rm -rf "$CONFIG_DIR"
    rm -rf "$DATA_DIR"
    rm -rf "$LOG_DIR"
    success "Budgie has been uninstalled successfully."
}

# --- Script Entrypoint ---

case "$1" in
    uninstall)
        uninstall
        ;;
    *)
        install
        ;;
esac