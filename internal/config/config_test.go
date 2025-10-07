package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	t.Run("has correct version", func(t *testing.T) {
		assert.Equal(t, "1.0", cfg.Version)
	})

	t.Run("has sensible defaults", func(t *testing.T) {
		assert.Equal(t, "false", cfg.AutoSaveBeforeSwitch) // Changed to false to match new workflow
		assert.False(t, cfg.VerifyAfterSwitch)
		assert.True(t, cfg.BackupBeforeSwitch)
		assert.Equal(t, 10, cfg.BackupRetention)
		assert.True(t, cfg.EnablePromptIntegration)
		assert.Equal(t, "({name})", cfg.PromptFormat)
		assert.Equal(t, "blue", cfg.PromptColor)
		assert.Equal(t, "warn", cfg.LogLevel)
		assert.True(t, cfg.ColorOutput)
		assert.True(t, cfg.ShowTimestamps)
	})

	t.Run("has empty exclude tools", func(t *testing.T) {
		assert.Empty(t, cfg.ExcludeTools)
	})

	t.Run("has log file path set", func(t *testing.T) {
		assert.NotEmpty(t, cfg.LogFile)
		assert.Contains(t, cfg.LogFile, ".envswitch")
		assert.Contains(t, cfg.LogFile, "envswitch.log")
	})
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()

	t.Run("returns valid path", func(t *testing.T) {
		assert.NotEmpty(t, path)
		assert.Contains(t, path, ".envswitch")
		assert.Contains(t, path, "config.yaml")
	})
}

func TestLoadConfig(t *testing.T) {
	// Setup temp home directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("returns default config when file doesn't exist", func(t *testing.T) {
		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "warn", cfg.LogLevel)
	})

	t.Run("loads config from file", func(t *testing.T) {
		// Create and save a config
		cfg := DefaultConfig()
		cfg.LogLevel = "debug"
		cfg.BackupRetention = 20
		err := cfg.Save()
		require.NoError(t, err)

		// Load it back
		loadedCfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "debug", loadedCfg.LogLevel)
		assert.Equal(t, 20, loadedCfg.BackupRetention)
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		// Create invalid config file
		configPath := GetConfigPath()
		os.MkdirAll(filepath.Dir(configPath), 0755)
		err := os.WriteFile(configPath, []byte("invalid: [yaml: content"), 0644)
		require.NoError(t, err)

		_, err = LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse config file")
	})
}

func TestConfigSave(t *testing.T) {
	// Setup temp home directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("saves config to file", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.LogLevel = "debug"

		err := cfg.Save()
		require.NoError(t, err)

		// Verify file exists
		configPath := GetConfigPath()
		_, err = os.Stat(configPath)
		assert.NoError(t, err)
	})

	t.Run("creates directory if it doesn't exist", func(t *testing.T) {
		// Use a fresh directory
		freshDir := filepath.Join(tempDir, "fresh")
		os.Setenv("HOME", freshDir)
		defer os.Setenv("HOME", tempDir)

		cfg := DefaultConfig()
		err := cfg.Save()
		require.NoError(t, err)

		// Verify directory was created
		configPath := GetConfigPath()
		_, err = os.Stat(filepath.Dir(configPath))
		assert.NoError(t, err)
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.LogLevel = "info"
		err := cfg.Save()
		require.NoError(t, err)

		// Update and save again
		cfg.LogLevel = "warn"
		err = cfg.Save()
		require.NoError(t, err)

		// Load and verify
		loadedCfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "warn", loadedCfg.LogLevel)
	})
}

func TestConfigGet(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LogLevel = "debug"
	cfg.BackupRetention = 15
	cfg.VerifyAfterSwitch = true

	t.Run("gets string value", func(t *testing.T) {
		value, err := cfg.Get("log_level")
		require.NoError(t, err)
		assert.Equal(t, "debug", value)
	})

	t.Run("gets int value", func(t *testing.T) {
		value, err := cfg.Get("backup_retention")
		require.NoError(t, err)
		assert.Equal(t, 15, value)
	})

	t.Run("gets bool value", func(t *testing.T) {
		value, err := cfg.Get("verify_after_switch")
		require.NoError(t, err)
		assert.Equal(t, true, value)
	})

	t.Run("returns error for unknown key", func(t *testing.T) {
		_, err := cfg.Get("unknown_key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown config key")
	})

	t.Run("gets all valid keys", func(t *testing.T) {
		keys := []string{
			"auto_save_before_switch",
			"verify_after_switch",
			"backup_before_switch",
			"backup_retention",
			"enable_prompt_integration",
			"prompt_format",
			"prompt_color",
			"log_level",
			"log_file",
			"color_output",
			"show_timestamps",
		}

		for _, key := range keys {
			_, err := cfg.Get(key)
			assert.NoError(t, err, "should get key: %s", key)
		}
	})
}

