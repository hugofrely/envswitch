package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hugofrely/envswitch/pkg/environment"
	"github.com/hugofrely/envswitch/pkg/tools"
	"github.com/spf13/cobra"
)

var (
	createFromCurrent bool
	createEmpty       bool
	createFrom        string
	createDescription string
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new environment",
	Long: `Create a new environment from the current system state,
another environment, or as an empty template.`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().BoolVar(&createFromCurrent, "from-current", false, "Create from current system state")
	createCmd.Flags().BoolVar(&createEmpty, "empty", false, "Create empty environment")
	createCmd.Flags().StringVar(&createFrom, "from", "", "Clone from existing environment")
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "Environment description")
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Validate name
	if name == "" {
		return fmt.Errorf("environment name cannot be empty")
	}

	// Check if environment already exists
	envDir, err := environment.GetEnvironmentsDir()
	if err != nil {
		return err
	}

	envPath := filepath.Join(envDir, name)
	if _, err := os.Stat(envPath); !os.IsNotExist(err) {
		return fmt.Errorf("environment '%s' already exists", name)
	}

	// Create environment directory structure
	if err := os.MkdirAll(envPath, 0755); err != nil {
		return fmt.Errorf("failed to create environment directory: %w", err)
	}

	snapshotsPath := filepath.Join(envPath, "snapshots")
	if err := os.MkdirAll(snapshotsPath, 0755); err != nil {
		return fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	// Create environment object
	env := &environment.Environment{
		Name:        name,
		Description: createDescription,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		LastUsed:    time.Time{},
		Tools:       make(map[string]environment.ToolConfig),
		EnvVars:     make(map[string]string),
		Path:        envPath,
	}

	// Initialize tools
	toolNames := []string{"gcloud", "kubectl", "aws", "azure", "docker", "terraform", "git"}
	for _, toolName := range toolNames {
		env.Tools[toolName] = environment.ToolConfig{
			Enabled:      createFromCurrent, // Only enable if creating from current
			SnapshotPath: filepath.Join("snapshots", toolName),
			Metadata:     make(map[string]interface{}),
		}
	}

	if createFromCurrent {
		fmt.Println("📸 Capturing current state...")
		fmt.Println()

		// Capture snapshots for each tool
		capturedCount := 0
		availableTools := map[string]tools.Tool{
			"gcloud":  tools.NewGCloudTool(),
			"kubectl": tools.NewKubectlTool(),
			"aws":     tools.NewAWSTool(),
			"docker":  tools.NewDockerTool(),
			"git":     tools.NewGitTool(),
		}

		for toolName, toolImpl := range availableTools {
			// Check if tool is installed
			if !toolImpl.IsInstalled() {
				fmt.Printf("  ⊘ %s (not installed)\n", toolName)
				env.Tools[toolName] = environment.ToolConfig{
					Enabled:      false,
					SnapshotPath: filepath.Join("snapshots", toolName),
					Metadata:     make(map[string]interface{}),
				}
				continue
			}

			// Create snapshot path
			snapshotPath := filepath.Join(envPath, "snapshots", toolName)

			// Capture snapshot
			if err := toolImpl.Snapshot(snapshotPath); err != nil {
				fmt.Printf("  ⚠ %s (failed: %v)\n", toolName, err)
				env.Tools[toolName] = environment.ToolConfig{
					Enabled:      false,
					SnapshotPath: filepath.Join("snapshots", toolName),
					Metadata:     make(map[string]interface{}),
				}
				continue
			}

			// Get metadata
			metadata, err := toolImpl.GetMetadata()
			if err != nil {
				metadata = make(map[string]interface{})
			}

			// Update environment config
			env.Tools[toolName] = environment.ToolConfig{
				Enabled:      true,
				SnapshotPath: filepath.Join("snapshots", toolName),
				Metadata:     metadata,
			}

			// Display success with metadata
			fmt.Printf("  ✓ %s", toolName)
			if len(metadata) > 0 {
				fmt.Print(" (")
				first := true
				for key, value := range metadata {
					if !first {
						fmt.Print(", ")
					}
					fmt.Printf("%s: %v", key, value)
					first = false
				}
				fmt.Print(")")
			}
			fmt.Println()

			capturedCount++
		}

		// Update snapshot info
		env.LastSnapshot = time.Now()
		fmt.Println()
		fmt.Printf("✅ Captured %d tool(s) successfully\n", capturedCount)
		fmt.Println()
	}

	// Save metadata
	if err := env.Save(); err != nil {
		return fmt.Errorf("failed to save environment: %w", err)
	}

	// Create empty env-vars.env file
	envVarsPath := filepath.Join(envPath, "env-vars.env")
	if err := os.WriteFile(envVarsPath, []byte("# Environment variables\n"), 0644); err != nil {
		return fmt.Errorf("failed to create env-vars.env: %w", err)
	}

	fmt.Printf("✅ Environment '%s' created successfully\n", name)
	fmt.Printf("   Path: %s\n", envPath)
	fmt.Println()
	fmt.Printf("Next: envswitch switch %s\n", name)

	return nil
}
