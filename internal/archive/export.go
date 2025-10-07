package archive

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hugofrely/envswitch/pkg/environment"
	"github.com/hugofrely/envswitch/pkg/spinner"
)

// ExportOptions defines options for exporting environments
type ExportOptions struct {
	OutputPath string   // Path to output file
	EnvNames   []string // Specific environments to export (empty = all)
	All        bool     // Export all environments
}

// ExportEnvironment exports a single environment to a file
func ExportEnvironment(envName, outputPath string) error {
	spin := spinner.New(fmt.Sprintf("Exporting '%s'", envName))
	spin.Start()

	// Load the environment
	env, err := environment.LoadEnvironment(envName)
	if err != nil {
		spin.Error(fmt.Sprintf("Failed to load environment '%s'", envName))
		return fmt.Errorf("failed to load environment '%s': %w", envName, err)
	}

	// Create archive
	spin.Update(fmt.Sprintf("Creating archive for '%s'", envName))
	archive, err := ArchiveEnvironment(env)
	if err != nil {
		spin.Error(fmt.Sprintf("Failed to create archive for '%s'", envName))
		return fmt.Errorf("failed to archive environment: %w", err)
	}

	// If no output path specified, use current directory
	if outputPath == "" {
		outputPath = fmt.Sprintf("%s-export.tar.gz", envName)
	}

	// Copy archive to output path
	spin.Update(fmt.Sprintf("Writing to %s", outputPath))
	if err := copyFile(archive.Path, outputPath); err != nil {
		spin.Error(fmt.Sprintf("Failed to write archive for '%s'", envName))
		return fmt.Errorf("failed to copy archive: %w", err)
	}

	spin.Success(fmt.Sprintf("Exported '%s' to %s", envName, outputPath))
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
	for i, env := range environments {
		spin := spinner.New(fmt.Sprintf("[%d/%d] Exporting '%s'", i+1, len(environments), env.Name))
		spin.Start()

		archive, err := ArchiveEnvironment(env)
		if err != nil {
			spin.Error(fmt.Sprintf("[%d/%d] Failed to export '%s'", i+1, len(environments), env.Name))
			continue
		}

		// Copy to output directory
		destPath := filepath.Join(outputDir, filepath.Base(archive.Path))
		spin.Update(fmt.Sprintf("[%d/%d] Writing '%s' to %s", i+1, len(environments), env.Name, destPath))
		if err := copyFile(archive.Path, destPath); err != nil {
			spin.Error(fmt.Sprintf("[%d/%d] Failed to write '%s'", i+1, len(environments), env.Name))
			continue
		}

		spin.Success(fmt.Sprintf("[%d/%d] Exported '%s'", i+1, len(environments), env.Name))
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
	for i, envName := range envNames {
		spin := spinner.New(fmt.Sprintf("[%d/%d] Exporting '%s'", i+1, len(envNames), envName))
		spin.Start()

		env, err := environment.LoadEnvironment(envName)
		if err != nil {
			spin.Error(fmt.Sprintf("[%d/%d] Failed to load '%s'", i+1, len(envNames), envName))
			continue
		}

		archive, err := ArchiveEnvironment(env)
		if err != nil {
			spin.Error(fmt.Sprintf("[%d/%d] Failed to export '%s'", i+1, len(envNames), envName))
			continue
		}

		// Copy to output directory
		destPath := filepath.Join(outputDir, filepath.Base(archive.Path))
		spin.Update(fmt.Sprintf("[%d/%d] Writing '%s' to %s", i+1, len(envNames), envName, destPath))
		if err := copyFile(archive.Path, destPath); err != nil {
			spin.Error(fmt.Sprintf("[%d/%d] Failed to write '%s'", i+1, len(envNames), envName))
			continue
		}

		spin.Success(fmt.Sprintf("[%d/%d] Exported '%s'", i+1, len(envNames), envName))
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
