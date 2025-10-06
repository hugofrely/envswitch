package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitTool_Name(t *testing.T) {
	tool := NewGitTool()
	if tool.Name() != "git" {
		t.Errorf("Expected name 'git', got '%s'", tool.Name())
	}
}

func TestGitTool_IsInstalled(t *testing.T) {
	tool := NewGitTool()
	// We can't reliably test this without knowing if git is installed
	// Just check that it doesn't panic
	_ = tool.IsInstalled()
}

func TestGitTool_Snapshot(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock gitconfig file
	mockGitConfig := filepath.Join(tmpDir, "gitconfig")
	testContent := "[user]\n\tname = Test User\n\temail = test@example.com\n"
	if err := os.WriteFile(mockGitConfig, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create mock gitconfig: %v", err)
	}

	// Create tool instance with mock config path
	tool := &GitTool{
		GitConfigPath: mockGitConfig,
	}

	// Create snapshot
	snapshotPath := filepath.Join(tmpDir, "snapshot")
	err = tool.Snapshot(snapshotPath)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Verify snapshot was created
	snapshotFile := filepath.Join(snapshotPath, "gitconfig")
	if _, err := os.Stat(snapshotFile); os.IsNotExist(err) {
		t.Error("Snapshot file was not created")
	}

	// Verify content
	content, err := os.ReadFile(snapshotFile)
	if err != nil {
		t.Fatalf("Failed to read snapshot file: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("Content mismatch: got %q, want %q", string(content), testContent)
	}
}

func TestGitTool_SnapshotWithLocal(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock gitconfig and gitconfig.local files
	mockGitConfig := filepath.Join(tmpDir, "gitconfig")
	mockGitConfigLocal := mockGitConfig + ".local"

	os.WriteFile(mockGitConfig, []byte("[user]\n\tname = Test User\n"), 0644)
	os.WriteFile(mockGitConfigLocal, []byte("[user]\n\temail = local@example.com\n"), 0644)

	// Create tool instance
	tool := &GitTool{
		GitConfigPath: mockGitConfig,
	}

	// Create snapshot
	snapshotPath := filepath.Join(tmpDir, "snapshot")
	err = tool.Snapshot(snapshotPath)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Verify both files were snapshotted
	if _, err := os.Stat(filepath.Join(snapshotPath, "gitconfig")); os.IsNotExist(err) {
		t.Error("gitconfig was not snapshotted")
	}
	if _, err := os.Stat(filepath.Join(snapshotPath, "gitconfig.local")); os.IsNotExist(err) {
		t.Error("gitconfig.local was not snapshotted")
	}
}

func TestGitTool_Restore(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create snapshot directory
	snapshotPath := filepath.Join(tmpDir, "snapshot")
	os.MkdirAll(snapshotPath, 0755)

	// Create test file in snapshot
	testContent := "[user]\n\tname = Restored User\n"
	os.WriteFile(filepath.Join(snapshotPath, "gitconfig"), []byte(testContent), 0644)

	// Create tool instance with mock restore path
	restorePath := filepath.Join(tmpDir, "gitconfig-restored")
	tool := &GitTool{
		GitConfigPath: restorePath,
	}

	// Restore from snapshot
	err = tool.Restore(snapshotPath)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify file was restored
	content, err := os.ReadFile(restorePath)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Content mismatch: got %q, want %q", string(content), testContent)
	}
}

func TestGitTool_ValidateSnapshot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tool := NewGitTool()

	// Test with non-existent directory
	err = tool.ValidateSnapshot(filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Error("Expected error for non-existent snapshot, got nil")
	}

	// Test with valid snapshot
	validSnapshot := filepath.Join(tmpDir, "valid")
	os.MkdirAll(validSnapshot, 0755)
	os.WriteFile(filepath.Join(validSnapshot, "gitconfig"), []byte("test"), 0644)

	err = tool.ValidateSnapshot(validSnapshot)
	if err != nil {
		t.Errorf("Unexpected error for valid snapshot: %v", err)
	}

	// Test with missing gitconfig
	invalidSnapshot := filepath.Join(tmpDir, "invalid")
	os.MkdirAll(invalidSnapshot, 0755)

	err = tool.ValidateSnapshot(invalidSnapshot)
	if err == nil {
		t.Error("Expected error for invalid snapshot, got nil")
	}
}

