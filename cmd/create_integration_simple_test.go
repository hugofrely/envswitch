package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hugofrely/envswitch/pkg/environment"
)

// TestCreateAndSwitchSimple is a simplified integration test
// that doesn't use sub-tests to avoid state issues on CI
func TestCreateAndSwitchSimple(t *testing.T) {
	// This test cannot run in parallel due to global flag manipulation
	// and HOME environment variable changes

	// Setup test environment
	originalHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

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

	// Create .kube directory
	kubeDir := filepath.Join(tempHome, ".kube")
	if err := os.MkdirAll(kubeDir, 0755); err != nil {
		t.Fatalf("Failed to create .kube directory: %v", err)
	}

	kubeConfig := filepath.Join(kubeDir, "config")

	// Initialize envswitch
	if err := os.MkdirAll(filepath.Join(tempHome, ".envswitch", "environments"), 0755); err != nil {
		t.Fatalf("Failed to create .envswitch directory: %v", err)
	}

	// ===== Step 1: Create work environment =====
	t.Log("Step 1: Create work environment with TEST_A")

	// Write test content A
	if err := os.WriteFile(kubeConfig, []byte("TEST_A\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create work environment
	createFromCurrent = true
	createEmpty = false
	createFrom = ""
	createDescription = "Work environment"

	if err := runCreate(createCmd, []string{"work"}); err != nil {
		t.Fatalf("Failed to create work environment: %v", err)
	}

	// Verify work environment was created and is active
	currentEnv, err := environment.GetCurrentEnvironment()
	if err != nil {
		t.Fatalf("Failed to get current environment: %v", err)
	}
	if currentEnv == nil || currentEnv.Name != "work" {
		t.Fatalf("Expected current environment to be 'work', got %v", currentEnv)
	}

	// Load work environment
	workEnv, err := environment.LoadEnvironment("work")
	if err != nil {
		t.Fatalf("Failed to load work environment: %v", err)
	}

	// Check if kubectl snapshot was created (depends on kubectl being installed)
	workSnapshot := filepath.Join(workEnv.Path, "snapshots", "kubectl", "config")
	if _, err := os.Stat(workSnapshot); err == nil {
		// kubectl was captured, verify content
		data, err := os.ReadFile(workSnapshot)
		if err != nil {
			t.Fatalf("Failed to read work snapshot: %v", err)
		}
		if string(data) != "TEST_A\n" {
			t.Errorf("Expected work snapshot 'TEST_A\\n', got %q", string(data))
		}
		t.Log("✅ Work environment created with TEST_A snapshot")
	} else {
		// kubectl not installed (like on CI), create snapshot manually for testing
		t.Log("⚠️  kubectl not installed, creating manual snapshot for testing")
		if err := os.MkdirAll(filepath.Dir(workSnapshot), 0755); err != nil {
			t.Fatalf("Failed to create snapshot dir: %v", err)
		}
		if err := os.WriteFile(workSnapshot, []byte("TEST_A\n"), 0644); err != nil {
			t.Fatalf("Failed to create manual snapshot: %v", err)
		}
		// Update environment to enable kubectl
		workEnv.Tools["kubectl"] = environment.ToolConfig{
			Enabled:      true,
			SnapshotPath: filepath.Join("snapshots", "kubectl"),
			Metadata:     make(map[string]interface{}),
		}
		if err := workEnv.Save(); err != nil {
			t.Fatalf("Failed to save environment: %v", err)
		}
		t.Log("✅ Work environment created with manual TEST_A snapshot")
	}

	// ===== Step 2: Create perso environment =====
	t.Log("Step 2: Create perso environment with TEST_B")

	// Write test content B
	if err := os.WriteFile(kubeConfig, []byte("TEST_B\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create perso environment
	createFromCurrent = true
	createEmpty = false
	createFrom = ""
	createDescription = "Personal environment"

	if err := runCreate(createCmd, []string{"perso"}); err != nil {
		t.Fatalf("Failed to create perso environment: %v", err)
	}

	// Verify perso environment was created and is active
	currentEnv, err = environment.GetCurrentEnvironment()
	if err != nil {
		t.Fatalf("Failed to get current environment: %v", err)
	}
	if currentEnv == nil || currentEnv.Name != "perso" {
		t.Fatalf("Expected current environment to be 'perso', got %v", currentEnv)
	}

	// Load perso environment
	persoEnv, err := environment.LoadEnvironment("perso")
	if err != nil {
		t.Fatalf("Failed to load perso environment: %v", err)
	}

	// Check if kubectl snapshot was created
	persoSnapshot := filepath.Join(persoEnv.Path, "snapshots", "kubectl", "config")
	if _, err := os.Stat(persoSnapshot); err == nil {
		// kubectl was captured, verify content
		data, err := os.ReadFile(persoSnapshot)
		if err != nil {
			t.Fatalf("Failed to read perso snapshot: %v", err)
		}
		if string(data) != "TEST_B\n" {
			t.Errorf("Expected perso snapshot 'TEST_B\\n', got %q", string(data))
		}
		t.Log("✅ Perso environment created with TEST_B snapshot")
	} else {
		// kubectl not installed, create snapshot manually for testing
		t.Log("⚠️  kubectl not installed, creating manual snapshot for testing")
		if err := os.MkdirAll(filepath.Dir(persoSnapshot), 0755); err != nil {
			t.Fatalf("Failed to create snapshot dir: %v", err)
		}
		if err := os.WriteFile(persoSnapshot, []byte("TEST_B\n"), 0644); err != nil {
			t.Fatalf("Failed to create manual snapshot: %v", err)
		}
		// Update environment to enable kubectl
		persoEnv.Tools["kubectl"] = environment.ToolConfig{
			Enabled:      true,
			SnapshotPath: filepath.Join("snapshots", "kubectl"),
			Metadata:     make(map[string]interface{}),
		}
		if err := persoEnv.Save(); err != nil {
			t.Fatalf("Failed to save environment: %v", err)
		}
		t.Log("✅ Perso environment created with manual TEST_B snapshot")
	}

	// ===== Step 3: Verify snapshots are independent =====
	t.Log("Step 3: Verify both snapshots are independent")

	// Re-read work snapshot to ensure it still has TEST_A
	workData, err := os.ReadFile(workSnapshot)
	if err != nil {
		t.Fatalf("Failed to re-read work snapshot: %v", err)
	}
	if string(workData) != "TEST_A\n" {
		t.Errorf("Work snapshot changed! Expected 'TEST_A\\n', got %q", string(workData))
	}

	// Verify perso snapshot still has TEST_B
	persoData, err := os.ReadFile(persoSnapshot)
	if err != nil {
		t.Fatalf("Failed to re-read perso snapshot: %v", err)
	}
	if string(persoData) != "TEST_B\n" {
		t.Errorf("Perso snapshot changed! Expected 'TEST_B\\n', got %q", string(persoData))
	}

	t.Log("✅ Both snapshots are correctly independent")

	// ===== Step 4: Switch to work and verify restore =====
	t.Log("Step 4: Switch to work and verify restore")

	if err := runSwitch(switchCmd, []string{"work"}); err != nil {
		t.Fatalf("Failed to switch to work: %v", err)
	}

	// Verify current environment
	currentEnv, err = environment.GetCurrentEnvironment()
	if err != nil {
		t.Fatalf("Failed to get current environment: %v", err)
	}
	if currentEnv == nil || currentEnv.Name != "work" {
		t.Fatalf("Expected current environment to be 'work', got %v", currentEnv)
	}

	// Verify file was restored to TEST_A
	restoredWorkData, err := os.ReadFile(kubeConfig)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	if string(restoredWorkData) != "TEST_A\n" {
		t.Errorf("Expected restored content 'TEST_A\\n', got %q", string(restoredWorkData))
	}

	t.Log("✅ Successfully switched to work and restored TEST_A")

	// ===== Step 5: Switch to perso and verify restore =====
	t.Log("Step 5: Switch to perso and verify restore")

	if err := runSwitch(switchCmd, []string{"perso"}); err != nil {
		t.Fatalf("Failed to switch to perso: %v", err)
	}

	// Verify current environment
	currentEnv, err = environment.GetCurrentEnvironment()
	if err != nil {
		t.Fatalf("Failed to get current environment: %v", err)
	}
	if currentEnv == nil || currentEnv.Name != "perso" {
		t.Fatalf("Expected current environment to be 'perso', got %v", currentEnv)
	}

	// Verify file was restored to TEST_B
	restoredPersoData, err := os.ReadFile(kubeConfig)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	if string(restoredPersoData) != "TEST_B\n" {
		t.Errorf("Expected restored content 'TEST_B\\n', got %q", string(restoredPersoData))
	}

	t.Log("✅ Successfully switched to perso and restored TEST_B")
	t.Log("🎉 All integration test steps passed!")
}
