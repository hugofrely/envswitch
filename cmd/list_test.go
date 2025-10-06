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

func TestRunList(t *testing.T) {
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

	t.Run("shows message when no environments exist", func(t *testing.T) {
		err := runList(listCmd, []string{})
		require.NoError(t, err)
	})

	t.Run("lists all environments", func(t *testing.T) {
		// Create test environments
		env1 := &environment.Environment{
			Name:        "env1",
			Description: "First environment",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Tools:       make(map[string]environment.ToolConfig),
			EnvVars:     make(map[string]string),
			Path:        filepath.Join(envsDir, "env1"),
		}
		os.MkdirAll(env1.Path, 0755)
		err := env1.Save()
		require.NoError(t, err)

		env2 := &environment.Environment{
			Name:        "env2",
			Description: "Second environment",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Tools:       make(map[string]environment.ToolConfig),
			EnvVars:     make(map[string]string),
			Path:        filepath.Join(envsDir, "env2"),
		}
		os.MkdirAll(env2.Path, 0755)
		err = env2.Save()
		require.NoError(t, err)

		err = runList(listCmd, []string{})
		require.NoError(t, err)

		// Clean up
		os.RemoveAll(env1.Path)
		os.RemoveAll(env2.Path)
	})

	t.Run("shows active environment marker", func(t *testing.T) {
		// Create environment
		env := &environment.Environment{
			Name:      "active-env",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars:   make(map[string]string),
			Path:      filepath.Join(envsDir, "active-env"),
		}
		os.MkdirAll(env.Path, 0755)
		err := env.Save()
		require.NoError(t, err)

		// Set as current
		err = environment.SetCurrentEnvironment("active-env")
		require.NoError(t, err)

		err = runList(listCmd, []string{})
		require.NoError(t, err)

		// Clean up
		os.RemoveAll(env.Path)
		os.Remove(filepath.Join(envswitchDir, "current.lock"))
	})

	t.Run("shows detailed information when flag is set", func(t *testing.T) {
		// Create environment with tools
		env := &environment.Environment{
			Name:      "detailed-env",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			LastUsed:  time.Now().Add(-24 * time.Hour),
			Tools: map[string]environment.ToolConfig{
				"gcloud": {
					Enabled:      true,
					SnapshotPath: "snapshots/gcloud",
					Metadata:     map[string]interface{}{"account": "test@example.com"},
				},
				"kubectl": {
					Enabled:      true,
					SnapshotPath: "snapshots/kubectl",
					Metadata:     map[string]interface{}{},
				},
			},
			EnvVars: make(map[string]string),
			Path:    filepath.Join(envsDir, "detailed-env"),
		}
		os.MkdirAll(env.Path, 0755)
		err := env.Save()
		require.NoError(t, err)

		// Set detailed flag
		listDetailed = true
		defer func() { listDetailed = false }()

		err = runList(listCmd, []string{})
		require.NoError(t, err)

		// Clean up
		os.RemoveAll(env.Path)
	})

	t.Run("handles environments without description", func(t *testing.T) {
		env := &environment.Environment{
			Name:      "no-desc",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tools:     make(map[string]environment.ToolConfig),
			EnvVars:   make(map[string]string),
			Path:      filepath.Join(envsDir, "no-desc"),
		}
		os.MkdirAll(env.Path, 0755)
		err := env.Save()
		require.NoError(t, err)

		err = runList(listCmd, []string{})
		require.NoError(t, err)

		// Clean up
		os.RemoveAll(env.Path)
	})
}

func TestFormatTimeAgo(t *testing.T) {
	// Test with a time 2 hours ago
	twoHoursAgo := time.Now().Add(-2 * time.Hour)
	formatted := formatTimeAgo(twoHoursAgo)
	assert.Contains(t, formatted, "ago")

	// Test with a time far in the past
	longAgo := time.Date(2020, 1, 15, 10, 30, 0, 0, time.UTC)
	formatted = formatTimeAgo(longAgo)
	assert.Contains(t, formatted, "ago")

	// Test with future time
	future := time.Now().Add(2 * time.Hour)
	formatted = formatTimeAgo(future)
	assert.Contains(t, formatted, "from now")
}

func TestListCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "list", listCmd.Use)
		assert.Contains(t, listCmd.Aliases, "ls")
		assert.NotEmpty(t, listCmd.Short)
		assert.NotEmpty(t, listCmd.Long)
	})

	t.Run("has detailed flag", func(t *testing.T) {
		flag := listCmd.Flags().Lookup("detailed")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})
}
