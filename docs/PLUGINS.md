# EnvSwitch Plugin System

The EnvSwitch plugin system allows you to extend functionality by adding support for additional tools beyond the built-in ones (gcloud, kubectl, aws, docker, git).

## Table of Contents

- [Overview](#overview)
- [Plugin Architecture](#plugin-architecture)
- [Creating a Plugin](#creating-a-plugin)
- [Plugin Manifest](#plugin-manifest)
- [Implementing the Plugin Interface](#implementing-the-plugin-interface)
- [Installing and Managing Plugins](#installing-and-managing-plugins)
- [Example: Terraform Plugin](#example-terraform-plugin)
- [Best Practices](#best-practices)
- [Testing Your Plugin](#testing-your-plugin)

---

## Overview

Plugins allow you to:

- Add support for new tools (e.g., Terraform, Ansible, Helm)
- Capture and restore tool-specific configurations
- Integrate with EnvSwitch's environment switching workflow
- Share custom tool integrations with the community

**Key Concepts:**

- **Plugin**: A directory containing a manifest and implementation
- **Manifest**: A `plugin.yaml` file describing the plugin
- **Tool Interface**: The Go interface plugins must implement
- **Snapshot/Restore**: Core operations for capturing and restoring tool state

---

## Plugin Architecture

```
~/.envswitch/
├── plugins/
│   ├── terraform/
│   │   ├── plugin.yaml       # Plugin manifest
│   │   └── ...               # Plugin files
│   ├── ansible/
│   │   ├── plugin.yaml
│   │   └── ...
│   └── custom-tool/
│       ├── plugin.yaml
│       └── ...
```

**Plugin Lifecycle:**

1. **Installation**: Copy plugin to `~/.envswitch/plugins/<plugin-name>`
2. **Discovery**: EnvSwitch reads `plugin.yaml` to understand the plugin
3. **Initialization**: Plugin is loaded when the tool is needed
4. **Snapshot**: Capture tool state during environment save
5. **Restore**: Restore tool state during environment switch

---

## Creating a Plugin

### Step 1: Create Plugin Directory

```bash
mkdir my-plugin
cd my-plugin
```

### Step 2: Create Plugin Manifest

Create `plugin.yaml`:

```yaml
metadata:
  name: my-tool
  version: 1.0.0
  description: Support for My Tool
  author: Your Name
  homepage: https://github.com/yourusername/envswitch-my-tool
  license: MIT
  tool_name: my-tool
  tags:
    - devops
    - cloud
```

### Step 3: Implement Plugin Logic

Plugins must implement the `Plugin` interface defined in `pkg/plugin/plugin.go`:

```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Initialize() error
    Snapshot(destPath string) error
    Restore(sourcePath string) error
    Validate(snapshotPath string) error
    IsInstalled() bool
    GetMetadata() (map[string]interface{}, error)
}
```

**Alternative Approach (Simpler):**

Plugins can also implement the `Tool` interface from `pkg/tools/tool.go`, which is simpler:

```go
type Tool interface {
    Name() string
    IsInstalled() bool
    Snapshot(snapshotPath string) error
    Restore(snapshotPath string) error
    GetMetadata() (map[string]interface{}, error)
    ValidateSnapshot(snapshotPath string) error
    Diff(snapshotPath string) ([]Change, error)
}
```

---

## Plugin Manifest

The `plugin.yaml` file contains metadata about your plugin.

### Required Fields

```yaml
metadata:
  name: my-plugin           # Unique plugin identifier (lowercase, hyphens)
  version: 1.0.0           # Semantic version
  tool_name: my-tool       # Name of the tool this plugin supports
```

### Optional Fields

```yaml
metadata:
  description: Brief description of what this plugin does
  author: Your Name or Organization
  homepage: https://github.com/username/plugin-repo
  license: MIT
  tags:
    - cloud
    - infrastructure
    - deployment
```

### Example: Terraform Plugin Manifest

```yaml
metadata:
  name: terraform
  version: 1.0.0
  description: Terraform workspace and state management
  author: EnvSwitch Community
  homepage: https://github.com/envswitch/terraform-plugin
  license: MIT
  tool_name: terraform
  tags:
    - terraform
    - infrastructure
    - iac
```

---

## Implementing the Plugin Interface

Here's a complete example of a Terraform plugin implementation:

### terraform_plugin.go

```go
package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "io"
)

type TerraformPlugin struct {
    ConfigDir string // ~/.terraform.d
}

func NewTerraformPlugin() *TerraformPlugin {
    home, _ := os.UserHomeDir()
    return &TerraformPlugin{
        ConfigDir: filepath.Join(home, ".terraform.d"),
    }
}

func (t *TerraformPlugin) Name() string {
    return "terraform"
}

func (t *TerraformPlugin) Version() string {
    return "1.0.0"
}

func (t *TerraformPlugin) Description() string {
    return "Terraform workspace and state management"
}

func (t *TerraformPlugin) Initialize() error {
    // Check if Terraform is installed
    if !t.IsInstalled() {
        return fmt.Errorf("terraform is not installed")
    }
    return nil
}

func (t *TerraformPlugin) IsInstalled() bool {
    _, err := exec.LookPath("terraform")
    return err == nil
}

func (t *TerraformPlugin) Snapshot(destPath string) error {
    // Create destination directory
    if err := os.MkdirAll(destPath, 0755); err != nil {
        return fmt.Errorf("failed to create snapshot directory: %w", err)
    }

    // Copy ~/.terraform.d directory
    if _, err := os.Stat(t.ConfigDir); !os.IsNotExist(err) {
        if err := copyDir(t.ConfigDir, filepath.Join(destPath, ".terraform.d")); err != nil {
            return fmt.Errorf("failed to copy terraform config: %w", err)
        }
    }

    // Get current workspace (if in a terraform directory)
    workspace, _ := t.getCurrentWorkspace()
    if workspace != "" {
        metadataFile := filepath.Join(destPath, "workspace.txt")
        if err := os.WriteFile(metadataFile, []byte(workspace), 0644); err != nil {
            return fmt.Errorf("failed to save workspace: %w", err)
        }
    }

    return nil
}

func (t *TerraformPlugin) Restore(sourcePath string) error {
    // Restore ~/.terraform.d
    terraformDir := filepath.Join(sourcePath, ".terraform.d")
    if _, err := os.Stat(terraformDir); !os.IsNotExist(err) {
        // Remove existing config
        if err := os.RemoveAll(t.ConfigDir); err != nil {
            return fmt.Errorf("failed to remove existing config: %w", err)
        }

        // Copy from snapshot
        if err := copyDir(terraformDir, t.ConfigDir); err != nil {
            return fmt.Errorf("failed to restore terraform config: %w", err)
        }
    }

    // Restore workspace (if saved)
    workspaceFile := filepath.Join(sourcePath, "workspace.txt")
    if _, err := os.Stat(workspaceFile); !os.IsNotExist(err) {
        workspace, err := os.ReadFile(workspaceFile)
        if err != nil {
            return fmt.Errorf("failed to read workspace: %w", err)
        }
        // Note: Workspace switching would need to be done in the actual terraform directory
        fmt.Printf("Workspace: %s\n", string(workspace))
    }

    return nil
}

func (t *TerraformPlugin) Validate(snapshotPath string) error {
    // Check if snapshot directory exists
    if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
        return fmt.Errorf("snapshot path does not exist: %s", snapshotPath)
    }
    return nil
}

func (t *TerraformPlugin) GetMetadata() (map[string]interface{}, error) {
    metadata := make(map[string]interface{})

    // Get terraform version
    cmd := exec.Command("terraform", "version", "-json")
    output, err := cmd.Output()
    if err == nil {
        metadata["version_output"] = string(output)
    }

    // Get current workspace
    workspace, _ := t.getCurrentWorkspace()
    if workspace != "" {
        metadata["workspace"] = workspace
    }

    return metadata, nil
}

func (t *TerraformPlugin) getCurrentWorkspace() (string, error) {
    cmd := exec.Command("terraform", "workspace", "show")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return string(output), nil
}

// Helper function to copy directories
func copyDir(src, dst string) error {
    return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        relPath, err := filepath.Rel(src, path)
        if err != nil {
            return err
        }

        targetPath := filepath.Join(dst, relPath)

        if info.IsDir() {
            return os.MkdirAll(targetPath, info.Mode())
        }

        return copyFile(path, targetPath)
    })
}

func copyFile(src, dst string) error {
    sourceFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer sourceFile.Close()

    destFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destFile.Close()

    _, err = io.Copy(destFile, sourceFile)
    return err
}
```

---

## Installing and Managing Plugins

### List Installed Plugins

```bash
envswitch plugin list
```

Output:
```
Installed plugins:

  • terraform v1.0.0
    Terraform workspace and state management
    Tool: terraform

  • ansible v1.0.0
    Ansible inventory and configuration management
    Tool: ansible

Total: 2 plugin(s)
```

### Install a Plugin

```bash
# Install from a local directory
envswitch plugin install ./terraform-plugin

# Install from a downloaded archive
envswitch plugin install ~/downloads/terraform-plugin.tar.gz
```

Output:
```
✅ Plugin 'terraform' v1.0.0 installed successfully
   Terraform workspace and state management
```

### View Plugin Info

```bash
envswitch plugin info terraform
```

Output:
```
Plugin: terraform
Version: 1.0.0
Description: Terraform workspace and state management
Author: EnvSwitch Community
Homepage: https://github.com/envswitch/terraform-plugin
License: MIT
Tool: terraform
Tags: [terraform infrastructure iac]
```

### Remove a Plugin

```bash
envswitch plugin remove terraform
# or
envswitch plugin rm terraform
envswitch plugin uninstall terraform
```

Output:
```
✅ Plugin 'terraform' removed successfully
```

---

## Example: Terraform Plugin

Here's the complete directory structure for a Terraform plugin:

```
terraform-plugin/
├── plugin.yaml
├── terraform_plugin.go
├── go.mod
├── go.sum
└── README.md
```

### plugin.yaml

```yaml
metadata:
  name: terraform
  version: 1.0.0
  description: Terraform workspace and state management
  author: EnvSwitch Community
  homepage: https://github.com/envswitch/terraform-plugin
  license: MIT
  tool_name: terraform
  tags:
    - terraform
    - infrastructure
    - iac
```

### What Gets Captured

The Terraform plugin captures:

1. **~/.terraform.d/** - Terraform CLI configuration
2. **Current workspace** - Active Terraform workspace name
3. **Plugin cache** - Downloaded provider plugins (optional)

### What Gets Restored

When switching environments:

1. Terraform configuration is restored
2. Workspace information is displayed
3. Plugin cache is restored (if captured)

---

## Best Practices

### 1. Configuration Files

- Capture tool-specific configuration directories
- Examples: `~/.terraform.d`, `~/.ansible`, `~/.helm`

### 2. State Files

- Be careful with state files (they may contain sensitive data)
- Consider excluding state files or encrypting snapshots

### 3. Credentials

- Never store plaintext credentials in snapshots
- Use environment variables or credential managers
- Consider using EnvSwitch's planned encryption feature

### 4. Error Handling

```go
func (p *MyPlugin) Snapshot(destPath string) error {
    if !p.IsInstalled() {
        return fmt.Errorf("tool not installed")
    }

    if err := os.MkdirAll(destPath, 0755); err != nil {
        return fmt.Errorf("failed to create snapshot dir: %w", err)
    }

    // ... snapshot logic
    return nil
}
```

### 5. Validation

Always validate snapshots before restoring:

```go
func (p *MyPlugin) Validate(snapshotPath string) error {
    if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
        return fmt.Errorf("snapshot does not exist")
    }

    // Check for required files
    requiredFiles := []string{"config.yaml", "credentials"}
    for _, file := range requiredFiles {
        path := filepath.Join(snapshotPath, file)
        if _, err := os.Stat(path); os.IsNotExist(err) {
            return fmt.Errorf("missing required file: %s", file)
        }
    }

    return nil
}
```

### 6. Metadata

Provide useful metadata for debugging:

```go
func (p *MyPlugin) GetMetadata() (map[string]interface{}, error) {
    return map[string]interface{}{
        "version":      p.getToolVersion(),
        "config_path":  p.ConfigPath,
        "last_updated": time.Now().Format(time.RFC3339),
    }, nil
}
```

---

## Testing Your Plugin

### Manual Testing

1. **Create test environment:**
```bash
envswitch create test-plugin --empty
```

2. **Configure your tool:**
```bash
# Set up your tool's configuration
my-tool config set ...
```

3. **Enable plugin in environment:**
```bash
vim ~/.envswitch/environments/test-plugin/metadata.yaml
```

Add:
```yaml
tools:
  my-tool:
    enabled: true
    snapshot_path: ""
```

4. **Test snapshot:**
```bash
envswitch switch test-plugin
```

5. **Verify snapshot was created:**
```bash
ls ~/.envswitch/environments/test-plugin/snapshots/my-tool/
```

6. **Test restore:**
```bash
# Modify tool configuration
my-tool config set something different

# Switch back
envswitch switch test-plugin

# Verify configuration was restored
my-tool config get
```

### Automated Testing

Create a test file `plugin_test.go`:

```go
package main

import (
    "os"
    "path/filepath"
    "testing"
)

func TestTerraformPlugin_Snapshot(t *testing.T) {
    plugin := NewTerraformPlugin()

    tempDir := t.TempDir()
    snapshotPath := filepath.Join(tempDir, "snapshot")

    err := plugin.Snapshot(snapshotPath)
    if err != nil {
        t.Fatalf("Snapshot failed: %v", err)
    }

    // Verify snapshot was created
    if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
        t.Error("Snapshot directory was not created")
    }
}

func TestTerraformPlugin_Restore(t *testing.T) {
    plugin := NewTerraformPlugin()

    tempDir := t.TempDir()
    snapshotPath := filepath.Join(tempDir, "snapshot")

    // Create snapshot first
    err := plugin.Snapshot(snapshotPath)
    if err != nil {
        t.Fatalf("Snapshot failed: %v", err)
    }

    // Test restore
    err = plugin.Restore(snapshotPath)
    if err != nil {
        t.Fatalf("Restore failed: %v", err)
    }
}

func TestTerraformPlugin_IsInstalled(t *testing.T) {
    plugin := NewTerraformPlugin()

    // This will depend on your test environment
    installed := plugin.IsInstalled()
    t.Logf("Terraform installed: %v", installed)
}
```

Run tests:
```bash
go test -v
```

---

## Plugin Distribution

### GitHub Repository

Create a repository for your plugin:

```
terraform-plugin/
├── plugin.yaml
├── terraform_plugin.go
├── go.mod
├── go.sum
├── README.md
├── LICENSE
└── examples/
    └── config.yaml
```

### Installation Instructions

In your README.md:

```markdown
# EnvSwitch Terraform Plugin

Support for Terraform workspace and state management in EnvSwitch.

## Installation

git clone https://github.com/yourusername/envswitch-terraform
cd envswitch-terraform
envswitch plugin install .

## Usage

The plugin automatically captures:
- Terraform CLI configuration (~/.terraform.d)
- Current workspace information
- Plugin cache

## Requirements

- EnvSwitch v0.1.0 or later
- Terraform CLI installed
```

---

## Future Enhancements

The plugin system is evolving. Planned features include:

- **Hot reloading**: Load plugins without restarting
- **Plugin marketplace**: Central repository of community plugins
- **Binary plugins**: Support for compiled binary plugins
- **Hook support**: Pre/post snapshot and restore hooks
- **Dependencies**: Plugins can depend on other plugins
- **Configuration UI**: Manage plugin settings via CLI

---

## Contributing Plugins

We encourage community contributions! To share your plugin:

1. Create a GitHub repository for your plugin
2. Follow the structure and best practices outlined above
3. Add comprehensive documentation
4. Submit to the EnvSwitch plugin registry (coming soon)

---

## Getting Help

- **Documentation**: [EnvSwitch README](../README.md)
- **Issues**: [GitHub Issues](https://github.com/hugofrely/envswitch/issues)
- **Discussions**: [GitHub Discussions](https://github.com/hugofrely/envswitch/discussions)

---

## Examples

See the [examples directory](../examples/plugins/) for complete plugin examples including:

- Terraform plugin
- Ansible plugin
- Helm plugin
- Custom tool plugin template
