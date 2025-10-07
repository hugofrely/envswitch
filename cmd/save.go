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

This captures snapshots of all enabled tools and updates the environment.

The save command works on the currently active environment. It will:
  - Capture current state of all enabled tools (gcloud, kubectl, aws, etc.)
  - Update snapshots in the active environment
  - Preserve tool configurations

Examples:
  # Save current state to active environment
  envswitch save

Note: You must have an active environment to use this command.
Use 'envswitch list' to see all environments and which one is active.`,
	Args: cobra.NoArgs,
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

	// Capture current state using the same function from create.go (which has a spinner)
	if err := captureCurrentState(currentEnv.Path, currentEnv); err != nil {
		return fmt.Errorf("failed to save current state: %w", err)
	}

	// Save environment metadata
	if err := currentEnv.Save(); err != nil {
		return fmt.Errorf("failed to save environment metadata: %w", err)
	}

	return nil
}
