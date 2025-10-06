#!/bin/bash
set -e

# EnvSwitch installation script

VERSION="latest"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="envswitch"

echo "🔧 Installing EnvSwitch..."

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case $OS in
    darwin)
        OS="darwin"
        ;;
    linux)
        OS="linux"
        ;;
    *)
        echo "❌ Unsupported OS: $OS"
        exit 1
        ;;
esac

echo "Detected: $OS/$ARCH"

# Download URL (adjust when releases are available)
DOWNLOAD_URL="https://github.com/hugofrely/envswitch/releases/download/${VERSION}/envswitch_${OS}_${ARCH}.tar.gz"

# Create temp directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

echo "📥 Downloading EnvSwitch..."
if command -v curl &> /dev/null; then
    curl -fsSL "$DOWNLOAD_URL" -o envswitch.tar.gz
elif command -v wget &> /dev/null; then
    wget -q "$DOWNLOAD_URL" -O envswitch.tar.gz
else
    echo "❌ Neither curl nor wget found. Please install one of them."
    exit 1
fi

echo "📦 Extracting..."
tar -xzf envswitch.tar.gz

echo "🔧 Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/"
else
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
fi

# Make executable
if [ -w "$INSTALL_DIR/$BINARY_NAME" ]; then
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# Cleanup
cd -
rm -rf "$TMP_DIR"

echo "✅ EnvSwitch installed successfully!"
echo ""
echo "Get started:"
echo "  1. Initialize: envswitch init"
echo "  2. Create environment: envswitch create myenv --from-current"
echo "  3. Switch: envswitch switch myenv"
echo ""
echo "For more information: envswitch --help"
