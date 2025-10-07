package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadManifest(t *testing.T) {
	t.Run("loads valid manifest", func(t *testing.T) {
		// Create temp file with manifest
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "plugin.yaml")

		manifestContent := `
metadata:
  name: test-plugin
  version: 1.0.0
  description: Test plugin
  tool_name: test
`
		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)

		manifest, err := LoadManifest(manifestPath)
		require.NoError(t, err)
		assert.Equal(t, "test-plugin", manifest.Metadata.Name)
		assert.Equal(t, "1.0.0", manifest.Metadata.Version)
		assert.Equal(t, "Test plugin", manifest.Metadata.Description)
		assert.Equal(t, "test", manifest.Metadata.ToolName)
	})

	t.Run("fails on missing name", func(t *testing.T) {
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "plugin.yaml")

		manifestContent := `
metadata:
  version: 1.0.0
  tool_name: test
`
		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)

		_, err = LoadManifest(manifestPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("fails on missing version", func(t *testing.T) {
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "plugin.yaml")

		manifestContent := `
metadata:
  name: test-plugin
  tool_name: test
`
		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)

		_, err = LoadManifest(manifestPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version is required")
	})

	t.Run("fails on missing tool_name", func(t *testing.T) {
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "plugin.yaml")

		manifestContent := `
metadata:
  name: test-plugin
  version: 1.0.0
`
		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)

		_, err = LoadManifest(manifestPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tool_name is required")
	})

	t.Run("fails on non-existent file", func(t *testing.T) {
		_, err := LoadManifest("/non/existent/path/plugin.yaml")
		assert.Error(t, err)
	})
}

func TestGetPluginsDir(t *testing.T) {
	t.Run("returns plugins directory path", func(t *testing.T) {
		pluginsDir, err := GetPluginsDir()
		require.NoError(t, err)
		assert.Contains(t, pluginsDir, ".envswitch")
		assert.Contains(t, pluginsDir, "plugins")
	})
}

func TestListInstalledPlugins(t *testing.T) {
	t.Run("returns empty list when no plugins installed", func(t *testing.T) {
		// This test uses the real plugins directory
		// In a real test, we'd want to mock this
		plugins, err := ListInstalledPlugins()
		require.NoError(t, err)
		assert.NotNil(t, plugins)
	})
}
