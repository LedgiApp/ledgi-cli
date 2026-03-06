#!/bin/bash
set -euo pipefail

REPO="LedgiApp/ledgi-cli"
BINARY="ledgi"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Error: Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "Error: Unsupported OS: $OS"; exit 1 ;;
esac

ASSET="${BINARY}-${OS}-${ARCH}"

echo "Detecting platform: ${OS}/${ARCH}"

# Get latest release download URL
DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

echo "Downloading ${BINARY} from ${DOWNLOAD_URL}..."
TMP="$(mktemp)"
if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP"; then
  echo "Error: Failed to download. Check https://github.com/${REPO}/releases for available binaries."
  rm -f "$TMP"
  exit 1
fi

chmod +x "$TMP"

echo "Installing to ${INSTALL_DIR}/${BINARY}..."
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "${INSTALL_DIR}/${BINARY}"
else
  sudo mv "$TMP" "${INSTALL_DIR}/${BINARY}"
fi

echo ""
echo "ledgi installed successfully!"
echo ""
echo "Get started:"
echo "  1. Get an API key from your Ledgi dashboard (Settings > API Keys)"
echo "  2. Run: ledgi login --api-key ledgi_sk_..."
echo "  3. Try: ledgi accounts list"
