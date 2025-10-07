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
	// Create a temporary directory for testing
	tempHome := t.TempDir()

	// Save original home and restore after test
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})

	os.Setenv("HOME", tempHome)

	// Initialize envswitch directory structure
	envswitchDir := filepath.Join(tempHome, ".envswitch")
	err := os.MkdirAll(filepath.Join(envswitchDir, "environments"), 0755)
	require.NoError(t, err)

	t.Run("fails when no active environment", func(t *testing.T) {
		// No current.lock file exists
		err := runSave(saveCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active environment")
	})

	t.Run("saves current state to active environment", func(t *testing.T) {
		// Create a test environment
		envPath := filepath.Join(envswitchDir, "environments", "test-env")
		err := os.MkdirAll(filepath.Join(envPath, "snapshots"), 0755)
		require.NoError(t, err)

		// Create environment metadata
		env := &environment.Environment{
			Name:        "test-env",
			Description: "Test environment",
			Tools:       make(map[string]environment.ToolConfig),
			EnvVars:     make(map[string]string),
			Path:        envPath,
		}

		// Enable kubectl for testing
		env.Tools["kubectl"] = environment.ToolConfig{
			Enabled:      true,
			SnapshotPath: filepath.Join("snapshots", "kubectl"),
			Metadata:     make(map[string]interface{}),
		}

		err = env.Save()
		require.NoError(t, err)

		// Set as current environment
		err = environment.SetCurrentEnvironment("test-env")
		require.NoError(t, err)

		// Create a kubectl config to snapshot
		kubeDir := filepath.Join(tempHome, ".kube")
		err = os.MkdirAll(kubeDir, 0755)
		require.NoError(t, err)
		kubeConfig := filepath.Join(kubeDir, "config")
		err = os.WriteFile(kubeConfig, []byte("test-config-content\n"), 0644)
		require.NoError(t, err)

		// Run save
		err = runSave(saveCmd, []string{})
		require.NoError(t, err)

		// Verify snapshot was created
		snapshotPath := filepath.Join(envPath, "snapshots", "kubectl", "config")
		assert.FileExists(t, snapshotPath)

		// Verify snapshot content
		data, err := os.ReadFile(snapshotPath)
		require.NoError(t, err)
		assert.Equal(t, "test-config-content\n", string(data))
	})

	t.Run("updates existing snapshot", func(t *testing.T) {
		// Create environment
		envPath := filepath.Join(envswitchDir, "environments", "update-env")
		err := os.MkdirAll(filepath.Join(envPath, "snapshots", "kubectl"), 0755)
		require.NoError(t, err)

		env := &environment.Environment{
			Name:        "update-env",
			Description: "Update test environment",
			Tools:       make(map[string]environment.ToolConfig),
			EnvVars:     make(map[string]string),
			Path:        envPath,
		}

		env.Tools["kubectl"] = environment.ToolConfig{
			Enabled:      true,
			SnapshotPath: filepath.Join("snapshots", "kubectl"),
			Metadata:     make(map[string]interface{}),
		}

		err = env.Save()
		require.NoError(t, err)

		// Set as current
		err = environment.SetCurrentEnvironment("update-env")
		require.NoError(t, err)

		// Create initial snapshot
		kubeDir := filepath.Join(tempHome, ".kube")
		err = os.MkdirAll(kubeDir, 0755)
		require.NoError(t, err)
		kubeConfig := filepath.Join(kubeDir, "config")
		err = os.WriteFile(kubeConfig, []byte("initial-content\n"), 0644)
		require.NoError(t, err)

		// First save
		err = runSave(saveCmd, []string{})
		require.NoError(t, err)

		// Verify initial snapshot
		snapshotPath := filepath.Join(envPath, "snapshots", "kubectl", "config")
		data, err := os.ReadFile(snapshotPath)
		require.NoError(t, err)
		assert.Equal(t, "initial-content\n", string(data))

		// Update kubectl config
		err = os.WriteFile(kubeConfig, []byte("updated-content\n"), 0644)
		require.NoError(t, err)

		// Save again
		err = runSave(saveCmd, []string{})
		require.NoError(t, err)

		// Verify snapshot was updated
		data, err = os.ReadFile(snapshotPath)
		require.NoError(t, err)
		assert.Equal(t, "updated-content\n", string(data))
	})

	t.Run("handles multiple tools", func(t *testing.T) {
		// Create environment
		envPath := filepath.Join(envswitchDir, "environments", "multi-env")
		err := os.MkdirAll(filepath.Join(envPath, "snapshots"), 0755)
		require.NoError(t, err)

		env := &environment.Environment{
			Name:        "multi-env",
			Description: "Multi-tool environment",
			Tools:       make(map[string]environment.ToolConfig),
			EnvVars:     make(map[string]string),
			Path:        envPath,
		}

		// Enable multiple tools
		env.Tools["kubectl"] = environment.ToolConfig{
			Enabled:      true,
			SnapshotPath: filepath.Join("snapshots", "kubectl"),
			Metadata:     make(map[string]interface{}),
		}

		env.Tools["git"] = environment.ToolConfig{
			Enabled:      true,
			SnapshotPath: filepath.Join("snapshots", "git"),
			Metadata:     make(map[string]interface{}),
		}

		err = env.Save()
		require.NoError(t, err)

		// Set as current
		err = environment.SetCurrentEnvironment("multi-env")
		require.NoError(t, err)

		// Create kubectl config
		kubeDir := filepath.Join(tempHome, ".kube")
		err = os.MkdirAll(kubeDir, 0755)
		require.NoError(t, err)
		kubeConfig := filepath.Join(kubeDir, "config")
		err = os.WriteFile(kubeConfig, []byte("kubectl-config\n"), 0644)
		require.NoError(t, err)

		// Create git config
		gitConfig := filepath.Join(tempHome, ".gitconfig")
		err = os.WriteFile(gitConfig, []byte("[user]\nname=Test\n"), 0644)
		require.NoError(t, err)

		// Run save
		err = runSave(saveCmd, []string{})
		require.NoError(t, err)

		// Verify both snapshots were created
		kubectlSnapshot := filepath.Join(envPath, "snapshots", "kubectl", "config")
		assert.FileExists(t, kubectlSnapshot)

		// Git saves as "gitconfig" without the dot
		gitSnapshot := filepath.Join(envPath, "snapshots", "git", "gitconfig")
		assert.FileExists(t, gitSnapshot)
	})

	t.Run("skips disabled tools", func(t *testing.T) {
		// Create environment
		envPath := filepath.Join(envswitchDir, "environments", "disabled-env")
		err := os.MkdirAll(filepath.Join(envPath, "snapshots"), 0755)
		require.NoError(t, err)

		env := &environment.Environment{
			Name:        "disabled-env",
			Description: "Disabled tools test",
			Tools:       make(map[string]environment.ToolConfig),
			EnvVars:     make(map[string]string),
			Path:        envPath,
		}

		// One enabled, one disabled
		env.Tools["kubectl"] = environment.ToolConfig{
			Enabled:      true,
			SnapshotPath: filepath.Join("snapshots", "kubectl"),
			Metadata:     make(map[string]interface{}),
		}

		env.Tools["aws"] = environment.ToolConfig{
			Enabled:      false, // Disabled
			SnapshotPath: filepath.Join("snapshots", "aws"),
			Metadata:     make(map[string]interface{}),
		}

		err = env.Save()
		require.NoError(t, err)

		// Set as current
		err = environment.SetCurrentEnvironment("disabled-env")
		require.NoError(t, err)

		// Create kubectl config
		kubeDir := filepath.Join(tempHome, ".kube")
		err = os.MkdirAll(kubeDir, 0755)
		require.NoError(t, err)
		kubeConfig := filepath.Join(kubeDir, "config")
		err = os.WriteFile(kubeConfig, []byte("kubectl-config\n"), 0644)
		require.NoError(t, err)

		// Run save
		err = runSave(saveCmd, []string{})
		require.NoError(t, err)

		// Verify kubectl snapshot was created
		kubectlSnapshot := filepath.Join(envPath, "snapshots", "kubectl", "config")
		assert.FileExists(t, kubectlSnapshot)

		// Verify aws snapshot was NOT created
		awsSnapshot := filepath.Join(envPath, "snapshots", "aws")
		_, err = os.Stat(awsSnapshot)
		// Should not exist or be empty
		if err == nil {
			// If directory exists, check it's empty
			entries, err := os.ReadDir(awsSnapshot)
			require.NoError(t, err)
			assert.Empty(t, entries, "AWS snapshot directory should be empty since tool is disabled")
		}
	})
}

func TestSaveCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "save", saveCmd.Use)
		assert.NotEmpty(t, saveCmd.Short)
		assert.NotEmpty(t, saveCmd.Long)
		assert.Contains(t, saveCmd.Short, "Save")
		assert.Contains(t, saveCmd.Short, "current")
	})

	t.Run("requires no arguments", func(t *testing.T) {
		// Save should work with no arguments
		assert.NotNil(t, saveCmd)
	})

	t.Run("has RunE function", func(t *testing.T) {
		assert.NotNil(t, saveCmd.RunE)
	})
}

// TestSaveIntegration is disabled in favor of TestSaveWorkflowSimple
// This version uses sub-tests which can be flaky on CI
// To re-enable, rename to TestSaveIntegration
func testSaveIntegrationDisabled(t *testing.T) {
	// This test cannot run in parallel due to global flag manipulation
	// and HOME environment variable changes

	// Create a temporary directory for testing
	tempHome := t.TempDir()

	// Save original home and restore after test
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})

	os.Setenv("HOME", tempHome)

	// Save and restore global flags
	origCreateFromCurrent := createFromCurrent
	origCreateEmpty := createEmpty
	origCreateFrom := createFrom
	origCreateDescription := createDescription
	defer func() {
		createFromCurrent = origCreateFromCurrent
		createEmpty = origCreateEmpty
		createFrom = origCreateFrom
		createDescription = origCreateDescription
	}()

	// Initialize envswitch
	envswitchDir := filepath.Join(tempHome, ".envswitch")
	err := os.MkdirAll(filepath.Join(envswitchDir, "environments"), 0755)
	require.NoError(t, err)

	t.Run("full workflow: create, modify, save", func(t *testing.T) {
		// Create .kube directory with initial config
		kubeDir := filepath.Join(tempHome, ".kube")
		err := os.MkdirAll(kubeDir, 0755)
		require.NoError(t, err)
		kubeConfig := filepath.Join(kubeDir, "config")
		err = os.WriteFile(kubeConfig, []byte("INITIAL_CONFIG\n"), 0644)
		require.NoError(t, err)

		// Create environment from current state
		createFromCurrent = true
		createEmpty = false
		createFrom = ""
		createDescription = "Integration test"

		err = runCreate(createCmd, []string{"integration-test"})
		require.NoError(t, err)

		// Verify environment was created and is active
		currentEnv, err := environment.GetCurrentEnvironment()
		require.NoError(t, err)
		require.NotNil(t, currentEnv)
		assert.Equal(t, "integration-test", currentEnv.Name)

		// Verify initial snapshot
		envPath := filepath.Join(envswitchDir, "environments", "integration-test")
		snapshotPath := filepath.Join(envPath, "snapshots", "kubectl", "config")
		data, err := os.ReadFile(snapshotPath)
		require.NoError(t, err)
		assert.Equal(t, "INITIAL_CONFIG\n", string(data))

		// Modify the kubectl config
		err = os.WriteFile(kubeConfig, []byte("MODIFIED_CONFIG\n"), 0644)
		require.NoError(t, err)

		// Save the changes
		err = runSave(saveCmd, []string{})
		require.NoError(t, err)

		// Verify snapshot was updated
		data, err = os.ReadFile(snapshotPath)
		require.NoError(t, err)
		assert.Equal(t, "MODIFIED_CONFIG\n", string(data))
	})

	t.Run("save preserves environment metadata", func(t *testing.T) {
		// Create environment
		envPath := filepath.Join(envswitchDir, "environments", "metadata-test")
		err := os.MkdirAll(filepath.Join(envPath, "snapshots"), 0755)
		require.NoError(t, err)

		env := &environment.Environment{
			Name:        "metadata-test",
			Description: "Test description that should be preserved",
			Tools:       make(map[string]environment.ToolConfig),
			EnvVars:     make(map[string]string),
			Path:        envPath,
		}

		env.Tools["kubectl"] = environment.ToolConfig{
			Enabled:      true,
			SnapshotPath: filepath.Join("snapshots", "kubectl"),
			Metadata:     make(map[string]interface{}),
		}

		err = env.Save()
		require.NoError(t, err)

		originalUpdatedAt := env.UpdatedAt

		// Set as current
		err = environment.SetCurrentEnvironment("metadata-test")
		require.NoError(t, err)

		// Create kubectl config
		kubeDir := filepath.Join(tempHome, ".kube")
		err = os.MkdirAll(kubeDir, 0755)
		require.NoError(t, err)
		kubeConfig := filepath.Join(kubeDir, "config")
		err = os.WriteFile(kubeConfig, []byte("test-config\n"), 0644)
		require.NoError(t, err)

		// Run save
		err = runSave(saveCmd, []string{})
		require.NoError(t, err)

		// Load environment and verify metadata was preserved
		loadedEnv, err := environment.LoadEnvironment("metadata-test")
		require.NoError(t, err)

		assert.Equal(t, "metadata-test", loadedEnv.Name)
		assert.Equal(t, "Test description that should be preserved", loadedEnv.Description)
		assert.True(t, loadedEnv.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
	})
}
