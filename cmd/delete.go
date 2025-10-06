package cmd

import (
	"fmt"
	"os"

	"github.com/hugofrely/envswitch/pkg/environment"
	"github.com/spf13/cobra"
)

var (
	deleteForce bool
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
			fmt.Println("Cancelled.")
			return nil
		}
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// TODO: Archive before deletion
	// archiveEnvironment(env)

	// Delete environment directory
	if err := os.RemoveAll(env.Path); err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	fmt.Printf("✅ Environment '%s' deleted successfully\n", name)

	return nil
}
