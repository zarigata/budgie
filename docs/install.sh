#!/bin/bash
# Budgie Installer for Linux and macOS
# Usage: curl -fsSL https://zarigata.github.io/budgie/install.sh | sudo bash
#
# This script downloads and installs the latest version of Budgie

set -e

VERSION="0.1.0"
REPO="zarigata/budgie"
BINARY_NAME="budgie"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/budgie"
DATA_DIR="/var/lib/budgie"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_banner() {
    echo -e "${CYAN}"
    echo "    ____            __     _      "
    echo "   / __ )__  ______/ /__  (_)__   "
    echo "  / __  / / / / __  / _ \/ / _ \  "
    echo " / /_/ / /_/ / /_/ /  __/ /  __/  "
    echo "/_____/\__,_/\__,_/\___/_/\___/   "
    echo ""
    echo -e "${NC}"
    echo -e "${GREEN}Budgie Installer v${VERSION}${NC}"
    echo "=================================="
    echo ""
}

error() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

warn() {
    echo -e "${YELLOW}Warning: $1${NC}"
}

info() {
    echo -e "${CYAN}$1${NC}"
}

success() {
    echo -e "${GREEN}$1${NC}"
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        error "This script must be run as root (use sudo)"
    fi
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)
            PLATFORM="linux"
            ;;
        darwin)
            PLATFORM="darwin"
            ;;
        *)
            error "Unsupported operating system: $OS"
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac

    info "Detected platform: ${PLATFORM}-${ARCH}"
}

# Check for required commands
check_dependencies() {
    for cmd in curl tar; do
        if ! command -v "$cmd" &> /dev/null; then
            error "Required command not found: $cmd"
        fi
    done
}

# Download and install Budgie
install_budgie() {
    BINARY="${BINARY_NAME}-${PLATFORM}-${ARCH}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${BINARY}.tar.gz"
    TMP_DIR=$(mktemp -d)

    info "Downloading Budgie v${VERSION}..."

    # Try to download from releases, fall back to building from source
    if curl -fsSL --head "$DOWNLOAD_URL" &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/budgie.tar.gz" || error "Download failed"
        tar -xzf "${TMP_DIR}/budgie.tar.gz" -C "${TMP_DIR}" || error "Extraction failed"

        if [ -f "${TMP_DIR}/${BINARY}" ]; then
            BINARY_PATH="${TMP_DIR}/${BINARY}"
        elif [ -f "${TMP_DIR}/budgie" ]; then
            BINARY_PATH="${TMP_DIR}/budgie"
        else
            error "Binary not found in archive"
        fi
    else
        warn "Pre-built binary not available. Attempting to build from source..."

        if ! command -v go &> /dev/null; then
            error "Go is required to build from source. Install Go 1.21+ and try again."
        fi

        info "Cloning repository..."
        git clone --depth 1 "https://github.com/${REPO}.git" "${TMP_DIR}/budgie-src" || error "Clone failed"

        info "Building Budgie..."
        cd "${TMP_DIR}/budgie-src"
        go build -ldflags="-s -w" -o "${TMP_DIR}/budgie" ./cmd/budgie || error "Build failed"
        BINARY_PATH="${TMP_DIR}/budgie"
    fi

    info "Installing binary to ${INSTALL_DIR}..."
    cp "$BINARY_PATH" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    # Cleanup
    rm -rf "$TMP_DIR"

    success "Binary installed to: ${INSTALL_DIR}/${BINARY_NAME}"
}

# Create directories
create_directories() {
    info "Creating directories..."
    mkdir -p "${CONFIG_DIR}"
    mkdir -p "${DATA_DIR}"
    mkdir -p "${DATA_DIR}/containers"
    mkdir -p "/var/log/budgie"
    success "Directories created"
}

# Create default configuration
create_config() {
    if [ ! -f "${CONFIG_DIR}/config.yaml" ]; then
        info "Creating default configuration..."
        cat > "${CONFIG_DIR}/config.yaml" << 'EOF'
# Budgie Configuration
# https://github.com/zarigata/budgie

data_dir: "/var/lib/budgie"

runtime:
  address: "/run/containerd/containerd.sock"

discovery:
  enabled: true
  port: 5353
  service_name: "_budgie._tcp"

proxy:
  type: "round-robin"  # Options: round-robin, least-connections
  health_check_interval: 30s
  health_check_timeout: 5s

sync:
  enabled: true
  port: 9876

logging:
  level: "info"  # Options: debug, info, warn, error
  file: "/var/log/budgie/budgie.log"
EOF
        success "Configuration created: ${CONFIG_DIR}/config.yaml"
    else
        warn "Configuration already exists, skipping..."
    fi
}

# Verify installation
verify_installation() {
    if command -v budgie &> /dev/null; then
        INSTALLED_VERSION=$(budgie --version 2>/dev/null || echo "unknown")
        success "Budgie installed successfully!"
        echo ""
        echo "Version: ${INSTALLED_VERSION}"
    else
        error "Installation verification failed"
    fi
}

# Print post-install instructions
print_instructions() {
    echo ""
    echo -e "${CYAN}Quick Start:${NC}"
    echo "  1. Run the setup wizard:  budgie nest"
    echo "  2. Or run a container:    budgie run example.bun"
    echo "  3. Discover on LAN:       budgie chirp"
    echo "  4. Get help:              budgie --help"
    echo ""
    echo -e "${CYAN}Documentation:${NC}"
    echo "  https://zarigata.github.io/budgie/"
    echo ""
    echo -e "${CYAN}Requirements:${NC}"
    echo "  - containerd (for container runtime)"
    echo "  - Ensure containerd is running: systemctl status containerd"
    echo ""
}

# Main installation flow
main() {
    print_banner
    check_root
    check_dependencies
    detect_platform
    install_budgie
    create_directories
    create_config
    verify_installation
    print_instructions
}

main "$@"
