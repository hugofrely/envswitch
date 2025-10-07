# NPM Plugin for EnvSwitch

Captures and restores NPM configuration (registry, authentication) across different environments.

**Pure YAML plugin - No Go code required!**

## What's Captured

- `~/.npmrc` - NPM configuration file containing:
  - Registry URL
  - Authentication tokens
  - Proxy settings
  - Custom configurations

## Installation

```bash
envswitch plugin install examples/npm-plugin-example
```

The plugin will be automatically activated in all your environments.

## Use Cases

### Multiple NPM Registries

```bash
# Work - Company private registry
npm config set registry https://npm.company.com
npm login
envswitch create work --from-current

# Personal - Public npm registry
npm config set registry https://registry.npmjs.org
npm login
envswitch create personal --from-current

# Switch between registries instantly
envswitch switch work      # → Company registry
envswitch switch personal  # → Public registry
```

### Client-Specific NPM Configs

```bash
# Client A - with proxy
npm config set registry https://npm.clientA.com
npm config set proxy http://proxy.clientA.com:8080
npm login
envswitch create clientA --from-current

# Client B - no proxy
npm config set registry https://npm.clientB.com
npm config delete proxy
npm login
envswitch create clientB --from-current

# Switch instantly
envswitch switch clientA  # → Client A registry + proxy
envswitch switch clientB  # → Client B registry, no proxy
```

## How It Works

This plugin uses **only** a `plugin.yaml` file:

```yaml
metadata:
  name: npm
  version: 1.0.0
  description: NPM registry and authentication management
  tool_name: npm
  config_path: $HOME/.npmrc  # ← Explicit path
```

EnvSwitch automatically:
1. Reads the `config_path`
2. Creates a GenericTool
3. Snapshots `~/.npmrc` during switches
4. Restores it when needed

**No Go code, no compilation, just YAML!**

## Testing

```bash
# 1. Configure npm
npm config set registry https://npm.company.com

# 2. Create environment
envswitch create test-npm --from-current

# 3. Change config
npm config set registry https://registry.npmjs.org

# 4. Switch back
envswitch switch test-npm
npm config get registry  # Should show: https://npm.company.com ✅
```

## Verification

```bash
# Check snapshot
cat ~/.envswitch/environments/work/snapshots/npm/.npmrc

# Debug mode
envswitch switch work --verbose
```

## Security Note

⚠️ `~/.npmrc` may contain auth tokens. Snapshots are stored in `~/.envswitch/` with standard Unix permissions. Consider using disk encryption.

## Support

- **Plugin Guide**: [docs/PLUGINS.md](../../docs/PLUGINS.md)
- **Issues**: [GitHub Issues](https://github.com/hugofrely/envswitch/issues)
