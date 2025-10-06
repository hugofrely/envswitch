package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hugofrely/envswitch/internal/archive"
	"github.com/hugofrely/envswitch/pkg/environment"
)

var (
	deleteForce     bool
	deleteNoArchive bool
)

var deleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"rm"},
	Short:   "Delete an environment",
	Long:    `Delete an environment and all its snapshots.`,
	Args:    cobra.ExactArgs(1),
	RunE:    runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation")
	deleteCmd.Flags().BoolVar(&deleteNoArchive, "no-archive", false, "Skip archiving before deletion")
}

func runDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Check if environment exists
	env, err := environment.LoadEnvironment(name)
	if err != nil {
		return fmt.Errorf("environment '%s' not found: %w", name, err)
	}

	// Check if it's the current environment
	current, _ := environment.GetCurrentEnvironment()
	if current != nil && current.Name == name {
		return fmt.Errorf("cannot delete active environment '%s'", name)
	}

	// Confirm deletion
	if !deleteForce {
		fmt.Printf("⚠️  Are you sure you want to delete '%s'? [y/N]: ", name)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			// If there's an error reading input, treat as "no"
			fmt.Println("Canceled.")
			return nil
		}
		if response != "y" && response != "Y" {
			fmt.Println("Canceled.")
			return nil
		}
	}

	// Archive before deletion (unless --no-archive is specified)
	var archivePath string
	if !deleteNoArchive {
		fmt.Println("📦 Archiving environment before deletion...")
		arch, err := archive.ArchiveEnvironment(env)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to archive environment: %v\n", err)
			fmt.Println("   Proceeding with deletion...")
		} else {
			archivePath = arch.Path
			fmt.Printf("✓ Archived to: %s\n", archivePath)
		}
	}

	// Delete environment directory
	if err := os.RemoveAll(env.Path); err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	fmt.Printf("✅ Environment '%s' deleted successfully\n", name)
	if archivePath != "" {
		fmt.Printf("   Archive saved at: %s\n", archivePath)
	}

	return nil
}
