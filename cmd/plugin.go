package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/hugofrely/envswitch/pkg/environment"
	"github.com/hugofrely/envswitch/pkg/plugin"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage plugins",
	Long: `Manage envswitch plugins.

Plugins extend envswitch functionality by adding support for additional tools,
custom integrations, and advanced features.

Available commands:
  list      List installed plugins
  install   Install a plugin
  remove    Remove a plugin
  info      Show plugin information`,
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed plugins",
	Long:  `List all installed plugins with their versions and descriptions.`,
	RunE:  runPluginList,
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install <path-to-plugin>",
	Short: "Install a plugin",
	Long: `Install a plugin from a local directory or archive.

The plugin must contain a plugin.yaml manifest file.

Examples:
  # Install from a directory
  envswitch plugin install ./my-plugin

  # Install from a downloaded archive
  envswitch plugin install ~/downloads/terraform-plugin.tar.gz`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginInstall,
}

var pluginRemoveCmd = &cobra.Command{
	Use:     "remove <plugin-name>",
	Aliases: []string{"rm", "uninstall"},
	Short:   "Remove an installed plugin",
	Long:    `Remove an installed plugin by name.`,
	Args:    cobra.ExactArgs(1),
	RunE:    runPluginRemove,
}

var pluginInfoCmd = &cobra.Command{
	Use:   "info <plugin-name>",
	Short: "Show plugin information",
	Long:  `Display detailed information about an installed plugin.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginInfo,
}

func init() {
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInstallCmd)
	pluginCmd.AddCommand(pluginRemoveCmd)
	pluginCmd.AddCommand(pluginInfoCmd)
}

func runPluginList(cmd *cobra.Command, args []string) error {
	plugins, err := plugin.ListInstalledPlugins()
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	if len(plugins) == 0 {
		fmt.Println("No plugins installed.")
		fmt.Println()
		fmt.Println("Install plugins with: envswitch plugin install <path>")
		return nil
	}

	fmt.Println("Installed plugins:")
	fmt.Println()

	for _, p := range plugins {
		fmt.Printf("  • %s v%s\n", p.Metadata.Name, p.Metadata.Version)
		if p.Metadata.Description != "" {
			fmt.Printf("    %s\n", p.Metadata.Description)
		}
		if p.Metadata.ToolName != "" {
			fmt.Printf("    Tool: %s\n", p.Metadata.ToolName)
		}
		fmt.Println()
	}

	fmt.Printf("Total: %d plugin(s)\n", len(plugins))
	return nil
}

func runPluginInstall(cmd *cobra.Command, args []string) error {
	sourcePath := args[0]

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("plugin path not found: %s", sourcePath)
	}

	// Check for plugin.yaml manifest
	manifestPath := filepath.Join(sourcePath, "plugin.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin.yaml not found in %s", sourcePath)
	}

	// Load manifest
	manifest, err := plugin.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest: %w", err)
	}

	// Check if plugin already installed
	installed, err := plugin.IsPluginInstalled(manifest.Metadata.Name)
	if err != nil {
		return fmt.Errorf("failed to check if plugin is installed: %w", err)
	}

	if installed {
		return fmt.Errorf("plugin '%s' is already installed (remove it first)", manifest.Metadata.Name)
	}

	// Get plugins directory
	pluginsDir, err := plugin.GetPluginsDir()
	if err != nil {
		return err
	}

	// Create plugins directory if it doesn't exist
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugins directory: %w", err)
	}

	// Copy plugin to plugins directory
	destPath := filepath.Join(pluginsDir, manifest.Metadata.Name)
	if err := copyDir(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	fmt.Printf("✅ Plugin '%s' v%s installed successfully\n", manifest.Metadata.Name, manifest.Metadata.Version)
	if manifest.Metadata.Description != "" {
		fmt.Printf("   %s\n", manifest.Metadata.Description)
	}

	// Synchroniser le plugin avec tous les environnements existants
	fmt.Println("🔄 Syncing plugin to existing environments...")
	if err := environment.SyncPluginsToEnvironments(); err != nil {
		fmt.Printf("⚠️  Warning: Failed to sync plugin to environments: %v\n", err)
	} else {
		fmt.Println("✅ Plugin enabled in all environments")
	}

	return nil
}

func runPluginRemove(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	// Check if plugin is installed
	installed, err := plugin.IsPluginInstalled(pluginName)
	if err != nil {
		return fmt.Errorf("failed to check plugin: %w", err)
	}

	if !installed {
		return fmt.Errorf("plugin '%s' is not installed", pluginName)
	}

	// Remove plugin
	if err := plugin.RemovePlugin(pluginName); err != nil {
		return err
	}

	fmt.Printf("✅ Plugin '%s' removed successfully\n", pluginName)
	return nil
}

func runPluginInfo(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	// Get plugins directory
	pluginsDir, err := plugin.GetPluginsDir()
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(pluginsDir, pluginName, "plugin.yaml")

	// Check if plugin exists
	if _, statErr := os.Stat(manifestPath); os.IsNotExist(statErr) {
		return fmt.Errorf("plugin '%s' is not installed", pluginName)
	}

	// Load manifest
	manifest, err := plugin.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	// Display info
	fmt.Printf("Plugin: %s\n", manifest.Metadata.Name)
	fmt.Printf("Version: %s\n", manifest.Metadata.Version)
	if manifest.Metadata.Description != "" {
		fmt.Printf("Description: %s\n", manifest.Metadata.Description)
	}
	if manifest.Metadata.Author != "" {
		fmt.Printf("Author: %s\n", manifest.Metadata.Author)
	}
	if manifest.Metadata.Homepage != "" {
		fmt.Printf("Homepage: %s\n", manifest.Metadata.Homepage)
	}
	if manifest.Metadata.License != "" {
		fmt.Printf("License: %s\n", manifest.Metadata.License)
	}
	if manifest.Metadata.ToolName != "" {
		fmt.Printf("Tool: %s\n", manifest.Metadata.ToolName)
	}
	if len(manifest.Metadata.Tags) > 0 {
		fmt.Printf("Tags: %v\n", manifest.Metadata.Tags)
	}

	return nil
}

// copyDir recursively copies a directory (helper function)
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, info.Mode())
	})
}

// Example plugin.yaml structure for documentation
var examplePluginManifest = `
metadata:
  name: terraform
  version: 1.0.0
  description: Terraform workspace and state management
  author: Your Name
  homepage: https://github.com/yourusername/envswitch-terraform
  license: MIT
  tool_name: terraform
  tags:
    - terraform
    - infrastructure
    - iac
`

func init() {
	_ = examplePluginManifest // For documentation
}
