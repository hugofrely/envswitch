package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDockerTool_Name(t *testing.T) {
	tool := NewDockerTool()
	if tool.Name() != "docker" {
		t.Errorf("Expected name 'docker', got '%s'", tool.Name())
	}
}

func TestDockerTool_IsInstalled(t *testing.T) {
	tool := NewDockerTool()
	// Just check that it doesn't panic
	_ = tool.IsInstalled()
}

func TestDockerTool_Snapshot(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock .docker directory
	mockDockerDir := filepath.Join(tmpDir, "docker-config")
	if err := os.MkdirAll(mockDockerDir, 0755); err != nil {
		t.Fatalf("Failed to create mock docker dir: %v", err)
	}

	// Create test config.json file
	configFile := filepath.Join(mockDockerDir, "config.json")
	configContent := `{
	"auths": {
		"https://index.docker.io/v1/": {
			"auth": "dGVzdDp0ZXN0"
		}
	},
	"credsStore": "desktop"
}
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config.json: %v", err)
	}

	// Create additional files
	contextsDir := filepath.Join(mockDockerDir, "contexts", "meta")
	os.MkdirAll(contextsDir, 0755)
	os.WriteFile(filepath.Join(contextsDir, "default.json"), []byte("{}"), 0644)

	// Create tool instance with mock config path
	tool := &DockerTool{
		DockerConfigDir: mockDockerDir,
	}

	// Create snapshot
	snapshotPath := filepath.Join(tmpDir, "snapshot")
	err = tool.Snapshot(snapshotPath)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Verify snapshot was created
	snapshotConfig := filepath.Join(snapshotPath, "config.json")
	if _, err := os.Stat(snapshotConfig); os.IsNotExist(err) {
		t.Error("Snapshot config.json was not created")
	}

	// Verify contexts were copied
	snapshotContext := filepath.Join(snapshotPath, "contexts", "meta", "default.json")
	if _, err := os.Stat(snapshotContext); os.IsNotExist(err) {
		t.Error("Snapshot contexts were not copied")
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

func TestDockerTool_Restore(t *testing.T) {
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

	// Create test config.json in snapshot
	configContent := `{"auths": {}}`
	os.WriteFile(filepath.Join(snapshotPath, "config.json"), []byte(configContent), 0644)

	// Create tool instance with mock restore path
	restorePath := filepath.Join(tmpDir, "docker-restored")
	tool := &DockerTool{
		DockerConfigDir: restorePath,
	}

	// Restore from snapshot
	err = tool.Restore(snapshotPath)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify file was restored
	restoredConfig := filepath.Join(restorePath, "config.json")
	content, err := os.ReadFile(restoredConfig)
	if err != nil {
		t.Fatalf("Failed to read restored config: %v", err)
	}

	if string(content) != configContent {
		t.Errorf("Config content mismatch: got %q, want %q", string(content), configContent)
	}
}

func TestDockerTool_ValidateSnapshot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tool := NewDockerTool()

	// Test with non-existent directory
	err = tool.ValidateSnapshot(filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Error("Expected error for non-existent snapshot, got nil")
	}

	// Test with valid snapshot
	validSnapshot := filepath.Join(tmpDir, "valid")
	os.MkdirAll(validSnapshot, 0755)
	os.WriteFile(filepath.Join(validSnapshot, "config.json"), []byte("{}"), 0644)

	err = tool.ValidateSnapshot(validSnapshot)
	if err != nil {
		t.Errorf("Unexpected error for valid snapshot: %v", err)
	}

	// Test with missing config.json
	invalidSnapshot := filepath.Join(tmpDir, "invalid")
	os.MkdirAll(invalidSnapshot, 0755)

	err = tool.ValidateSnapshot(invalidSnapshot)
	if err == nil {
		t.Error("Expected error for invalid snapshot, got nil")
	}
}

func TestDockerTool_GetMetadata(t *testing.T) {
	tool := NewDockerTool()

	// This test will only pass if docker is installed
	if !tool.IsInstalled() {
		t.Skip("docker is not installed, skipping metadata test")
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

func TestDockerTool_Diff(t *testing.T) {
	tool := NewDockerTool()

	// This test will only pass if docker is installed
	if !tool.IsInstalled() {
		t.Skip("docker is not installed, skipping diff test")
	}

	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create snapshot
	snapshotPath := filepath.Join(tmpDir, "snapshot")
	os.MkdirAll(snapshotPath, 0755)
	os.WriteFile(filepath.Join(snapshotPath, "config.json"), []byte("{}"), 0644)

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

func TestDockerTool_SnapshotWithMultipleFiles(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock .docker directory with multiple files
	mockDockerDir := filepath.Join(tmpDir, "docker-config")
	os.MkdirAll(mockDockerDir, 0755)

	// Create config.json
	os.WriteFile(filepath.Join(mockDockerDir, "config.json"), []byte("{}"), 0644)

	// Create contexts
	contextsDir := filepath.Join(mockDockerDir, "contexts", "meta")
	os.MkdirAll(contextsDir, 0755)
	os.WriteFile(filepath.Join(contextsDir, "context1.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(contextsDir, "context2.json"), []byte("{}"), 0644)

	// Create buildx
	buildxDir := filepath.Join(mockDockerDir, "buildx", "instances")
	os.MkdirAll(buildxDir, 0755)
	os.WriteFile(filepath.Join(buildxDir, "default"), []byte("builder"), 0644)

	// Create tool instance
	tool := &DockerTool{
		DockerConfigDir: mockDockerDir,
	}

	// Create snapshot
	snapshotPath := filepath.Join(tmpDir, "snapshot")
	err = tool.Snapshot(snapshotPath)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Verify all files were snapshotted
	expectedFiles := []string{
		"config.json",
		"contexts/meta/context1.json",
		"contexts/meta/context2.json",
		"buildx/instances/default",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(snapshotPath, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("File %s was not snapshotted", file)
		}
	}
}
