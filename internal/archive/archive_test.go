package archive

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hugofrely/envswitch/pkg/environment"
)

func TestGetArchiveDir(t *testing.T) {
	dir, err := GetArchiveDir()
	if err != nil {
		t.Fatalf("GetArchiveDir failed: %v", err)
	}

	if dir == "" {
		t.Error("Expected non-empty archive directory path")
	}

	if !filepath.IsAbs(dir) {
		t.Error("Expected absolute path for archive directory")
	}

	// Should end with /archives
	if filepath.Base(dir) != "archives" {
		t.Errorf("Expected directory to end with 'archives', got: %s", filepath.Base(dir))
	}
}

func TestArchiveEnvironment(t *testing.T) {
	// Create temporary test environment
	tmpDir, err := os.MkdirTemp("", "envswitch-archive-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test environment directory structure
	envPath := filepath.Join(tmpDir, "environments", "test-env")
	if err := os.MkdirAll(envPath, 0755); err != nil {
		t.Fatalf("Failed to create env directory: %v", err)
	}

	// Create some test files
	testFile := filepath.Join(envPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	subDir := filepath.Join(envPath, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	subFile := filepath.Join(subDir, "subfile.txt")
	if err := os.WriteFile(subFile, []byte("sub content"), 0644); err != nil {
		t.Fatalf("Failed to create subfile: %v", err)
	}

	// Create test environment
	env := &environment.Environment{
		Name:      "test-env",
		Path:      envPath,
		CreatedAt: time.Now(),
	}

	// Override archive directory for testing
	archiveDir := filepath.Join(tmpDir, "archives")
	originalGetArchiveDirFunc := getArchiveDirFunc
	getArchiveDirFunc = func() (string, error) {
		return archiveDir, nil
	}
	defer func() { getArchiveDirFunc = originalGetArchiveDirFunc }()

	// Archive the environment
	archive, err := ArchiveEnvironment(env)
	if err != nil {
		t.Fatalf("ArchiveEnvironment failed: %v", err)
	}

	// Verify archive was created
	if archive == nil {
		t.Fatal("Expected non-nil archive")
	}

	if archive.EnvName != "test-env" {
		t.Errorf("Expected env name 'test-env', got: %s", archive.EnvName)
	}

	if _, err := os.Stat(archive.Path); os.IsNotExist(err) {
		t.Errorf("Archive file was not created: %s", archive.Path)
	}

	// Verify archive file has .tar.gz extension
	if filepath.Ext(filepath.Base(archive.Path)) != ".gz" {
		t.Errorf("Expected .tar.gz extension, got: %s", filepath.Ext(archive.Path))
	}

	// Verify archive is not empty
	info, err := os.Stat(archive.Path)
	if err != nil {
		t.Fatalf("Failed to stat archive: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Archive file is empty")
	}
}

func TestArchiveEnvironment_NilEnvironment(t *testing.T) {
	_, err := ArchiveEnvironment(nil)
	if err == nil {
		t.Error("Expected error when archiving nil environment")
	}
}

func TestListArchives(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envswitch-list-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	archiveDir := filepath.Join(tmpDir, "archives")
	originalGetArchiveDirFunc := getArchiveDirFunc
	getArchiveDirFunc = func() (string, error) {
		return archiveDir, nil
	}
	defer func() { getArchiveDirFunc = originalGetArchiveDirFunc }()

	// Test with no archives directory
	archives, err := ListArchives()
	if err != nil {
		t.Fatalf("ListArchives failed: %v", err)
	}

	if len(archives) != 0 {
		t.Errorf("Expected 0 archives, got: %d", len(archives))
	}

	// Create archive directory with test files
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		t.Fatalf("Failed to create archive dir: %v", err)
	}

	// Create test archive files
	testArchive1 := filepath.Join(archiveDir, "env1-20240101-120000.tar.gz")
	testArchive2 := filepath.Join(archiveDir, "env2-20240102-130000.tar.gz")
	testNonArchive := filepath.Join(archiveDir, "readme.txt")

	for _, path := range []string{testArchive1, testArchive2, testNonArchive} {
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// List archives
	archives, err = ListArchives()
	if err != nil {
		t.Fatalf("ListArchives failed: %v", err)
	}

	if len(archives) != 2 {
		t.Errorf("Expected 2 archives, got: %d", len(archives))
	}
}

func TestDeleteArchive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envswitch-delete-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test archive file
	testArchive := filepath.Join(tmpDir, "test-archive.tar.gz")
	if err := os.WriteFile(testArchive, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testArchive); os.IsNotExist(err) {
		t.Fatal("Test archive was not created")
	}

	// Delete archive
	if err := DeleteArchive(testArchive); err != nil {
		t.Fatalf("DeleteArchive failed: %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(testArchive); !os.IsNotExist(err) {
		t.Error("Archive file still exists after deletion")
	}
}

func TestRestoreArchive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envswitch-restore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test environment
	envPath := filepath.Join(tmpDir, "environments", "test-env")
	if err := os.MkdirAll(envPath, 0755); err != nil {
		t.Fatalf("Failed to create env directory: %v", err)
	}

	// Create test files
	testFile := filepath.Join(envPath, "test.txt")
	testContent := "original content"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	subDir := filepath.Join(envPath, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	subFile := filepath.Join(subDir, "subfile.txt")
	if err := os.WriteFile(subFile, []byte("sub content"), 0644); err != nil {
		t.Fatalf("Failed to create subfile: %v", err)
	}

	// Create environment
	env := &environment.Environment{
		Name:      "test-env",
		Path:      envPath,
		CreatedAt: time.Now(),
	}

	// Archive the environment
	archiveDir := filepath.Join(tmpDir, "archives")
	originalGetArchiveDirFunc := getArchiveDirFunc
	getArchiveDirFunc = func() (string, error) {
		return archiveDir, nil
	}
	defer func() { getArchiveDirFunc = originalGetArchiveDirFunc }()

	archive, err := ArchiveEnvironment(env)
	if err != nil {
		t.Fatalf("ArchiveEnvironment failed: %v", err)
	}

	// Remove original environment
	if err := os.RemoveAll(envPath); err != nil {
		t.Fatalf("Failed to remove environment: %v", err)
	}

	// Restore from archive
	restorePath := filepath.Join(tmpDir, "restored")
	if err := RestoreArchive(archive.Path, restorePath); err != nil {
		t.Fatalf("RestoreArchive failed: %v", err)
	}

	// Verify restored files
	restoredFile := filepath.Join(restorePath, "test-env", "test.txt")
	content, err := os.ReadFile(restoredFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected content '%s', got: '%s'", testContent, string(content))
	}

	// Verify subdirectory and file
	restoredSubFile := filepath.Join(restorePath, "test-env", "subdir", "subfile.txt")
	if _, err := os.Stat(restoredSubFile); os.IsNotExist(err) {
		t.Error("Restored subdirectory file not found")
	}
}

func TestArchiveDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envswitch-archivedir-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test directory structure
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create some files
	file1 := filepath.Join(testDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	subDir := filepath.Join(testDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	file2 := filepath.Join(subDir, "file2.txt")
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// This is an internal function, so we test it via ArchiveEnvironment
	// The test above (TestRestoreArchive) already validates this functionality
}
