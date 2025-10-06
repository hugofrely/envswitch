package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/pkg/environment"
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

	t.Run("creates environment from existing environment", func(t *testing.T) {
		// Create source environment with some data
		sourceEnvPath := filepath.Join(envDir, "source-env")
		err := os.MkdirAll(sourceEnvPath, 0755)
		require.NoError(t, err)

		// Create snapshots directory with some files
		sourceSnapshots := filepath.Join(sourceEnvPath, "snapshots")
		err = os.MkdirAll(filepath.Join(sourceSnapshots, "gcloud"), 0755)
		require.NoError(t, err)
		err = os.MkdirAll(filepath.Join(sourceSnapshots, "kubectl"), 0755)
		require.NoError(t, err)

		// Create some test files in snapshots
		testFile1 := filepath.Join(sourceSnapshots, "gcloud", "config.yaml")
		err = os.WriteFile(testFile1, []byte("test-config"), 0644)
		require.NoError(t, err)

		testFile2 := filepath.Join(sourceSnapshots, "kubectl", "config")
		err = os.WriteFile(testFile2, []byte("test-kube-config"), 0644)
		require.NoError(t, err)

		// Create env-vars.env file
		envVarsContent := "# Test env vars\nTEST_VAR=value123\n"
		err = os.WriteFile(filepath.Join(sourceEnvPath, "env-vars.env"), []byte(envVarsContent), 0644)
		require.NoError(t, err)

		// Create source environment metadata
		sourceEnv := &environment.Environment{
			Name:        "source-env",
			Description: "Source environment",
			Path:        sourceEnvPath,
			Tools: map[string]environment.ToolConfig{
				"gcloud": {
					Enabled:      true,
					SnapshotPath: "snapshots/gcloud",
					Metadata: map[string]interface{}{
						"project": "test-project",
					},
				},
				"kubectl": {
					Enabled:      true,
					SnapshotPath: "snapshots/kubectl",
					Metadata: map[string]interface{}{
						"context": "test-context",
					},
				},
			},
			EnvVars: map[string]string{
				"TEST_VAR": "value123",
			},
		}
		err = sourceEnv.Save()
		require.NoError(t, err)

		// Now clone from source environment
		createFrom = "source-env"
		defer func() { createFrom = "" }()

		err = runCreate(createCmd, []string{"cloned-env"})
		require.NoError(t, err)

		// Verify cloned environment
		clonedEnv, err := environment.LoadEnvironment("cloned-env")
		require.NoError(t, err)

		// Check that tools were copied
		assert.Equal(t, 2, len(clonedEnv.Tools))
		assert.True(t, clonedEnv.Tools["gcloud"].Enabled)
		assert.True(t, clonedEnv.Tools["kubectl"].Enabled)
		assert.Equal(t, "test-project", clonedEnv.Tools["gcloud"].Metadata["project"])
		assert.Equal(t, "test-context", clonedEnv.Tools["kubectl"].Metadata["context"])

		// Check that env vars were copied
		assert.Equal(t, "value123", clonedEnv.EnvVars["TEST_VAR"])

		// Verify snapshot files were copied
		clonedPath := filepath.Join(envDir, "cloned-env")
		assert.FileExists(t, filepath.Join(clonedPath, "snapshots", "gcloud", "config.yaml"))
		assert.FileExists(t, filepath.Join(clonedPath, "snapshots", "kubectl", "config"))

		// Verify file contents
		content, err := os.ReadFile(filepath.Join(clonedPath, "snapshots", "gcloud", "config.yaml"))
		require.NoError(t, err)
		assert.Equal(t, "test-config", string(content))

		content, err = os.ReadFile(filepath.Join(clonedPath, "snapshots", "kubectl", "config"))
		require.NoError(t, err)
		assert.Equal(t, "test-kube-config", string(content))

		// Verify env-vars.env was copied
		envVarsContent2, err := os.ReadFile(filepath.Join(clonedPath, "env-vars.env"))
		require.NoError(t, err)
		assert.Equal(t, envVarsContent, string(envVarsContent2))
	})

	t.Run("fails when cloning from non-existent environment", func(t *testing.T) {
		createFrom = "non-existent"
		defer func() { createFrom = "" }()

		err := runCreate(createCmd, []string{"should-fail"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})
}
