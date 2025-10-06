package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	testContent := "Hello, EnvSwitch!"
	if err := os.WriteFile(srcFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	dstFile := filepath.Join(tmpDir, "destination.txt")
	if err := CopyFile(srcFile, dstFile); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify destination file exists and has correct content
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Content mismatch: got %q, want %q", string(content), testContent)
	}

	// Verify file permissions
	srcInfo, _ := os.Stat(srcFile)
	dstInfo, _ := os.Stat(dstFile)
	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("Permission mismatch: got %v, want %v", dstInfo.Mode(), srcInfo.Mode())
	}
}

func TestCopyFileNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "nonexistent.txt")
	dstFile := filepath.Join(tmpDir, "destination.txt")

	err = CopyFile(srcFile, dstFile)
	if err == nil {
		t.Error("Expected error when copying non-existent file, got nil")
	}
}

func TestCopyDir(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source directory with files and subdirectories
	srcDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Create test files
	files := map[string]string{
		"file1.txt":         "Content 1",
		"file2.txt":         "Content 2",
		"subdir/file3.txt":  "Content 3",
		"subdir/file4.txt":  "Content 4",
		"subdir2/file5.txt": "Content 5",
	}

	for path, content := range files {
		fullPath := filepath.Join(srcDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	// Copy directory
	dstDir := filepath.Join(tmpDir, "destination")
	if err := CopyDir(srcDir, dstDir); err != nil {
		t.Fatalf("CopyDir failed: %v", err)
	}

	// Verify all files were copied correctly
	for path, expectedContent := range files {
		dstFile := filepath.Join(dstDir, path)
		content, err := os.ReadFile(dstFile)
		if err != nil {
			t.Errorf("Failed to read destination file %s: %v", path, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("Content mismatch for %s: got %q, want %q", path, string(content), expectedContent)
		}
	}
}

func TestCopyDirNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "nonexistent")
	dstDir := filepath.Join(tmpDir, "destination")

	err = CopyDir(srcDir, dstDir)
	if err == nil {
		t.Error("Expected error when copying non-existent directory, got nil")
	}
}

func TestDirSize(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create files with known sizes
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	content1 := "Hello"    // 5 bytes
	content2 := "World123" // 8 bytes
	expectedSize := int64(len(content1) + len(content2))

	os.WriteFile(file1, []byte(content1), 0644)
	os.WriteFile(file2, []byte(content2), 0644)

	size, err := DirSize(tmpDir)
	if err != nil {
		t.Fatalf("DirSize failed: %v", err)
	}

	if size != expectedSize {
		t.Errorf("Size mismatch: got %d, want %d", size, expectedSize)
	}
}

func TestCountFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envswitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create files and subdirectories
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)

	subdir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subdir, 0755)
	os.WriteFile(filepath.Join(subdir, "file3.txt"), []byte("test"), 0644)

	count, err := CountFiles(tmpDir)
	if err != nil {
		t.Fatalf("CountFiles failed: %v", err)
	}

	expectedCount := 3
	if count != expectedCount {
		t.Errorf("Count mismatch: got %d, want %d", count, expectedCount)
	}
}
