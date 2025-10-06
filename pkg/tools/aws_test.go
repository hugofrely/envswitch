package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAWSTool_Name(t *testing.T) {
	tool := NewAWSTool()
	if tool.Name() != "aws" {
		t.Errorf("Expected name 'aws', got '%s'", tool.Name())
	}
}

func TestAWSTool_IsInstalled(t *testing.T) {
	tool := NewAWSTool()
	// Just check that it doesn't panic
	_ = tool.IsInstalled()
}

func TestAWSTool_Snapshot(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock AWS config directory
	mockConfigDir := filepath.Join(tmpDir, "aws-config")
	if err := os.MkdirAll(mockConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create mock config: %v", err)
	}

	// Create test config and credentials files
	configFile := filepath.Join(mockConfigDir, "config")
	credentialsFile := filepath.Join(mockConfigDir, "credentials")

	configContent := "[default]\nregion = us-east-1\noutput = json\n"
	credentialsContent := "[default]\naws_access_key_id = AKIAIOSFODNN7EXAMPLE\naws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY\n"

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	if err := os.WriteFile(credentialsFile, []byte(credentialsContent), 0600); err != nil {
		t.Fatalf("Failed to create credentials file: %v", err)
	}

	// Create tool instance with mock config path
	tool := &AWSTool{
		AWSConfigDir: mockConfigDir,
	}

	// Create snapshot
	snapshotPath := filepath.Join(tmpDir, "snapshot")
	err = tool.Snapshot(snapshotPath)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Verify snapshot files were created
	snapshotConfig := filepath.Join(snapshotPath, "config")
	snapshotCredentials := filepath.Join(snapshotPath, "credentials")

	if _, err := os.Stat(snapshotConfig); os.IsNotExist(err) {
		t.Error("Snapshot config file was not created")
	}
	if _, err := os.Stat(snapshotCredentials); os.IsNotExist(err) {
		t.Error("Snapshot credentials file was not created")
	}

	// Verify content
	content, err := os.ReadFile(snapshotConfig)
	if err != nil {
		t.Fatalf("Failed to read snapshot config: %v", err)
	}
	if string(content) != configContent {
		t.Errorf("Config content mismatch: got %q, want %q", string(content), configContent)
	}

	content, err = os.ReadFile(snapshotCredentials)
	if err != nil {
		t.Fatalf("Failed to read snapshot credentials: %v", err)
	}
	if string(content) != credentialsContent {
		t.Errorf("Credentials content mismatch: got %q, want %q", string(content), credentialsContent)
	}
}

func TestAWSTool_Restore(t *testing.T) {
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

	// Create test files in snapshot
	configContent := "[default]\nregion = us-west-2\n"
	credentialsContent := "[default]\naws_access_key_id = TEST\n"

	os.WriteFile(filepath.Join(snapshotPath, "config"), []byte(configContent), 0644)
	os.WriteFile(filepath.Join(snapshotPath, "credentials"), []byte(credentialsContent), 0600)

	// Create tool instance with mock restore path
	restorePath := filepath.Join(tmpDir, "aws-restored")
	tool := &AWSTool{
		AWSConfigDir: restorePath,
	}

	// Restore from snapshot
	err = tool.Restore(snapshotPath)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify files were restored
	restoredConfig := filepath.Join(restorePath, "config")
	content, err := os.ReadFile(restoredConfig)
	if err != nil {
		t.Fatalf("Failed to read restored config: %v", err)
	}

	if string(content) != configContent {
		t.Errorf("Config content mismatch: got %q, want %q", string(content), configContent)
	}

	restoredCredentials := filepath.Join(restorePath, "credentials")
	content, err = os.ReadFile(restoredCredentials)
	if err != nil {
		t.Fatalf("Failed to read restored credentials: %v", err)
	}

	if string(content) != credentialsContent {
		t.Errorf("Credentials content mismatch: got %q, want %q", string(content), credentialsContent)
	}
}

func TestAWSTool_ValidateSnapshot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tool := NewAWSTool()

	// Test with non-existent directory
	err = tool.ValidateSnapshot(filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Error("Expected error for non-existent snapshot, got nil")
	}

	// Test with valid snapshot (config only)
	validSnapshot1 := filepath.Join(tmpDir, "valid1")
	os.MkdirAll(validSnapshot1, 0755)
	os.WriteFile(filepath.Join(validSnapshot1, "config"), []byte("test"), 0644)

	err = tool.ValidateSnapshot(validSnapshot1)
	if err != nil {
		t.Errorf("Unexpected error for valid snapshot with config: %v", err)
	}

	// Test with valid snapshot (credentials only)
	validSnapshot2 := filepath.Join(tmpDir, "valid2")
	os.MkdirAll(validSnapshot2, 0755)
	os.WriteFile(filepath.Join(validSnapshot2, "credentials"), []byte("test"), 0600)

	err = tool.ValidateSnapshot(validSnapshot2)
	if err != nil {
		t.Errorf("Unexpected error for valid snapshot with credentials: %v", err)
	}

	// Test with valid snapshot (both files)
	validSnapshot3 := filepath.Join(tmpDir, "valid3")
	os.MkdirAll(validSnapshot3, 0755)
	os.WriteFile(filepath.Join(validSnapshot3, "config"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(validSnapshot3, "credentials"), []byte("test"), 0600)

	err = tool.ValidateSnapshot(validSnapshot3)
	if err != nil {
		t.Errorf("Unexpected error for valid snapshot with both files: %v", err)
	}

	// Test with missing both files
	invalidSnapshot := filepath.Join(tmpDir, "invalid")
	os.MkdirAll(invalidSnapshot, 0755)

	err = tool.ValidateSnapshot(invalidSnapshot)
	if err == nil {
		t.Error("Expected error for invalid snapshot, got nil")
	}
}

func TestAWSTool_GetMetadata(t *testing.T) {
	tool := NewAWSTool()

	// This test will only pass if aws is installed
	if !tool.IsInstalled() {
		t.Skip("aws is not installed, skipping metadata test")
	}

	metadata, err := tool.GetMetadata()
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}

	// Just verify we got a map back
	if metadata == nil {
		t.Error("Expected non-nil metadata map")
	}

	// Should at least have profile
	if _, ok := metadata["profile"]; !ok {
		t.Error("Expected metadata to contain 'profile'")
	}
}

func TestAWSTool_Diff(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create snapshot
	snapshotPath := filepath.Join(tmpDir, "snapshot")
	os.MkdirAll(snapshotPath, 0755)
	os.WriteFile(filepath.Join(snapshotPath, "config"), []byte("test"), 0644)

	tool := &AWSTool{
		AWSConfigDir: filepath.Join(tmpDir, "aws"),
	}

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
