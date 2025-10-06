package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/pkg/environment"
)

func TestRunSwitch(t *testing.T) {
	// Setup test environment
	originalHome := os.Getenv("HOME")
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Initialize envswitch directory
	envswitchDir := filepath.Join(tmpDir, ".envswitch")
	envsDir := filepath.Join(envswitchDir, "environments")
	os.MkdirAll(envsDir, 0755)

	t.Run("switches to target environment", func(t *testing.T) {
		// Create target environment
		targetEnv := &environment.Environment{
			Name:      "target",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars:   make(map[string]string),
			Path:      filepath.Join(envsDir, "target"),
		}
		os.MkdirAll(targetEnv.Path, 0755)
		err := targetEnv.Save()
		require.NoError(t, err)

		err = runSwitch(switchCmd, []string{"target"})
		require.NoError(t, err)

		// Verify current environment is set
		current, err := environment.GetCurrentEnvironment()
		require.NoError(t, err)
		assert.Equal(t, "target", current.Name)

		// Clean up
		os.RemoveAll(targetEnv.Path)
		os.Remove(filepath.Join(envswitchDir, "current.lock"))
	})

	t.Run("switches from one environment to another", func(t *testing.T) {
		// Create source environment
		sourceEnv := &environment.Environment{
			Name:      "source",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars:   make(map[string]string),
			Path:      filepath.Join(envsDir, "source"),
		}
		os.MkdirAll(sourceEnv.Path, 0755)
		err := sourceEnv.Save()
		require.NoError(t, err)

		// Create target environment
		targetEnv := &environment.Environment{
			Name:      "destination",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars:   make(map[string]string),
			Path:      filepath.Join(envsDir, "destination"),
		}
		os.MkdirAll(targetEnv.Path, 0755)
		err = targetEnv.Save()
		require.NoError(t, err)

		// Set source as current
		err = environment.SetCurrentEnvironment("source")
		require.NoError(t, err)

		// Switch to destination
		err = runSwitch(switchCmd, []string{"destination"})
		require.NoError(t, err)

		// Verify switch
		current, err := environment.GetCurrentEnvironment()
		require.NoError(t, err)
		assert.Equal(t, "destination", current.Name)

		// Clean up
		os.RemoveAll(sourceEnv.Path)
		os.RemoveAll(targetEnv.Path)
		os.Remove(filepath.Join(envswitchDir, "current.lock"))
	})

	t.Run("shows message when already on target environment", func(t *testing.T) {
		// Create environment
		env := &environment.Environment{
			Name:      "already-active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars:   make(map[string]string),
			Path:      filepath.Join(envsDir, "already-active"),
		}
		os.MkdirAll(env.Path, 0755)
		err := env.Save()
		require.NoError(t, err)

		// Set as current
		err = environment.SetCurrentEnvironment("already-active")
		require.NoError(t, err)

		// Try to switch to same environment
		err = runSwitch(switchCmd, []string{"already-active"})
		require.NoError(t, err)

		// Clean up
		os.RemoveAll(env.Path)
		os.Remove(filepath.Join(envswitchDir, "current.lock"))
	})

	t.Run("performs dry run without making changes", func(t *testing.T) {
		// Create environments
		env1 := &environment.Environment{
			Name:      "env1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars:   make(map[string]string),
			Path:      filepath.Join(envsDir, "env1"),
		}
		os.MkdirAll(env1.Path, 0755)
		err := env1.Save()
		require.NoError(t, err)

		env2 := &environment.Environment{
			Name:      "env2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars:   make(map[string]string),
			Path:      filepath.Join(envsDir, "env2"),
		}
		os.MkdirAll(env2.Path, 0755)
		err = env2.Save()
		require.NoError(t, err)

		// Set env1 as current
		err = environment.SetCurrentEnvironment("env1")
		require.NoError(t, err)

		// Dry run switch to env2
		switchDryRun = true
		defer func() { switchDryRun = false }()

		err = runSwitch(switchCmd, []string{"env2"})
		require.NoError(t, err)

		// Verify current didn't change
		current, err := environment.GetCurrentEnvironment()
		require.NoError(t, err)
		assert.Equal(t, "env1", current.Name)

		// Clean up
		os.RemoveAll(env1.Path)
		os.RemoveAll(env2.Path)
		os.Remove(filepath.Join(envswitchDir, "current.lock"))
	})

	t.Run("shows verification message when verify flag is set", func(t *testing.T) {
		// Create environment
		env := &environment.Environment{
			Name:      "verify-env",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars:   make(map[string]string),
			Path:      filepath.Join(envsDir, "verify-env"),
		}
		os.MkdirAll(env.Path, 0755)
		err := env.Save()
		require.NoError(t, err)

		// Set verify flag
		switchVerify = true
		defer func() { switchVerify = false }()

		err = runSwitch(switchCmd, []string{"verify-env"})
		require.NoError(t, err)

		// Clean up
		os.RemoveAll(env.Path)
		os.Remove(filepath.Join(envswitchDir, "current.lock"))
	})

	t.Run("returns error for non-existent environment", func(t *testing.T) {
		err := runSwitch(switchCmd, []string{"non-existent"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load environment")
	})

	t.Run("handles switching from no current environment", func(t *testing.T) {
		// Create target environment
		env := &environment.Environment{
			Name:      "first-switch",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars:   make(map[string]string),
			Path:      filepath.Join(envsDir, "first-switch"),
		}
		os.MkdirAll(env.Path, 0755)
		err := env.Save()
		require.NoError(t, err)

		// Make sure no current environment is set
		os.Remove(filepath.Join(envswitchDir, "current.lock"))

		// Switch to environment
		err = runSwitch(switchCmd, []string{"first-switch"})
		require.NoError(t, err)

		// Verify it's now current
		current, err := environment.GetCurrentEnvironment()
		require.NoError(t, err)
		assert.Equal(t, "first-switch", current.Name)

		// Clean up
		os.RemoveAll(env.Path)
		os.Remove(filepath.Join(envswitchDir, "current.lock"))
	})

	t.Run("handles tools with missing configuration gracefully", func(t *testing.T) {
		// Create source environment with kubectl enabled (but no .kube directory)
		sourceEnv := &environment.Environment{
			Name:      "source-with-missing-tool",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools: map[string]environment.ToolConfig{
				"kubectl": {
					Enabled: true,
				},
			},
			EnvVars: make(map[string]string),
			Path:    filepath.Join(envsDir, "source-with-missing-tool"),
		}
		os.MkdirAll(sourceEnv.Path, 0755)
		err := sourceEnv.Save()
		require.NoError(t, err)

		// Create target environment
		targetEnv := &environment.Environment{
			Name:      "target-simple",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars:   make(map[string]string),
			Path:      filepath.Join(envsDir, "target-simple"),
		}
		os.MkdirAll(targetEnv.Path, 0755)
		err = targetEnv.Save()
		require.NoError(t, err)

		// Set source as current
		err = environment.SetCurrentEnvironment("source-with-missing-tool")
		require.NoError(t, err)

		// Switch should succeed despite kubectl config missing
		err = runSwitch(switchCmd, []string{"target-simple"})
		require.NoError(t, err)

		// Verify switch succeeded
		current, err := environment.GetCurrentEnvironment()
		require.NoError(t, err)
		assert.Equal(t, "target-simple", current.Name)

		// Clean up
		os.RemoveAll(sourceEnv.Path)
		os.RemoveAll(targetEnv.Path)
		os.Remove(filepath.Join(envswitchDir, "current.lock"))
	})

	t.Run("handles invalid snapshots gracefully during restore", func(t *testing.T) {
		// Create target environment with invalid/empty snapshot
		targetEnv := &environment.Environment{
			Name:      "target-with-invalid-snapshot",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools: map[string]environment.ToolConfig{
				"kubectl": {
					Enabled: true,
				},
			},
			EnvVars: make(map[string]string),
			Path:    filepath.Join(envsDir, "target-with-invalid-snapshot"),
		}
		os.MkdirAll(targetEnv.Path, 0755)

		// Create empty snapshot directory (invalid)
		emptySnapshotPath := filepath.Join(targetEnv.Path, "snapshots", "kubectl")
		os.MkdirAll(emptySnapshotPath, 0755)

		err := targetEnv.Save()
		require.NoError(t, err)

		// Switch should succeed despite invalid snapshot
		err = runSwitch(switchCmd, []string{"target-with-invalid-snapshot"})
		require.NoError(t, err)

		// Verify switch succeeded
		current, err := environment.GetCurrentEnvironment()
		require.NoError(t, err)
		assert.Equal(t, "target-with-invalid-snapshot", current.Name)

		// Clean up
		os.RemoveAll(targetEnv.Path)
		os.Remove(filepath.Join(envswitchDir, "current.lock"))
	})
}

func TestSwitchCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "switch <name>", switchCmd.Use)
		assert.NotEmpty(t, switchCmd.Short)
		assert.NotEmpty(t, switchCmd.Long)
	})

	t.Run("has verify flag", func(t *testing.T) {
		flag := switchCmd.Flags().Lookup("verify")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})

	t.Run("has dry-run flag", func(t *testing.T) {
		flag := switchCmd.Flags().Lookup("dry-run")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})

	t.Run("requires exactly one argument", func(t *testing.T) {
		// Setup test environment
		originalHome := os.Getenv("HOME")
		tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalHome)

		// Initialize envswitch directory
		envswitchDir := filepath.Join(tmpDir, ".envswitch")
		envsDir := filepath.Join(envswitchDir, "environments")
		os.MkdirAll(envsDir, 0755)

		// Test with no arguments
		err = switchCmd.Args(switchCmd, []string{})
		assert.Error(t, err)

		// Test with two arguments
		err = switchCmd.Args(switchCmd, []string{"env1", "env2"})
		assert.Error(t, err)

		// Test with one argument
		err = switchCmd.Args(switchCmd, []string{"env1"})
		assert.NoError(t, err)
	})
}
