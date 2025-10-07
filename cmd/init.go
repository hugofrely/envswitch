package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize EnvSwitch",
	Long:  `Initialize EnvSwitch by creating the configuration directory and default config file.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	envswitchDir := filepath.Join(home, ".envswitch")

	// Create main directory
	fmt.Println("Creating ~/.envswitch/...")
	if err := os.MkdirAll(envswitchDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{"environments", "auto-backups"}
	for _, dir := range dirs {
		dirPath := filepath.Join(envswitchDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	// Create default config
	configPath := filepath.Join(envswitchDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := map[string]interface{}{
			"version":                   "1.0",
			"auto_save_before_switch":   true,
			"verify_after_switch":       false,
			"backup_retention":          10,
			"enable_prompt_integration": true,
			"prompt_format":             "({name})",
			"prompt_color":              "blue",
			"log_level":                 "info",
			"log_file":                  filepath.Join(envswitchDir, "envswitch.log"),
			"exclude_tools":             []string{},
			"color_output":              true,
			"show_timestamps":           false,
			"backup_before_switch":      true,
		}

		data, err := yaml.Marshal(defaultConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
	}

	// Create history log
	historyPath := filepath.Join(envswitchDir, "history.log")
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		if err := os.WriteFile(historyPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create history log: %w", err)
		}
	}

	fmt.Println("✓ Configuration directory created")
	fmt.Println("✓ Default config created")
	fmt.Println("✓ Shell integration available")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Create your first environment:")
	fmt.Println("     envswitch create work --from-current")
	fmt.Println()
	fmt.Println("  2. Install shell integration:")
	fmt.Println("     envswitch shell install")
	fmt.Println()
	fmt.Println("  3. Read the docs:")
	fmt.Println("     envswitch help")

	return nil
}
