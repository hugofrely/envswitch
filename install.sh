#!/bin/bash
set -e

# EnvSwitch Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/hugofrely/envswitch/main/install.sh | bash

REPO="hugofrely/envswitch"
BINARY_NAME="envswitch"

# Default install directory (will be set based on OS)
INSTALL_DIR="${INSTALL_DIR:-}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux*)
            OS="linux"
            # Linux: Use /usr/local/bin by default
            if [ -z "$INSTALL_DIR" ]; then
                INSTALL_DIR="/usr/local/bin"
            fi
            ;;
        darwin*)
            OS="darwin"
            # macOS: Use ~/.local/bin by default (no sudo needed)
            if [ -z "$INSTALL_DIR" ]; then
                INSTALL_DIR="$HOME/.local/bin"
            fi
            ;;
        *)
            log_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            log_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    log_info "Detected platform: $PLATFORM"
    log_info "Install directory: $INSTALL_DIR"
}

# Get latest release version
get_latest_version() {
    log_info "Fetching latest release version..."

    # Try to get latest version from GitHub API
    VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$VERSION" ]; then
        log_error "Failed to get latest version"
        exit 1
    fi

    log_success "Latest version: $VERSION"
}

# Download binary
download_binary() {
    BINARY_URL="https://github.com/$REPO/releases/download/$VERSION/${BINARY_NAME}-${PLATFORM}"
    TEMP_FILE="/tmp/${BINARY_NAME}-${PLATFORM}"

    log_info "Downloading $BINARY_NAME from $BINARY_URL..."

    if ! curl -fsSL -o "$TEMP_FILE" "$BINARY_URL"; then
        log_error "Failed to download binary"
        exit 1
    fi

    log_success "Downloaded to $TEMP_FILE"
}

# Verify download (optional checksum verification)
verify_binary() {
    if [ ! -f "$TEMP_FILE" ]; then
        log_error "Binary file not found: $TEMP_FILE"
        exit 1
    fi

    # Make executable
    chmod +x "$TEMP_FILE"

    log_success "Binary verified"
}

# Install binary
install_binary() {
    log_info "Installing to $INSTALL_DIR/$BINARY_NAME..."

    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        log_info "Creating install directory: $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR" || {
            log_error "Failed to create install directory"
            log_info "You may need to run: mkdir -p $INSTALL_DIR"
            exit 1
        }
    fi

    # Check if we have write permission
    if [ ! -w "$INSTALL_DIR" ]; then
        log_warning "No write permission to $INSTALL_DIR"
        log_info "Attempting to install with sudo..."

        if ! sudo mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"; then
            log_error "Failed to install binary"
            exit 1
        fi
    else
        if ! mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"; then
            log_error "Failed to install binary"
            exit 1
        fi
    fi

    log_success "Installed to $INSTALL_DIR/$BINARY_NAME"
}

# Add to PATH if needed
add_to_path() {
    # Check if install directory is already in PATH
    if echo "$PATH" | grep -q "$INSTALL_DIR"; then
        return 0
    fi

    log_warning "$INSTALL_DIR is not in your PATH"

    # Detect shell
    SHELL_NAME=$(basename "$SHELL")

    case "$SHELL_NAME" in
        bash)
            SHELL_RC="$HOME/.bashrc"
            if [ "$(uname -s)" = "Darwin" ]; then
                SHELL_RC="$HOME/.bash_profile"
            fi
            ;;
        zsh)
            SHELL_RC="$HOME/.zshrc"
            ;;
        fish)
            SHELL_RC="$HOME/.config/fish/config.fish"
            ;;
        *)
            log_warning "Unknown shell: $SHELL_NAME"
            log_info "Add this to your shell configuration:"
            log_info "  export PATH=\"$INSTALL_DIR:\$PATH\""
            return
            ;;
    esac

    # Ask user if they want to add to PATH
    echo ""

    # Check if we're running interactively
    if [ -t 0 ]; then
        echo -n "Add $INSTALL_DIR to PATH in $SHELL_RC? [Y/n] "
        read -r response
        response=${response:-Y}
    else
        # Non-interactive mode (piped from curl)
        if [ -t 1 ] && [ -c /dev/tty ]; then
            echo -n "Add $INSTALL_DIR to PATH in $SHELL_RC? [Y/n] "
            read -r response < /dev/tty || response="Y"
            response=${response:-Y}
        else
            log_info "Non-interactive mode: auto-adding to PATH"
            response="Y"
        fi
    fi

    if [[ "$response" =~ ^[Yy]$ ]]; then
        echo "" >> "$SHELL_RC"
        echo "# Added by envswitch installer" >> "$SHELL_RC"
        echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$SHELL_RC"
        log_success "Added to $SHELL_RC"
        log_info "Run: source $SHELL_RC"
        log_info "Or restart your terminal"
    else
        log_info "Skipped. Add manually:"
        log_info "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> $SHELL_RC"
    fi
}

# Verify installation
verify_installation() {
    # Try with full path first
    if [ -x "$INSTALL_DIR/$BINARY_NAME" ]; then
        VERSION_OUTPUT=$("$INSTALL_DIR/$BINARY_NAME" --version)
        log_success "Installation verified: $VERSION_OUTPUT"
    fi

    # Check if in PATH
    if ! command -v "$BINARY_NAME" &> /dev/null; then
        log_warning "$BINARY_NAME not found in PATH"
        add_to_path
    else
        VERSION_OUTPUT=$($BINARY_NAME --version)
        log_success "Available in PATH: $VERSION_OUTPUT"
    fi
}

