package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hugofrely/envswitch/internal/archive"
)

var (
	importName  string
	importForce bool
	importAll   bool
)

var importCmd = &cobra.Command{
	Use:   "import <archive-path>",
	Short: "Import environment from archive file",
	Long: `Import an environment from a compressed archive file.

This allows you to:
  - Restore backed up environments
  - Import environments shared by colleagues
  - Migrate environments from other machines
  - Restore archived environments

The archive must be a .tar.gz file created by 'envswitch export'.

Examples:
  # Import an environment
  envswitch import work-backup.tar.gz

  # Import with a new name
  envswitch import work-backup.tar.gz --name work-restored

  # Import and overwrite existing environment
  envswitch import work-backup.tar.gz --force

  # Import all environments from a directory
  envswitch import ~/backups/ --all`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVarP(&importName, "name", "n", "", "New name for the imported environment")
	importCmd.Flags().BoolVarP(&importForce, "force", "f", false, "Overwrite existing environment")
	importCmd.Flags().BoolVar(&importAll, "all", false, "Import all archives from directory")
}

func runImport(cmd *cobra.Command, args []string) error {
	archivePath := args[0]

	fmt.Println("📥 Importing environment(s)...")
	fmt.Println()

	// Import all from directory
	if importAll {
		if err := archive.ImportAll(archivePath, importForce); err != nil {
			return fmt.Errorf("failed to import environments: %w", err)
		}

		fmt.Printf("✅ Environments imported from: %s\n", archivePath)
		return nil
	}

	// Validate single archive
	if !strings.HasSuffix(archivePath, ".tar.gz") && !strings.HasSuffix(archivePath, ".tgz") {
		return fmt.Errorf("invalid archive format: must be .tar.gz or .tgz")
	}

	// Import single archive
	options := archive.ImportOptions{
		ArchivePath: archivePath,
		NewName:     importName,
		Force:       importForce,
	}

	if err := archive.ImportEnvironment(archivePath, options); err != nil {
		return fmt.Errorf("failed to import environment: %w", err)
	}

	envName := importName
	if envName == "" {
		// Extract name from archive filename
		base := filepath.Base(archivePath)
		// Remove extensions
		envName = strings.TrimSuffix(base, ".tar.gz")
		envName = strings.TrimSuffix(envName, ".tgz")
		// Remove timestamp if present (format: envname-YYYYMMDD-HHMMSS)
		parts := strings.Split(envName, "-")
		if len(parts) >= 3 {
			// Remove last 2 parts if they look like timestamp
			lastPart := parts[len(parts)-1]
			secondLastPart := parts[len(parts)-2]
			if len(lastPart) == 6 && len(secondLastPart) == 8 {
				envName = strings.Join(parts[:len(parts)-2], "-")
			}
		}
	}

	fmt.Printf("✅ Environment '%s' imported successfully\n", envName)
	fmt.Printf("   You can now switch to it with: envswitch switch %s\n", envName)

	return nil
}
