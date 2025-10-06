package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hugofrely/envswitch/internal/storage"
)

// KubectlTool implements the Tool interface for Kubectl
type KubectlTool struct {
	KubeConfigDir string // ~/.kube
}

// NewKubectlTool creates a new Kubectl tool instance
func NewKubectlTool() *KubectlTool {
	home, _ := os.UserHomeDir()
	return &KubectlTool{
		KubeConfigDir: filepath.Join(home, ".kube"),
	}
}

func (k *KubectlTool) Name() string {
	return "kubectl"
}

func (k *KubectlTool) IsInstalled() bool {
	_, err := exec.LookPath("kubectl")
	return err == nil
}

func (k *KubectlTool) Snapshot(snapshotPath string) error {
	// Check if .kube directory exists
	if _, err := os.Stat(k.KubeConfigDir); os.IsNotExist(err) {
		return fmt.Errorf("kubectl config directory does not exist: %s", k.KubeConfigDir)
	}

	// Create snapshot directory
	if err := os.MkdirAll(snapshotPath, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Copy the entire .kube directory to snapshot
	if err := storage.CopyDir(k.KubeConfigDir, snapshotPath); err != nil {
		return fmt.Errorf("failed to copy kubectl config: %w", err)
	}

	return nil
}

func (k *KubectlTool) Restore(snapshotPath string) error {
	// Validate snapshot first
	if err := k.ValidateSnapshot(snapshotPath); err != nil {
		return fmt.Errorf("invalid snapshot: %w", err)
	}

	// Create parent directory if it doesn't exist
	configParent := filepath.Dir(k.KubeConfigDir)
	if err := os.MkdirAll(configParent, 0755); err != nil {
		return fmt.Errorf("failed to create config parent directory: %w", err)
	}

	// Remove existing config directory if it exists
	if _, err := os.Stat(k.KubeConfigDir); err == nil {
		if err := os.RemoveAll(k.KubeConfigDir); err != nil {
			return fmt.Errorf("failed to remove existing config: %w", err)
		}
	}

	// Restore from snapshot
	if err := storage.CopyDir(snapshotPath, k.KubeConfigDir); err != nil {
		return fmt.Errorf("failed to restore kubectl config: %w", err)
	}

	return nil
}

func (k *KubectlTool) GetMetadata() (map[string]interface{}, error) {
	if !k.IsInstalled() {
		return nil, fmt.Errorf("kubectl is not installed")
	}

	metadata := make(map[string]interface{})

	// Get current context
	if context := k.execCommand("kubectl", "config", "current-context"); context != "" {
		metadata["current_context"] = context
	}

	// Get cluster info
	if cluster := k.execCommand("kubectl", "config", "view", "--minify", "-o", "jsonpath={.clusters[0].cluster.server}"); cluster != "" {
		metadata["cluster"] = cluster
	}

	// Get current namespace
	if namespace := k.execCommand("kubectl", "config", "view", "--minify", "-o", "jsonpath={.contexts[0].context.namespace}"); namespace != "" {
		metadata["namespace"] = namespace
	} else {
		metadata["namespace"] = "default"
	}

	return metadata, nil
}

func (k *KubectlTool) ValidateSnapshot(snapshotPath string) error {
	// Check if snapshot directory exists
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot directory does not exist")
	}

	// Check for config file
	configPath := filepath.Join(snapshotPath, "config")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("missing required file: config")
	}

	return nil
}

func (k *KubectlTool) Diff(snapshotPath string) ([]Change, error) {
	// Get current metadata
	currentMeta, err := k.GetMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to get current metadata: %w", err)
	}

	changes := []Change{}

	// TODO: Read metadata from snapshot and compare
	_ = currentMeta

	return changes, nil
}

// execCommand executes a command and returns the output
func (k *KubectlTool) execCommand(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
