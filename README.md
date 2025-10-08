# EnvSwitch 🔄

[![CI](https://img.shields.io/github/actions/workflow/status/hugofrely/envswitch/ci.yml?logo=github)](https://github.com/hugofrely/envswitch/actions)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![Project Status](https://img.shields.io/badge/status-stable-green.svg)](PROJECT_STATUS.md)

EnvSwitch is a powerful CLI tool that captures, saves, and restores the complete state of your development environments. Switch instantly between work and personal projects, different client environments, or testing scenarios—without losing your authentication, configurations, or contexts.

![EnvSwitch Demo](demo.gif)

---

## 🎯 The Problem

As a developer, you probably work across multiple environments:

```bash
# Work - Morning
gcloud auth login work@company.com
gcloud config set project company-prod-123
kubectl config use-context gke-work-cluster

# Personal - Evening
gcloud auth login personal@gmail.com
gcloud config set project personal-project
kubectl config use-context minikube
```

**This is exhausting.** And error-prone. What if you forget to switch? Deploy to the wrong environment? Lose hours troubleshooting?

## 💡 The Solution

**EnvSwitch creates snapshots of your entire dev environment.**

```bash
# Setup your work environment
gcloud auth login work@company.com
envswitch create work --from-current    # Captures AND switches to 'work'

# Setup your personal environment
gcloud auth login personal@gmail.com
envswitch create personal --from-current # Captures AND switches to 'personal'

# Then switch instantly, anytime
envswitch switch work        # All your work configs restored
envswitch switch personal    # All your personal configs restored

# Save changes to active environment
envswitch save              # Updates current environment with latest changes
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

```bash
curl -fsSL https://raw.githubusercontent.com/hugofrely/envswitch/main/install.sh | bash
```

See [INSTALL.md](INSTALL.md) for other installation methods.

### First Steps

```bash
# Initialize
envswitch init

# Capture current state
envswitch create work --from-current

# Make changes and capture another
envswitch create personal --from-current

# Switch instantly
envswitch switch work
```

See [Quick Start Guide](docs/QUICKSTART.md) for detailed walkthrough.

---

## 📖 Usage

### Creating Environments

```bash
# Create from current system state (auto-switches to it)
envswitch create myenv --from-current

# Create empty environment
envswitch create myenv --empty

# Clone existing environment (auto-switches to it)
envswitch create dev --from prod

# With description
envswitch create staging --from-current \
    --description "Staging environment for testing"
```

### Saving Environment Changes

```bash
# Save current system state to the active environment
envswitch save

# This updates the active environment with any changes you've made
# (authentication, configurations, etc.)
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

### Plugin Management

```bash
# List installed plugins
envswitch plugin list

# Install plugin
envswitch plugin install ./my-plugin
```

**📖 Plugin Development**: Create plugins for any tool in 2 minutes—no code required! See [Plugin Documentation](docs/PLUGINS.md).

---

## 🛠️ Supported Tools

| Tool           | Status         | What's Captured                                                 |
| -------------- | -------------- | --------------------------------------------------------------- |
| **GCloud CLI** | ✅ Implemented | Authentication, active account, project, region, configurations |
| **Kubectl**    | ✅ Implemented | Contexts, clusters, current namespace, kubeconfig               |
| **AWS CLI**    | ✅ Implemented | Profiles, credentials, default region, config                   |
| **Docker**     | ✅ Implemented | Registry authentication, config.json                            |
| **Git**        | ✅ Implemented | User name, email, signing keys                                  |
| **Plugins**    | ✅ Implemented | Any tool via plugin system (npm, vim, terraform, etc.)          |

**All built-in tools are fully implemented!** ✅

Add support for additional tools using the [Plugin System](docs/PLUGINS.md) - no code required!

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

```bash
# View all config
envswitch config list

# Set config values
envswitch config set auto_save_before_switch false
envswitch config set log_level debug
```

See [Quick Start Guide](docs/QUICKSTART.md#configuration) for all configuration options.

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

---

## 🎓 Real-World Examples

### Example 1: Work vs Personal

```bash
# Setup work environment
gcloud auth login work@company.com
gcloud config set project company-prod-123
kubectl config use-context gke-company-cluster
envswitch create work --from-current    # Captures and switches to 'work'

# Setup personal environment
gcloud auth login personal@gmail.com
gcloud config set project my-side-project
kubectl config use-context minikube
envswitch create personal --from-current # Captures and switches to 'personal'

# Daily usage
envswitch switch work      # 9am - Start work
envswitch switch personal  # 6pm - Side projects

# If you make changes to your setup
envswitch save            # Save changes to active environment
```

### Example 2: Multi-Client Consulting

```bash
# Setup Client A environment
gcloud auth login consultant@clientA.com
aws configure  # Setup AWS for Client A
envswitch create clientA --from-current

# Setup Client B environment
gcloud auth login consultant@clientB.com
aws configure  # Setup AWS for Client B
envswitch create clientB --from-current

# Switch throughout the day
envswitch switch clientA   # Morning meeting
envswitch switch clientB   # Afternoon development
```

### Example 3: Production vs Staging

```bash
# Setup Production environment
gcloud auth login ops@company.com
gcloud config set project company-prod
kubectl config use-context production-cluster
envswitch create production --from-current

# Setup Staging environment
gcloud config set project company-staging
kubectl config use-context staging-cluster
envswitch create staging --from-current

# Safe switching with verification
envswitch switch production --verify
envswitch switch staging
```

---

## ✅ Production Ready

EnvSwitch is **production-ready** and stable! All core features are fully implemented and tested.

**Features:**

- ✅ Environment creation, listing, switching, and deletion
- ✅ Complete snapshot support for gcloud, kubectl, aws, docker, and git
- ✅ Automatic backup system with retention policy
- ✅ Environment variables capture and restore
- ✅ History tracking with rollback capability
- ✅ Import/Export for backup and sharing
- ✅ Shell integration with prompt indicators
- ✅ Auto-completion for bash, zsh, and fish
- ✅ Pre/post switch hooks for automation
- ✅ Plugin system for custom tool support
- ✅ Comprehensive configuration options
- ✅ Dry-run and verification modes

See [PROJECT_STATUS.md](PROJECT_STATUS.md) for details.

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

**Help wanted:**

- Creating new tool plugins (Terraform, Ansible, Helm, etc.)
- Writing tests
- Documentation and examples
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

## ⚠️ Best Practices

While EnvSwitch is production-ready, we recommend:

**Initial Setup:**

```bash
# Create a safety backup before first use
cp -r ~/.config/gcloud ~/.config/gcloud.backup
cp -r ~/.kube ~/.kube.backup
cp -r ~/.aws ~/.aws.backup
```

**Regular Backups:**

```bash
# Export your environments regularly
envswitch export --all --output ~/envswitch-backups
```

---

**Made with ❤️ by developers, for developers.**

**Stop manually switching environments. Start using EnvSwitch.** 🚀
