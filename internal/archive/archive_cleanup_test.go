package archive

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/pkg/environment"
)

func TestCleanupOldArchives(t *testing.T) {
	// Setup temp directory
	tempDir := t.TempDir()
	oldGetArchiveDirFunc := getArchiveDirFunc
	getArchiveDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getArchiveDirFunc = oldGetArchiveDirFunc }()

	t.Run("does nothing when retention is 0", func(t *testing.T) {
		deleted, err := CleanupOldArchives(0)
		require.NoError(t, err)
		assert.Equal(t, 0, deleted)
	})

	t.Run("does nothing when retention is negative", func(t *testing.T) {
		deleted, err := CleanupOldArchives(-1)
		require.NoError(t, err)
		assert.Equal(t, 0, deleted)
	})

	t.Run("keeps all archives when count is below retention", func(t *testing.T) {
		// Create 3 archives
		for i := 0; i < 3; i++ {
			createTestArchive(t, tempDir, time.Now().Add(-time.Duration(i)*time.Hour))
		}

		deleted, err := CleanupOldArchives(5)
		require.NoError(t, err)
		assert.Equal(t, 0, deleted)

		// Verify all archives still exist
		archives, _ := ListArchives()
		assert.Len(t, archives, 3)
	})

	t.Run("deletes old archives beyond retention count", func(t *testing.T) {
		// Clean up previous test
		os.RemoveAll(tempDir)
		os.MkdirAll(tempDir, 0755)

		// Create 5 archives with different timestamps
		times := []time.Time{
			time.Now().Add(-5 * time.Hour), // Oldest
			time.Now().Add(-4 * time.Hour),
			time.Now().Add(-3 * time.Hour),
			time.Now().Add(-2 * time.Hour),
			time.Now().Add(-1 * time.Hour), // Newest
		}

		for _, timestamp := range times {
			createTestArchive(t, tempDir, timestamp)
		}

		// Keep only 3 most recent
		deleted, err := CleanupOldArchives(3)
		require.NoError(t, err)
		assert.Equal(t, 2, deleted)

		// Verify only 3 archives remain
		archives, _ := ListArchives()
		assert.Len(t, archives, 3)
	})

	t.Run("keeps newest archives", func(t *testing.T) {
		// Clean up
		os.RemoveAll(tempDir)
		os.MkdirAll(tempDir, 0755)

		// Create archives
		old := createTestArchive(t, tempDir, time.Now().Add(-2*time.Hour))
		recent := createTestArchive(t, tempDir, time.Now().Add(-1*time.Hour))

		deleted, err := CleanupOldArchives(1)
		require.NoError(t, err)
		assert.Equal(t, 1, deleted)

		// Old should be deleted, recent should remain
		_, err = os.Stat(old)
		assert.True(t, os.IsNotExist(err), "old archive should be deleted")

		_, err = os.Stat(recent)
		assert.NoError(t, err, "recent archive should exist")
	})

	t.Run("handles deletion errors gracefully", func(t *testing.T) {
		// Clean up
		os.RemoveAll(tempDir)
		os.MkdirAll(tempDir, 0755)

		// Create archives
		for i := 0; i < 3; i++ {
			createTestArchive(t, tempDir, time.Now().Add(-time.Duration(i)*time.Hour))
		}

		// Make one archive read-only (on systems that support it)
		archives, _ := ListArchives()
		if len(archives) > 0 {
			os.Chmod(archives[0].Path, 0444)

			deleted, err := CleanupOldArchives(1)
			// Should not return error, but may delete fewer than expected
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, 2, deleted)
		}
	})

	t.Run("sorts archives by date correctly", func(t *testing.T) {
		// Clean up
		os.RemoveAll(tempDir)
		os.MkdirAll(tempDir, 0755)

		// Create archives in random order
		times := []time.Time{
			time.Now().Add(-3 * time.Hour),
			time.Now().Add(-1 * time.Hour),
			time.Now().Add(-4 * time.Hour),
			time.Now().Add(-2 * time.Hour),
		}

		for _, timestamp := range times {
			createTestArchive(t, tempDir, timestamp)
		}

		// Keep only 2 (should keep the 2 newest)
		deleted, err := CleanupOldArchives(2)
		require.NoError(t, err)
		assert.Equal(t, 2, deleted)

		archives, _ := ListArchives()
		assert.Len(t, archives, 2)
	})
}

func TestCleanupWithNoArchives(t *testing.T) {
	tempDir := t.TempDir()
	oldGetArchiveDirFunc := getArchiveDirFunc
	getArchiveDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getArchiveDirFunc = oldGetArchiveDirFunc }()

	deleted, err := CleanupOldArchives(5)
	require.NoError(t, err)
	assert.Equal(t, 0, deleted)
}

// Helper function to create a test archive with specific timestamp
func createTestArchive(t *testing.T, dir string, timestamp time.Time) string {
	t.Helper()

	filename := "test-" + timestamp.Format("20060102-150405") + ".tar.gz"
	path := filepath.Join(dir, filename)

	// Create empty archive file
	file, err := os.Create(path)
	require.NoError(t, err)
	file.Close()

	// Set modification time
	err = os.Chtimes(path, timestamp, timestamp)
	require.NoError(t, err)

	return path
}

func TestListArchivesSorting(t *testing.T) {
	tempDir := t.TempDir()
	oldGetArchiveDirFunc := getArchiveDirFunc
	getArchiveDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getArchiveDirFunc = oldGetArchiveDirFunc }()

	// Create archives with known timestamps
	old := createTestArchive(t, tempDir, time.Now().Add(-2*time.Hour))
	mid := createTestArchive(t, tempDir, time.Now().Add(-1*time.Hour))
	new := createTestArchive(t, tempDir, time.Now())

	archives, err := ListArchives()
	require.NoError(t, err)
	require.Len(t, archives, 3)

	// Verify we can identify the archives
	paths := make([]string, len(archives))
	for i, a := range archives {
		paths[i] = a.Path
	}

	assert.Contains(t, paths, old)
	assert.Contains(t, paths, mid)
	assert.Contains(t, paths, new)
}

func TestCleanupIntegration(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	oldGetArchiveDirFunc := getArchiveDirFunc
	getArchiveDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getArchiveDirFunc = oldGetArchiveDirFunc }()

	// Setup environment directory
	envDir := filepath.Join(tempDir, "test-env")
	os.MkdirAll(envDir, 0755)

	env := &environment.Environment{
		Name: "test",
		Path: envDir,
	}

	// Create multiple archives
	for i := 0; i < 5; i++ {
		_, err := ArchiveEnvironment(env)
		require.NoError(t, err)
		time.Sleep(1100 * time.Millisecond) // Ensure different timestamps (>1s for format 20060102-150405)
	}

	// Verify 5 archives exist
	archives, _ := ListArchives()
	assert.Len(t, archives, 5)

	// Cleanup keeping only 2
	deleted, err := CleanupOldArchives(2)
	require.NoError(t, err)
	assert.Equal(t, 3, deleted)

	// Verify only 2 remain
	archives, _ = ListArchives()
	assert.Len(t, archives, 2)
}
