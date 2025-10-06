package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/internal/config"
)

func TestConfigCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "config", configCmd.Use)
		assert.Contains(t, configCmd.Short, "configuration")
	})

	t.Run("has list subcommand", func(t *testing.T) {
		found := false
		for _, cmd := range configCmd.Commands() {
			if cmd.Name() == "list" {
				found = true
				break
			}
		}
		assert.True(t, found, "config list subcommand should exist")
	})

	t.Run("has get subcommand", func(t *testing.T) {
		found := false
		for _, cmd := range configCmd.Commands() {
			if cmd.Name() == "get" {
				found = true
				break
			}
		}
		assert.True(t, found, "config get subcommand should exist")
	})

	t.Run("has set subcommand", func(t *testing.T) {
		found := false
		for _, cmd := range configCmd.Commands() {
			if cmd.Name() == "set" {
				found = true
				break
			}
		}
		assert.True(t, found, "config set subcommand should exist")
	})
}

func TestRunConfigList(t *testing.T) {
	// Setup temp home directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("lists default configuration when no config exists", func(t *testing.T) {
		// Just ensure it doesn't error
		err := runConfigList(configListCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("lists existing configuration", func(t *testing.T) {
		// Create a config
		cfg := config.DefaultConfig()
		cfg.LogLevel = "debug"
		cfg.ColorOutput = false

		// Save it
		err := cfg.Save()
		require.NoError(t, err)

		// List it
		err = runConfigList(configListCmd, []string{})
		assert.NoError(t, err)
	})
}

func TestRunConfigGet(t *testing.T) {
	// Setup temp home directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("gets default value for non-existent config", func(t *testing.T) {
		err := runConfigGet(configGetCmd, []string{"log_level"})
		assert.NoError(t, err)
	})

	t.Run("gets custom value from saved config", func(t *testing.T) {
		// Create and save config
		cfg := config.DefaultConfig()
		cfg.LogLevel = "debug"
		err := cfg.Save()
		require.NoError(t, err)

		// Get the value
		err = runConfigGet(configGetCmd, []string{"log_level"})
		assert.NoError(t, err)
	})

	t.Run("returns error for unknown key", func(t *testing.T) {
		err := runConfigGet(configGetCmd, []string{"nonexistent_key"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown config key")
	})

	t.Run("gets all supported config keys", func(t *testing.T) {
		cfg := config.DefaultConfig()
		err := cfg.Save()
		require.NoError(t, err)

		keys := []string{
			"auto_save_before_switch",
			"verify_after_switch",
			"backup_retention",
			"default_editor",
			"enable_prompt_integration",
			"prompt_format",
			"prompt_color",
			"log_level",
			"log_file",
			"encryption_enabled",
			"encryption_use_keyring",
			"color_output",
			"show_timestamps",
		}

		for _, key := range keys {
			err := runConfigGet(configGetCmd, []string{key})
			assert.NoError(t, err, "should get key: %s", key)
		}
	})
}

func TestRunConfigSet(t *testing.T) {
	// Setup temp home directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("sets boolean value", func(t *testing.T) {
		err := runConfigSet(configSetCmd, []string{"verify_after_switch", "true"})
		assert.NoError(t, err)

		// Verify it was saved
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.True(t, cfg.VerifyAfterSwitch)
	})

	t.Run("sets string value", func(t *testing.T) {
		err := runConfigSet(configSetCmd, []string{"log_level", "debug"})
		assert.NoError(t, err)

		// Verify it was saved
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "debug", cfg.LogLevel)
	})

	t.Run("sets integer value", func(t *testing.T) {
		err := runConfigSet(configSetCmd, []string{"backup_retention", "20"})
		assert.NoError(t, err)

		// Verify it was saved
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, 20, cfg.BackupRetention)
	})

	t.Run("sets auto_save_before_switch with valid values", func(t *testing.T) {
		validValues := []string{"true", "false", "prompt"}

		for _, value := range validValues {
			err := runConfigSet(configSetCmd, []string{"auto_save_before_switch", value})
			assert.NoError(t, err, "should accept value: %s", value)

			// Verify it was saved
			cfg, err := config.LoadConfig()
			require.NoError(t, err)
			assert.Equal(t, value, cfg.AutoSaveBeforeSwitch)
		}
	})

	t.Run("rejects invalid auto_save_before_switch value", func(t *testing.T) {
		err := runConfigSet(configSetCmd, []string{"auto_save_before_switch", "invalid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value")
	})

	t.Run("sets log_level with valid values", func(t *testing.T) {
		validValues := []string{"debug", "info", "warn", "error"}

		for _, value := range validValues {
			err := runConfigSet(configSetCmd, []string{"log_level", value})
			assert.NoError(t, err, "should accept value: %s", value)

			// Verify it was saved
			cfg, err := config.LoadConfig()
			require.NoError(t, err)
			assert.Equal(t, value, cfg.LogLevel)
		}
	})

	t.Run("rejects invalid log_level value", func(t *testing.T) {
		err := runConfigSet(configSetCmd, []string{"log_level", "invalid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value")
	})

	t.Run("returns error for unknown key", func(t *testing.T) {
		err := runConfigSet(configSetCmd, []string{"nonexistent_key", "value"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown")
	})

	t.Run("creates config directory if it doesn't exist", func(t *testing.T) {
		// Use a fresh temp directory
		freshDir := filepath.Join(tempDir, "fresh")
		os.Setenv("HOME", freshDir)
		defer os.Setenv("HOME", tempDir)

		err := runConfigSet(configSetCmd, []string{"log_level", "info"})
		assert.NoError(t, err)

		// Verify config file was created
		configPath := filepath.Join(freshDir, ".envswitch", "config.yaml")
		_, err = os.Stat(configPath)
		assert.NoError(t, err, "config file should be created")
	})

	t.Run("sets editor value", func(t *testing.T) {
		err := runConfigSet(configSetCmd, []string{"default_editor", "nano"})
		assert.NoError(t, err)

		// Verify it was saved
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "nano", cfg.DefaultEditor)
	})

	t.Run("sets prompt format", func(t *testing.T) {
		err := runConfigSet(configSetCmd, []string{"prompt_format", "[{name}]"})
		assert.NoError(t, err)

		// Verify it was saved
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "[{name}]", cfg.PromptFormat)
	})

	t.Run("sets prompt color", func(t *testing.T) {
		err := runConfigSet(configSetCmd, []string{"prompt_color", "green"})
		assert.NoError(t, err)

		// Verify it was saved
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "green", cfg.PromptColor)
	})

	t.Run("toggles encryption_enabled", func(t *testing.T) {
		err := runConfigSet(configSetCmd, []string{"encryption_enabled", "true"})
		assert.NoError(t, err)

		// Verify it was saved
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.True(t, cfg.EncryptionEnabled)

		// Toggle back
		err = runConfigSet(configSetCmd, []string{"encryption_enabled", "false"})
		assert.NoError(t, err)

		cfg, err = config.LoadConfig()
		require.NoError(t, err)
		assert.False(t, cfg.EncryptionEnabled)
	})

	t.Run("toggles color_output", func(t *testing.T) {
		err := runConfigSet(configSetCmd, []string{"color_output", "false"})
		assert.NoError(t, err)

		// Verify it was saved
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.False(t, cfg.ColorOutput)
	})

	t.Run("toggles enable_prompt_integration", func(t *testing.T) {
		err := runConfigSet(configSetCmd, []string{"enable_prompt_integration", "false"})
		assert.NoError(t, err)

		// Verify it was saved
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.False(t, cfg.EnablePromptIntegration)
	})
}

