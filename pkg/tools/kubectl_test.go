package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKubectlTool_Name(t *testing.T) {
	tool := NewKubectlTool()
	if tool.Name() != "kubectl" {
		t.Errorf("Expected name 'kubectl', got '%s'", tool.Name())
	}
}

func TestKubectlTool_IsInstalled(t *testing.T) {
	tool := NewKubectlTool()
	// Just check that it doesn't panic
	_ = tool.IsInstalled()
}

func TestKubectlTool_Snapshot(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock .kube directory
	mockKubeDir := filepath.Join(tmpDir, "kube-config")
	if err := os.MkdirAll(mockKubeDir, 0755); err != nil {
		t.Fatalf("Failed to create mock kube dir: %v", err)
	}

	// Create test config file
	configFile := filepath.Join(mockKubeDir, "config")
	configContent := `apiVersion: v1
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: minikube
contexts:
- context:
    cluster: minikube
    user: minikube
  name: minikube
current-context: minikube
kind: Config
users:
- name: minikube
  user:
    client-certificate: /path/to/cert
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create additional files
	cacheDir := filepath.Join(mockKubeDir, "cache")
	os.MkdirAll(cacheDir, 0755)
	os.WriteFile(filepath.Join(cacheDir, "discovery.json"), []byte("{}"), 0644)

	// Create tool instance with mock config path
	tool := &KubectlTool{
		KubeConfigDir: mockKubeDir,
	}

	// Create snapshot
	snapshotPath := filepath.Join(tmpDir, "snapshot")
	err = tool.Snapshot(snapshotPath)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Verify snapshot was created
	snapshotConfig := filepath.Join(snapshotPath, "config")
	if _, err := os.Stat(snapshotConfig); os.IsNotExist(err) {
		t.Error("Snapshot config file was not created")
	}

	// Verify cache directory was copied
	snapshotCache := filepath.Join(snapshotPath, "cache", "discovery.json")
	if _, err := os.Stat(snapshotCache); os.IsNotExist(err) {
		t.Error("Snapshot cache was not copied")
	}

	// Verify content
	content, err := os.ReadFile(snapshotConfig)
	if err != nil {
		t.Fatalf("Failed to read snapshot config: %v", err)
	}
	if string(content) != configContent {
		t.Errorf("Config content mismatch")
	}
}

func TestKubectlTool_Restore(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create snapshot directory
	snapshotPath := filepath.Join(tmpDir, "snapshot")
	if err := os.MkdirAll(snapshotPath, 0755); err != nil {
		t.Fatalf("Failed to create snapshot dir: %v", err)
	}

	// Create test config in snapshot
	configContent := "apiVersion: v1\nkind: Config\ncurrent-context: test-context\n"
	os.WriteFile(filepath.Join(snapshotPath, "config"), []byte(configContent), 0644)

	// Create tool instance with mock restore path
	restorePath := filepath.Join(tmpDir, "kube-restored")
	tool := &KubectlTool{
		KubeConfigDir: restorePath,
	}

	// Restore from snapshot
	err = tool.Restore(snapshotPath)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify file was restored
	restoredConfig := filepath.Join(restorePath, "config")
	content, err := os.ReadFile(restoredConfig)
	if err != nil {
		t.Fatalf("Failed to read restored config: %v", err)
	}

	if string(content) != configContent {
		t.Errorf("Config content mismatch: got %q, want %q", string(content), configContent)
	}
}

func TestKubectlTool_ValidateSnapshot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tool := NewKubectlTool()

	// Test with non-existent directory
	err = tool.ValidateSnapshot(filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Error("Expected error for non-existent snapshot, got nil")
	}

	// Test with valid snapshot
	validSnapshot := filepath.Join(tmpDir, "valid")
	os.MkdirAll(validSnapshot, 0755)
	os.WriteFile(filepath.Join(validSnapshot, "config"), []byte("test"), 0644)

	err = tool.ValidateSnapshot(validSnapshot)
	if err != nil {
		t.Errorf("Unexpected error for valid snapshot: %v", err)
	}

	// Test with missing config file
	invalidSnapshot := filepath.Join(tmpDir, "invalid")
	os.MkdirAll(invalidSnapshot, 0755)

	err = tool.ValidateSnapshot(invalidSnapshot)
	if err == nil {
		t.Error("Expected error for invalid snapshot, got nil")
	}
}

func TestKubectlTool_GetMetadata(t *testing.T) {
	tool := NewKubectlTool()

	// This test will only pass if kubectl is installed
	if !tool.IsInstalled() {
		t.Skip("kubectl is not installed, skipping metadata test")
	}

	metadata, err := tool.GetMetadata()
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}

	// Just verify we got a map back
	if metadata == nil {
		t.Error("Expected non-nil metadata map")
	}
}

func TestKubectlTool_Diff(t *testing.T) {
	tool := NewKubectlTool()

	// This test will only pass if kubectl is installed
	if !tool.IsInstalled() {
		t.Skip("kubectl is not installed, skipping diff test")
	}

	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create snapshot
	snapshotPath := filepath.Join(tmpDir, "snapshot")
	os.MkdirAll(snapshotPath, 0755)
	os.WriteFile(filepath.Join(snapshotPath, "config"), []byte("test"), 0644)

	// Call Diff (currently returns empty changes)
	changes, err := tool.Diff(snapshotPath)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Verify we got a slice back (even if empty)
	if changes == nil {
		t.Error("Expected non-nil changes slice")
	}
}
