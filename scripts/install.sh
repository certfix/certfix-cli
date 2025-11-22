#!/bin/bash
set -e

# GitHub repository and release info
REPO_OWNER="certfix"
REPO_NAME="certfix-cli"

# Installation paths
BIN_PATH="/usr/local/bin/certfix"

# Detect architecture
detect_arch() {
    local arch
    arch=$(uname -m)
    case $arch in
        x86_64)
            echo "amd64"
            ;;
        aarch64)
            echo "arm64"
            ;;
        armv7l)
            echo "armv7"
            ;;
        *)
            echo "[ERROR] Unsupported architecture: $arch"
            echo "Supported architectures: x86_64, aarch64, armv7l"
            exit 1
            ;;
    esac
}

ARCH=$(detect_arch)
BINARY_NAME="certfix-cli-linux-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/${BINARY_NAME}"

echo "[INFO] Installing Certfix CLI for $(uname -m) (${ARCH})..."

# Check if running as root or with sudo
if [[ $EUID -ne 0 ]]; then
   echo "[ERROR] This script must be run as root or with sudo"
   exit 1
fi

# Download the latest binary
echo "[INFO] Downloading latest release for ${ARCH}..."
if ! curl -fsSL "$DOWNLOAD_URL" -o "$BIN_PATH"; then
  echo "[ERROR] Failed to download binary from $DOWNLOAD_URL"
  echo "[INFO] Make sure you have created a release with the binary attached"
  exit 1
fi

chmod +x "$BIN_PATH"
echo "[INFO] Binary installed to $BIN_PATH"

echo "[SUCCESS] Certfix CLI installed successfully!"
echo "Architecture: $(uname -m) (${ARCH})"
echo ""
echo "Next steps:"
echo "1. Run 'certfix configure' to set up your configuration"
echo "2. Use 'certfix --help' to see available commands"