func TestConfigSetCmd(t *testing.T) {
	t.Run("requires exactly 2 arguments", func(t *testing.T) {
		err := configSetCmd.Args(configSetCmd, []string{"key", "value"})
		assert.NoError(t, err)

		err = configSetCmd.Args(configSetCmd, []string{"key"})
		assert.Error(t, err)

		err = configSetCmd.Args(configSetCmd, []string{})
		assert.Error(t, err)
	})
}

func TestConfigGetCmd(t *testing.T) {
	t.Run("requires exactly 1 argument", func(t *testing.T) {
		err := configGetCmd.Args(configGetCmd, []string{"key"})
		assert.NoError(t, err)

		err = configGetCmd.Args(configGetCmd, []string{})
		assert.Error(t, err)

		err = configGetCmd.Args(configGetCmd, []string{"key", "extra"})
		assert.Error(t, err)
	})
}

func TestConfigIntegration(t *testing.T) {
	// Setup temp home directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("full workflow: set, get, list", func(t *testing.T) {
		// Set a value
		err := runConfigSet(configSetCmd, []string{"log_level", "debug"})
		require.NoError(t, err)

		// Get the value
		err = runConfigGet(configGetCmd, []string{"log_level"})
		assert.NoError(t, err)

		// List all values
		err = runConfigList(configListCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("updates persist across multiple operations", func(t *testing.T) {
		// Set multiple values
		err := runConfigSet(configSetCmd, []string{"log_level", "warn"})
		require.NoError(t, err)

		err = runConfigSet(configSetCmd, []string{"backup_retention", "15"})
		require.NoError(t, err)

		err = runConfigSet(configSetCmd, []string{"verify_after_switch", "true"})
		require.NoError(t, err)

		// Verify all values are correct
		cfg, err := config.LoadConfig()
		require.NoError(t, err)

		assert.Equal(t, "warn", cfg.LogLevel)
		assert.Equal(t, 15, cfg.BackupRetention)
		assert.True(t, cfg.VerifyAfterSwitch)
	})

	t.Run("default values are used when not set", func(t *testing.T) {
		// Load fresh config (no file exists yet in this test)
		freshDir := filepath.Join(tempDir, "fresh-defaults")
		os.Setenv("HOME", freshDir)
		defer os.Setenv("HOME", tempDir)

		cfg, err := config.LoadConfig()
		require.NoError(t, err)

		// Check some default values
		assert.Equal(t, "info", cfg.LogLevel)
		assert.Equal(t, 10, cfg.BackupRetention)
		assert.Equal(t, "vim", cfg.DefaultEditor)
		assert.Equal(t, "true", cfg.AutoSaveBeforeSwitch)
	})
}
