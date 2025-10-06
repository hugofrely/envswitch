package cmd

import (
	"github.com/spf13/cobra"

	"github.com/hugofrely/envswitch/pkg/environment"
)

// completeEnvironmentNames provides completion for environment names
func completeEnvironmentNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	envs, err := environment.ListEnvironments()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, env := range envs {
		names = append(names, env.Name)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}
