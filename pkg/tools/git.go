package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hugofrely/envswitch/internal/storage"
)

// GitTool implements the Tool interface for Git
type GitTool struct {
	GitConfigPath string // ~/.gitconfig
}

// NewGitTool creates a new Git tool instance
func NewGitTool() *GitTool {
	home, _ := os.UserHomeDir()
	return &GitTool{
		GitConfigPath: filepath.Join(home, ".gitconfig"),
	}
}

func (g *GitTool) Name() string {
	return "git"
}

func (g *GitTool) IsInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func (g *GitTool) Snapshot(snapshotPath string) error {
	if !g.IsInstalled() {
		return fmt.Errorf("git is not installed")
	}

	// Check if .gitconfig exists
	if _, err := os.Stat(g.GitConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("git config file does not exist: %s", g.GitConfigPath)
	}

	// Create snapshot directory
	if err := os.MkdirAll(snapshotPath, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Copy .gitconfig to snapshot
	destPath := filepath.Join(snapshotPath, "gitconfig")
	if err := storage.CopyFile(g.GitConfigPath, destPath); err != nil {
		return fmt.Errorf("failed to copy git config: %w", err)
	}

	// Also copy .gitconfig.local if it exists
	gitConfigLocal := g.GitConfigPath + ".local"
	if _, err := os.Stat(gitConfigLocal); err == nil {
		destPath := filepath.Join(snapshotPath, "gitconfig.local")
		if err := storage.CopyFile(gitConfigLocal, destPath); err != nil {
			// Not critical, just log warning
			fmt.Fprintf(os.Stderr, "Warning: failed to copy .gitconfig.local: %v\n", err)
		}
	}

	return nil
}

func (g *GitTool) Restore(snapshotPath string) error {
	if !g.IsInstalled() {
		return fmt.Errorf("git is not installed")
	}

	// Validate snapshot first
	if err := g.ValidateSnapshot(snapshotPath); err != nil {
		return fmt.Errorf("invalid snapshot: %w", err)
	}

	// Restore .gitconfig
	srcPath := filepath.Join(snapshotPath, "gitconfig")
	if err := storage.CopyFile(srcPath, g.GitConfigPath); err != nil {
		return fmt.Errorf("failed to restore git config: %w", err)
	}

	// Restore .gitconfig.local if it exists in snapshot
	srcLocalPath := filepath.Join(snapshotPath, "gitconfig.local")
	if _, err := os.Stat(srcLocalPath); err == nil {
		destLocalPath := g.GitConfigPath + ".local"
		if err := storage.CopyFile(srcLocalPath, destLocalPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to restore .gitconfig.local: %v\n", err)
		}
	}

	return nil
}

func (g *GitTool) GetMetadata() (map[string]interface{}, error) {
	if !g.IsInstalled() {
		return nil, fmt.Errorf("git is not installed")
	}

	metadata := make(map[string]interface{})

	// Get user name
	if name := g.execCommand("git", "config", "--global", "user.name"); name != "" {
		metadata["user_name"] = name
	}

	// Get user email
	if email := g.execCommand("git", "config", "--global", "user.email"); email != "" {
		metadata["user_email"] = email
	}

	// Get signing key if configured
	if signingKey := g.execCommand("git", "config", "--global", "user.signingkey"); signingKey != "" {
		metadata["signing_key"] = signingKey
	}

	return metadata, nil
}

func (g *GitTool) ValidateSnapshot(snapshotPath string) error {
	// Check if snapshot directory exists
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot directory does not exist")
	}

	// Check for gitconfig file
	configPath := filepath.Join(snapshotPath, "gitconfig")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("missing required file: gitconfig")
	}

	return nil
}

func (g *GitTool) Diff(snapshotPath string) ([]Change, error) {
	// Get current metadata
	currentMeta, err := g.GetMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to get current metadata: %w", err)
	}

	changes := []Change{}

	// TODO: Read metadata from snapshot and compare
	_ = currentMeta

	return changes, nil
}

// execCommand executes a command and returns the output
func (g *GitTool) execCommand(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