func TestGitTool_GetMetadata(t *testing.T) {
	tool := NewGitTool()

	// This test will only pass if git is installed
	if !tool.IsInstalled() {
		t.Skip("git is not installed, skipping metadata test")
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

func TestGitTool_getSnapshotMetadata(t *testing.T) {
	t.Run("reads metadata from snapshot gitconfig", func(t *testing.T) {
		// Create temp directory for snapshot
		tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create snapshot gitconfig with test data
		gitconfigContent := `[user]
	name = Test User
	email = test@example.com
	signingkey = ABC123
[core]
	editor = vim
`
		snapshotPath := filepath.Join(tmpDir, "snapshot")
		os.MkdirAll(snapshotPath, 0755)
		os.WriteFile(filepath.Join(snapshotPath, "gitconfig"), []byte(gitconfigContent), 0644)

		tool := NewGitTool()
		metadata, err := tool.getSnapshotMetadata(snapshotPath)

		if err != nil {
			t.Fatalf("getSnapshotMetadata failed: %v", err)
		}

		// Verify all user fields were extracted
		if metadata["user_name"] != "Test User" {
			t.Errorf("Expected user_name 'Test User', got '%v'", metadata["user_name"])
		}
		if metadata["user_email"] != "test@example.com" {
			t.Errorf("Expected user_email 'test@example.com', got '%v'", metadata["user_email"])
		}
		if metadata["signing_key"] != "ABC123" {
			t.Errorf("Expected signing_key 'ABC123', got '%v'", metadata["signing_key"])
		}
	})

	t.Run("handles missing gitconfig file", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		tool := NewGitTool()
		metadata, err := tool.getSnapshotMetadata(tmpDir)

		// Should not error, just return empty metadata
		if err != nil {
			t.Fatalf("getSnapshotMetadata should not error on missing file: %v", err)
		}
		if len(metadata) != 0 {
			t.Errorf("Expected empty metadata, got %v", metadata)
		}
	})

	t.Run("handles partial metadata", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Only name, no email or signingkey
		gitconfigContent := `[user]
	name = Partial User
`
		snapshotPath := filepath.Join(tmpDir, "snapshot")
		os.MkdirAll(snapshotPath, 0755)
		os.WriteFile(filepath.Join(snapshotPath, "gitconfig"), []byte(gitconfigContent), 0644)

		tool := NewGitTool()
		metadata, err := tool.getSnapshotMetadata(snapshotPath)

		if err != nil {
			t.Fatalf("getSnapshotMetadata failed: %v", err)
		}

		if metadata["user_name"] != "Partial User" {
			t.Errorf("Expected user_name 'Partial User', got '%v'", metadata["user_name"])
		}
		if _, exists := metadata["user_email"]; exists {
			t.Error("Expected user_email to not exist")
		}
		if _, exists := metadata["signing_key"]; exists {
			t.Error("Expected signing_key to not exist")
		}
	})
}

func TestGitTool_Diff(t *testing.T) {
	t.Run("detects changes between snapshots", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create snapshot with old metadata
		oldGitconfigContent := `[user]
	name = Old User
	email = old@example.com
`
		snapshotPath := filepath.Join(tmpDir, "snapshot")
		os.MkdirAll(snapshotPath, 0755)
		os.WriteFile(filepath.Join(snapshotPath, "gitconfig"), []byte(oldGitconfigContent), 0644)

		// Create second snapshot with new metadata
		newGitconfigContent := `[user]
	name = New User
	email = old@example.com
	signingkey = XYZ789
`
		newSnapshotPath := filepath.Join(tmpDir, "new-snapshot")
		os.MkdirAll(newSnapshotPath, 0755)
		os.WriteFile(filepath.Join(newSnapshotPath, "gitconfig"), []byte(newGitconfigContent), 0644)

		tool := NewGitTool()

		// Get metadata from both snapshots
		oldMeta, err := tool.getSnapshotMetadata(snapshotPath)
		if err != nil {
			t.Fatalf("Failed to get old metadata: %v", err)
		}

		newMeta, err := tool.getSnapshotMetadata(newSnapshotPath)
		if err != nil {
			t.Fatalf("Failed to get new metadata: %v", err)
		}

		// Compare manually using compareMetadataField
		var changes []Change
		changes = append(changes, compareMetadataField("user_name", oldMeta, newMeta)...)
		changes = append(changes, compareMetadataField("user_email", oldMeta, newMeta)...)
		changes = append(changes, compareMetadataField("signing_key", oldMeta, newMeta)...)

		// Should detect:
		// - Modified: user_name (Old User -> New User)
		// - Added: signing_key (XYZ789)
		// - Unchanged: user_email

		if len(changes) != 2 {
			t.Errorf("Expected 2 changes, got %d", len(changes))
		}

		// Find the changes
		var nameChange, keyChange *Change
		for i := range changes {
			if changes[i].Path == "user_name" {
				nameChange = &changes[i]
			}
			if changes[i].Path == "signing_key" {
				keyChange = &changes[i]
			}
		}

		if nameChange == nil {
			t.Error("Expected user_name change")
		} else {
			if nameChange.Type != ChangeTypeModified {
				t.Errorf("Expected Modified type for user_name, got %v", nameChange.Type)
			}
			if nameChange.OldValue != "Old User" {
				t.Errorf("Expected OldValue 'Old User', got '%s'", nameChange.OldValue)
			}
			if nameChange.NewValue != "New User" {
				t.Errorf("Expected NewValue 'New User', got '%s'", nameChange.NewValue)
			}
		}

		if keyChange == nil {
			t.Error("Expected signing_key change")
		} else {
			if keyChange.Type != ChangeTypeAdded {
				t.Errorf("Expected Added type for signing_key, got %v", keyChange.Type)
			}
			if keyChange.NewValue != "XYZ789" {
				t.Errorf("Expected NewValue 'XYZ789', got '%s'", keyChange.NewValue)
			}
		}
	})
}
