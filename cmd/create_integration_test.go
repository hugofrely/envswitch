package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hugofrely/envswitch/pkg/environment"
)

// TestCreateWorkflowIntegration tests the complete workflow:
// 1. Create a test file in .kube with TEST_A
// 2. Run: envswitch create work --from-current (should auto-switch)
// 3. Change test file to TEST_B
// 4. Run: envswitch create perso --from-current (should auto-switch)
// 5. Verify both environments have their correct snapshots
func TestCreateWorkflowIntegration(t *testing.T) {
	// Setup test environment
	originalHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create .kube directory
	kubeDir := filepath.Join(tempHome, ".kube")
	if err := os.MkdirAll(kubeDir, 0755); err != nil {
		t.Fatalf("Failed to create .kube directory: %v", err)
	}

	// Use the standard kubectl config file name
	testFile := filepath.Join(kubeDir, "config")

	// Initialize envswitch
	if err := os.MkdirAll(filepath.Join(tempHome, ".envswitch", "environments"), 0755); err != nil {
		t.Fatalf("Failed to create .envswitch directory: %v", err)
	}

	t.Run("Step 1: Create test file with TEST_A content", func(t *testing.T) {
		content := "TEST_A\n"
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Verify file was created
		data, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}
		if string(data) != content {
			t.Errorf("Expected content %q, got %q", content, string(data))
		}
	})

	t.Run("Step 2: Create work environment from current (should auto-switch)", func(t *testing.T) {
		// Set flags for create command
		createFromCurrent = true
		createEmpty = false
		createFrom = ""
		createDescription = "Work environment"

		// Run create command directly
		if err := runCreate(createCmd, []string{"work"}); err != nil {
			t.Fatalf("Failed to create work environment: %v", err)
		}

		// Verify environment was created
		workEnv, err := environment.LoadEnvironment("work")
		if err != nil {
			t.Fatalf("Failed to load work environment: %v", err)
		}

		if workEnv.Name != "work" {
			t.Errorf("Expected environment name 'work', got %q", workEnv.Name)
		}

		// Verify auto-switch happened
		currentEnv, err := environment.GetCurrentEnvironment()
		if err != nil {
			t.Fatalf("Failed to get current environment: %v", err)
		}
		if currentEnv == nil {
			t.Fatal("Expected current environment to be set, got nil")
		}
		if currentEnv.Name != "work" {
			t.Errorf("Expected current environment to be 'work', got %q", currentEnv.Name)
		}

		// Verify kubectl snapshot exists and contains TEST_A
		kubectlSnapshot := filepath.Join(workEnv.Path, "snapshots", "kubectl", "config")
		if _, err := os.Stat(kubectlSnapshot); os.IsNotExist(err) {
			t.Errorf("kubectl snapshot not found: %s", kubectlSnapshot)
		} else {
			data, err := os.ReadFile(kubectlSnapshot)
			if err != nil {
				t.Fatalf("Failed to read kubectl snapshot: %v", err)
			}
			if string(data) != "TEST_A\n" {
				t.Errorf("Expected snapshot content 'TEST_A\\n', got %q", string(data))
			}
		}

		t.Logf("✅ Work environment created and auto-switched successfully")
	})

	t.Run("Step 3: Change test file to TEST_B content", func(t *testing.T) {
		content := "TEST_B\n"
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Verify file was updated
		data, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}
		if string(data) != content {
			t.Errorf("Expected content %q, got %q", content, string(data))
		}
	})

	t.Run("Step 4: Create perso environment from current (should auto-switch)", func(t *testing.T) {
		// Reset flags for create command
		createFromCurrent = true
		createEmpty = false
		createFrom = ""
		createDescription = "Personal environment"

		// Run create command directly
		if err := runCreate(createCmd, []string{"perso"}); err != nil {
			t.Fatalf("Failed to create perso environment: %v", err)
		}

		// Verify environment was created
		persoEnv, err := environment.LoadEnvironment("perso")
		if err != nil {
			t.Fatalf("Failed to load perso environment: %v", err)
		}

		if persoEnv.Name != "perso" {
			t.Errorf("Expected environment name 'perso', got %q", persoEnv.Name)
		}

		// Verify auto-switch happened
		currentEnv, err := environment.GetCurrentEnvironment()
		if err != nil {
			t.Fatalf("Failed to get current environment: %v", err)
		}
		if currentEnv == nil {
			t.Fatal("Expected current environment to be set, got nil")
		}
		if currentEnv.Name != "perso" {
			t.Errorf("Expected current environment to be 'perso', got %q", currentEnv.Name)
		}

		// Verify kubectl snapshot exists and contains TEST_B
		kubectlSnapshot := filepath.Join(persoEnv.Path, "snapshots", "kubectl", "config")
		if _, err := os.Stat(kubectlSnapshot); os.IsNotExist(err) {
			t.Errorf("kubectl snapshot not found: %s", kubectlSnapshot)
		} else {
			data, err := os.ReadFile(kubectlSnapshot)
			if err != nil {
				t.Fatalf("Failed to read kubectl snapshot: %v", err)
			}
			if string(data) != "TEST_B\n" {
				t.Errorf("Expected snapshot content 'TEST_B\\n', got %q", string(data))
			}
		}

		t.Logf("✅ Perso environment created and auto-switched successfully")
	})

	t.Run("Step 5: Verify .envswitch structure", func(t *testing.T) {
		envswitchDir := filepath.Join(tempHome, ".envswitch")

		// Check that .envswitch directory exists
		if _, err := os.Stat(envswitchDir); os.IsNotExist(err) {
			t.Fatalf(".envswitch directory does not exist: %s", envswitchDir)
		}

		// Check that environments directory exists
		envsDir := filepath.Join(envswitchDir, "environments")
		if _, err := os.Stat(envsDir); os.IsNotExist(err) {
			t.Fatalf("environments directory does not exist: %s", envsDir)
		}

		// Check that work environment directory exists
		workDir := filepath.Join(envsDir, "work")
		if _, err := os.Stat(workDir); os.IsNotExist(err) {
			t.Fatalf("work environment directory does not exist: %s", workDir)
		}

		// Check that perso environment directory exists
		persoDir := filepath.Join(envsDir, "perso")
		if _, err := os.Stat(persoDir); os.IsNotExist(err) {
			t.Fatalf("perso environment directory does not exist: %s", persoDir)
		}

		// Check current.lock points to perso
		currentLock := filepath.Join(envswitchDir, "current.lock")
		if _, err := os.Stat(currentLock); os.IsNotExist(err) {
			t.Fatalf("current.lock file does not exist: %s", currentLock)
		}
		data, err := os.ReadFile(currentLock)
		if err != nil {
			t.Fatalf("Failed to read current.lock: %v", err)
		}
		if string(data) != "perso" {
			t.Errorf("Expected current.lock to contain 'perso', got %q", string(data))
		}

		// Verify work environment still has TEST_A in its snapshot
		workKubectlSnapshot := filepath.Join(workDir, "snapshots", "kubectl", "config")
		if data, err := os.ReadFile(workKubectlSnapshot); err != nil {
			t.Errorf("Failed to read work kubectl snapshot: %v", err)
		} else if string(data) != "TEST_A\n" {
			t.Errorf("Work environment snapshot changed! Expected 'TEST_A\\n', got %q", string(data))
		}

		// Verify perso environment has TEST_B in its snapshot
		persoKubectlSnapshot := filepath.Join(persoDir, "snapshots", "kubectl", "config")
		if data, err := os.ReadFile(persoKubectlSnapshot); err != nil {
			t.Errorf("Failed to read perso kubectl snapshot: %v", err)
		} else if string(data) != "TEST_B\n" {
			t.Errorf("Perso environment snapshot incorrect! Expected 'TEST_B\\n', got %q", string(data))
		}

		t.Logf("✅ .envswitch structure is correct:")
		t.Logf("   - work environment: TEST_A")
		t.Logf("   - perso environment: TEST_B")
		t.Logf("   - current: perso")
	})

	t.Run("Step 6: Switch back to work and verify restore", func(t *testing.T) {
		// Switch to work environment
		if err := runSwitch(switchCmd, []string{"work"}); err != nil {
			t.Fatalf("Failed to switch to work environment: %v", err)
		}

		// Verify current environment is work
		currentEnv, err := environment.GetCurrentEnvironment()
		if err != nil {
			t.Fatalf("Failed to get current environment: %v", err)
		}
		if currentEnv.Name != "work" {
			t.Errorf("Expected current environment to be 'work', got %q", currentEnv.Name)
		}

		// Verify test file was restored to TEST_A
		data, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read test file after restore: %v", err)
		}
		if string(data) != "TEST_A\n" {
			t.Errorf("Expected restored content 'TEST_A\\n', got %q", string(data))
		}

		t.Logf("✅ Successfully switched to work and restored TEST_A")
	})

	t.Run("Step 7: Switch to perso and verify restore", func(t *testing.T) {
		// Switch to perso environment
		if err := runSwitch(switchCmd, []string{"perso"}); err != nil {
			t.Fatalf("Failed to switch to perso environment: %v", err)
		}

		// Verify current environment is perso
		currentEnv, err := environment.GetCurrentEnvironment()
		if err != nil {
			t.Fatalf("Failed to get current environment: %v", err)
		}
		if currentEnv.Name != "perso" {
			t.Errorf("Expected current environment to be 'perso', got %q", currentEnv.Name)
		}

		// Verify test file was restored to TEST_B
		data, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read test file after restore: %v", err)
		}
		if string(data) != "TEST_B\n" {
			t.Errorf("Expected restored content 'TEST_B\\n', got %q", string(data))
		}

		t.Logf("✅ Successfully switched to perso and restored TEST_B")
	})
}
