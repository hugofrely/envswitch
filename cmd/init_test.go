package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRunInit(t *testing.T) {
	// Create a temporary directory for testing
	tempHome := t.TempDir()

	// Save original home and restore after test
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})

	os.Setenv("HOME", tempHome)

	t.Run("creates directory structure", func(t *testing.T) {
		err := runInit(initCmd, []string{})
		require.NoError(t, err)

		envswitchDir := filepath.Join(tempHome, ".envswitch")

		// Check main directory exists
		assert.DirExists(t, envswitchDir)

		// Check subdirectories exist
		assert.DirExists(t, filepath.Join(envswitchDir, "environments"))
		assert.DirExists(t, filepath.Join(envswitchDir, "auto-backups"))
	})

	t.Run("creates default config file", func(t *testing.T) {
		err := runInit(initCmd, []string{})
		require.NoError(t, err)

		envswitchDir := filepath.Join(tempHome, ".envswitch")
		configPath := filepath.Join(envswitchDir, "config.yaml")

		assert.FileExists(t, configPath)

		// Read and validate config
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var config map[string]interface{}
		err = yaml.Unmarshal(data, &config)
		require.NoError(t, err)

		assert.Equal(t, "1.0", config["version"])
		assert.Equal(t, true, config["auto_save_before_switch"])
		assert.Equal(t, "vim", config["default_editor"])
	})

	t.Run("creates history log", func(t *testing.T) {
		err := runInit(initCmd, []string{})
		require.NoError(t, err)

		envswitchDir := filepath.Join(tempHome, ".envswitch")
		historyPath := filepath.Join(envswitchDir, "history.log")

		assert.FileExists(t, historyPath)
	})

	t.Run("does not overwrite existing config", func(t *testing.T) {
		// First init
		err := runInit(initCmd, []string{})
		require.NoError(t, err)

		envswitchDir := filepath.Join(tempHome, ".envswitch")
		configPath := filepath.Join(envswitchDir, "config.yaml")

		// Modify config
		customConfig := map[string]interface{}{
			"version": "2.0",
			"custom":  "value",
		}
		data, err := yaml.Marshal(customConfig)
		require.NoError(t, err)
		err = os.WriteFile(configPath, data, 0644)
		require.NoError(t, err)

		// Second init should not overwrite
		err = runInit(initCmd, []string{})
		require.NoError(t, err)

		// Read config and verify it wasn't overwritten
		data, err = os.ReadFile(configPath)
		require.NoError(t, err)

		var config map[string]interface{}
		err = yaml.Unmarshal(data, &config)
		require.NoError(t, err)

		assert.Equal(t, "2.0", config["version"])
		assert.Equal(t, "value", config["custom"])
	})
}
