package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/pkg/environment"
)

// TestSaveWorkflowSimple is a simplified integration test for the save command
// that doesn't use sub-tests to avoid state issues on CI
func TestSaveWorkflowSimple(t *testing.T) {
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

	// ===== Step 1: Create environment with initial config =====
	t.Log("Step 1: Create environment with initial config")

	// Create .kube directory with initial config
	kubeDir := filepath.Join(tempHome, ".kube")
	err = os.MkdirAll(kubeDir, 0755)
	require.NoError(t, err)

	kubeConfig := filepath.Join(kubeDir, "config")
	err = os.WriteFile(kubeConfig, []byte("INITIAL_CONFIG\n"), 0644)
	require.NoError(t, err)

	// Create environment from current state
	createFromCurrent = true
	createEmpty = false
	createFrom = ""
	createDescription = "Integration test"

	err = runCreate(createCmd, []string{"test-save"})
	require.NoError(t, err)

	// Verify environment was created and is active
	currentEnv, err := environment.GetCurrentEnvironment()
	require.NoError(t, err)
	require.NotNil(t, currentEnv)
	assert.Equal(t, "test-save", currentEnv.Name)

	// Load environment and check for snapshot
	envPath := filepath.Join(envswitchDir, "environments", "test-save")
	snapshotPath := filepath.Join(envPath, "snapshots", "kubectl", "config")

	// Check if kubectl snapshot was created
	if _, err := os.Stat(snapshotPath); err != nil {
		// kubectl not installed, create snapshot manually for testing
		t.Log("⚠️  kubectl not installed, creating manual snapshot for testing")
		if err := os.MkdirAll(filepath.Dir(snapshotPath), 0755); err != nil {
			t.Fatalf("Failed to create snapshot dir: %v", err)
		}
		if err := os.WriteFile(snapshotPath, []byte("INITIAL_CONFIG\n"), 0644); err != nil {
			t.Fatalf("Failed to create manual snapshot: %v", err)
		}
		// Update environment to enable kubectl
		testEnv, err := environment.LoadEnvironment("test-save")
		require.NoError(t, err)
		testEnv.Tools["kubectl"] = environment.ToolConfig{
			Enabled:      true,
			SnapshotPath: filepath.Join("snapshots", "kubectl"),
			Metadata:     make(map[string]interface{}),
		}
		require.NoError(t, testEnv.Save())
	}

	// Verify snapshot content
	data, err := os.ReadFile(snapshotPath)
	require.NoError(t, err)
	assert.Equal(t, "INITIAL_CONFIG\n", string(data))

	t.Log("✅ Environment created with INITIAL_CONFIG")

	// ===== Step 2: Modify config and save =====
	t.Log("Step 2: Modify config and use save command")

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

	t.Log("✅ Config modified and saved successfully")

	// ===== Step 3: Verify metadata is preserved =====
	t.Log("Step 3: Verify environment metadata is preserved")

	// Load environment and check metadata
	loadedEnv, err := environment.LoadEnvironment("test-save")
	require.NoError(t, err)

	assert.Equal(t, "test-save", loadedEnv.Name)
	assert.Equal(t, "Integration test", loadedEnv.Description)
	assert.False(t, loadedEnv.UpdatedAt.IsZero(), "UpdatedAt should be set")

	t.Log("✅ Metadata preserved correctly")

	// ===== Step 4: Multiple saves work correctly =====
	t.Log("Step 4: Test multiple saves")

	// Modify again
	err = os.WriteFile(kubeConfig, []byte("THIRD_CONFIG\n"), 0644)
	require.NoError(t, err)

	// Save again
	err = runSave(saveCmd, []string{})
	require.NoError(t, err)

	// Verify third snapshot
	data, err = os.ReadFile(snapshotPath)
	require.NoError(t, err)
	assert.Equal(t, "THIRD_CONFIG\n", string(data))

	t.Log("✅ Multiple saves work correctly")

	t.Log("🎉 All save integration steps passed!")
}
