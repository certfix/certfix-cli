#!/bin/bash
set -e

# GitHub repository and release info
REPO_OWNER="certfix"
REPO_NAME="certfix-cli"

# Installation path
BIN_PATH="/usr/local/bin/certfix"

# Detect operating system
detect_os() {
    local os
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case $os in
        linux)
            echo "linux"
            ;;
        darwin)
            echo "darwin"
            ;;
        *)
            echo "[ERROR] Unsupported operating system: $os" >&2
            echo "Supported operating systems: Linux, macOS" >&2
            exit 1
            ;;
    esac
}

# Detect architecture
detect_arch() {
    local arch
    arch=$(uname -m)
    case $arch in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        armv7l)
            echo "armv7"
            ;;
        *)
            echo "[ERROR] Unsupported architecture: $arch" >&2
            echo "Supported architectures: x86_64, aarch64/arm64, armv7l" >&2
            exit 1
            ;;
    esac
}

OS=$(detect_os)
ARCH=$(detect_arch)
BINARY_NAME="certfix-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/${BINARY_NAME}"

echo "[INFO] Installing Certfix CLI for ${OS}/$(uname -m) (${ARCH})..."

# Check if running as root or with sudo
if [[ $EUID -ne 0 ]]; then
    echo "[ERROR] This script must be run as root or with sudo" >&2
    exit 1
fi

# Check for curl or wget
if command -v curl &>/dev/null; then
    DOWNLOADER="curl"
elif command -v wget &>/dev/null; then
    DOWNLOADER="wget"
else
    echo "[ERROR] Neither curl nor wget found. Please install one and retry." >&2
    exit 1
fi

# Download the latest binary
echo "[INFO] Downloading latest release from ${DOWNLOAD_URL}..."
TMP_FILE=$(mktemp)
if [ "$DOWNLOADER" = "curl" ]; then
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"; then
        echo "[ERROR] Failed to download binary from $DOWNLOAD_URL" >&2
        rm -f "$TMP_FILE"
        exit 1
    fi
else
    if ! wget -qO "$TMP_FILE" "$DOWNLOAD_URL"; then
        echo "[ERROR] Failed to download binary from $DOWNLOAD_URL" >&2
        rm -f "$TMP_FILE"
        exit 1
    fi
fi

chmod +x "$TMP_FILE"
mv "$TMP_FILE" "$BIN_PATH"
echo "[INFO] Binary installed to $BIN_PATH"

# Verify installation
if ! command -v certfix &>/dev/null; then
    echo "[WARN] certfix not found in PATH. You may need to add /usr/local/bin to your PATH." >&2
else
    echo "[SUCCESS] Certfix CLI installed successfully!"
    echo ""
    certfix version
    echo ""
fi

echo "Next steps:"
echo "  certfix configure --api-url https://<your-certfix-server>"
echo "  certfix login"
echo "  certfix --help"
