package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hugofrely/envswitch/pkg/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCreate(t *testing.T) {
	// Create a temporary directory for testing
	tempHome := t.TempDir()

	// Save original home and restore after test
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})

	os.Setenv("HOME", tempHome)

	// Create envswitch directory structure
	envswitchDir := filepath.Join(tempHome, ".envswitch")
	envDir := filepath.Join(envswitchDir, "environments")
	err := os.MkdirAll(envDir, 0755)
	require.NoError(t, err)

	t.Run("creates new environment", func(t *testing.T) {
		err := runCreate(createCmd, []string{"test-env"})
		require.NoError(t, err)

		// Check environment directory exists
		envPath := filepath.Join(envDir, "test-env")
		assert.DirExists(t, envPath)

		// Check snapshots directory exists
		assert.DirExists(t, filepath.Join(envPath, "snapshots"))

		// Check metadata file exists
		assert.FileExists(t, filepath.Join(envPath, "metadata.yaml"))

		// Check env-vars.env file exists
		assert.FileExists(t, filepath.Join(envPath, "env-vars.env"))
	})

	t.Run("validates environment name", func(t *testing.T) {
		err := runCreate(createCmd, []string{""})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("prevents duplicate environment", func(t *testing.T) {
		// Create first environment
		err := runCreate(createCmd, []string{"duplicate"})
		require.NoError(t, err)

		// Try to create duplicate
		err = runCreate(createCmd, []string{"duplicate"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("creates environment with description", func(t *testing.T) {
		createDescription = "Test description"
		defer func() { createDescription = "" }()

		err := runCreate(createCmd, []string{"with-desc"})
		require.NoError(t, err)

		// Load environment and check description
		env, err := environment.LoadEnvironment("with-desc")
		require.NoError(t, err)

		assert.Equal(t, "with-desc", env.Name)
		assert.Equal(t, "Test description", env.Description)
	})

	t.Run("creates environment from current state", func(t *testing.T) {
		createFromCurrent = true
		defer func() { createFromCurrent = false }()

		err := runCreate(createCmd, []string{"from-current"})
		require.NoError(t, err)

		// Load environment and verify it was created
		env, err := environment.LoadEnvironment("from-current")
		require.NoError(t, err)

		// In test environment, tools won't have config dirs, so they should be disabled
		// This is the correct behavior - tools are only enabled if snapshot succeeds
		assert.Equal(t, "from-current", env.Name)

		// Verify tools exist in config (even if disabled)
		for _, tool := range []string{"gcloud", "kubectl", "aws", "docker", "git"} {
			_, exists := env.Tools[tool]
			assert.True(t, exists, "Tool %s should exist in config", tool)
		}
	})

	t.Run("creates empty environment", func(t *testing.T) {
		createEmpty = true
		defer func() { createEmpty = false }()

		err := runCreate(createCmd, []string{"empty-env"})
		require.NoError(t, err)

		// Load environment and check tools are disabled
		env, err := environment.LoadEnvironment("empty-env")
		require.NoError(t, err)

		// Tools should be disabled when creating empty
		for _, tool := range []string{"gcloud", "kubectl", "aws", "docker", "git"} {
			assert.False(t, env.Tools[tool].Enabled, "Tool %s should be disabled", tool)
		}
	})
}
