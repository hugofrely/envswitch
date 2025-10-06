package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hugofrely/envswitch/pkg/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunShow(t *testing.T) {
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

	t.Run("shows environment details", func(t *testing.T) {
		// Create test environment
		env := &environment.Environment{
			Name:        "test-env",
			Description: "Test environment",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			LastUsed:    time.Now().Add(-24 * time.Hour),
			Tools:       make(map[string]environment.ToolConfig),
			EnvVars:     make(map[string]string),
			Path:        filepath.Join(envsDir, "test-env"),
		}
		os.MkdirAll(env.Path, 0755)
		err := env.Save()
		require.NoError(t, err)

		err = runShow(showCmd, []string{"test-env"})
		require.NoError(t, err)

		// Clean up
		os.RemoveAll(env.Path)
	})

	t.Run("shows environment with tools", func(t *testing.T) {
		env := &environment.Environment{
			Name:      "with-tools",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools: map[string]environment.ToolConfig{
				"gcloud": {
					Enabled:      true,
					SnapshotPath: "snapshots/gcloud",
					Metadata: map[string]interface{}{
						"account": "user@example.com",
						"project": "my-project",
					},
				},
				"kubectl": {
					Enabled:      true,
					SnapshotPath: "snapshots/kubectl",
					Metadata: map[string]interface{}{
						"context": "minikube",
					},
				},
				"aws": {
					Enabled:      false, // Should not be shown
					SnapshotPath: "snapshots/aws",
					Metadata:     map[string]interface{}{},
				},
			},
			EnvVars: make(map[string]string),
			Path:    filepath.Join(envsDir, "with-tools"),
		}
		os.MkdirAll(env.Path, 0755)
		err := env.Save()
		require.NoError(t, err)

		err = runShow(showCmd, []string{"with-tools"})
		require.NoError(t, err)

		// Clean up
		os.RemoveAll(env.Path)
	})

	t.Run("shows environment with environment variables", func(t *testing.T) {
		env := &environment.Environment{
			Name:      "with-vars",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars: map[string]string{
				"AWS_REGION":   "us-east-1",
				"DEBUG":        "true",
				"API_ENDPOINT": "https://api.example.com",
			},
			Path: filepath.Join(envsDir, "with-vars"),
		}
		os.MkdirAll(env.Path, 0755)
		err := env.Save()
		require.NoError(t, err)

		err = runShow(showCmd, []string{"with-vars"})
		require.NoError(t, err)

		// Clean up
		os.RemoveAll(env.Path)
	})

	t.Run("shows environment with tags", func(t *testing.T) {
		env := &environment.Environment{
			Name:      "with-tags",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars:   make(map[string]string),
			Tags:      []string{"production", "client-a", "critical"},
			Path:      filepath.Join(envsDir, "with-tags"),
		}
		os.MkdirAll(env.Path, 0755)
		err := env.Save()
		require.NoError(t, err)

		err = runShow(showCmd, []string{"with-tags"})
		require.NoError(t, err)

		// Clean up
		os.RemoveAll(env.Path)
	})

	t.Run("shows environment with snapshot timestamp", func(t *testing.T) {
		env := &environment.Environment{
			Name:         "with-snapshot",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastSnapshot: time.Now().Add(-2 * time.Hour),
			Tools:        make(map[string]environment.ToolConfig),
			EnvVars:      make(map[string]string),
			Path:         filepath.Join(envsDir, "with-snapshot"),
		}
		os.MkdirAll(env.Path, 0755)
		err := env.Save()
		require.NoError(t, err)

		err = runShow(showCmd, []string{"with-snapshot"})
		require.NoError(t, err)

		// Clean up
		os.RemoveAll(env.Path)
	})

	t.Run("returns error for non-existent environment", func(t *testing.T) {
		err := runShow(showCmd, []string{"non-existent"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load environment")
	})

	t.Run("shows environment without optional fields", func(t *testing.T) {
		env := &environment.Environment{
			Name:      "minimal",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars:   make(map[string]string),
			Path:      filepath.Join(envsDir, "minimal"),
		}
		os.MkdirAll(env.Path, 0755)
		err := env.Save()
		require.NoError(t, err)

		err = runShow(showCmd, []string{"minimal"})
		require.NoError(t, err)

		// Clean up
		os.RemoveAll(env.Path)
	})
}

func TestShowCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "show <name>", showCmd.Use)
		assert.NotEmpty(t, showCmd.Short)
		assert.NotEmpty(t, showCmd.Long)
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

		// Test with no arguments - should fail validation
		err = showCmd.Args(showCmd, []string{})
		assert.Error(t, err)

		// Test with two arguments - should fail validation
		err = showCmd.Args(showCmd, []string{"env1", "env2"})
		assert.Error(t, err)

		// Test with one argument - should pass validation
		err = showCmd.Args(showCmd, []string{"env1"})
		assert.NoError(t, err)
	})
}
