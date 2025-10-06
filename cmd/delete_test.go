package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/pkg/environment"
)

const (
	gzipExt = ".gz"
)

func TestRunDelete(t *testing.T) {
	// Create a temporary directory for testing
	tempHome := t.TempDir()

	// Save original home and restore after test
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})

	os.Setenv("HOME", tempHome)

	// Create envswitch directory structure
	envswitchDir := filepath.Join(tempHome, ".envswitch")
	envDir := filepath.Join(envswitchDir, "environments")
	err := os.MkdirAll(envDir, 0755)
	require.NoError(t, err)

	t.Run("deletes environment with force flag", func(t *testing.T) {
		// Create test environment
		env := &environment.Environment{
			Name: "to-delete",
			Path: filepath.Join(envDir, "to-delete"),
		}
		err := os.MkdirAll(env.Path, 0755)
		require.NoError(t, err)
		err = env.Save()
		require.NoError(t, err)

		// Delete with force flag
		deleteForce = true
		defer func() { deleteForce = false }()

		err = runDelete(deleteCmd, []string{"to-delete"})
		require.NoError(t, err)

		// Verify environment is deleted
		_, err = os.Stat(env.Path)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("returns error for non-existent environment", func(t *testing.T) {
		deleteForce = true
		defer func() { deleteForce = false }()

		err := runDelete(deleteCmd, []string{"non-existent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("prevents deleting active environment", func(t *testing.T) {
		// Create and set as current
		env := &environment.Environment{
			Name: "current-env",
			Path: filepath.Join(envDir, "current-env"),
		}
		err := os.MkdirAll(env.Path, 0755)
		require.NoError(t, err)
		err = env.Save()
		require.NoError(t, err)

		err = environment.SetCurrentEnvironment("current-env")
		require.NoError(t, err)

		// Try to delete
		deleteForce = true
		defer func() { deleteForce = false }()

		err = runDelete(deleteCmd, []string{"current-env"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete active environment")

		// Verify environment still exists
		_, err = os.Stat(env.Path)
		assert.NoError(t, err)
	})

	t.Run("deletes environment directory and all contents", func(t *testing.T) {
		// Create environment with some files
		env := &environment.Environment{
			Name: "with-files",
			Path: filepath.Join(envDir, "with-files"),
		}
		err := os.MkdirAll(env.Path, 0755)
		require.NoError(t, err)
		err = env.Save()
		require.NoError(t, err)

		// Create some files
		snapshotsDir := filepath.Join(env.Path, "snapshots")
		err = os.MkdirAll(snapshotsDir, 0755)
		require.NoError(t, err)

		testFile := filepath.Join(snapshotsDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		// Delete
		deleteForce = true
		defer func() { deleteForce = false }()

		err = runDelete(deleteCmd, []string{"with-files"})
		require.NoError(t, err)

		// Verify all deleted
		_, err = os.Stat(env.Path)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("archives environment before deletion", func(t *testing.T) {
		// Create environment
		env := &environment.Environment{
			Name: "to-archive",
			Path: filepath.Join(envDir, "to-archive"),
		}
		err := os.MkdirAll(env.Path, 0755)
		require.NoError(t, err)
		err = env.Save()
		require.NoError(t, err)

		// Create test file
		testFile := filepath.Join(env.Path, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		// Delete with archiving
		deleteForce = true
		deleteNoArchive = false
		defer func() {
			deleteForce = false
			deleteNoArchive = false
		}()

		err = runDelete(deleteCmd, []string{"to-archive"})
		require.NoError(t, err)

		// Verify environment is deleted
		_, err = os.Stat(env.Path)
		assert.True(t, os.IsNotExist(err))

		// Verify archive was created
		archiveDir := filepath.Join(envswitchDir, "archives")
		entries, err := os.ReadDir(archiveDir)
		require.NoError(t, err)

		// Should have at least one archive
		archiveFound := false
		for _, entry := range entries {
			if filepath.Ext(entry.Name()) == gzipExt {
				archiveFound = true
				break
			}
		}
		assert.True(t, archiveFound, "Archive file should have been created")
	})

	t.Run("skips archiving with --no-archive flag", func(t *testing.T) {
		// Create environment
		env := &environment.Environment{
			Name: "no-archive",
			Path: filepath.Join(envDir, "no-archive"),
		}
		err := os.MkdirAll(env.Path, 0755)
		require.NoError(t, err)
		err = env.Save()
		require.NoError(t, err)

		// Count existing archives
		archiveDir := filepath.Join(envswitchDir, "archives")
		initialCount := 0
		if entries, err := os.ReadDir(archiveDir); err == nil {
			for _, entry := range entries {
				if filepath.Ext(entry.Name()) == gzipExt {
					initialCount++
				}
			}
		}

		// Delete without archiving
		deleteForce = true
		deleteNoArchive = true
		defer func() {
			deleteForce = false
			deleteNoArchive = false
		}()

		err = runDelete(deleteCmd, []string{"no-archive"})
		require.NoError(t, err)

		// Verify environment is deleted
		_, err = os.Stat(env.Path)
		assert.True(t, os.IsNotExist(err))

		// Verify no new archive was created
		finalCount := 0
		if entries, err := os.ReadDir(archiveDir); err == nil {
			for _, entry := range entries {
				if filepath.Ext(entry.Name()) == gzipExt {
					finalCount++
				}
			}
		}
		assert.Equal(t, initialCount, finalCount, "No new archive should have been created")
	})
}
