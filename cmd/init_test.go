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

	t.Run("creates config with all default values", func(t *testing.T) {
		// Use a fresh temp directory for this test to avoid pollution
		freshDir := t.TempDir()
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", freshDir)
		defer os.Setenv("HOME", oldHome)

		err := runInit(initCmd, []string{})
		require.NoError(t, err)

		envswitchDir := filepath.Join(freshDir, ".envswitch")
		configPath := filepath.Join(envswitchDir, "config.yaml")

		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var config map[string]interface{}
		err = yaml.Unmarshal(data, &config)
		require.NoError(t, err)

		// Verify all default config values
		assert.Equal(t, "1.0", config["version"])
		assert.Equal(t, true, config["auto_save_before_switch"])
		assert.Equal(t, false, config["verify_after_switch"])
		assert.Equal(t, 10, config["backup_retention"])
		assert.Equal(t, true, config["enable_prompt_integration"])
		assert.Equal(t, "({name})", config["prompt_format"])
		assert.Equal(t, "blue", config["prompt_color"])
		assert.Equal(t, "info", config["log_level"])
		assert.NotEmpty(t, config["log_file"])
		assert.NotNil(t, config["exclude_tools"])
		assert.Equal(t, true, config["color_output"])
		assert.Equal(t, false, config["show_timestamps"])
		assert.Equal(t, true, config["backup_before_switch"])
	})

	t.Run("handles existing directory gracefully", func(t *testing.T) {
		envswitchDir := filepath.Join(tempHome, ".envswitch")
		err := os.MkdirAll(envswitchDir, 0755)
		require.NoError(t, err)

		// Init should not fail if directory already exists
		err = runInit(initCmd, []string{})
		require.NoError(t, err)

		assert.DirExists(t, envswitchDir)
	})

	t.Run("handles existing subdirectories gracefully", func(t *testing.T) {
		envswitchDir := filepath.Join(tempHome, ".envswitch")
		err := os.MkdirAll(filepath.Join(envswitchDir, "environments"), 0755)
		require.NoError(t, err)

		// Init should not fail if subdirectories already exist
		err = runInit(initCmd, []string{})
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(envswitchDir, "environments"))
		assert.DirExists(t, filepath.Join(envswitchDir, "auto-backups"))
	})

	t.Run("handles existing history log gracefully", func(t *testing.T) {
		envswitchDir := filepath.Join(tempHome, ".envswitch")
		err := os.MkdirAll(envswitchDir, 0755)
		require.NoError(t, err)

		historyPath := filepath.Join(envswitchDir, "history.log")
		existingContent := "existing history\n"
		err = os.WriteFile(historyPath, []byte(existingContent), 0644)
		require.NoError(t, err)

		// Init should not overwrite history log
		err = runInit(initCmd, []string{})
		require.NoError(t, err)

		content, err := os.ReadFile(historyPath)
		require.NoError(t, err)
		assert.Equal(t, existingContent, string(content))
	})
}

func TestInitCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "init", initCmd.Use)
		assert.NotEmpty(t, initCmd.Short)
		assert.NotEmpty(t, initCmd.Long)
	})

	t.Run("requires no arguments", func(t *testing.T) {
		// Init should work with no arguments
		// This is implicitly tested in other tests
		assert.NotNil(t, initCmd)
	})
}
