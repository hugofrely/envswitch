package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hugofrely/envswitch/internal/storage"
)

// DockerTool implements the Tool interface for Docker
type DockerTool struct {
	DockerConfigDir string // ~/.docker
}

// NewDockerTool creates a new Docker tool instance
func NewDockerTool() *DockerTool {
	home, _ := os.UserHomeDir()
	return &DockerTool{
		DockerConfigDir: filepath.Join(home, ".docker"),
	}
}

func (d *DockerTool) Name() string {
	return "docker"
}

func (d *DockerTool) IsInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func (d *DockerTool) Snapshot(snapshotPath string) error {
	// Check if .docker directory exists
	if _, err := os.Stat(d.DockerConfigDir); os.IsNotExist(err) {
		return fmt.Errorf("docker config directory does not exist: %s", d.DockerConfigDir)
	}

	// Create snapshot directory
	if err := os.MkdirAll(snapshotPath, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Copy the entire .docker directory to snapshot
	if err := storage.CopyDir(d.DockerConfigDir, snapshotPath); err != nil {
		return fmt.Errorf("failed to copy docker config: %w", err)
	}

	return nil
}

func (d *DockerTool) Restore(snapshotPath string) error {
	// Validate snapshot first
	if err := d.ValidateSnapshot(snapshotPath); err != nil {
		return fmt.Errorf("invalid snapshot: %w", err)
	}

	// Create parent directory if it doesn't exist
	configParent := filepath.Dir(d.DockerConfigDir)
	if err := os.MkdirAll(configParent, 0755); err != nil {
		return fmt.Errorf("failed to create config parent directory: %w", err)
	}

	// Remove existing config directory if it exists
	if _, err := os.Stat(d.DockerConfigDir); err == nil {
		if err := os.RemoveAll(d.DockerConfigDir); err != nil {
			return fmt.Errorf("failed to remove existing config: %w", err)
		}
	}

	// Restore from snapshot
	if err := storage.CopyDir(snapshotPath, d.DockerConfigDir); err != nil {
		return fmt.Errorf("failed to restore docker config: %w", err)
	}

	return nil
}

func (d *DockerTool) GetMetadata() (map[string]interface{}, error) {
	if !d.IsInstalled() {
		return nil, fmt.Errorf("docker is not installed")
	}

	metadata := make(map[string]interface{})

	// Get Docker version
	if version := d.execCommand("docker", "version", "--format", "{{.Server.Version}}"); version != "" {
		metadata["version"] = version
	}

	// Get current context
	if context := d.execCommand("docker", "context", "show"); context != "" {
		metadata["context"] = context
	}

	return metadata, nil
}

func (d *DockerTool) ValidateSnapshot(snapshotPath string) error {
	// Check if snapshot directory exists
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot directory does not exist")
	}

	// Check for config.json file
	configPath := filepath.Join(snapshotPath, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("missing required file: config.json")
	}

	return nil
}

func (d *DockerTool) Diff(snapshotPath string) ([]Change, error) {
	// Get current metadata
	currentMeta, err := d.GetMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to get current metadata: %w", err)
	}

	// Get snapshot metadata
	snapshotMeta, err := d.getSnapshotMetadata(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot metadata: %w", err)
	}

	changes := []Change{}

	// Compare context
	changes = append(changes, compareMetadataField("context", snapshotMeta, currentMeta)...)

	// Note: We don't compare version as it's about the Docker server version,
	// not about the configuration state

	return changes, nil
}

// getSnapshotMetadata reads metadata from a snapshot by parsing config.json
func (d *DockerTool) getSnapshotMetadata(snapshotPath string) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})

	configPath := filepath.Join(snapshotPath, "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		// Parse JSON to extract current context
		var config map[string]interface{}
		if err := json.Unmarshal(data, &config); err == nil {
			if currentContext, ok := config["currentContext"].(string); ok {
				metadata["context"] = currentContext
			}
		}
	}

	return metadata, nil
}

// execCommand executes a command and returns the output
func (d *DockerTool) execCommand(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
