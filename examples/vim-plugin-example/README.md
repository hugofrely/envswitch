# Vim Plugin for EnvSwitch

Captures and restores Vim configuration across different environments.

**Pure YAML plugin - No Go code required!**

## What's Captured

- `~/.vimrc` - Vim configuration file  
- `~/.vim/` - Vim plugins directory

Both paths are captured using the `config_paths` feature.

## Installation

```bash
envswitch plugin install examples/vim-plugin-example
```

## How It Works

This plugin uses **multiple config paths**:

```yaml
metadata:
  name: vim
  tool_name: vim
  config_paths:
    - $HOME/.vimrc
    - $HOME/.vim
```

**Pure YAML, captures multiple paths!**

## Testing

```bash
# Create test vimrc
echo "set number" > ~/.vimrc
envswitch create test-vim --from-current

# Change vimrc
echo "set nonumber" > ~/.vimrc

# Switch back
envswitch switch test-vim
cat ~/.vimrc  # Should show: set number
```

## Support

- [Plugin Guide](../../docs/PLUGINS.md)
- [GitHub Issues](https://github.com/hugofrely/envswitch/issues)
