package cmd

import (
	"fmt"
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

	// Success message is already displayed by the spinner in ImportEnvironment
	return nil
}
