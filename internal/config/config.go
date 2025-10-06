package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the global configuration for envswitch
type Config struct {
	Version string `yaml:"version"`

	// Behavior settings
	AutoSaveBeforeSwitch string `yaml:"auto_save_before_switch"` // "true" | "false" | "prompt"
	VerifyAfterSwitch    bool   `yaml:"verify_after_switch"`
	BackupBeforeSwitch   bool   `yaml:"backup_before_switch"`
	BackupRetention      int    `yaml:"backup_retention"`

	// Shell integration
	EnablePromptIntegration bool   `yaml:"enable_prompt_integration"`
	PromptFormat            string `yaml:"prompt_format"`
	PromptColor             string `yaml:"prompt_color"`

	// Logging
	LogLevel string `yaml:"log_level"` // debug | info | warn | error
	LogFile  string `yaml:"log_file"`

	// Tools
	ExcludeTools []string `yaml:"exclude_tools"`

	// UI
	ColorOutput    bool `yaml:"color_output"`
	ShowTimestamps bool `yaml:"show_timestamps"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Version:                 "1.0",
		AutoSaveBeforeSwitch:    "true",
		VerifyAfterSwitch:       false,
		BackupBeforeSwitch:      true,
		BackupRetention:         10,
		EnablePromptIntegration: true,
		PromptFormat:            "({name})",
		PromptColor:             "blue",
		LogLevel:                "warn",
		LogFile:                 filepath.Join(home, ".envswitch", "envswitch.log"),
		ExcludeTools:            []string{},
		ColorOutput:             true,
		ShowTimestamps:          true,
	}
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".envswitch", "config.yaml")
}

// LoadConfig loads the configuration from file
func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()

	// If config doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configPath := GetConfigPath()

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Get retrieves a configuration value by key
func (c *Config) Get(key string) (interface{}, error) {
	switch key {
	case "auto_save_before_switch":
		return c.AutoSaveBeforeSwitch, nil
	case "verify_after_switch":
		return c.VerifyAfterSwitch, nil
	case "backup_before_switch":
		return c.BackupBeforeSwitch, nil
	case "backup_retention":
		return c.BackupRetention, nil
	case "enable_prompt_integration":
		return c.EnablePromptIntegration, nil
	case "prompt_format":
		return c.PromptFormat, nil
	case "prompt_color":
		return c.PromptColor, nil
	case "log_level":
		return c.LogLevel, nil
	case "log_file":
		return c.LogFile, nil
	case "color_output":
		return c.ColorOutput, nil
	case "show_timestamps":
		return c.ShowTimestamps, nil
	default:
		return nil, fmt.Errorf("unknown config key: %s", key)
	}
}

// Set updates a configuration value by key
func (c *Config) Set(key string, value interface{}) error {
	switch key {
	case "auto_save_before_switch":
		return c.setAutoSaveBeforeSwitch(value)
	case "verify_after_switch":
		return c.setBoolValue(&c.VerifyAfterSwitch, value, key)
	case "backup_before_switch":
		return c.setBoolValue(&c.BackupBeforeSwitch, value, key)
	case "backup_retention":
		return c.setIntValue(&c.BackupRetention, value, key)
	case "enable_prompt_integration":
		return c.setBoolValue(&c.EnablePromptIntegration, value, key)
	case "prompt_format":
		return c.setStringValue(&c.PromptFormat, value, key)
	case "prompt_color":
		return c.setStringValue(&c.PromptColor, value, key)
	case "log_level":
		return c.setLogLevel(value)
	case "color_output":
		return c.setBoolValue(&c.ColorOutput, value, key)
	case "show_timestamps":
		return c.setBoolValue(&c.ShowTimestamps, value, key)
	default:
		return fmt.Errorf("unknown or read-only config key: %s", key)
	}
}

func (c *Config) setAutoSaveBeforeSwitch(value interface{}) error {
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid type for auto_save_before_switch: expected string")
	}
	if v != "true" && v != "false" && v != "prompt" {
		return fmt.Errorf("invalid value for auto_save_before_switch: must be 'true', 'false', or 'prompt'")
	}
	c.AutoSaveBeforeSwitch = v
	return nil
}

func (c *Config) setLogLevel(value interface{}) error {
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid type for log_level: expected string")
	}
	if v != "debug" && v != "info" && v != "warn" && v != "error" {
		return fmt.Errorf("invalid value for log_level: must be 'debug', 'info', 'warn', or 'error'")
	}
	c.LogLevel = v
	return nil
}

func (c *Config) setStringValue(field *string, value interface{}, key string) error {
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid type for %s: expected string", key)
	}
	*field = v
	return nil
}

func (c *Config) setBoolValue(field *bool, value interface{}, key string) error {
	v, ok := value.(bool)
	if !ok {
		return fmt.Errorf("invalid type for %s: expected bool", key)
	}
	*field = v
	return nil
}

func (c *Config) setIntValue(field *int, value interface{}, key string) error {
	v, ok := value.(int)
	if !ok {
		return fmt.Errorf("invalid type for %s: expected int", key)
	}
	*field = v
	return nil
}
