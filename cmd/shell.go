package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hugofrely/envswitch/internal/config"
	"github.com/hugofrely/envswitch/internal/shell"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Shell integration commands",
	Long:  `Commands for integrating envswitch with your shell (bash, zsh, fish).`,
}

var shellInitCmd = &cobra.Command{
	Use:   "init [bash|zsh|fish]",
	Short: "Generate shell initialization script",
	Long: `Generate shell initialization script to enable prompt integration.

Add the output to your shell's configuration file:
  bash: ~/.bashrc or ~/.bash_profile
  zsh:  ~/.zshrc
  fish: ~/.config/fish/config.fish

Example:
  envswitch shell init bash >> ~/.bashrc`,
	Args:              cobra.ExactArgs(1),
	ValidArgs:         []string{"bash", "zsh", "fish"},
	RunE:              runShellInit,
	DisableAutoGenTag: true,
}

var shellInstallCmd = &cobra.Command{
	Use:   "install [bash|zsh|fish]",
	Short: "Install shell integration automatically",
	Long: `Automatically install shell integration by appending the initialization
script to your shell's configuration file.

This command will:
  1. Generate the appropriate shell script
  2. Append it to your shell's config file
  3. Display instructions to reload your shell`,
	Args:              cobra.ExactArgs(1),
	ValidArgs:         []string{"bash", "zsh", "fish"},
	RunE:              runShellInstall,
	DisableAutoGenTag: true,
}

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.AddCommand(shellInitCmd)
	shellCmd.AddCommand(shellInstallCmd)
}

func runShellInit(cmd *cobra.Command, args []string) error {
	shellType := args[0]

	// Load config for prompt settings
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	script, err := shell.GenerateInitScript(shellType, cfg)
	if err != nil {
		return fmt.Errorf("failed to generate init script: %w", err)
	}

	fmt.Print(script)
	return nil
}

func runShellInstall(cmd *cobra.Command, args []string) error {
	shellType := args[0]

	// Load config for prompt settings
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	configFile, err := shell.InstallShellIntegration(shellType, cfg)
	if err != nil {
		return fmt.Errorf("failed to install shell integration: %w", err)
	}

	fmt.Printf("✅ Shell integration installed successfully!\n\n")
	fmt.Printf("Configuration file updated: %s\n\n", configFile)
	fmt.Printf("To activate the changes, run:\n")

	switch shellType {
	case "bash":
		fmt.Printf("  source %s\n", configFile)
	case "zsh":
		fmt.Printf("  source %s\n", configFile)
	case "fish":
		fmt.Printf("  source %s\n", configFile)
	}

	fmt.Println("\nOr simply restart your shell.")

	return nil
}
