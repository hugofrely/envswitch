package cmd

import (
	"fmt"

	"github.com/hugofrely/envswitch/pkg/environment"
	"github.com/spf13/cobra"
)

var (
	switchVerify bool
	switchDryRun bool
)

var switchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch to another environment",
	Long: `Switch to another environment by saving the current state
and restoring the target environment's snapshot.`,
	Args: cobra.ExactArgs(1),
	RunE: runSwitch,
}

func init() {
	rootCmd.AddCommand(switchCmd)
	switchCmd.Flags().BoolVar(&switchVerify, "verify", false, "Verify connectivity after switch")
	switchCmd.Flags().BoolVar(&switchDryRun, "dry-run", false, "Preview changes without applying")
}

func runSwitch(cmd *cobra.Command, args []string) error {
	targetName := args[0]

	// Load target environment
	_, err := environment.LoadEnvironment(targetName)
	if err != nil {
		return fmt.Errorf("failed to load environment '%s': %w", targetName, err)
	}

	// Get current environment
	currentEnv, err := environment.GetCurrentEnvironment()
	if err != nil {
		return fmt.Errorf("failed to get current environment: %w", err)
	}

	if currentEnv != nil && currentEnv.Name == targetName {
		fmt.Printf("Already on '%s'\n", targetName)
		return nil
	}

	var fromName string
	if currentEnv != nil {
		fromName = currentEnv.Name
	} else {
		fromName = "(none)"
	}

	if switchDryRun {
		fmt.Printf("Preview of changes (DRY RUN):\n\n")
		fmt.Printf("Would switch: %s → %s\n", fromName, targetName)
		fmt.Println()
		fmt.Println("No changes will be applied (use without --dry-run to apply)")
		return nil
	}

	fmt.Printf("🔄 Switching from '%s' to '%s'...\n", fromName, targetName)
	fmt.Println()

	// TODO: Implement actual switch logic
	// 1. Create security backup
	// 2. Save current state if exists
	// 3. Restore target state
	// 4. Execute post-switch hooks
	// 5. Update current.lock and history

	// For now, just update the current environment
	if err := environment.SetCurrentEnvironment(targetName); err != nil {
		return fmt.Errorf("failed to set current environment: %w", err)
	}

	fmt.Println("⚠️  Note: Full switch implementation coming soon")
	fmt.Println("         This currently only updates the active environment marker")
	fmt.Println()

	fmt.Printf("✅ Successfully switched to '%s'\n", targetName)

	if switchVerify {
		fmt.Println()
		fmt.Println("🔍 Verification:")
		fmt.Println("   (Verification not yet implemented)")
	}

	return nil
}
