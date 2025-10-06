package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/internal/archive"
	"github.com/hugofrely/envswitch/internal/config"
	"github.com/hugofrely/envswitch/pkg/environment"
)

func TestSwitchWithVerifyAfterSwitch(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Initialize envswitch
	require.NoError(t, runInit(initCmd, []string{}))

	// Create config with verify_after_switch enabled
	cfg := config.DefaultConfig()
	cfg.VerifyAfterSwitch = true
	require.NoError(t, cfg.Save())

	// Create test environments
	createTestEnv(t, tempDir, "test-verify")
	createTestEnv(t, tempDir, "target-verify")

	// Switch (should verify automatically)
	err := runSwitch(switchCmd, []string{"target-verify"})
	assert.NoError(t, err)
}

func TestSwitchWithBackupRetention(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Initialize
	require.NoError(t, runInit(initCmd, []string{}))

	// Configure retention to keep only 2 backups
	cfg := config.DefaultConfig()
	cfg.BackupRetention = 2
	require.NoError(t, cfg.Save())

	// Create environments
	env1 := createTestEnv(t, tempDir, "env1")
	env2 := createTestEnv(t, tempDir, "env2")
	env3 := createTestEnv(t, tempDir, "env3")

	// Switch multiple times to create backups
	require.NoError(t, environment.SetCurrentEnvironment("env1"))
	time.Sleep(1100 * time.Millisecond)

	err := runSwitch(switchCmd, []string{"env2"})
	require.NoError(t, err)
	time.Sleep(1100 * time.Millisecond)

	err = runSwitch(switchCmd, []string{"env3"})
	require.NoError(t, err)
	time.Sleep(1100 * time.Millisecond)

	err = runSwitch(switchCmd, []string{"env1"})
	require.NoError(t, err)

	// Check that old backups were cleaned up
	archives, err := archive.ListArchives()
	require.NoError(t, err)

	// Should have at most 2 archives (retention policy)
	assert.LessOrEqual(t, len(archives), 2)

	// Cleanup
	_ = os.RemoveAll(env1.Path)
	_ = os.RemoveAll(env2.Path)
	_ = os.RemoveAll(env3.Path)
}

func TestSwitchWithExcludeTools(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Initialize
	require.NoError(t, runInit(initCmd, []string{}))

	// Configure to exclude docker and kubectl
	cfg := config.DefaultConfig()
	cfg.ExcludeTools = []string{"docker", "kubectl"}
	require.NoError(t, cfg.Save())

	// Create environments
	env1 := createTestEnv(t, tempDir, "exclude-test1")
	env2 := createTestEnv(t, tempDir, "exclude-test2")

	// Enable all tools
	env1.Tools["docker"] = environment.ToolConfig{Enabled: true}
	env1.Tools["kubectl"] = environment.ToolConfig{Enabled: true}
	env1.Tools["git"] = environment.ToolConfig{Enabled: true}
	env1.Save()

	env2.Tools["docker"] = environment.ToolConfig{Enabled: true}
	env2.Tools["kubectl"] = environment.ToolConfig{Enabled: true}
	env2.Tools["git"] = environment.ToolConfig{Enabled: true}
	env2.Save()

	// Switch
	require.NoError(t, environment.SetCurrentEnvironment("exclude-test1"))
	err := runSwitch(switchCmd, []string{"exclude-test2"})

	// Should succeed even though docker/kubectl might not be installed
	// because they're excluded
	assert.NoError(t, err)

	// Cleanup
	_ = os.RemoveAll(env1.Path)
	_ = os.RemoveAll(env2.Path)
}

func TestGetToolRegistryFiltering(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("returns all tools when no exclusions", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.ExcludeTools = []string{}
		require.NoError(t, cfg.Save())

		tools := getToolRegistry()
		assert.Len(t, tools, 5) // git, aws, gcloud, kubectl, docker
		assert.Contains(t, tools, "git")
		assert.Contains(t, tools, "aws")
		assert.Contains(t, tools, "gcloud")
		assert.Contains(t, tools, "kubectl")
		assert.Contains(t, tools, "docker")
	})

	t.Run("excludes specified tools", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.ExcludeTools = []string{"docker", "kubectl"}
		require.NoError(t, cfg.Save())

		tools := getToolRegistry()
		assert.Len(t, tools, 3)
		assert.Contains(t, tools, "git")
		assert.Contains(t, tools, "aws")
		assert.Contains(t, tools, "gcloud")
		assert.NotContains(t, tools, "docker")
		assert.NotContains(t, tools, "kubectl")
	})

	t.Run("excludes all tools", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.ExcludeTools = []string{"git", "aws", "gcloud", "kubectl", "docker"}
		require.NoError(t, cfg.Save())

		tools := getToolRegistry()
		assert.Len(t, tools, 0)
	})
}

func TestConfigLoadingInSwitch(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Initialize
	require.NoError(t, runInit(initCmd, []string{}))

	// Test with missing config (should use defaults)
	os.Remove(config.GetConfigPath())

	env1 := createTestEnv(t, tempDir, "config-test1")
	env2 := createTestEnv(t, tempDir, "config-test2")

	require.NoError(t, environment.SetCurrentEnvironment("config-test1"))
	err := runSwitch(switchCmd, []string{"config-test2"})

	// Should work with default config
	assert.NoError(t, err)

	// Cleanup
	_ = os.RemoveAll(env1.Path)
	_ = os.RemoveAll(env2.Path)
}

func TestBackupRetentionZero(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Initialize
	require.NoError(t, runInit(initCmd, []string{}))

	// Set retention to 0 (no cleanup)
	cfg := config.DefaultConfig()
	cfg.BackupRetention = 0
	require.NoError(t, cfg.Save())

	// Create environments and switch
	env1 := createTestEnv(t, tempDir, "ret-zero1")
	env2 := createTestEnv(t, tempDir, "ret-zero2")

	require.NoError(t, environment.SetCurrentEnvironment("ret-zero1"))
	time.Sleep(1100 * time.Millisecond)

	err := runSwitch(switchCmd, []string{"ret-zero2"})
	require.NoError(t, err)

	// With retention=0, no cleanup should occur
	archives, _ := archive.ListArchives()
	// At least 1 archive should exist (the one we just created)
	assert.GreaterOrEqual(t, len(archives), 1)

	// Cleanup
	_ = os.RemoveAll(env1.Path)
	_ = os.RemoveAll(env2.Path)
}

// Helper to create a test environment
func createTestEnv(t *testing.T, tempDir, name string) *environment.Environment {
	t.Helper()

	envswitchDir := filepath.Join(tempDir, ".envswitch")
	envPath := filepath.Join(envswitchDir, "environments", name)
	os.MkdirAll(envPath, 0755)

	env := &environment.Environment{
		Name:      name,
		Path:      envPath,
		CreatedAt: time.Now(),
		Tools:     make(map[string]environment.ToolConfig),
	}

	require.NoError(t, env.Save())
	return env
}
