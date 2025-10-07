# EnvSwitch Plugin Guide

Create plugins to add support for any tool in 2 minutes - **no code required!**

## Quick Start: Create a Plugin in 2 Minutes

### 1. Create Plugin Directory

```bash
mkdir my-plugin
cd my-plugin
```

### 2. Create `plugin.yaml`

```yaml
metadata:
  name: npm
  version: 1.0.0
  description: NPM registry and authentication
  tool_name: npm
  author: Your Name
  config_path: $HOME/.npmrc
```

### 3. Install

```bash
envswitch plugin install .
```

**That's it!** The plugin:
- ✅ Is installed in `~/.envswitch/plugins/npm/`
- ✅ Is automatically activated in ALL environments
- ✅ Captures `~/.npmrc` during every switch

## Configuration Options

You have **three options** for specifying config paths:

### Option 1: Auto-detection (Simple)

Let EnvSwitch automatically detect the config path:

```yaml
metadata:
  tool_name: npm
```

EnvSwitch will:
1. Check if `~/.npm/` exists (directory) → use it
2. Otherwise → use `~/.npmrc` (file)

**Best for**: Standard tools following `~/.TOOLNAME` or `~/.TOOLNAMErc` convention

### Option 2: Single Path (Recommended)

Specify an explicit path:

```yaml
metadata:
  tool_name: npm
  config_path: $HOME/.npmrc
```

You can use:
- Environment variables: `$HOME`, `$XDG_CONFIG_HOME`
- Absolute paths: `/etc/hosts`
- Paths outside `$HOME/`: `/usr/local/etc/app.conf`

**Best for**: Most use cases where you want explicit control

### Option 3: Multiple Paths (Advanced)

Capture multiple files/directories:

```yaml
metadata:
  tool_name: vim
  config_paths:
    - $HOME/.vimrc
    - $HOME/.vim
```

**Best for**: Tools with configs in multiple locations

## How It Works

### Auto-Detection Flow

```
plugin.yaml (tool_name: "npm")
         ↓
Load installed plugin
         ↓
Auto-detect path: "npm" → "~/.npmrc"
         ↓
Create GenericTool with this path
         ↓
Snapshot: copy ~/.npmrc → snapshots/npm/
         ↓
Restore: copy snapshots/npm/ → ~/.npmrc
```

### Custom Path Flow

```
plugin.yaml (config_path: "/etc/hosts")
         ↓
Load installed plugin
         ↓
Use custom path: "/etc/hosts"
         ↓
Create GenericTool with this path
         ↓
Snapshot: copy /etc/hosts → snapshots/hosts/
         ↓
Restore: copy snapshots/hosts/ → /etc/hosts
```

### Multiple Paths Flow

```
plugin.yaml (config_paths: ["~/.vimrc", "~/.vim"])
         ↓
Load installed plugin
         ↓
Create MultiPathTool with these paths
         ↓
Snapshot: copy both ~/.vimrc and ~/.vim/
         ↓
Restore: restore both files/directories
```

## Example Plugins

### NPM Plugin

**File**: `examples/npm-plugin-example/`

```yaml
metadata:
  name: npm
  version: 1.0.0
  description: NPM registry and authentication management
  tool_name: npm
  config_path: $HOME/.npmrc
```

Captures: `~/.npmrc`

**Use case**: Switch between company and public npm registries

### Vim Plugin

**File**: `examples/vim-plugin-example/`

```yaml
metadata:
  name: vim
  version: 1.0.0
  description: Vim editor configuration and plugins
  tool_name: vim
  config_paths:
    - $HOME/.vimrc
    - $HOME/.vim
```

Captures: Both `~/.vimrc` AND `~/.vim/`

**Use case**: Different vim configs for different projects

### Hosts Plugin

**File**: `examples/hosts-plugin-example/`

```yaml
metadata:
  name: hosts
  version: 1.0.0
  description: System hosts file (/etc/hosts)
  tool_name: hosts
  config_path: /etc/hosts
```

Captures: `/etc/hosts` (system file outside `$HOME/`)

**Use case**: Switch between development and production DNS entries

## Complete Workflow

### 1. Install Plugin

```bash
envswitch plugin install ./npm-plugin
```

Output:
```
✅ Plugin 'npm' v1.0.0 installed successfully
   NPM registry and authentication
🔄 Syncing plugin to existing environments...
✅ Plugin enabled in all environments
```

### 2. Configure Tool

```bash
npm config set registry https://npm.mycompany.com
npm login
```

### 3. Create Environment

```bash
envswitch create work --from-current
```

The NPM config (`~/.npmrc`) is automatically captured!

### 4. Switch Between Environments

```bash
# Change config
npm config set registry https://registry.npmjs.org/
npm login

# Create another environment
envswitch create personal --from-current

# Switch easily
envswitch switch work      # → Company config
envswitch switch personal  # → Personal config
```

## Real-World Use Cases

### Multiple Clients

```bash
# Client A
npm config set registry https://npm.clientA.com
npm login
envswitch create clientA --from-current

# Client B
npm config set registry https://npm.clientB.com
npm login
envswitch create clientB --from-current

# Switch instantly
envswitch switch clientA  # All Client A config
envswitch switch clientB  # All Client B config
```

### Work vs Personal

```bash
# Work (with proxy)
npm config set registry https://npm.company.com
npm config set proxy http://proxy:8080
npm login
envswitch create work --from-current

# Personal (no proxy)
npm config set registry https://registry.npmjs.org/
npm config delete proxy
npm login
envswitch create personal --from-current

# Switch
envswitch switch work      # Company config + proxy
envswitch switch personal  # Personal config
```

