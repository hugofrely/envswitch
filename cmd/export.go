package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hugofrely/envswitch/internal/archive"
)

var (
	exportOutput string
	exportAll    bool
)

var exportCmd = &cobra.Command{
	Use:   "export [environment-name...]",
	Short: "Export environments to archive files",
	Long: `Export one or more environments to compressed archive files.

This allows you to:
  - Backup environments
  - Share environments with colleagues
  - Migrate environments between machines
  - Archive environments for long-term storage

Examples:
  # Export a single environment
  envswitch export work --output work-backup.tar.gz

  # Export multiple environments
  envswitch export work personal --output ~/backups/

  # Export all environments
  envswitch export --all --output all-envs/

  # Export to current directory (default)
  envswitch export work`,
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output path (file or directory)")
	exportCmd.Flags().BoolVar(&exportAll, "all", false, "Export all environments")
}

func runExport(cmd *cobra.Command, args []string) error {
	// Validate arguments
	if exportAll && len(args) > 0 {
		return fmt.Errorf("cannot specify environment names with --all flag")
	}

	if !exportAll && len(args) == 0 {
		return fmt.Errorf("must specify at least one environment name or use --all flag")
	}

	fmt.Println("📦 Exporting environments...")
	fmt.Println()

	// Export all environments
	if exportAll {
		output := exportOutput
		if output == "" {
			output = "envswitch-export"
		}

		if err := archive.ExportAllEnvironments(output); err != nil {
			return fmt.Errorf("failed to export environments: %w", err)
		}

		fmt.Printf("✅ All environments exported to: %s\n", output)
		return nil
	}

	// Export single environment
	if len(args) == 1 {
		envName := args[0]
		output := exportOutput
		if output == "" {
			output = fmt.Sprintf("%s-export.tar.gz", envName)
		}

		if err := archive.ExportEnvironment(envName, output); err != nil {
			return fmt.Errorf("failed to export environment: %w", err)
		}

		fmt.Printf("✅ Environment '%s' exported to: %s\n", envName, output)
		return nil
	}

	// Export multiple environments
	output := exportOutput
	if output == "" {
		output = "envswitch-export"
	}

	if err := archive.ExportEnvironments(args, output); err != nil {
		return fmt.Errorf("failed to export environments: %w", err)
	}

	fmt.Printf("✅ %d environment(s) exported to: %s\n", len(args), output)
	return nil
}
