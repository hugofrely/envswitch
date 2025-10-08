# Installation Guide

## Quick Install (Recommended)

### One-line Install

```bash
curl -fsSL https://raw.githubusercontent.com/hugofrely/envswitch/main/install.sh | bash
```

This will:
1. ✅ Detect your OS and architecture automatically
2. ✅ Download the latest release binary
3. ✅ Install to the appropriate location:
   - **macOS**: `~/.local/bin/envswitch` (no sudo needed)
   - **Linux**: `/usr/local/bin/envswitch` (may need sudo)
4. ✅ Add to PATH if needed (will ask for confirmation)
5. ✅ Install shell auto-completion (optional, will prompt)
6. ✅ Verify the installation

### Custom Install Directory

```bash
# Install to a specific directory
INSTALL_DIR="/usr/local/bin" curl -fsSL https://raw.githubusercontent.com/hugofrely/envswitch/main/install.sh | bash

# Install to user directory (no sudo)
INSTALL_DIR="$HOME/bin" curl -fsSL https://raw.githubusercontent.com/hugofrely/envswitch/main/install.sh | bash
```

## Manual Installation

### Download Binary

Visit [GitHub Releases](https://github.com/hugofrely/envswitch/releases/latest) and download the appropriate binary:

#### macOS (Apple Silicon) - No sudo needed

```bash
curl -LO https://github.com/hugofrely/envswitch/releases/latest/download/envswitch-darwin-arm64
chmod +x envswitch-darwin-arm64
mkdir -p ~/.local/bin
mv envswitch-darwin-arm64 ~/.local/bin/envswitch

# Add to PATH if not already there
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

#### macOS (Intel) - No sudo needed

```bash
curl -LO https://github.com/hugofrely/envswitch/releases/latest/download/envswitch-darwin-amd64
chmod +x envswitch-darwin-amd64
mkdir -p ~/.local/bin
mv envswitch-darwin-amd64 ~/.local/bin/envswitch

# Add to PATH if not already there
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

<details>
<summary>Install to /usr/local/bin (requires sudo)</summary>

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/hugofrely/envswitch/releases/latest/download/envswitch-darwin-arm64
chmod +x envswitch-darwin-arm64
sudo mv envswitch-darwin-arm64 /usr/local/bin/envswitch

# macOS (Intel)
curl -LO https://github.com/hugofrely/envswitch/releases/latest/download/envswitch-darwin-amd64
chmod +x envswitch-darwin-amd64
sudo mv envswitch-darwin-amd64 /usr/local/bin/envswitch
```

</details>

#### Linux (AMD64)

```bash
curl -LO https://github.com/hugofrely/envswitch/releases/latest/download/envswitch-linux-amd64
chmod +x envswitch-linux-amd64
sudo mv envswitch-linux-amd64 /usr/local/bin/envswitch
```

#### Linux (ARM64)

```bash
curl -LO https://github.com/hugofrely/envswitch/releases/latest/download/envswitch-linux-arm64
chmod +x envswitch-linux-arm64
sudo mv envswitch-linux-arm64 /usr/local/bin/envswitch
```

#### Windows (AMD64)

Download `envswitch-windows-amd64.exe` from the releases page and add it to your PATH.

### Verify Installation

```bash
envswitch --version
```

## Install from Source

Requires Go 1.21+:

```bash
# Clone repository
git clone https://github.com/hugofrely/envswitch.git
cd envswitch

# Build and install
make install

# Or just build locally
make build
./bin/envswitch --version
```

## Post-Installation Setup

### 1. Initialize EnvSwitch

```bash
envswitch init
```

### 2. Install Shell Integration (Optional)

```bash
# For bash
envswitch shell install bash
source ~/.bashrc

# For zsh
envswitch shell install zsh
source ~/.zshrc

# For fish
envswitch shell install fish
source ~/.config/fish/config.fish
```

### 3. Enable Auto-completion (Optional)

```bash
# Bash
envswitch completion bash > /usr/local/etc/bash_completion.d/envswitch

# Zsh
envswitch completion zsh > "${fpath[1]}/_envswitch"

# Fish
envswitch completion fish > ~/.config/fish/completions/envswitch.fish
```

## Supported Platforms

| OS      | Architecture | Binary Name                  | Supported |
|---------|--------------|------------------------------|-----------|
| macOS   | ARM64        | envswitch-darwin-arm64      | ✅        |
| macOS   | AMD64        | envswitch-darwin-amd64      | ✅        |
| Linux   | AMD64        | envswitch-linux-amd64       | ✅        |
| Linux   | ARM64        | envswitch-linux-arm64       | ✅        |
| Windows | AMD64        | envswitch-windows-amd64.exe | ✅        |

## Updating

To update to the latest version, simply run the install script again:

```bash
curl -fsSL https://raw.githubusercontent.com/hugofrely/envswitch/main/install.sh | bash
```

Or download the latest binary manually from the releases page.

## Uninstalling

```bash
# Remove binary
sudo rm /usr/local/bin/envswitch

# Remove data (optional)
rm -rf ~/.envswitch
```

## Troubleshooting

### "Permission denied" when installing

The script will automatically attempt to use `sudo` if it doesn't have write permission. If this fails:

```bash
# Install to a directory you own
INSTALL_DIR="$HOME/.local/bin" curl -fsSL https://raw.githubusercontent.com/hugofrely/envswitch/main/install.sh | bash

# Then add to PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### "envswitch: command not found" after installation

Make sure `/usr/local/bin` is in your PATH:

```bash
echo $PATH | grep -q "/usr/local/bin" || echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Install script fails to download

1. Check you have an internet connection
2. Verify GitHub releases exist: https://github.com/hugofrely/envswitch/releases
3. Try manual installation instead

### Wrong architecture downloaded

The script auto-detects your platform. If it fails, you can download manually:

```bash
# Check your architecture
uname -m
# x86_64 = amd64, arm64 = arm64

# Check your OS
uname -s
# Darwin = macOS, Linux = linux
```

## Next Steps

After installation, see the [Quick Start Guide](docs/QUICKSTART.md) to get started!

## Getting Help

- 📖 [Documentation](README.md)
- 🚀 [Quick Start Guide](docs/QUICKSTART.md)
- 🐛 [Report Issues](https://github.com/hugofrely/envswitch/issues)
- 💬 [Discussions](https://github.com/hugofrely/envswitch/discussions)