## Plugin Management

### List Plugins

```bash
envswitch plugin list
```

Output:
```
Installed plugins:

  • npm v1.0.0
    NPM registry and authentication
    Tool: npm

  • vim v1.0.0
    Vim editor configuration
    Tool: vim

Total: 2 plugin(s)
```

### Show Plugin Info

```bash
envswitch plugin info npm
```

Output:
```
Plugin: npm
Version: 1.0.0
Description: NPM registry and authentication
Tool: npm
```

### Remove Plugin

```bash
envswitch plugin remove npm
```

## Testing Your Plugin

```bash
# 1. Create test environment
envswitch create test-plugin --from-current

# 2. Verify snapshot was created
ls -la ~/.envswitch/environments/test-plugin/snapshots/TOOLNAME/

# 3. Modify tool config
# (use your tool's config commands)

# 4. Switch to restore
envswitch switch test-plugin

# 5. Verify config was restored
# (check your tool's config)
```

## Debugging

### Verify Snapshots

```bash
# List snapshots
ls -la ~/.envswitch/environments/*/snapshots/npm/

# View snapshot content
cat ~/.envswitch/environments/work/snapshots/npm/.npmrc
```

### Verbose Mode

```bash
envswitch switch work --verbose
```

Output:
```
[DEBUG] Using custom config path for 'npm': /Users/you/.npmrc
[DEBUG] Loaded plugin 'npm' for tool 'npm'
[DEBUG] Snapshotting npm...
[DEBUG] Restoring npm...
✅ Successfully switched to 'work' (1.02s)
```

### Common Issues

**Plugin not activated:**

Reinstall it - plugins auto-activate on install:
```bash
envswitch plugin install ./my-plugin
```

**Config not captured:**

Check that the config file exists:
```bash
ls -la ~/.npmrc  # For NPM example
```

**Config in non-standard location:**

Use the `config_path` field in `plugin.yaml`:
```yaml
config_path: /custom/path/to/config
```

## When Do You Need Go Code?

**You DON'T need Go code for**:
- ✅ Single file/directory in `$HOME/`
- ✅ Single file/directory outside `$HOME/` (use `config_path`)
- ✅ Multiple files/directories (use `config_paths`)
- ✅ Environment variable expansion
- ✅ 95% of all tools!

**You ONLY need Go code for**:
- Complex logic (parsing, transforming configs)
- Running commands before/after snapshot
- Conditional behavior based on system state
- Custom validation or verification

For 95% of tools, **YAML is enough**!

## Distributing Your Plugin

### On GitHub

```bash
# Create repo
git init
git add plugin.yaml README.md
git commit -m "Initial commit"
git remote add origin https://github.com/YOU/plugin-name
git push -u origin main
```

### Install from GitHub

```bash
git clone https://github.com/YOU/plugin-name
envswitch plugin install ./plugin-name
```

## Suggested Plugins

Help the community by creating these plugins:

### Package Managers
- **yarn** - Node.js alternative
- **pnpm** - Performant Node.js
- **pip** - Python
- **poetry** - Modern Python
- **gem** - Ruby
- **cargo** - Rust
- **composer** - PHP
- **go** - Go modules

### Infrastructure
- **terraform** - Infrastructure as Code
- **terragrunt** - Terraform wrapper
- **ansible** - Configuration management
- **pulumi** - Modern IaC
- **helm** - Kubernetes packages

### Cloud
- **azure** - Azure CLI
- **doctl** - DigitalOcean CLI
- **heroku** - Heroku CLI
- **fly** - Fly.io CLI

### Dev Tools
- **gh** - GitHub CLI
- **git-credential** - Git credentials
- **ssh** - SSH config/keys
- **gpg** - GPG keys

## FAQ

**Q: Do I need to write code?**

No! A simple `plugin.yaml` file is enough for 95% of cases.

**Q: Is the plugin automatically activated?**

Yes! Upon installation, it's activated in ALL existing environments.

**Q: Can I disable a plugin in one environment?**

Yes, edit `~/.envswitch/environments/NAME/metadata.yaml`:
```yaml
tools:
  npm:
    enabled: false
```

**Q: Where are snapshots stored?**

```
~/.envswitch/environments/ENV_NAME/snapshots/TOOL_NAME/
```

**Q: How do I share my plugin?**

Publish on GitHub and share the link. Users can:
```bash
git clone https://github.com/YOU/plugin
envswitch plugin install ./plugin
```

**Q: My tool doesn't use ~/.TOOLRC, what do I do?**

Use the `config_path` field:
```yaml
config_path: /custom/path/to/config
```

**Q: Can I capture multiple files?**

Yes! Use `config_paths`:
```yaml
config_paths:
  - $HOME/.vimrc
  - $HOME/.vim
```

**Q: Are credentials secure?**

Snapshots are in `~/.envswitch/` with standard Unix permissions. Use FileVault (macOS) or equivalent for encryption.

## Support

- **Main Documentation**: [README.md](../README.md)
- **Examples**: [examples/](../examples/)
- **Issues**: [GitHub Issues](https://github.com/hugofrely/envswitch/issues)
- **Discussions**: [GitHub Discussions](https://github.com/hugofrely/envswitch/discussions)

## Summary

1. Create `plugin.yaml` with `tool_name` and optional `config_path`/`config_paths`
2. Run `envswitch plugin install .`
3. **Done!** Plugin is active everywhere automatically

**No Go code needed for most plugins!**
