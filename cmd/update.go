package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hugofrely/envswitch/internal/updater"
	"github.com/hugofrely/envswitch/internal/version"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates and get update instructions",
	Long: `Check if a new version of envswitch is available.
If an update is available, provides instructions on how to update.`,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	fmt.Println("Checking for updates...")

	info, err := updater.CheckForUpdate()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if info.CurrentVersion == version.DevVersion {
		fmt.Println("⚠️  Running development version - update check skipped")
		return nil
	}

	if !info.Available {
		fmt.Printf("✓ You are already running the latest version (%s)\n", info.CurrentVersion)
		return nil
	}

	fmt.Printf("\n🎉 A new version is available!\n\n")
	fmt.Printf("  Current version: %s\n", info.CurrentVersion)
	fmt.Printf("  Latest version:  %s\n", info.LatestVersion)
	fmt.Printf("  Release URL:     %s\n\n", info.ReleaseURL)

	fmt.Println("To update, run one of the following:")
	fmt.Println()
	fmt.Printf("  # Using curl:\n")
	fmt.Printf("  %s\n\n", updater.GetUpdateCommand())
	fmt.Printf("  # Or using wget:\n")
	fmt.Printf("  wget -qO- https://raw.githubusercontent.com/hugofrely/envswitch/main/install.sh | bash\n\n")

	if info.DownloadURL != "" {
		fmt.Printf("Or download the binary directly for your platform:\n")
		fmt.Printf("  %s\n\n", info.DownloadURL)
	}

	fmt.Printf("For more installation options, visit:\n")
	fmt.Printf("  https://github.com/hugofrely/envswitch#installation\n")

	return nil
}
