package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/hugofrely/envswitch/internal/archive"
	"github.com/hugofrely/envswitch/internal/config"
	"github.com/hugofrely/envswitch/internal/history"
	"github.com/hugofrely/envswitch/internal/hooks"
	"github.com/hugofrely/envswitch/internal/logger"
	"github.com/hugofrely/envswitch/pkg/environment"
	"github.com/hugofrely/envswitch/pkg/tools"
)

var (
	switchVerify   bool
	switchDryRun   bool
	switchNoBackup bool
	switchNoHooks  bool
)

var switchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch to another environment",
	Long: `Switch to another environment by saving the current state
and restoring the target environment's snapshot.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeEnvironmentNames,
	RunE:              runSwitch,
}

func init() {
	rootCmd.AddCommand(switchCmd)
	switchCmd.Flags().BoolVar(&switchVerify, "verify", false, "Verify connectivity after switch")
	switchCmd.Flags().BoolVar(&switchDryRun, "dry-run", false, "Preview changes without applying")
	switchCmd.Flags().BoolVar(&switchNoBackup, "no-backup", false, "Skip creating backup archive")
	switchCmd.Flags().BoolVar(&switchNoHooks, "no-hooks", false, "Skip executing pre/post hooks")
}

func runSwitch(cmd *cobra.Command, args []string) error {
	targetName := args[0]

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Warn("Failed to load config, using defaults: %v", err)
		cfg = config.DefaultConfig()
	}

	// Initialize logger and output
	if logErr := logger.InitLogger(cfg); logErr != nil {
		logger.Warn("Failed to initialize logger: %v", logErr)
	}
	defer logger.Close()

	// Load target environment
	if _, loadErr := environment.LoadEnvironment(targetName); loadErr != nil {
		return fmt.Errorf("failed to load environment '%s': %w", targetName, loadErr)
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

	fromName := getFromName(currentEnv)

	if switchDryRun {
		return handleDryRun(fromName, targetName)
	}

	// Check auto-save configuration
	if currentEnv != nil && cfg.AutoSaveBeforeSwitch == "prompt" {
		fmt.Printf("\n💾 Save current environment '%s' before switching? (y/N): ", currentEnv.Name)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			logger.Info("Skipping auto-save as per user choice")
		}
	}

	return performSwitch(currentEnv, targetName, fromName, cfg)
}

func getFromName(currentEnv *environment.Environment) string {
	if currentEnv != nil {
		return currentEnv.Name
	}
	return "(none)"
}

func handleDryRun(fromName, targetName string) error {
	fmt.Printf("Preview of changes (DRY RUN):\n\n")
	fmt.Printf("Would switch: %s → %s\n", fromName, targetName)
	fmt.Println()
	fmt.Println("No changes will be applied (use without --dry-run to apply)")
	return nil
}

func performSwitch(currentEnv *environment.Environment, targetName, fromName string, cfg *config.Config) error {
	startTime := time.Now()

	targetEnv, err := environment.LoadEnvironment(targetName)
	if err != nil {
		return err
	}

	logger.Info("Switching from '%s' to '%s'", fromName, targetName)
	fmt.Printf("🔄 Switching from '%s' to '%s'...\n", fromName, targetName)
	fmt.Println()

	historyEntry := history.SwitchEntry{
		Timestamp: startTime,
		From:      fromName,
		To:        targetName,
		Success:   false,
	}

	backupPath, err := createBackup(currentEnv, &historyEntry)
	if err != nil {
		return err
	}

	if saveErr := saveCurrentState(currentEnv); saveErr != nil {
		return saveErr
	}

	if hookErr := executePreSwitchHooks(targetEnv, targetName, &historyEntry, startTime); hookErr != nil {
		return hookErr
	}

	toolCount, err := restoreTargetState(targetEnv, &historyEntry, startTime)
	if err != nil {
		return err
	}
	historyEntry.ToolsCount = toolCount

	executePostSwitchHooks(targetEnv, targetName)

	if err := finalizeSwitch(targetEnv, targetName, &historyEntry, startTime, backupPath); err != nil {
		return err
	}

	return nil
}

func createBackup(currentEnv *environment.Environment, entry *history.SwitchEntry) (string, error) {
	if currentEnv == nil || switchNoBackup {
		return "", nil
	}

	fmt.Println("📦 Creating security backup...")
	backup, backupErr := archive.ArchiveEnvironment(currentEnv)
	if backupErr != nil {
		fmt.Printf("⚠️  Warning: Failed to create backup: %v\n", backupErr)
		fmt.Println("   Proceeding with switch...")
		return "", nil
	}

	entry.BackupPath = backup.Path
	fmt.Printf("✓ Backup created: %s\n\n", filepath.Base(backup.Path))
	return backup.Path, nil
}

func saveCurrentState(currentEnv *environment.Environment) error {
	if currentEnv == nil {
		return nil
	}

	fmt.Println("💾 Saving current state...")
	if err := snapshotCurrentEnvironment(currentEnv); err != nil {
		return fmt.Errorf("failed to save current state: %w", err)
	}
	fmt.Println("✓ Current state saved")
	fmt.Println()
	return nil
}

func executePreSwitchHooks(targetEnv *environment.Environment, targetName string, entry *history.SwitchEntry, startTime time.Time) error {
	if switchNoHooks || len(targetEnv.Hooks.PreSwitch) == 0 {
		return nil
	}

	fmt.Println("🔧 Running pre-switch hooks...")
	if err := hooks.ExecuteHooks(targetEnv.Hooks.PreSwitch, targetName); err != nil {
		entry.ErrorMsg = fmt.Sprintf("pre-switch hook failed: %v", err)
		entry.DurationMs = time.Since(startTime).Milliseconds()
		recordHistory(entry)
		return fmt.Errorf("pre-switch hook failed: %w", err)
	}
	fmt.Println()
	return nil
}

func restoreTargetState(targetEnv *environment.Environment, entry *history.SwitchEntry, startTime time.Time) (int, error) {
	fmt.Println("🔄 Restoring target environment state...")
	toolCount, err := restoreEnvironment(targetEnv)
	if err != nil {
		entry.ErrorMsg = fmt.Sprintf("restore failed: %v", err)
		entry.DurationMs = time.Since(startTime).Milliseconds()
		recordHistory(entry)
		return 0, fmt.Errorf("failed to restore target state: %w", err)
	}
	fmt.Printf("✓ Restored %d tool(s)\n\n", toolCount)
	return toolCount, nil
}

func executePostSwitchHooks(targetEnv *environment.Environment, targetName string) {
	if switchNoHooks || len(targetEnv.Hooks.PostSwitch) == 0 {
		return
	}

	fmt.Println("🔧 Running post-switch hooks...")
	if err := hooks.ExecuteHooks(targetEnv.Hooks.PostSwitch, targetName); err != nil {
		fmt.Printf("⚠️  Warning: Post-switch hook failed: %v\n", err)
	}
	fmt.Println()
}

func finalizeSwitch(targetEnv *environment.Environment, targetName string, entry *history.SwitchEntry, startTime time.Time, backupPath string) error {
	// Load config for verification settings
	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	if err := environment.SetCurrentEnvironment(targetName); err != nil {
		return fmt.Errorf("failed to update current environment: %w", err)
	}

	targetEnv.LastUsed = time.Now()
	if err := targetEnv.Save(); err != nil {
		fmt.Printf("⚠️  Warning: Failed to update environment metadata: %v\n", err)
	}

	entry.Success = true
	entry.DurationMs = time.Since(startTime).Milliseconds()
	recordHistory(entry)

	fmt.Printf("✅ Successfully switched to '%s' (%.2fs)\n", targetName, time.Since(startTime).Seconds())
	logger.Info("Successfully switched to '%s' in %.2fs", targetName, time.Since(startTime).Seconds())

	if backupPath != "" {
		fmt.Printf("   Backup: %s\n", filepath.Base(backupPath))
	}

	// Cleanup old backups based on retention policy
	if cfg.BackupRetention > 0 {
		deleted, err := archive.CleanupOldArchives(cfg.BackupRetention)
		if err != nil {
			logger.Warn("Failed to cleanup old archives: %v", err)
		} else if deleted > 0 {
			logger.Debug("Cleaned up %d old archive(s)", deleted)
		}
	}

	// Verify after switch if configured or flag is set
	if cfg.VerifyAfterSwitch || switchVerify {
		fmt.Println()
		fmt.Println("🔍 Verification:")
		verifyEnvironment(targetEnv)
	}

	return nil
}

// snapshotCurrentEnvironment creates snapshots of all enabled tools in the current environment
func snapshotCurrentEnvironment(env *environment.Environment) error {
	toolRegistry := getToolRegistry()
	snapshotCount := 0

	for toolName, config := range env.Tools {
		if !config.Enabled {
			continue
		}

		tool, exists := toolRegistry[toolName]
		if !exists {
			fmt.Printf("  ⚠️  Unknown tool '%s', skipping\n", toolName)
			continue
		}

		snapshotPath := filepath.Join(env.Path, "snapshots", toolName)
		if err := os.MkdirAll(snapshotPath, 0755); err != nil {
			fmt.Printf("  ⚠️  Failed to create snapshot directory for %s: %v, skipping\n", toolName, err)
			continue
		}

		fmt.Printf("  Snapshotting %s...\n", toolName)
		if err := tool.Snapshot(snapshotPath); err != nil {
			fmt.Printf("  ⚠️  Failed to snapshot %s: %v, skipping\n", toolName, err)
			continue
		}

		// Update snapshot metadata
		config.SnapshotPath = snapshotPath
		env.Tools[toolName] = config
		snapshotCount++
	}

	// Capture and save environment variables if configured
	if len(env.EnvVars) > 0 {
		fmt.Println("  Capturing environment variables...")
		varNames := make([]string, 0, len(env.EnvVars))
		for varName := range env.EnvVars {
			varNames = append(varNames, varName)
		}

		capturedVars, captureErr := environment.CaptureEnvVars(varNames)
		if captureErr != nil {
			fmt.Printf("  ⚠️  Failed to capture environment variables: %v\n", captureErr)
		} else {
			if saveErr := env.SaveEnvVars(capturedVars); saveErr != nil {
				fmt.Printf("  ⚠️  Failed to save environment variables: %v\n", saveErr)
			} else {
				fmt.Printf("  ✓ Captured %d environment variable(s)\n", len(capturedVars))
			}
		}
	}

	if snapshotCount > 0 {
		env.LastSnapshot = time.Now()
	}
	return env.Save()
}

// restoreEnvironment restores all enabled tools from the target environment
func restoreEnvironment(env *environment.Environment) (int, error) {
	toolRegistry := getToolRegistry()
	restoredCount := 0

	for toolName, config := range env.Tools {
		if !config.Enabled {
			continue
		}

		tool, exists := toolRegistry[toolName]
		if !exists {
			fmt.Printf("  ⚠️  Unknown tool '%s', skipping\n", toolName)
			continue
		}

		snapshotPath := filepath.Join(env.Path, "snapshots", toolName)

		// Check if snapshot exists and is valid
		if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
			fmt.Printf("  ⚠️  No snapshot found for %s, skipping\n", toolName)
			continue
		}

		// Validate snapshot before restoring
		if err := tool.ValidateSnapshot(snapshotPath); err != nil {
			fmt.Printf("  ⚠️  Invalid snapshot for %s: %v, skipping\n", toolName, err)
			continue
		}

		fmt.Printf("  Restoring %s...\n", toolName)
		if err := tool.Restore(snapshotPath); err != nil {
			fmt.Printf("  ⚠️  Failed to restore %s: %v, skipping\n", toolName, err)
			continue
		}
		restoredCount++
	}

	// Restore environment variables if available
	envVars, loadErr := env.LoadEnvVars()
	if loadErr != nil {
		fmt.Printf("  ⚠️  Failed to load environment variables: %v\n", loadErr)
	} else if len(envVars) > 0 {
		fmt.Println("  Restoring environment variables...")
		if restoreErr := environment.RestoreEnvVars(envVars); restoreErr != nil {
			fmt.Printf("  ⚠️  Failed to restore environment variables: %v\n", restoreErr)
		} else {
			fmt.Printf("  ✓ Restored %d environment variable(s)\n", len(envVars))
		}
	}

	return restoredCount, nil
}

// verifyEnvironment performs verification checks on the environment
func verifyEnvironment(env *environment.Environment) {
	toolRegistry := getToolRegistry()

	for toolName, config := range env.Tools {
		if !config.Enabled {
			continue
		}

		tool, exists := toolRegistry[toolName]
		if !exists {
			continue
		}

		// Check if tool is installed
		if tool.IsInstalled() {
			fmt.Printf("   ✓ %s is installed\n", toolName)
		} else {
			fmt.Printf("   ✗ %s is NOT installed\n", toolName)
		}
	}
}

// recordHistory saves a switch entry to the history
func recordHistory(entry *history.SwitchEntry) {
	hist, err := history.LoadHistory()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to load history: %v\n", err)
		return
	}

	if err := hist.AddEntry(entry); err != nil {
		fmt.Printf("⚠️  Warning: Failed to save history: %v\n", err)
	}
}

// getToolRegistry returns a map of all available tools, filtered by config
func getToolRegistry() map[string]tools.Tool {
	allTools := map[string]tools.Tool{
		"git":     tools.NewGitTool(),
		"aws":     tools.NewAWSTool(),
		"gcloud":  tools.NewGCloudTool(),
		"kubectl": tools.NewKubectlTool(),
		"docker":  tools.NewDockerTool(),
	}

	// Load config to check for excluded tools
	cfg, err := config.LoadConfig()
	if err != nil || cfg == nil || len(cfg.ExcludeTools) == 0 {
		return allTools
	}

	// Filter out excluded tools
	filteredTools := make(map[string]tools.Tool)
	for name, tool := range allTools {
		excluded := false
		for _, excludedTool := range cfg.ExcludeTools {
			if name == excludedTool {
				excluded = true
				logger.Debug("Excluding tool '%s' as per configuration", name)
				break
			}
		}
		if !excluded {
			filteredTools[name] = tool
		}
	}

	return filteredTools
}
