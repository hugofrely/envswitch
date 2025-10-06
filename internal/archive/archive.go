package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/hugofrely/envswitch/pkg/environment"
)

// Archive represents an archived environment
type Archive struct {
	Path        string
	EnvName     string
	ArchivedAt  time.Time
	OriginalEnv *environment.Environment
}

// getArchiveDirFunc is a function variable that can be overridden in tests
var getArchiveDirFunc = getArchiveDirDefault

// getArchiveDirDefault returns the path to the archive directory
func getArchiveDirDefault() (string, error) {
	envswitchDir, err := environment.GetEnvswitchDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(envswitchDir, "archives"), nil
}

// GetArchiveDir returns the path to the archive directory
func GetArchiveDir() (string, error) {
	return getArchiveDirFunc()
}

// ArchiveEnvironment creates a compressed archive of an environment before deletion
func ArchiveEnvironment(env *environment.Environment) (*Archive, error) {
	if env == nil {
		return nil, fmt.Errorf("environment cannot be nil")
	}

	// Ensure archive directory exists
	archiveDir, err := GetArchiveDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get archive directory: %w", err)
	}

	if mkdirErr := os.MkdirAll(archiveDir, 0755); mkdirErr != nil {
		return nil, fmt.Errorf("failed to create archive directory: %w", mkdirErr)
	}

	// Create archive filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	archiveFilename := fmt.Sprintf("%s-%s.tar.gz", env.Name, timestamp)
	archivePath := filepath.Join(archiveDir, archiveFilename)

	// Create archive file
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create archive file: %w", err)
	}
	defer func() { _ = archiveFile.Close() }()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(archiveFile)
	defer func() { _ = gzipWriter.Close() }()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer func() { _ = tarWriter.Close() }()

	// Archive the entire environment directory
	if err := archiveDirectory(tarWriter, env.Path, env.Name); err != nil {
		// Clean up partial archive on error
		_ = os.Remove(archivePath)
		return nil, fmt.Errorf("failed to archive environment: %w", err)
	}

	archive := &Archive{
		Path:        archivePath,
		EnvName:     env.Name,
		ArchivedAt:  time.Now(),
		OriginalEnv: env,
	}

	return archive, nil
}

// archiveDirectory recursively adds a directory to a tar archive
func archiveDirectory(tarWriter *tar.Writer, sourcePath, basePath string) error {
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header: %w", err)
		}

		// Update header name to be relative to base path
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		header.Name = filepath.Join(basePath, relPath)

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		// If it's a file (not a directory), write the content
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer func() { _ = file.Close() }()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return fmt.Errorf("failed to write file content: %w", err)
			}
		}

		return nil
	})
}

// ListArchives returns all archived environments
func ListArchives() ([]*Archive, error) {
	archiveDir, err := GetArchiveDir()
	if err != nil {
		return nil, err
	}

	// Check if archive directory exists
	if _, statErr := os.Stat(archiveDir); os.IsNotExist(statErr) {
		return []*Archive{}, nil
	}

	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read archive directory: %w", err)
	}

	archives := make([]*Archive, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only include .tar.gz files
		if filepath.Ext(entry.Name()) != ".gz" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		archives = append(archives, &Archive{
			Path:       filepath.Join(archiveDir, entry.Name()),
			EnvName:    entry.Name(),
			ArchivedAt: info.ModTime(),
		})
	}

	return archives, nil
}

// DeleteArchive removes an archive file
func DeleteArchive(archivePath string) error {
	if err := os.Remove(archivePath); err != nil {
		return fmt.Errorf("failed to delete archive: %w", err)
	}
	return nil
}

// CleanupOldArchives removes old archives based on retention policy
func CleanupOldArchives(retentionCount int) (int, error) {
	if retentionCount <= 0 {
		return 0, nil // No cleanup if retention is 0 or negative
	}

	archives, err := ListArchives()
	if err != nil {
		return 0, fmt.Errorf("failed to list archives: %w", err)
	}

	// If we have fewer archives than the retention limit, nothing to do
	if len(archives) <= retentionCount {
		return 0, nil
	}

	// Sort archives by date (newest first)
	// Using a simple bubble sort since the list is usually small
	for i := 0; i < len(archives)-1; i++ {
		for j := 0; j < len(archives)-i-1; j++ {
			if archives[j].ArchivedAt.Before(archives[j+1].ArchivedAt) {
				archives[j], archives[j+1] = archives[j+1], archives[j]
			}
		}
	}

	// Delete archives beyond retention count
	deletedCount := 0
	for i := retentionCount; i < len(archives); i++ {
		if err := DeleteArchive(archives[i].Path); err != nil {
			// Continue deleting others even if one fails
			continue
		}
		deletedCount++
	}

	return deletedCount, nil
}

// RestoreArchive extracts an archived environment (for future use)
func RestoreArchive(archivePath, destPath string) error {
	// Open archive file
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer func() { _ = archiveFile.Close() }()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() { _ = gzipReader.Close() }()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// #nosec G305 - Archive extraction is intentional and from trusted source
		targetPath := filepath.Join(destPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// #nosec G115 - File mode conversion is safe in this context
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// Create parent directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			// Create file
			// #nosec G115 - File mode conversion is safe in this context
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			// #nosec G110 - Decompression bomb risk is acceptable for trusted archives
			if _, err := io.Copy(outFile, tarReader); err != nil {
				_ = outFile.Close()
				return fmt.Errorf("failed to write file content: %w", err)
			}
			_ = outFile.Close()
		}
	}

	return nil
}
