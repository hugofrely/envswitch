# EnvSwitch Quick Start Guide

This guide will help you get started with EnvSwitch in 5 minutes.

## Installation

### Option 1: Install from source (requires Go)

```bash
git clone https://github.com/hugofrely/envswitch.git
cd envswitch
make install
```

### Option 2: Download binary

Visit the [releases page](https://github.com/hugofrely/envswitch/releases) and download the binary for your platform.

### Option 3: Install script

```bash
curl -fsSL https://raw.githubusercontent.com/hugofrely/envswitch/main/install.sh | bash
```

## First Time Setup

### 1. Initialize EnvSwitch

```bash
envswitch init
```

This creates `~/.envswitch/` directory with default configuration.

### 2. Create Your First Environment

Capture your current development environment:

```bash
envswitch create work --from-current --description "Work environment"
```

This creates a snapshot of your current:
- GCloud authentication and configuration
- Kubectl contexts
- AWS credentials and profiles
- Docker registry authentication
- Git configuration

### 3. Make Changes and Create Another Environment

```bash
# Switch to personal GCloud account
gcloud auth login personal@gmail.com
gcloud config set project my-personal-project

# Create personal environment
envswitch create personal --from-current --description "Personal projects"
```

### 4. Switch Between Environments

Now you can instantly switch between your work and personal environments:

```bash
# Switch to work
envswitch switch work

# Switch to personal
envswitch switch personal

# Or use the shortcut
envswitch work
```

## Common Workflows

### List All Environments

```bash
envswitch list
# or
envswitch ls --detailed
```

### Show Environment Details

```bash
envswitch show work
```

### Create Environment from Another

```bash
envswitch create work-dev --from work --description "Development environment"
```

### Delete an Environment

```bash
envswitch delete old-env
# or with force
envswitch rm old-env --force
```

## What's Next?

- Read the [full documentation](./README.md)
- Learn about [hooks and automation](./HOOKS.md)
- Configure [shell integration](./SHELL_INTEGRATION.md)
- Explore [advanced features](./ADVANCED.md)

## Troubleshooting

### Environment not switching?

EnvSwitch is in early development. The full snapshot/restore functionality is being implemented. Currently:
- ✅ Environment creation works
- ✅ Environment listing works
- 🚧 Snapshot capture is in progress
- 🚧 Full switching logic is in progress

### Want to contribute?

See [CONTRIBUTING.md](../CONTRIBUTING.md) for how to help build EnvSwitch!

## Getting Help

- 📖 [Full Documentation](../README.md)
- 🐛 [Report Issues](https://github.com/hugofrely/envswitch/issues)
- 💬 [Discussions](https://github.com/hugofrely/envswitch/discussions)
