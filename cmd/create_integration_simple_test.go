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

	// Verify work snapshot contains TEST_A
	workEnv, err := environment.LoadEnvironment("work")
	if err != nil {
		t.Fatalf("Failed to load work environment: %v", err)
	}

	workSnapshot := filepath.Join(workEnv.Path, "snapshots", "kubectl", "config")
	data, err := os.ReadFile(workSnapshot)
	if err != nil {
		t.Fatalf("Failed to read work snapshot: %v", err)
	}
	if string(data) != "TEST_A\n" {
		t.Errorf("Expected work snapshot 'TEST_A\\n', got %q", string(data))
	}

	t.Log("✅ Work environment created successfully with TEST_A")

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

	// Verify perso snapshot contains TEST_B
	persoEnv, err := environment.LoadEnvironment("perso")
	if err != nil {
		t.Fatalf("Failed to load perso environment: %v", err)
	}

	persoSnapshot := filepath.Join(persoEnv.Path, "snapshots", "kubectl", "config")
	data, err = os.ReadFile(persoSnapshot)
	if err != nil {
		t.Fatalf("Failed to read perso snapshot: %v", err)
	}
	if string(data) != "TEST_B\n" {
		t.Errorf("Expected perso snapshot 'TEST_B\\n', got %q", string(data))
	}

	t.Log("✅ Perso environment created successfully with TEST_B")

	// ===== Step 3: Verify snapshots are independent =====
	t.Log("Step 3: Verify both snapshots are independent")

	// Re-read work snapshot to ensure it still has TEST_A
	data, err = os.ReadFile(workSnapshot)
	if err != nil {
		t.Fatalf("Failed to re-read work snapshot: %v", err)
	}
	if string(data) != "TEST_A\n" {
		t.Errorf("Work snapshot changed! Expected 'TEST_A\\n', got %q", string(data))
	}

	// Verify perso snapshot still has TEST_B
	data, err = os.ReadFile(persoSnapshot)
	if err != nil {
		t.Fatalf("Failed to re-read perso snapshot: %v", err)
	}
	if string(data) != "TEST_B\n" {
		t.Errorf("Perso snapshot changed! Expected 'TEST_B\\n', got %q", string(data))
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
	data, err = os.ReadFile(kubeConfig)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	if string(data) != "TEST_A\n" {
		t.Errorf("Expected restored content 'TEST_A\\n', got %q", string(data))
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
	data, err = os.ReadFile(kubeConfig)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	if string(data) != "TEST_B\n" {
		t.Errorf("Expected restored content 'TEST_B\\n', got %q", string(data))
	}

	t.Log("✅ Successfully switched to perso and restored TEST_B")
	t.Log("🎉 All integration test steps passed!")
}
