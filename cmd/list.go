package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/hugofrely/envswitch/pkg/environment"
)

var (
	listDetailed bool
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all environments",
	Long:    `List all available environments with their status and basic information.`,
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&listDetailed, "detailed", false, "Show detailed information")
}

func runList(cmd *cobra.Command, args []string) error {
	environments, err := environment.ListEnvironments()
	if err != nil {
		return err
	}

	if len(environments) == 0 {
		fmt.Println("No environments found.")
		fmt.Println()
		fmt.Println("Create your first environment:")
		fmt.Println("  envswitch create myenv --from-current")
		return nil
	}

	current, _ := environment.GetCurrentEnvironment()
	var currentName string
	if current != nil {
		currentName = current.Name
	}

	fmt.Println("Available environments:")
	fmt.Println()

	for _, env := range environments {
		prefix := "  "
		suffix := ""

		if env.Name == currentName {
			prefix = "  * "
			suffix = " (active)"
		}

		fmt.Printf("%s%s%s", prefix, env.Name, suffix)

		if env.Description != "" {
			fmt.Printf(" - %s", env.Description)
		}
		fmt.Println()

		if listDetailed {
			if !env.LastUsed.IsZero() {
				fmt.Printf("                       Last used: %s\n", formatTimeAgo(env.LastUsed))
			}

			// Show enabled tools
			var enabledTools []string
			for toolName, toolConfig := range env.Tools {
				if toolConfig.Enabled {
					enabledTools = append(enabledTools, toolName)
				}
			}
			if len(enabledTools) > 0 {
				fmt.Printf("                       Tools: %s\n", strings.Join(enabledTools, ", "))
			}
			fmt.Println()
		}
	}

	fmt.Printf("Total: %d environment", len(environments))
	if len(environments) != 1 {
		fmt.Print("s")
	}
	fmt.Println()

	return nil
}

func formatTimeAgo(t time.Time) string {
	// Simple time ago formatting
	// TODO: Implement more sophisticated time formatting
	return t.Format("2006-01-02 15:04")
}
