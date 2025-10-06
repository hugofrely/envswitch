package environment

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnvswitchDir(t *testing.T) {
	t.Run("returns correct directory path", func(t *testing.T) {
		dir, err := GetEnvswitchDir()
		require.NoError(t, err)

		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".envswitch")
		assert.Equal(t, expected, dir)
	})
}

func TestGetEnvironmentsDir(t *testing.T) {
	t.Run("returns correct environments directory path", func(t *testing.T) {
		dir, err := GetEnvironmentsDir()
		require.NoError(t, err)

		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".envswitch", "environments")
		assert.Equal(t, expected, dir)
	})
}

func TestEnvironmentSaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tempHome := t.TempDir()

	// Save original home and restore after test
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})

	os.Setenv("HOME", tempHome)

	// Create directory structure
	envDir := filepath.Join(tempHome, ".envswitch", "environments", "test-env")
	err := os.MkdirAll(envDir, 0755)
	require.NoError(t, err)

	t.Run("saves and loads environment", func(t *testing.T) {
		now := time.Now()
		env := &Environment{
			Name:        "test-env",
			Description: "Test environment",
			CreatedAt:   now,
			UpdatedAt:   now,
			LastUsed:    now,
			Tools: map[string]ToolConfig{
				"gcloud": {
					Enabled:      true,
					SnapshotPath: "snapshots/gcloud",
					Metadata:     map[string]interface{}{"version": "1.0"},
				},
			},
			EnvVars: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			Path: envDir,
		}

		// Save
		err := env.Save()
		require.NoError(t, err)

		// Check metadata file exists
		metadataPath := filepath.Join(envDir, "metadata.yaml")
		assert.FileExists(t, metadataPath)

		// Load
		loaded, err := LoadEnvironment("test-env")
		require.NoError(t, err)

		assert.Equal(t, env.Name, loaded.Name)
		assert.Equal(t, env.Description, loaded.Description)
		assert.Equal(t, env.Tools["gcloud"].Enabled, loaded.Tools["gcloud"].Enabled)
		assert.Equal(t, env.EnvVars["KEY1"], loaded.EnvVars["KEY1"])
	})

	t.Run("updates UpdatedAt on save", func(t *testing.T) {
		env := &Environment{
			Name:      "test-env2",
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now().Add(-time.Hour),
			Path:      filepath.Join(tempHome, ".envswitch", "environments", "test-env2"),
		}

		err := os.MkdirAll(env.Path, 0755)
		require.NoError(t, err)

		oldUpdatedAt := env.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		err = env.Save()
		require.NoError(t, err)

		assert.True(t, env.UpdatedAt.After(oldUpdatedAt))
	})

	t.Run("returns error for non-existent environment", func(t *testing.T) {
		_, err := LoadEnvironment("non-existent")
		assert.Error(t, err)
	})
}

func TestListEnvironments(t *testing.T) {
	// Create a temporary directory for testing
	tempHome := t.TempDir()

	// Save original home and restore after test
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})

	os.Setenv("HOME", tempHome)

	envswitchDir := filepath.Join(tempHome, ".envswitch")
	envDir := filepath.Join(envswitchDir, "environments")

	t.Run("returns empty list when no environments exist", func(t *testing.T) {
		err := os.MkdirAll(envDir, 0755)
		require.NoError(t, err)

		envs, err := ListEnvironments()
		require.NoError(t, err)
		assert.Empty(t, envs)
	})

	t.Run("lists multiple environments", func(t *testing.T) {
		// Create test environments
		env1 := &Environment{
			Name: "env1",
			Path: filepath.Join(envDir, "env1"),
		}
		env2 := &Environment{
			Name: "env2",
			Path: filepath.Join(envDir, "env2"),
		}

		err := os.MkdirAll(env1.Path, 0755)
		require.NoError(t, err)
		err = env1.Save()
		require.NoError(t, err)

		err = os.MkdirAll(env2.Path, 0755)
		require.NoError(t, err)
		err = env2.Save()
		require.NoError(t, err)

		// List environments
		envs, err := ListEnvironments()
		require.NoError(t, err)
		assert.Len(t, envs, 2)

		names := []string{envs[0].Name, envs[1].Name}
		assert.Contains(t, names, "env1")
		assert.Contains(t, names, "env2")
	})

	t.Run("skips invalid directories", func(t *testing.T) {
		// Create a directory without metadata
		invalidDir := filepath.Join(envDir, "invalid")
		err := os.MkdirAll(invalidDir, 0755)
		require.NoError(t, err)

		// Should not error, just skip invalid
		envs, err := ListEnvironments()
		require.NoError(t, err)

		// Should only have valid environments
		for _, env := range envs {
			assert.NotEqual(t, "invalid", env.Name)
		}
	})
}

func TestCurrentEnvironment(t *testing.T) {
	// Create a temporary directory for testing
	tempHome := t.TempDir()

	// Save original home and restore after test
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})

	os.Setenv("HOME", tempHome)

	envswitchDir := filepath.Join(tempHome, ".envswitch")
	envDir := filepath.Join(envswitchDir, "environments")

	t.Run("returns nil when no current environment", func(t *testing.T) {
		err := os.MkdirAll(envswitchDir, 0755)
		require.NoError(t, err)

		current, err := GetCurrentEnvironment()
		require.NoError(t, err)
		assert.Nil(t, current)
	})

	t.Run("sets and gets current environment", func(t *testing.T) {
		// Create test environment
		env := &Environment{
			Name: "current-test",
			Path: filepath.Join(envDir, "current-test"),
		}

		err := os.MkdirAll(env.Path, 0755)
		require.NoError(t, err)
		err = env.Save()
		require.NoError(t, err)

		// Set current
		err = SetCurrentEnvironment("current-test")
		require.NoError(t, err)

		// Get current
		current, err := GetCurrentEnvironment()
		require.NoError(t, err)
		assert.NotNil(t, current)
		assert.Equal(t, "current-test", current.Name)
	})

	t.Run("returns error when current environment doesn't exist", func(t *testing.T) {
		err := SetCurrentEnvironment("non-existent")
		require.NoError(t, err)

		_, err = GetCurrentEnvironment()
		assert.Error(t, err)
	})
}

func TestToolConfig(t *testing.T) {
	t.Run("creates tool config with defaults", func(t *testing.T) {
		config := ToolConfig{
			Enabled:      true,
			SnapshotPath: "snapshots/test",
			Metadata:     make(map[string]interface{}),
		}

		assert.True(t, config.Enabled)
		assert.Equal(t, "snapshots/test", config.SnapshotPath)
		assert.NotNil(t, config.Metadata)
	})
}
