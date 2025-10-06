# EnvSwitch 🔄

> **Snapshots for your development environments**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![Project Status](https://img.shields.io/badge/status-beta-yellow.svg)](PROJECT_STATUS.md)

EnvSwitch is a powerful CLI tool that captures, saves, and restores the complete state of your development environments. Switch instantly between work and personal projects, different client environments, or testing scenarios—without losing your authentication, configurations, or contexts.

---

## 🎯 The Problem

As a developer, you probably work across multiple environments:

```bash
# Client A - Morning
gcloud auth login user@companyA.com
gcloud config set project companyA-prod-123
kubectl config use-context gke-companyA-cluster
export ENV=production

# Client B - Afternoon
gcloud auth login user@companyB.com
gcloud config set project companyB-dev-456
kubectl config use-context gke-companyB-cluster
export ENV=development

# Personal projects - Evening
gcloud auth login personal@gmail.com
kubectl config use-context minikube
export ENV=local
```

**This is exhausting.** And error-prone. What if you forget to switch? Deploy to the wrong environment? Lose hours troubleshooting?

## 💡 The Solution

**EnvSwitch creates snapshots of your entire dev environment.**

```bash
# Create environment snapshots once
envswitch create work --from-current
envswitch create personal --from-current
envswitch create clientA --from-current

# Then switch instantly, anytime
envswitch switch work        # All your work configs restored
envswitch switch personal    # All your personal configs restored
envswitch switch clientA     # All clientA configs restored
```

One command. Everything restored. **Instantly.**

---

## ✨ Features

### 🎯 **Environment Management**

- Create unlimited environment snapshots
- Clone environments for quick variations
- Tag and organize environments
- Delete old environments safely

### 📸 **Comprehensive Snapshots**

Captures complete state of:

- **GCloud CLI** - Authentication, projects, configurations
- **Kubectl** - Contexts, clusters, namespaces, configs
- **AWS CLI** - Profiles, credentials, regions
- **Docker** - Registry authentication
- **Git** - User config (name, email, signing keys)
- **Environment Variables** - Custom variables per environment

### 🔄 **Smart Switching**

- Automatic backup before switch
- Atomic operations (all or nothing)
- Rollback on failure
- Verification after switch
- History tracking

### 🛡️ **Safety First**

- Automatic backups before every switch
- Dry-run mode to preview changes
- Diff to see what would change
- Never lose your configurations

### 🎨 **Developer Experience**

- Beautiful CLI output
- Shell integration (prompt indicator)
- Auto-completion (bash/zsh/fish)
- Hooks for automation
- Detailed logging

---

## 🚀 Quick Start

### Installation

#### Option 1: Install Script (Recommended for macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/hugofrely/envswitch/main/install.sh | bash
```

Or with wget:

```bash
wget -qO- https://raw.githubusercontent.com/hugofrely/envswitch/main/install.sh | bash
```

#### Option 2: Go Install

```bash
go install github.com/hugofrely/envswitch@latest
```

#### Option 3: Build from Source

```bash
git clone https://github.com/hugofrely/envswitch.git
cd envswitch
make install
```

#### Option 4: Download Binary

Download the latest release from [GitHub Releases](https://github.com/hugofrely/envswitch/releases).

### First Steps

```bash
# 1. Initialize EnvSwitch
envswitch init

# 2. Create your first environment (captures current state)
envswitch create work --from-current \
    --description "Work environment with company auth"

# 3. Make changes to your environment
gcloud auth login personal@gmail.com
kubectl config use-context minikube

# 4. Create another environment
envswitch create personal --from-current \
    --description "Personal projects"

# 5. Switch between environments instantly!
envswitch switch work      # Back to work configs
envswitch switch personal  # Back to personal configs
```

That's it! 🎉

---

## 📖 Usage

### Creating Environments

```bash
# Create from current system state
envswitch create myenv --from-current

# Create empty environment
envswitch create myenv --empty

# Clone existing environment
envswitch create dev --from prod

# With description
envswitch create staging --from-current \
    --description "Staging environment for testing"
```

### Listing Environments

```bash
# Simple list
envswitch list

# Detailed view
envswitch ls --detailed

# Output shows active environment with *
#   * work - Work environment
#     personal - Personal projects
#     clientA - Client A production
```

### Switching Environments

```bash
# Switch to environment (with loading spinner)
envswitch switch myenv

# Preview changes without applying
envswitch switch myenv --dry-run

# Switch with verification
envswitch switch myenv --verify

# Skip backup during switch
envswitch switch myenv --no-backup

# Verbose mode (shows detailed logs)
envswitch switch myenv --verbose
```

### Viewing Environment Details

```bash
# Show detailed information
envswitch show work

# Output:
# Environment: work
# Description: Work environment
# Created: 2024-01-15 09:30:00
# Last used: 2024-01-15 14:22:15
#
# 📸 Snapshot Contents:
#   ✓ gcloud
#     - account: user@company.com
#     - project: company-prod-123
#   ✓ kubectl
#     - context: gke-company-cluster
#   ✓ aws
#   ✓ docker
#   ✓ git
```

### Deleting Environments

```bash
# Delete with confirmation
envswitch delete myenv

# Force delete without confirmation
envswitch rm myenv --force
```

### Viewing Switch History

```bash
# Show last 10 switches (default)
envswitch history

# Show last 20 switches
envswitch history --limit 20

# Show all history
envswitch history --all

# Show detailed view with full information
envswitch history show

# Clear history
envswitch history clear
```

### Import/Export Environments

```bash
# Export single environment
envswitch export myenv --output myenv-backup.tar.gz

# Export all environments
envswitch export --all --output ./backups

# Import environment
envswitch import myenv-backup.tar.gz

# Import with different name
envswitch import myenv-backup.tar.gz --name new-env

# Force overwrite existing
envswitch import myenv-backup.tar.gz --force

# Import all from directory
envswitch import --all ./backups
```

### Terminal UI (Interactive Mode)

```bash
# Launch interactive TUI
envswitch tui

# Navigate with keyboard:
#   ↑/↓ - Move selection
#   Enter - View details / Switch
#   r - Refresh
#   d - Delete
#   q/Esc - Quit
```

### Plugin Management

```bash
# List installed plugins
envswitch plugin list

# Install plugin
envswitch plugin install ./my-plugin

# Show plugin information
envswitch plugin info terraform

# Remove plugin
envswitch plugin remove terraform
```

**📖 Plugin Development**: See [Plugin Documentation](docs/PLUGINS.md) for how to create and distribute your own plugins.

---

## 🛠️ Supported Tools

| Tool           | Status         | What's Captured                                                 |
| -------------- | -------------- | --------------------------------------------------------------- |
| **GCloud CLI** | 🚧 In Progress | Authentication, active account, project, region, configurations |
| **Kubectl**    | 🚧 In Progress | Contexts, clusters, current namespace, kubeconfig               |
| **AWS CLI**    | 🚧 In Progress | Profiles, credentials, default region, config                   |
| **Docker**     | 🚧 In Progress | Registry authentication, config.json                            |
| **Git**        | 🚧 In Progress | User name, email, signing keys                                  |
| **Azure CLI**  | 📅 Planned     | Authentication, subscriptions, defaults                         |
| **Terraform**  | 📅 Planned     | Workspaces, backend config                                      |
| **SSH**        | 📅 Planned     | SSH keys, config                                                |

**Legend:**

- ✅ Implemented
- 🚧 In Progress
- 📅 Planned

---

## 📁 How It Works

EnvSwitch stores environment snapshots in `~/.envswitch/`:

```
~/.envswitch/
├── config.yaml              # Global configuration
├── environments/            # All your environments
│   ├── work/
│   │   ├── metadata.yaml    # Environment info
│   │   ├── snapshots/       # Tool configurations
│   │   │   ├── gcloud/      # Copy of ~/.config/gcloud/
│   │   │   ├── kubectl/     # Copy of ~/.kube/
│   │   │   ├── aws/         # Copy of ~/.aws/
│   │   │   ├── docker/      # Copy of ~/.docker/
│   │   │   └── git/         # Git configuration
│   │   └── env-vars.env     # Environment variables
│   │
│   ├── personal/
│   │   └── ...
│   │
│   └── clientA/
│       └── ...
│
├── auto-backups/            # Safety backups
├── current.lock             # Active environment marker
└── history.log              # Switch history
```

### When You Switch

1. 🔒 **Creates safety backup** of current state
2. 💾 **Saves current state** to the active environment
3. 🔄 **Restores target environment** from its snapshot
4. ✅ **Updates tracking** (current.lock, history)

If anything goes wrong, your data is safe in auto-backups!

---

## ⚙️ Configuration

Global config at `~/.envswitch/config.yaml`:

```yaml
# Behavior
auto_save_before_switch: true # Auto-save before switching
verify_after_switch: false # Verify connectivity after switch
backup_before_switch: true # Create backup before each switch
backup_retention: 10 # Keep last 10 auto-backups

# UI
color_output: true # Colored output
show_timestamps: false # Show timestamps in output

# Shell Integration
enable_prompt_integration: true # Show env in prompt
prompt_format: "({name})" # Format: (work)
prompt_color: blue # Prompt color

# Logging
log_level: warn # debug, info, warn, error (default: warn)
log_file: ~/.envswitch/envswitch.log

# Tools
exclude_tools: [] # Skip specific tools (e.g., ["docker", "aws"])
```

---

## 🔧 Advanced Usage

### Environment Variables

Each environment can have custom variables:

```bash
# Edit environment variables
vim ~/.envswitch/environments/work/env-vars.env

# Add variables:
# AWS_REGION=us-east-1
# DEBUG=true
# API_URL=https://api.company.com
```

Variables are automatically loaded when switching.

### Hooks

Run commands before/after switching:

```yaml
# In environment metadata.yaml
hooks:
  pre_switch:
    - command: "echo 'Switching to work...'"
  post_switch:
    - command: "kubectl get nodes"
      verify: true
```

### Diff Environments

```bash
# Compare current state with environment
envswitch diff work

# Shows:
# Modified:
#   gcloud.account: personal@gmail.com → user@company.com
#   kubectl.context: minikube → gke-company-cluster
```

---

## 🎓 Examples

### Multi-Client Consulting

```bash
# Setup
envswitch create clientA --from-current
envswitch create clientB --from-current
envswitch create clientC --from-current

# Daily work
envswitch switch clientA   # Morning meeting
envswitch switch clientB   # Afternoon development
envswitch switch clientC   # Code review
```

### Work vs Personal

```bash
# Work hours
envswitch switch work

# After hours
envswitch switch personal
```

### Production vs Staging vs Dev

```bash
envswitch create prod --from-current
envswitch create staging --from prod
envswitch create dev --empty

# Safe switching
envswitch switch prod --verify
```

---

## 🚧 Development Status

**Current Version:** `0.1.0-alpha`

This project is in **early development**. Core features are being implemented.

**What Works:**

- ✅ Environment creation
- ✅ Environment listing & detailed view
- ✅ Environment deletion with archives
- ✅ Environment switching with loading spinner
- ✅ Tool snapshot capture (gcloud, kubectl, aws, docker, git)
- ✅ Backup system with retention policy
- ✅ Environment variables capture/restore
- ✅ Configuration system
- ✅ History tracking with detailed view
- ✅ Import/Export environments
- ✅ Shell integration (bash, zsh, fish)
- ✅ Auto-completion
- ✅ Hooks system (pre/post switch)
- ✅ Verbose mode for detailed logging

**Planned:**

- 📅 TUI (Terminal UI)
- 📅 Plugin system
- 📅 Encryption support
- 📅 Git sync
- 📅 Diff functionality

See [PROJECT_STATUS.md](PROJECT_STATUS.md) for detailed roadmap.

---

## 🤝 Contributing

We'd love your help! EnvSwitch is open source and welcoming contributors.

**Ways to contribute:**

- 🐛 Report bugs
- 💡 Suggest features
- 📝 Improve documentation
- 💻 Submit pull requests
- ⭐ Star the project

**Getting started:**

1. Read [CONTRIBUTING.md](CONTRIBUTING.md)
2. Check [good first issues](https://github.com/hugofrely/envswitch/labels/good%20first%20issue)
3. Read [GETTING_STARTED.md](GETTING_STARTED.md) for dev setup
4. Create plugins - see [Plugin Documentation](docs/PLUGINS.md)

**High-priority help needed:**

- Creating new tool plugins (Terraform, Ansible, Helm, etc.)
- Writing tests
- Documentation and examples
- TUI development
- Testing on different platforms

---

## 📚 Documentation

- **[Getting Started](GETTING_STARTED.md)** - Development setup
- **[Quick Start](docs/QUICKSTART.md)** - User guide
- **[Contributing](CONTRIBUTING.md)** - How to contribute
- **[Project Status](PROJECT_STATUS.md)** - Current progress

---

## 🙏 Acknowledgments

Inspired by:

- **nvm** - Node version management done right
- **rbenv** - Ruby environment management
- **direnv** - Directory-based environments
- **kubectl** - Kubernetes context switching

Built for developers tired of manual environment switching.

---

## 📄 License

[MIT License](LICENSE) - see LICENSE file for details.

---

## 💬 Support & Community

- **Documentation:** [README.md](README.md)
- **Issues:** [GitHub Issues](https://github.com/hugofrely/envswitch/issues)
- **Discussions:** [GitHub Discussions](https://github.com/hugofrely/envswitch/discussions)
- **Twitter:** [@hugofrely](https://twitter.com/hugofrely)

---

## ⚠️ Important Notice

**This is alpha software.** Not recommended for production use yet.

**Always backup your configurations** before using EnvSwitch:

```bash
# Backup your configs
cp -r ~/.config/gcloud ~/.config/gcloud.backup
cp -r ~/.kube ~/.kube.backup
cp -r ~/.aws ~/.aws.backup
```

---

## 🌟 Star History

If you find EnvSwitch useful, please consider starring the repository!

[![Star History Chart](https://api.star-history.com/svg?repos=hugofrely/envswitch&type=Date)](https://star-history.com/#hugofrely/envswitch&Date)

---

**Made with ❤️ by developers, for developers.**

**Stop manually switching environments. Start using EnvSwitch.** 🚀
