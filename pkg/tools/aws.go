package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hugofrely/envswitch/internal/storage"
)

// AWSTool implements the Tool interface for AWS CLI
type AWSTool struct {
	AWSConfigDir string // ~/.aws
}

// NewAWSTool creates a new AWS tool instance
func NewAWSTool() *AWSTool {
	home, _ := os.UserHomeDir()
	return &AWSTool{
		AWSConfigDir: filepath.Join(home, ".aws"),
	}
}

func (a *AWSTool) Name() string {
	return "aws"
}

func (a *AWSTool) IsInstalled() bool {
	_, err := exec.LookPath("aws")
	return err == nil
}

func (a *AWSTool) Snapshot(snapshotPath string) error {
	if !a.IsInstalled() {
		return fmt.Errorf("aws cli is not installed")
	}

	// Check if .aws directory exists
	if _, err := os.Stat(a.AWSConfigDir); os.IsNotExist(err) {
		return fmt.Errorf("aws config directory does not exist: %s", a.AWSConfigDir)
	}

	// Create snapshot directory
	if err := os.MkdirAll(snapshotPath, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Copy the entire .aws directory to snapshot
	if err := storage.CopyDir(a.AWSConfigDir, snapshotPath); err != nil {
		return fmt.Errorf("failed to copy aws config: %w", err)
	}

	return nil
}

func (a *AWSTool) Restore(snapshotPath string) error {
	if !a.IsInstalled() {
		return fmt.Errorf("aws cli is not installed")
	}

	// Validate snapshot first
	if err := a.ValidateSnapshot(snapshotPath); err != nil {
		return fmt.Errorf("invalid snapshot: %w", err)
	}

	// Create parent directory if it doesn't exist
	configParent := filepath.Dir(a.AWSConfigDir)
	if err := os.MkdirAll(configParent, 0755); err != nil {
		return fmt.Errorf("failed to create config parent directory: %w", err)
	}

	// Remove existing config directory if it exists
	if _, err := os.Stat(a.AWSConfigDir); err == nil {
		if err := os.RemoveAll(a.AWSConfigDir); err != nil {
			return fmt.Errorf("failed to remove existing config: %w", err)
		}
	}

	// Restore from snapshot
	if err := storage.CopyDir(snapshotPath, a.AWSConfigDir); err != nil {
		return fmt.Errorf("failed to restore aws config: %w", err)
	}

	return nil
}

func (a *AWSTool) GetMetadata() (map[string]interface{}, error) {
	if !a.IsInstalled() {
		return nil, fmt.Errorf("aws cli is not installed")
	}

	metadata := make(map[string]interface{})

	// Get current profile from environment or default
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "default"
	}
	metadata["profile"] = profile

	// Get region
	if region := a.execCommand("aws", "configure", "get", "region"); region != "" {
		metadata["region"] = region
	}

	// Try to get account ID (requires valid credentials)
	if accountID := a.execCommand("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text"); accountID != "" {
		metadata["account_id"] = accountID
	}

	return metadata, nil
}

func (a *AWSTool) ValidateSnapshot(snapshotPath string) error {
	// Check if snapshot directory exists
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot directory does not exist")
	}

	// Check for essential files (at least one should exist)
	configPath := filepath.Join(snapshotPath, "config")
	credentialsPath := filepath.Join(snapshotPath, "credentials")

	_, configErr := os.Stat(configPath)
	_, credErr := os.Stat(credentialsPath)

	if os.IsNotExist(configErr) && os.IsNotExist(credErr) {
		return fmt.Errorf("missing required files: config and credentials")
	}

	return nil
}

func (a *AWSTool) Diff(snapshotPath string) ([]Change, error) {
	// Get current metadata
	currentMeta, err := a.GetMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to get current metadata: %w", err)
	}

	changes := []Change{}

	// TODO: Read metadata from snapshot and compare
	_ = currentMeta

	return changes, nil
}

// execCommand executes a command and returns the output
func (a *AWSTool) execCommand(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
