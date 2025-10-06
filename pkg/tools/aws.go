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

	// Get snapshot metadata by temporarily creating a new AWSTool pointing to snapshot
	snapshotMeta, err := a.getSnapshotMetadata(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot metadata: %w", err)
	}

	changes := []Change{}

	// Compare profile
	if currentMeta["profile"] != snapshotMeta["profile"] {
		changes = append(changes, Change{
			Type:     ChangeTypeModified,
			Path:     "profile",
			OldValue: fmt.Sprintf("%v", snapshotMeta["profile"]),
			NewValue: fmt.Sprintf("%v", currentMeta["profile"]),
		})
	}

	// Compare region
	currentRegion, currentHasRegion := currentMeta["region"]
	snapshotRegion, snapshotHasRegion := snapshotMeta["region"]

	if currentHasRegion && !snapshotHasRegion {
		changes = append(changes, Change{
			Type:     ChangeTypeAdded,
			Path:     "region",
			NewValue: fmt.Sprintf("%v", currentRegion),
		})
	} else if !currentHasRegion && snapshotHasRegion {
		changes = append(changes, Change{
			Type:     ChangeTypeRemoved,
			Path:     "region",
			OldValue: fmt.Sprintf("%v", snapshotRegion),
		})
	} else if currentHasRegion && snapshotHasRegion && currentRegion != snapshotRegion {
		changes = append(changes, Change{
			Type:     ChangeTypeModified,
			Path:     "region",
			OldValue: fmt.Sprintf("%v", snapshotRegion),
			NewValue: fmt.Sprintf("%v", currentRegion),
		})
	}

	// Compare account ID
	currentAccountID, currentHasAccountID := currentMeta["account_id"]
	snapshotAccountID, snapshotHasAccountID := snapshotMeta["account_id"]

	if currentHasAccountID && !snapshotHasAccountID {
		changes = append(changes, Change{
			Type:     ChangeTypeAdded,
			Path:     "account_id",
			NewValue: fmt.Sprintf("%v", currentAccountID),
		})
	} else if !currentHasAccountID && snapshotHasAccountID {
		changes = append(changes, Change{
			Type:     ChangeTypeRemoved,
			Path:     "account_id",
			OldValue: fmt.Sprintf("%v", snapshotAccountID),
		})
	} else if currentHasAccountID && snapshotHasAccountID && currentAccountID != snapshotAccountID {
		changes = append(changes, Change{
			Type:     ChangeTypeModified,
			Path:     "account_id",
			OldValue: fmt.Sprintf("%v", snapshotAccountID),
			NewValue: fmt.Sprintf("%v", currentAccountID),
		})
	}

	return changes, nil
}

// getSnapshotMetadata reads metadata from a snapshot by parsing the config files
func (a *AWSTool) getSnapshotMetadata(snapshotPath string) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})

	// Read profile from environment or default
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "default"
	}
	metadata["profile"] = profile

	// Try to read region from snapshot config file
	configPath := filepath.Join(snapshotPath, "config")
	if data, err := os.ReadFile(configPath); err == nil {
		content := string(data)
		// Simple parsing for region (this is a basic implementation)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "region") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					metadata["region"] = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	// Note: We cannot get account_id from snapshot files alone as it requires API call
	// So we skip account_id for snapshot metadata

	return metadata, nil
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
