package environment

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Environment represents a saved development environment
type Environment struct {
	Name             string                 `yaml:"name"`
	Description      string                 `yaml:"description"`
	CreatedAt        time.Time              `yaml:"created_at"`
	UpdatedAt        time.Time              `yaml:"updated_at"`
	LastUsed         time.Time              `yaml:"last_used"`
	LastSnapshot     time.Time              `yaml:"last_snapshot"`
	Tools            map[string]ToolConfig  `yaml:"tools"`
	EnvVars          map[string]string      `yaml:"environment_variables"`
	Hooks            Hooks                  `yaml:"hooks,omitempty"`
	Tags             []string               `yaml:"tags,omitempty"`
	Metadata         MetadataInfo           `yaml:"metadata,omitempty"`
	SnapshotInfo     SnapshotInfo           `yaml:"snapshot_info,omitempty"`
	Path             string                 `yaml:"-"`
}

// ToolConfig represents configuration for a specific tool
type ToolConfig struct {
	Enabled      bool                   `yaml:"enabled"`
	SnapshotPath string                 `yaml:"snapshot_path"`
	Metadata     map[string]interface{} `yaml:"metadata,omitempty"`
}

// Hooks represents pre/post hooks for environment operations
type Hooks struct {
	PreSwitch    []Hook `yaml:"pre_switch,omitempty"`
	PostSwitch   []Hook `yaml:"post_switch,omitempty"`
	PreSnapshot  []Hook `yaml:"pre_snapshot,omitempty"`
	PostSnapshot []Hook `yaml:"post_snapshot,omitempty"`
}

// Hook represents a single hook command or script
type Hook struct {
	Command     string `yaml:"command,omitempty"`
	Script      string `yaml:"script,omitempty"`
	Description string `yaml:"description,omitempty"`
	Verify      bool   `yaml:"verify,omitempty"`
}

// MetadataInfo contains additional metadata about the environment
type MetadataInfo struct {
	Color string `yaml:"color,omitempty"`
	Icon  string `yaml:"icon,omitempty"`
}

// SnapshotInfo contains information about the snapshot
type SnapshotInfo struct {
	SizeBytes int64 `yaml:"size_bytes"`
	FileCount int   `yaml:"file_count"`
	Encrypted bool  `yaml:"encrypted"`
}

// GetEnvswitchDir returns the path to the .envswitch directory
func GetEnvswitchDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".envswitch"), nil
}

// GetEnvironmentsDir returns the path to the environments directory
func GetEnvironmentsDir() (string, error) {
	dir, err := GetEnvswitchDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "environments"), nil
}

// LoadEnvironment loads an environment from disk
func LoadEnvironment(name string) (*Environment, error) {
	envDir, err := GetEnvironmentsDir()
	if err != nil {
		return nil, err
	}

	envPath := filepath.Join(envDir, name)
	metadataPath := filepath.Join(envPath, "metadata.yaml")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var env Environment
	if err := yaml.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	env.Path = envPath
	return &env, nil
}

// Save saves the environment metadata to disk
func (e *Environment) Save() error {
	metadataPath := filepath.Join(e.Path, "metadata.yaml")

	e.UpdatedAt = time.Now()

	data, err := yaml.Marshal(e)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// ListEnvironments returns all available environments
func ListEnvironments() ([]*Environment, error) {
	envDir, err := GetEnvironmentsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(envDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Environment{}, nil
		}
		return nil, fmt.Errorf("failed to read environments directory: %w", err)
	}

	var environments []*Environment
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		env, err := LoadEnvironment(entry.Name())
		if err != nil {
			// Skip invalid environments
			continue
		}

		environments = append(environments, env)
	}

	return environments, nil
}

// GetCurrentEnvironment returns the currently active environment
func GetCurrentEnvironment() (*Environment, error) {
	dir, err := GetEnvswitchDir()
	if err != nil {
		return nil, err
	}

	lockPath := filepath.Join(dir, "current.lock")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read current.lock: %w", err)
	}

	name := string(data)
	return LoadEnvironment(name)
}

// SetCurrentEnvironment sets the currently active environment
func SetCurrentEnvironment(name string) error {
	dir, err := GetEnvswitchDir()
	if err != nil {
		return err
	}

	lockPath := filepath.Join(dir, "current.lock")
	return os.WriteFile(lockPath, []byte(name), 0644)
}
