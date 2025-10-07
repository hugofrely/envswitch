package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hugofrely/envswitch/pkg/environment"
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save the current system state to the active environment",
	Long: `Save the current system state to the active environment.
This captures snapshots of all enabled tools and updates the environment.`,
	RunE: runSave,
}

func init() {
	rootCmd.AddCommand(saveCmd)
}

func runSave(cmd *cobra.Command, args []string) error {
	// Get current environment
	currentEnv, err := environment.GetCurrentEnvironment()
	if err != nil {
		return fmt.Errorf("failed to get current environment: %w", err)
	}

	if currentEnv == nil {
		return fmt.Errorf("no active environment. Use 'envswitch create' to create one first")
	}

	fmt.Printf("💾 Saving current state to '%s'...\n", currentEnv.Name)
	fmt.Println()

	// Snapshot the current environment
	if err := snapshotCurrentEnvironment(currentEnv); err != nil {
		return fmt.Errorf("failed to save current state: %w", err)
	}

	fmt.Printf("✅ Successfully saved current state to '%s'\n", currentEnv.Name)
	return nil
}
