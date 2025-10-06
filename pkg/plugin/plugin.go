package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Plugin represents a plugin that extends envswitch functionality
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Description returns a short description
	Description() string

	// Initialize initializes the plugin
	Initialize() error

	// Snapshot creates a snapshot for this plugin's tool
	Snapshot(destPath string) error

	// Restore restores from a snapshot
	Restore(sourcePath string) error

	// Validate validates a snapshot
	Validate(snapshotPath string) error

	// IsInstalled checks if the tool is installed
	IsInstalled() bool

	// GetMetadata returns metadata about the current state
	GetMetadata() (map[string]interface{}, error)
}

// Metadata represents plugin metadata
type Metadata struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Homepage    string   `yaml:"homepage,omitempty"`
	License     string   `yaml:"license,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	ToolName    string   `yaml:"tool_name"` // The tool this plugin supports
}

// Manifest represents the plugin manifest file
type Manifest struct {
	Metadata Metadata `yaml:"metadata"`
	// Future: add hooks, dependencies, etc.
}

// LoadManifest loads a plugin manifest from a file
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Validate required fields
	if manifest.Metadata.Name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}
	if manifest.Metadata.Version == "" {
		return nil, fmt.Errorf("plugin version is required")
	}
	if manifest.Metadata.ToolName == "" {
		return nil, fmt.Errorf("tool_name is required")
	}

	return &manifest, nil
}

// GetPluginsDir returns the plugins directory path
func GetPluginsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	pluginsDir := filepath.Join(home, ".envswitch", "plugins")
	return pluginsDir, nil
}

// ListInstalledPlugins lists all installed plugins
func ListInstalledPlugins() ([]*Manifest, error) {
	pluginsDir, err := GetPluginsDir()
	if err != nil {
		return nil, err
	}

	// Check if plugins directory exists
	if _, statErr := os.Stat(pluginsDir); os.IsNotExist(statErr) {
		return []*Manifest{}, nil
	}

	// Read plugins directory
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugins directory: %w", err)
	}

	var plugins []*Manifest
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Look for plugin.yaml in the directory
		manifestPath := filepath.Join(pluginsDir, entry.Name(), "plugin.yaml")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		manifest, err := LoadManifest(manifestPath)
		if err != nil {
			fmt.Printf("Warning: Failed to load plugin '%s': %v\n", entry.Name(), err)
			continue
		}

		plugins = append(plugins, manifest)
	}

	return plugins, nil
}

// IsPluginInstalled checks if a plugin is installed
func IsPluginInstalled(pluginName string) (bool, error) {
	pluginsDir, err := GetPluginsDir()
	if err != nil {
		return false, err
	}

	pluginDir := filepath.Join(pluginsDir, pluginName)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")

	_, err = os.Stat(manifestPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// RemovePlugin removes an installed plugin
func RemovePlugin(pluginName string) error {
	pluginsDir, err := GetPluginsDir()
	if err != nil {
		return err
	}

	pluginDir := filepath.Join(pluginsDir, pluginName)

	// Check if plugin exists
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin '%s' is not installed", pluginName)
	}

	// Remove plugin directory
	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("failed to remove plugin: %w", err)
	}

	return nil
}
