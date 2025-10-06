package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hugofrely/envswitch/pkg/environment"
)

var showCmd = &cobra.Command{
	Use:               "show <name>",
	Short:             "Show details of an environment",
	Long:              `Display detailed information about a specific environment.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeEnvironmentNames,
	RunE:              runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	env, err := environment.LoadEnvironment(name)
	if err != nil {
		return fmt.Errorf("failed to load environment '%s': %w", name, err)
	}

	fmt.Printf("Environment: %s\n", env.Name)
	if env.Description != "" {
		fmt.Printf("Description: %s\n", env.Description)
	}
	fmt.Printf("Created: %s\n", env.CreatedAt.Format("2006-01-02 15:04:05"))
	if !env.LastUsed.IsZero() {
		fmt.Printf("Last used: %s\n", env.LastUsed.Format("2006-01-02 15:04:05"))
	}
	if !env.LastSnapshot.IsZero() {
		fmt.Printf("Last snapshot: %s\n", env.LastSnapshot.Format("2006-01-02 15:04:05"))
	}
	fmt.Println()

	fmt.Println("📸 Snapshot Contents:")
	fmt.Println()

	for toolName, toolConfig := range env.Tools {
		if !toolConfig.Enabled {
			continue
		}

		fmt.Printf("  ✓ %s\n", toolName)
		if len(toolConfig.Metadata) > 0 {
			for key, value := range toolConfig.Metadata {
				fmt.Printf("    - %s: %v\n", key, value)
			}
		}
		fmt.Println()
	}

	if len(env.EnvVars) > 0 {
		fmt.Printf("  ✓ Environment Variables (%d)\n", len(env.EnvVars))
		for key, value := range env.EnvVars {
			fmt.Printf("    %s=%s\n", key, value)
		}
		fmt.Println()
	}

	if len(env.Tags) > 0 {
		fmt.Printf("Tags: %v\n", env.Tags)
	}

	return nil
}