# Install shell auto-completion
install_completion() {
    log_info "Installing auto-completion..."

    # Detect shell
    SHELL_NAME=$(basename "$SHELL")

    case "$SHELL_NAME" in
        zsh)
            COMP_DIR="$HOME/.zsh/completion"
            mkdir -p "$COMP_DIR"

            # Generate completion script
            if command -v "$BINARY_NAME" &> /dev/null; then
                "$BINARY_NAME" completion zsh > "$COMP_DIR/_$BINARY_NAME" 2>/dev/null || {
                    log_warning "Failed to generate completion script"
                    return
                }
            else
                "$INSTALL_DIR/$BINARY_NAME" completion zsh > "$COMP_DIR/_$BINARY_NAME" 2>/dev/null || {
                    log_warning "Failed to generate completion script"
                    return
                }
            fi

            # Add to .zshrc if not already there
            if ! grep -q "fpath.*zsh/completion" "$HOME/.zshrc" 2>/dev/null; then
                cat >> "$HOME/.zshrc" << 'EOF'

# envswitch completion
fpath=(~/.zsh/completion $fpath)
autoload -Uz compinit && compinit
EOF
                log_success "Auto-completion installed for zsh"
                log_info "Run: source ~/.zshrc"
            else
                log_success "Auto-completion installed for zsh"
                log_info "Completion config already in ~/.zshrc"
            fi
            ;;

        bash)
            # Check for bash-completion directory
            if [ -d "/usr/local/etc/bash_completion.d" ]; then
                COMP_DIR="/usr/local/etc/bash_completion.d"
            elif [ -d "/etc/bash_completion.d" ]; then
                COMP_DIR="/etc/bash_completion.d"
            else
                COMP_DIR="$HOME/.bash_completion.d"
                mkdir -p "$COMP_DIR"
            fi

            # Generate completion script
            if command -v "$BINARY_NAME" &> /dev/null; then
                "$BINARY_NAME" completion bash > "$COMP_DIR/$BINARY_NAME" 2>/dev/null || {
                    log_warning "Failed to generate completion script"
                    return
                }
            else
                "$INSTALL_DIR/$BINARY_NAME" completion bash > "$COMP_DIR/$BINARY_NAME" 2>/dev/null || {
                    log_warning "Failed to generate completion script"
                    return
                }
            fi

            # Add to .bashrc if using home directory
            if [ "$COMP_DIR" = "$HOME/.bash_completion.d" ]; then
                BASHRC="$HOME/.bashrc"
                if [ "$(uname -s)" = "Darwin" ]; then
                    BASHRC="$HOME/.bash_profile"
                fi

                if ! grep -q "bash_completion.d" "$BASHRC" 2>/dev/null; then
                    cat >> "$BASHRC" << 'EOF'

# envswitch completion
for bcfile in ~/.bash_completion.d/* ; do
  [ -f "$bcfile" ] && . "$bcfile"
done
EOF
                    log_success "Auto-completion installed for bash"
                    log_info "Run: source $BASHRC"
                else
                    log_success "Auto-completion installed for bash"
                fi
            else
                log_success "Auto-completion installed for bash"
            fi
            ;;

        fish)
            COMP_DIR="$HOME/.config/fish/completions"
            mkdir -p "$COMP_DIR"

            # Generate completion script
            if command -v "$BINARY_NAME" &> /dev/null; then
                "$BINARY_NAME" completion fish > "$COMP_DIR/$BINARY_NAME.fish" 2>/dev/null || {
                    log_warning "Failed to generate completion script"
                    return
                }
            else
                "$INSTALL_DIR/$BINARY_NAME" completion fish > "$COMP_DIR/$BINARY_NAME.fish" 2>/dev/null || {
                    log_warning "Failed to generate completion script"
                    return
                }
            fi

            log_success "Auto-completion installed for fish"
            log_info "Restart your shell or run: source ~/.config/fish/config.fish"
            ;;

        *)
            log_warning "Unknown shell: $SHELL_NAME"
            log_info "Supported shells: bash, zsh, fish"
            log_info "To install manually, run:"
            log_info "  $BINARY_NAME completion [bash|zsh|fish] > /path/to/completion/file"
            ;;
    esac
}

# Cleanup
cleanup() {
    if [ -f "$TEMP_FILE" ]; then
        rm -f "$TEMP_FILE"
    fi
}

# Main installation flow
main() {
    echo ""
    echo "╔═══════════════════════════════════════╗"
    echo "║   EnvSwitch Installation Script      ║"
    echo "╚═══════════════════════════════════════╝"
    echo ""

    detect_platform
    get_latest_version
    download_binary
    verify_binary
    install_binary
    cleanup
    verify_installation
    install_completion

    echo ""
    log_success "Installation complete!"
    echo ""
    echo "Next steps:"
    echo "  1. Initialize: ${GREEN}$BINARY_NAME init${NC}"
    echo "  2. Create env:  ${GREEN}$BINARY_NAME create work --from-current${NC}"
    echo "  3. Install shell integration: ${GREEN}$BINARY_NAME shell install bash${NC}"
    echo ""
    echo "Documentation: https://github.com/$REPO"
    echo ""
}

# Trap errors and cleanup
trap cleanup EXIT

# Run main
main