func TestConfigSet(t *testing.T) {
	t.Run("sets auto_save_before_switch with valid values", func(t *testing.T) {
		cfg := DefaultConfig()
		validValues := []string{"true", "false", "prompt"}

		for _, value := range validValues {
			err := cfg.Set("auto_save_before_switch", value)
			assert.NoError(t, err, "should accept value: %s", value)
			assert.Equal(t, value, cfg.AutoSaveBeforeSwitch)
		}
	})

	t.Run("rejects invalid auto_save_before_switch value", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Set("auto_save_before_switch", "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value")
	})

	t.Run("sets verify_after_switch", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Set("verify_after_switch", true)
		assert.NoError(t, err)
		assert.True(t, cfg.VerifyAfterSwitch)
	})

	t.Run("rejects wrong type for verify_after_switch", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Set("verify_after_switch", "not a bool")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid type")
	})

	t.Run("sets backup_retention", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Set("backup_retention", 20)
		assert.NoError(t, err)
		assert.Equal(t, 20, cfg.BackupRetention)
	})

	t.Run("rejects wrong type for backup_retention", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Set("backup_retention", "not an int")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid type")
	})

	t.Run("rejects unknown key", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Set("unknown_key", "value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown")
	})

	t.Run("sets enable_prompt_integration", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Set("enable_prompt_integration", false)
		assert.NoError(t, err)
		assert.False(t, cfg.EnablePromptIntegration)
	})

	t.Run("sets prompt_format", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Set("prompt_format", "[{name}]")
		assert.NoError(t, err)
		assert.Equal(t, "[{name}]", cfg.PromptFormat)
	})

	t.Run("sets prompt_color", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Set("prompt_color", "green")
		assert.NoError(t, err)
		assert.Equal(t, "green", cfg.PromptColor)
	})

	t.Run("sets log_level with valid values", func(t *testing.T) {
		cfg := DefaultConfig()
		validValues := []string{"debug", "info", "warn", "error"}

		for _, value := range validValues {
			err := cfg.Set("log_level", value)
			assert.NoError(t, err, "should accept value: %s", value)
			assert.Equal(t, value, cfg.LogLevel)
		}
	})

	t.Run("rejects invalid log_level value", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Set("log_level", "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value")
	})

	t.Run("sets backup_before_switch", func(t *testing.T) {
		cfg := DefaultConfig()

		// Test setting to false
		err := cfg.Set("backup_before_switch", false)
		assert.NoError(t, err)
		assert.False(t, cfg.BackupBeforeSwitch)

		// Test setting to true
		err = cfg.Set("backup_before_switch", true)
		assert.NoError(t, err)
		assert.True(t, cfg.BackupBeforeSwitch)
	})

	t.Run("sets color_output", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Set("color_output", false)
		assert.NoError(t, err)
		assert.False(t, cfg.ColorOutput)
	})

	t.Run("returns error for unknown key", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Set("unknown_key", "value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown")
	})

	t.Run("returns error for read-only key", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Set("version", "2.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown")
	})
}

func TestConfigPersistence(t *testing.T) {
	// Setup temp home directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("changes persist after save and load", func(t *testing.T) {
		// Create config and modify it
		cfg := DefaultConfig()
		cfg.LogLevel = "debug"
		cfg.BackupRetention = 25
		cfg.VerifyAfterSwitch = true
		cfg.PromptColor = "red"

		// Save
		err := cfg.Save()
		require.NoError(t, err)

		// Load
		loadedCfg, err := LoadConfig()
		require.NoError(t, err)

		// Verify all changes persisted
		assert.Equal(t, "debug", loadedCfg.LogLevel)
		assert.Equal(t, 25, loadedCfg.BackupRetention)
		assert.True(t, loadedCfg.VerifyAfterSwitch)
		assert.Equal(t, "red", loadedCfg.PromptColor)
	})
}

func TestConfigYAMLFormat(t *testing.T) {
	// Setup temp home directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("saves as valid YAML", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Save()
		require.NoError(t, err)

		// Read the file
		configPath := GetConfigPath()
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		// Check it contains expected YAML keys
		content := string(data)
		assert.Contains(t, content, "version:")
		assert.Contains(t, content, "auto_save_before_switch:")
		assert.Contains(t, content, "log_level:")
		assert.Contains(t, content, "backup_retention:")
	})
}

func TestConfigEdgeCases(t *testing.T) {
	t.Run("handles multiple Set operations", func(t *testing.T) {
		cfg := DefaultConfig()

		err := cfg.Set("log_level", "debug")
		require.NoError(t, err)

		err = cfg.Set("log_level", "info")
		require.NoError(t, err)

		err = cfg.Set("log_level", "warn")
		require.NoError(t, err)

		assert.Equal(t, "warn", cfg.LogLevel)
	})

	t.Run("Get and Set work together", func(t *testing.T) {
		cfg := DefaultConfig()

		// Set a value
		err := cfg.Set("backup_retention", 30)
		require.NoError(t, err)

		// Get it back
		value, err := cfg.Get("backup_retention")
		require.NoError(t, err)
		assert.Equal(t, 30, value)
	})
}
