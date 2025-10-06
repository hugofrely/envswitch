package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GCloudTool implements the Tool interface for Google Cloud CLI
type GCloudTool struct {
	ConfigPath string // ~/.config/gcloud
}

// NewGCloudTool creates a new GCloud tool instance
func NewGCloudTool() *GCloudTool {
	home, _ := os.UserHomeDir()
	return &GCloudTool{
		ConfigPath: filepath.Join(home, ".config", "gcloud"),
	}
}

func (g *GCloudTool) Name() string {
	return "gcloud"
}

func (g *GCloudTool) IsInstalled() bool {
	_, err := exec.LookPath("gcloud")
	return err == nil
}

func (g *GCloudTool) Snapshot(snapshotPath string) error {
	if !g.IsInstalled() {
		return fmt.Errorf("gcloud is not installed")
	}

	// Check if config directory exists
	if _, err := os.Stat(g.ConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("gcloud config directory does not exist")
	}

	// Create snapshot directory
	if err := os.MkdirAll(snapshotPath, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Copy the entire gcloud config directory
	// TODO: Implement recursive copy utility
	return fmt.Errorf("snapshot not yet implemented - use copyDir utility")
}

func (g *GCloudTool) Restore(snapshotPath string) error {
	if !g.IsInstalled() {
		return fmt.Errorf("gcloud is not installed")
	}

	// Validate snapshot first
	if err := g.ValidateSnapshot(snapshotPath); err != nil {
		return fmt.Errorf("invalid snapshot: %w", err)
	}

	// Remove existing config directory
	if err := os.RemoveAll(g.ConfigPath); err != nil {
		return fmt.Errorf("failed to remove existing config: %w", err)
	}

	// Restore from snapshot
	// TODO: Implement recursive copy utility
	return fmt.Errorf("restore not yet implemented - use copyDir utility")
}

func (g *GCloudTool) GetMetadata() (map[string]interface{}, error) {
	if !g.IsInstalled() {
		return nil, fmt.Errorf("gcloud is not installed")
	}

	metadata := make(map[string]interface{})

	// Get account
	if account := g.execCommand("gcloud", "config", "get-value", "account"); account != "" {
		metadata["account"] = account
	}

	// Get project
	if project := g.execCommand("gcloud", "config", "get-value", "project"); project != "" {
		metadata["project"] = project
	}

	// Get region
	if region := g.execCommand("gcloud", "config", "get-value", "compute/region"); region != "" {
		metadata["region"] = region
	}

	// Get active configuration
	if config := g.execCommand("gcloud", "config", "configurations", "list", "--filter=is_active:true", "--format=value(name)"); config != "" {
		metadata["config_name"] = config
	}

	return metadata, nil
}

func (g *GCloudTool) ValidateSnapshot(snapshotPath string) error {
	// Check if snapshot directory exists
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot directory does not exist")
	}

	// Check for essential files/directories
	requiredPaths := []string{
		"configurations",
	}

	for _, path := range requiredPaths {
		fullPath := filepath.Join(snapshotPath, path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("missing required path: %s", path)
		}
	}

	return nil
}

func (g *GCloudTool) Diff(snapshotPath string) ([]Change, error) {
	// Get current metadata
	currentMeta, err := g.GetMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to get current metadata: %w", err)
	}

	// TODO: Read metadata from snapshot
	// For now, return empty changes
	var changes []Change

	// This would compare currentMeta with snapshotMeta
	// and generate a list of changes

	return changes, nil
}

// execCommand executes a command and returns the output
func (g *GCloudTool) execCommand(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
