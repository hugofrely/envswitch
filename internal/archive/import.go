package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hugofrely/envswitch/pkg/environment"
	"github.com/hugofrely/envswitch/pkg/spinner"
)

// ImportOptions defines options for importing environments
type ImportOptions struct {
	ArchivePath string // Path to archive file
	NewName     string // Optional: new name for the environment
	Force       bool   // Overwrite existing environment
}

// ImportEnvironment imports an environment from an archive file
func ImportEnvironment(archivePath string, options ImportOptions) error {
	spin := spinner.New(fmt.Sprintf("Importing %s", filepath.Base(archivePath)))
	spin.Start()

	// Check if archive exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		spin.Error(fmt.Sprintf("Archive file not found: %s", archivePath))
		return fmt.Errorf("archive file not found: %s", archivePath)
	}

	// Validate archive format
	if !strings.HasSuffix(archivePath, ".tar.gz") && !strings.HasSuffix(archivePath, ".tgz") {
		spin.Error("Invalid archive format")
		return fmt.Errorf("invalid archive format: must be .tar.gz or .tgz")
	}

	spin.Update("Opening archive...")
	// Open archive
	file, err := os.Open(archivePath)
	if err != nil {
		spin.Error("Failed to open archive")
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		spin.Error("Failed to read archive")
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract to temporary directory first
	tempDir, err := os.MkdirTemp("", "envswitch-import-*")
	if err != nil {
		spin.Error("Failed to create temp directory")
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract archive
	spin.Update("Extracting archive...")
	envName, err := extractTarArchive(tarReader, tempDir)
	if err != nil {
		spin.Error("Failed to extract archive")
		return err
	}

	// Use new name if specified
	finalEnvName := envName
	if options.NewName != "" {
		finalEnvName = options.NewName
	}

	// Check if environment already exists
	envDir, err := environment.GetEnvironmentsDir()
	if err != nil {
		spin.Error("Failed to get environments directory")
		return fmt.Errorf("failed to get environments directory: %w", err)
	}

	finalEnvPath := filepath.Join(envDir, finalEnvName)
	if _, err := os.Stat(finalEnvPath); err == nil {
		if !options.Force {
			spin.Error(fmt.Sprintf("Environment '%s' already exists", finalEnvName))
			return fmt.Errorf("environment '%s' already exists (use --force to overwrite)", finalEnvName)
		}
		// Remove existing environment
		spin.Update(fmt.Sprintf("Removing existing environment '%s'", finalEnvName))
		if err := os.RemoveAll(finalEnvPath); err != nil {
			spin.Error("Failed to remove existing environment")
			return fmt.Errorf("failed to remove existing environment: %w", err)
		}
	}

	// Move from temp to final location
	spin.Update(fmt.Sprintf("Installing environment '%s'", finalEnvName))
	extractedPath := filepath.Join(tempDir, envName)
	if err := os.Rename(extractedPath, finalEnvPath); err != nil {
		// If rename fails (cross-device), copy instead
		if err := copyDir(extractedPath, finalEnvPath); err != nil {
			spin.Error("Failed to install environment")
			return fmt.Errorf("failed to move environment: %w", err)
		}
	}

	// Update metadata if name changed
	if options.NewName != "" && options.NewName != envName {
		env, err := environment.LoadEnvironment(finalEnvName)
		if err == nil {
			env.Name = finalEnvName
			env.Path = finalEnvPath
			if err := env.Save(); err != nil {
				spin.Error("Failed to update environment metadata")
				return fmt.Errorf("failed to update environment name in metadata: %w", err)
			}
		}
	}

	spin.Success(fmt.Sprintf("Imported environment '%s'", finalEnvName))
	return nil
}

// ImportAll imports all archives from a directory
func ImportAll(dirPath string, force bool) error {
	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return fmt.Errorf("directory not found: %s", dirPath)
	}

	// Find all .tar.gz files
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	imported := 0
	archives := []string{}

	// First, collect all archives
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".tar.gz") && !strings.HasSuffix(name, ".tgz") {
			continue
		}
		archives = append(archives, name)
	}

	// Import each archive with progress
	for i, name := range archives {
		archivePath := filepath.Join(dirPath, name)
		options := ImportOptions{
			ArchivePath: archivePath,
			Force:       force,
		}

		// ImportEnvironment has its own spinner
		if err := ImportEnvironment(archivePath, options); err != nil {
			fmt.Printf("✗ [%d/%d] Failed to import %s\n", i+1, len(archives), name)
			continue
		}

		imported++
	}

	if imported == 0 {
		return fmt.Errorf("no archives were imported successfully")
	}

	return nil
}

// extractTarArchive extracts a tar archive and returns the environment name
func extractTarArchive(tarReader *tar.Reader, tempDir string) (string, error) {
	var envName string
	for {
		header, nextErr := tarReader.Next()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			return "", fmt.Errorf("failed to read tar: %w", nextErr)
		}

		// Extract environment name from first directory
		if envName == "" && header.Typeflag == tar.TypeDir {
			envName = filepath.Base(header.Name)
		}

		target := filepath.Join(tempDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			if err := extractTarFile(tarReader, target, header); err != nil {
				return "", err
			}
		}
	}

	if envName == "" {
		return "", fmt.Errorf("could not determine environment name from archive")
	}

	return envName, nil
}

// extractTarFile extracts a single file from tar archive
func extractTarFile(tarReader *tar.Reader, target string, header *tar.Header) error {
	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Create file
	outFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, tarReader); err != nil {
		return fmt.Errorf("failed to extract file: %w", err)
	}

	// Set permissions
	if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath)
	})
}
