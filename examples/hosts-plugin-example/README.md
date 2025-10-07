# Hosts Plugin for EnvSwitch

Captures and restores the system `/etc/hosts` file across different environments.

**Pure YAML plugin - No Go code required!**

## What's Captured

- `/etc/hosts` - System hosts file for DNS resolution

## Installation

```bash
envswitch plugin install examples/hosts-plugin-example
```

## How It Works

This plugin uses the `config_path` field to specify a file **outside** `$HOME/`:

```yaml
metadata:
  name: hosts
  tool_name: hosts
  config_path: /etc/hosts  # System file!
```

**Pure YAML, works with system files!**

## Use Cases

### Development vs Production DNS

```bash
# Development - local services
cat /etc/hosts
# 127.0.0.1  api.local
# 127.0.0.1  db.local

envswitch create dev --from-current

# Production - different DNS
sudo vim /etc/hosts
# 10.0.1.50  api.company.com

envswitch create prod --from-current

# Switch between DNS configs
envswitch switch dev   # → Local DNS
envswitch switch prod  # → Production DNS
```

## Testing

```bash
# Create environment (captures /etc/hosts)
envswitch create test-hosts --from-current

# Check snapshot
cat ~/.envswitch/environments/test-hosts/snapshots/hosts/hosts
```

## Permissions Note

**Reading** `/etc/hosts` works fine - snapshots are created successfully.

**Writing** to `/etc/hosts` requires sudo privileges. The current implementation doesn't handle sudo automatically, so restore operations will read the snapshot but may fail to write.

### Workaround

Manually restore if needed:
```bash
sudo cp ~/.envswitch/environments/ENVNAME/snapshots/hosts/hosts /etc/hosts
```

## Other System Files

This same approach works for any system file:

```yaml
# Application configs
config_path: /usr/local/etc/myapp.conf

# System configs  
config_path: /etc/nginx/nginx.conf
```

## Support

- [Plugin Guide](../../docs/PLUGINS.md)
- [GitHub Issues](https://github.com/hugofrely/envswitch/issues)
