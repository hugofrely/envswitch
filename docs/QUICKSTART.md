# EnvSwitch Quick Start Guide

This guide will help you get started with EnvSwitch in 5 minutes.

## Installation

### Option 1: Download Binary (Recommended)

Visit the [releases page](https://github.com/hugofrely/envswitch/releases/latest) and download the binary for your platform:

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/hugofrely/envswitch/releases/latest/download/envswitch-darwin-arm64
chmod +x envswitch-darwin-arm64
sudo mv envswitch-darwin-arm64 /usr/local/bin/envswitch

# macOS (Intel)
curl -LO https://github.com/hugofrely/envswitch/releases/latest/download/envswitch-darwin-amd64
chmod +x envswitch-darwin-amd64
sudo mv envswitch-darwin-amd64 /usr/local/bin/envswitch

# Linux
curl -LO https://github.com/hugofrely/envswitch/releases/latest/download/envswitch-linux-amd64
chmod +x envswitch-linux-amd64
sudo mv envswitch-linux-amd64 /usr/local/bin/envswitch

# Verify installation
envswitch --version
```

### Option 2: Install from Source (requires Go)

```bash
git clone https://github.com/hugofrely/envswitch.git
cd envswitch
make install
```

### Option 3: Install Script (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/hugofrely/envswitch/main/install.sh | bash
```

This automatically:

- Detects your platform
- Installs the latest version
- Configures PATH
- **Optionally installs auto-completion** (will prompt)

See [INSTALL.md](../INSTALL.md) for more options.

## First Time Setup

### 1. Initialize EnvSwitch

```bash
envswitch init
```

This creates:

- `~/.envswitch/` directory
- `~/.envswitch/config.yaml` with default configuration
- `~/.envswitch/environments/` for your environments
- `~/.envswitch/archives/` for deleted environment backups
- `~/.envswitch/history.log` for tracking switches

### 2. Install Shell Integration (Optional but Recommended)

Get environment name in your shell prompt and auto-completion:

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

Your prompt will now show the current environment: `(work) user@machine$`

### 3. Enable Auto-completion (Optional)

**If you used the install script:**
Auto-completion should already be installed (the installer prompts you).

**If you installed manually:**

```bash
# Bash
envswitch completion bash > /usr/local/etc/bash_completion.d/envswitch

# Zsh
mkdir -p ~/.zsh/completion
envswitch completion zsh > ~/.zsh/completion/_envswitch
echo 'fpath=(~/.zsh/completion $fpath)' >> ~/.zshrc
echo 'autoload -Uz compinit && compinit' >> ~/.zshrc
source ~/.zshrc

# Fish
mkdir -p ~/.config/fish/completions
envswitch completion fish > ~/.config/fish/completions/envswitch.fish
```

Now you can use `envswitch switch <TAB>` to see available environments!

📖 See [COMPLETION_SETUP.md](../COMPLETION_SETUP.md) for detailed instructions.

## Creating Your First Environment

### Capture Current State

Capture your current development environment:

```bash
envswitch create work --from-current --description "Work environment"
```

This creates a snapshot of your current:

- ✅ **GCloud** - authentication and configuration
- ✅ **Kubectl** - contexts and clusters
- ✅ **AWS** - credentials and profiles
- ✅ **Docker** - registry authentication
- ✅ **Git** - configuration
- ✅ **Environment Variables** - (if configured in metadata.yaml)

### Make Changes and Create Another Environment

```bash
# Switch to personal GCloud account
gcloud auth login personal@gmail.com
gcloud config set project my-personal-project

# Switch kubectl context
kubectl config use-context minikube

# Create personal environment
envswitch create personal --from-current --description "Personal projects"
```

## Switching Between Environments

Now you can instantly switch between your work and personal environments:

```bash
# Switch to work (shows loading spinner)
envswitch switch work

# Switch to personal
envswitch switch personal

# Switch with verification
envswitch switch work --verify

# Skip backup during switch
envswitch switch work --no-backup

# Verbose mode (see detailed logs)
envswitch switch work --verbose
```

**What happens during a switch:**

1. 🔄 Shows loading spinner with progress message
2. 📦 Creates backup of current environment (if enabled in config)
3. 💾 Saves current state to current environment
4. 🔄 Restores target environment state
5. ✅ Updates current.lock file and metadata
6. 🧹 Cleans up old backups (based on `backup_retention` config)
7. 📝 Records switch in history log

**Output modes:**
- **Normal mode**: Shows only success message and spinner
- **Verbose mode** (`--verbose`): Shows detailed debug logs for all operations

## Common Workflows

### List All Environments

```bash
# Simple list
envswitch list

# Detailed view
envswitch list --detailed

# Shows:
# - Environment names and descriptions
# - Active environment (marked with *)
# - Last used timestamp
# - Enabled tools
```

### Show Environment Details

```bash
envswitch show work

# Shows:
# - Environment metadata
# - Snapshot contents
# - Environment variables
# - Enabled tools with their metadata
# - Tags
```

### Configure Environment Variables

Environment variables are captured and restored automatically:

```bash
# Edit environment metadata
vim ~/.envswitch/environments/work/metadata.yaml

# Add environment variables:
environment_variables:
  API_KEY: ""
  DATABASE_URL: ""
  DEBUG: ""

# Next switch will capture these variables
envswitch switch personal
envswitch switch work  # Variables are restored!
```

### View Switch History

```bash
# Show recent switches (default: last 10)
envswitch history

# Show last 5 switches
envswitch history --limit 5

# Show all history
envswitch history --all

# Show detailed view with full information
envswitch history show

# Clear history
envswitch history clear
```

**History format:**
```
✅ 2025-10-06 19:36:07  personal → work  1.41s
✅ 2025-10-06 18:22:15  work → personal  1.23s
```

### Rollback to Previous Environment

```bash
# Undo last switch
envswitch rollback

# Rollback to specific history entry
envswitch rollback --to 3
```

### Compare Environments

```bash
# Compare two environments
envswitch diff work personal

# Shows differences in:
# - Tool configurations
# - Environment variables
# - Metadata
```

### Delete an Environment

```bash
# Delete with archive (default)
envswitch delete old-env

# Force delete without confirmation
envswitch delete old-env --force

# Delete without creating archive
envswitch delete old-env --no-archive
```

**Note:** Deleted environments are automatically archived to `~/.envswitch/archives/` unless `--no-archive` is used.

### Restore Deleted Environment

```bash
# List archives
envswitch archive list

# Restore from archive
envswitch archive restore old-env-20250106-120000.tar.gz
```

## Configuration

### View Configuration

```bash
# List all config
envswitch config list

# Get specific value
envswitch config get log_level
```

### Set Configuration

```bash
# Set auto-save behavior
envswitch config set auto_save_before_switch true

# Set backup retention (number of backups to keep)
envswitch config set backup_retention 10

# Enable verification after switch
envswitch config set verify_after_switch true

# Set log level (debug, info, warn, error) - default: warn
envswitch config set log_level warn

# Enable/disable backup before each switch (default: true)
envswitch config set backup_before_switch true

# Enable color output
envswitch config set color_output true

# Show timestamps in output
envswitch config set show_timestamps false

# Customize prompt
envswitch config set prompt_format "[{env}] "
envswitch config set prompt_color cyan

# Exclude tools from snapshots
envswitch config set exclude_tools docker,git
```

### Configuration File

Edit `~/.envswitch/config.yaml` directly:

```yaml
version: "1.0"
auto_save_before_switch: true # true, false, or prompt
verify_after_switch: false
backup_retention: 10
backup_before_switch: true # Create backup before each switch
log_level: warn # Default log level (debug, info, warn, error)
log_file: ~/.envswitch/envswitch.log
color_output: true
show_timestamps: false

# Shell integration
enable_prompt_integration: true
prompt_format: "({name})"
prompt_color: blue

# Tool exclusions
exclude_tools: [] # Skip specific tools
```

## Advanced Features

### Pre/Post Switch Hooks

Add custom commands to run before/after switching:

```bash
# Edit environment metadata
vim ~/.envswitch/environments/work/metadata.yaml

# Add hooks:
hooks:
  pre_switch:
    - command: "echo Switching to work environment"
    - script: "./scripts/pre-switch.sh"
      verify: true

  post_switch:
    - command: "kubectl config current-context"
    - command: "gcloud config get-value project"
```

### Dry Run Mode

Preview what would happen without making changes:

```bash
envswitch switch work --dry-run
```

### Verification

Verify environment after switching:

```bash
# One-time verification
envswitch switch work --verify

# Always verify (in config)
envswitch config set verify_after_switch true
```

## Troubleshooting

### Check Version

```bash
envswitch --version
# Shows: version, git commit, build date
```

### View Logs

```bash
# Default log location
tail -f ~/.envswitch/envswitch.log

# Enable debug logging
envswitch config set log_level debug
```

### Common Issues

**Environment not switching?**

- ✅ All core functionality is implemented
- ✅ Full snapshot/restore works for all 5 tools
- ✅ Environment variables capture/restore works
- ✅ Hooks system works
- ✅ History and rollback work

**Missing tool configurations?**

- Some tools may not have active configurations
- Check tool status with `envswitch show <env-name>`
- Only enabled tools are snapshotted

**Shell prompt not updating?**

- Make sure you ran `envswitch shell install`
- Reload your shell: `source ~/.bashrc` (or .zshrc)
- Check config: `envswitch config get enable_prompt_integration`

## Plugin System

Want to add support for new tools like npm, terraform, or pip? EnvSwitch makes it incredibly simple!

### Creating a Plugin (2 minutes, no code required!)

Most plugins need **zero Go code**—just a simple YAML file:

```bash
# 1. Create plugin directory
mkdir my-tool-plugin
cd my-tool-plugin

# 2. Create plugin.yaml
cat > plugin.yaml << 'EOF'
metadata:
  name: my-tool
  version: 1.0.0
  description: Support for my-tool
  tool_name: my-tool
  author: Your Name
EOF

# 3. Install plugin
envswitch plugin install .
```

**That's it!** The plugin is automatically:
- ✅ Installed
- ✅ Activated in ALL your environments
- ✅ Capturing config during every switch

EnvSwitch automatically detects config file locations:
- `npm` → `~/.npmrc`
- `yarn` → `~/.yarnrc`
- `pip` → `~/.pip/pip.conf`
- `terraform` → `~/.terraform.d/`
- Custom tools → `~/.TOOLNAME` or `~/.TOOLNAMErc`

### Managing Plugins

```bash
# List installed plugins
envswitch plugin list

# Show plugin details
envswitch plugin info npm

# Remove plugin
envswitch plugin remove npm
```

📖 **Full plugin guide**: See [Plugin Documentation](../docs/PLUGINS.md) for complete examples and advanced features.

## What's Next?

- 📖 Read the [full documentation](../README.md)
- 🔌 Create your own plugins with [Plugin Documentation](../docs/PLUGINS.md)
- 🚀 Check out [versioning system](../VERSIONING.md)
- 🐛 [Report issues](https://github.com/hugofrely/envswitch/issues)
- 💬 [Join discussions](https://github.com/hugofrely/envswitch/discussions)

## Getting Help

- 📖 **Full Documentation**: See [README.md](../README.md)
- 🐛 **Report Issues**: [GitHub Issues](https://github.com/hugofrely/envswitch/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/hugofrely/envswitch/discussions)
- 📝 **Project Status**: See [PROJECT_STATUS.md](../PROJECT_STATUS.md)

## Project Status

- ✅ **Phase 1 (MVP)**: COMPLETED

  - All 5 tools implemented (gcloud, kubectl, aws, docker, git)
  - Full switching logic with snapshot/restore
  - Configuration system
  - History tracking with detailed view
  - Hooks system (pre/post switch)
  - Archive system
  - Import/Export functionality

- ✅ **Phase 2 (Essential Features)**: COMPLETED

  - Environment variables handling
  - Shell integration (bash, zsh, fish)
  - Auto-completion
  - Prompt customization
  - Loading spinner during switch
  - Verbose mode for detailed logging
  - Backup configuration options

- 🚧 **Phase 3 (Advanced Features)**: IN PROGRESS
  - ✅ Plugin system with auto-activation (no code required)
  - Encryption support
  - TUI (Terminal UI)
  - Template system
  - Git sync

## Contributing

Want to help make EnvSwitch better? See [CONTRIBUTING.md](../CONTRIBUTING.md)!

---

**Happy switching! 🚀**
