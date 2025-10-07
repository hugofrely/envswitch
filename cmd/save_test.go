package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/pkg/environment"
)

func TestRunSave(t *testing.T) {
	// Setup temp home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})
	os.Setenv("HOME", tempHome)

	t.Run("fails when no active environment", func(t *testing.T) {
		// Ensure no current environment by removing current.lock if it exists
		dir, _ := environment.GetEnvswitchDir()
		lockPath := filepath.Join(dir, "current.lock")
		os.Remove(lockPath)

		err := runSave(saveCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active environment")
	})

	t.Run("saves current state", func(t *testing.T) {
		// Create test environment
		envDir, _ := environment.GetEnvironmentsDir()
		testEnvPath := filepath.Join(envDir, "test")
		os.MkdirAll(filepath.Join(testEnvPath, "snapshots"), 0755)

		env := &environment.Environment{
			Name:  "test",
			Path:  testEnvPath,
			Tools: make(map[string]environment.ToolConfig),
		}

		// Initialize tools as disabled (since they may not be installed)
		toolNames := []string{"gcloud", "kubectl", "aws", "docker", "git"}
		for _, tool := range toolNames {
			env.Tools[tool] = environment.ToolConfig{
				Enabled:      false,
				SnapshotPath: filepath.Join("snapshots", tool),
				Metadata:     make(map[string]interface{}),
			}
		}

		err := env.Save()
		require.NoError(t, err)

		// Set as current
		err = environment.SetCurrentEnvironment("test")
		require.NoError(t, err)

		// Run save
		err = runSave(saveCmd, []string{})
		assert.NoError(t, err)

		// Verify environment still exists and has metadata
		savedEnv, err := environment.LoadEnvironment("test")
		require.NoError(t, err)
		assert.Equal(t, "test", savedEnv.Name)
		assert.NotNil(t, savedEnv.Tools)
	})

	t.Run("updates existing snapshot", func(t *testing.T) {
		// Create test environment
		envDir, _ := environment.GetEnvironmentsDir()
		testEnvPath := filepath.Join(envDir, "update-test")
		snapshotsPath := filepath.Join(testEnvPath, "snapshots")
		os.MkdirAll(snapshotsPath, 0755)

		env := &environment.Environment{
			Name:  "update-test",
			Path:  testEnvPath,
			Tools: make(map[string]environment.ToolConfig),
		}

		// Initialize tools
		toolNames := []string{"gcloud", "kubectl", "aws", "docker", "git"}
		for _, tool := range toolNames {
			env.Tools[tool] = environment.ToolConfig{
				Enabled:      false,
				SnapshotPath: filepath.Join("snapshots", tool),
				Metadata:     make(map[string]interface{}),
			}
		}

		err := env.Save()
		require.NoError(t, err)

		// Set as current
		err = environment.SetCurrentEnvironment("update-test")
		require.NoError(t, err)

		// First save
		err = runSave(saveCmd, []string{})
		require.NoError(t, err)

		// Second save (update)
		err = runSave(saveCmd, []string{})
		assert.NoError(t, err)

		// Verify environment still exists
		savedEnv, err := environment.LoadEnvironment("update-test")
		require.NoError(t, err)
		assert.Equal(t, "update-test", savedEnv.Name)
	})

	t.Run("handles multiple tools", func(t *testing.T) {
		// Create test environment with multiple tools
		envDir, _ := environment.GetEnvironmentsDir()
		testEnvPath := filepath.Join(envDir, "multi-tool")
		os.MkdirAll(filepath.Join(testEnvPath, "snapshots"), 0755)

		env := &environment.Environment{
			Name:  "multi-tool",
			Path:  testEnvPath,
			Tools: make(map[string]environment.ToolConfig),
		}

		// Initialize multiple tools
		toolNames := []string{"gcloud", "kubectl", "aws", "docker", "git"}
		for _, tool := range toolNames {
			env.Tools[tool] = environment.ToolConfig{
				Enabled:      false,
				SnapshotPath: filepath.Join("snapshots", tool),
				Metadata:     make(map[string]interface{}),
			}
		}

		err := env.Save()
		require.NoError(t, err)

		// Set as current
		err = environment.SetCurrentEnvironment("multi-tool")
		require.NoError(t, err)

		// Run save
		err = runSave(saveCmd, []string{})
		assert.NoError(t, err)

		// Verify all tools are still configured
		savedEnv, err := environment.LoadEnvironment("multi-tool")
		require.NoError(t, err)
		assert.Len(t, savedEnv.Tools, len(toolNames))
	})

	t.Run("skips disabled tools", func(t *testing.T) {
		// Create test environment
		envDir, _ := environment.GetEnvironmentsDir()
		testEnvPath := filepath.Join(envDir, "skip-test")
		os.MkdirAll(filepath.Join(testEnvPath, "snapshots"), 0755)

		env := &environment.Environment{
			Name:  "skip-test",
			Path:  testEnvPath,
			Tools: make(map[string]environment.ToolConfig),
		}

		// Initialize with some disabled tools
		env.Tools["kubectl"] = environment.ToolConfig{
			Enabled:      false,
			SnapshotPath: filepath.Join("snapshots", "kubectl"),
			Metadata:     make(map[string]interface{}),
		}

		err := env.Save()
		require.NoError(t, err)

		// Set as current
		err = environment.SetCurrentEnvironment("skip-test")
		require.NoError(t, err)

		// Run save - should not fail even with disabled tools
		err = runSave(saveCmd, []string{})
		assert.NoError(t, err)
	})
}

func TestSaveCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "save", saveCmd.Use)
		assert.NotEmpty(t, saveCmd.Short)
		assert.NotEmpty(t, saveCmd.Long)
	})

	t.Run("accepts no arguments", func(t *testing.T) {
		assert.NotNil(t, saveCmd.Args)
		// Test that it rejects arguments
		err := saveCmd.Args(saveCmd, []string{"extra-arg"})
		assert.Error(t, err)
	})

	t.Run("has runE function", func(t *testing.T) {
		assert.NotNil(t, saveCmd.RunE)
	})
}
