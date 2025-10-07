package archive

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hugofrely/envswitch/pkg/environment"
)

// ExportOptions defines options for exporting environments
type ExportOptions struct {
	OutputPath string   // Path to output file
	EnvNames   []string // Specific environments to export (empty = all)
	All        bool     // Export all environments
}

// ExportEnvironment exports a single environment to a file
func ExportEnvironment(envName, outputPath string) error {
	// Load the environment
	env, err := environment.LoadEnvironment(envName)
	if err != nil {
		return fmt.Errorf("failed to load environment '%s': %w", envName, err)
	}

	// Create archive
	archive, err := ArchiveEnvironment(env)
	if err != nil {
		return fmt.Errorf("failed to archive environment: %w", err)
	}

	// If no output path specified, use current directory
	if outputPath == "" {
		outputPath = fmt.Sprintf("%s-export.tar.gz", envName)
	}

	// Copy archive to output path
	if err := copyFile(archive.Path, outputPath); err != nil {
		return fmt.Errorf("failed to copy archive: %w", err)
	}

	return nil
}

// ExportAllEnvironments exports all environments to a single archive or directory
func ExportAllEnvironments(outputPath string) error {
	// Load all environments
	environments, err := environment.ListEnvironments()
	if err != nil {
		return fmt.Errorf("failed to list environments: %w", err)
	}

	if len(environments) == 0 {
		return fmt.Errorf("no environments to export")
	}

	// Create output directory
	outputDir := outputPath
	if outputDir == "" {
		outputDir = "envswitch-export"
	}

	// Remove .tar.gz extension if present and use as directory
	if filepath.Ext(outputDir) == ".gz" {
		outputDir = outputDir[:len(outputDir)-7] // Remove .tar.gz
	} else if filepath.Ext(outputDir) == ".tar" {
		outputDir = outputDir[:len(outputDir)-4] // Remove .tar
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Export each environment
	exported := 0
	for _, env := range environments {
		archive, err := ArchiveEnvironment(env)
		if err != nil {
			fmt.Printf("Warning: Failed to export '%s': %v\n", env.Name, err)
			continue
		}

		// Copy to output directory
		destPath := filepath.Join(outputDir, filepath.Base(archive.Path))
		if err := copyFile(archive.Path, destPath); err != nil {
			fmt.Printf("Warning: Failed to copy archive for '%s': %v\n", env.Name, err)
			continue
		}

		exported++
	}

	if exported == 0 {
		return fmt.Errorf("no environments were exported successfully")
	}

	return nil
}

// ExportEnvironments exports multiple specific environments
func ExportEnvironments(envNames []string, outputDir string) error {
	if len(envNames) == 0 {
		return fmt.Errorf("no environments specified")
	}

	// Create output directory
	if outputDir == "" {
		outputDir = "envswitch-export"
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Export each specified environment
	exported := 0
	for _, envName := range envNames {
		env, err := environment.LoadEnvironment(envName)
		if err != nil {
			fmt.Printf("Warning: Failed to load '%s': %v\n", envName, err)
			continue
		}

		archive, err := ArchiveEnvironment(env)
		if err != nil {
			fmt.Printf("Warning: Failed to export '%s': %v\n", envName, err)
			continue
		}

		// Copy to output directory
		destPath := filepath.Join(outputDir, filepath.Base(archive.Path))
		if err := copyFile(archive.Path, destPath); err != nil {
			fmt.Printf("Warning: Failed to copy archive for '%s': %v\n", envName, err)
			continue
		}

		exported++
	}

	if exported == 0 {
		return fmt.Errorf("no environments were exported successfully")
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, copyErr := destFile.ReadFrom(sourceFile); copyErr != nil {
		return copyErr
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}
