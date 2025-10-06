package environment

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const envVarsFileName = "env-vars.env"

// EnvVar represents an environment variable
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CaptureEnvVars captures specified environment variables
func CaptureEnvVars(varNames []string) ([]EnvVar, error) {
	var envVars []EnvVar

	for _, name := range varNames {
		value := os.Getenv(name)
		// Only capture if the variable is set
		if value != "" {
			envVars = append(envVars, EnvVar{
				Key:   name,
				Value: value,
			})
		}
	}

	return envVars, nil
}

// SaveEnvVars saves environment variables to a file in the environment's snapshot directory
func (e *Environment) SaveEnvVars(envVars []EnvVar) error {
	if len(envVars) == 0 {
		return nil
	}

	envFilePath := filepath.Join(e.Path, "snapshots", envVarsFileName)

	// Create snapshots directory if it doesn't exist
	snapshotsDir := filepath.Dir(envFilePath)
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	file, err := os.Create(envFilePath)
	if err != nil {
		return fmt.Errorf("failed to create env vars file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, envVar := range envVars {
		// Escape values that contain special characters
		value := escapeEnvValue(envVar.Value)
		line := fmt.Sprintf("%s=%s\n", envVar.Key, value)
		if _, err := writer.WriteString(line); err != nil {
			return fmt.Errorf("failed to write env var: %w", err)
		}
	}

	return nil
}

// LoadEnvVars loads environment variables from the environment's snapshot directory
func (e *Environment) LoadEnvVars() ([]EnvVar, error) {
	envFilePath := filepath.Join(e.Path, "snapshots", envVarsFileName)

	// If file doesn't exist, return empty slice (not an error)
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		return []EnvVar{}, nil
	}

	file, err := os.Open(envFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open env vars file: %w", err)
	}
	defer file.Close()

	var envVars []EnvVar
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := unescapeEnvValue(strings.TrimSpace(parts[1]))

		envVars = append(envVars, EnvVar{
			Key:   key,
			Value: value,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read env vars file: %w", err)
	}

	return envVars, nil
}

// RestoreEnvVars sets environment variables in the current process
// Note: This only affects the current process, not the parent shell
// For shell integration, use the shell init script
func RestoreEnvVars(envVars []EnvVar) error {
	for _, envVar := range envVars {
		if err := os.Setenv(envVar.Key, envVar.Value); err != nil {
			return fmt.Errorf("failed to set env var %s: %w", envVar.Key, err)
		}
	}
	return nil
}

// GenerateShellExports generates shell export commands for the environment variables
func GenerateShellExports(envVars []EnvVar) string {
	var builder strings.Builder

	for _, envVar := range envVars {
		// Use shell-safe quoting
		value := shellQuote(envVar.Value)
		builder.WriteString(fmt.Sprintf("export %s=%s\n", envVar.Key, value))
	}

	return builder.String()
}

// escapeEnvValue escapes special characters in environment variable values
func escapeEnvValue(value string) string {
	// If value contains spaces, newlines, or quotes, wrap in quotes and escape
	if strings.ContainsAny(value, " \t\n\"'") {
		value = strings.ReplaceAll(value, "\\", "\\\\")
		value = strings.ReplaceAll(value, "\"", "\\\"")
		value = strings.ReplaceAll(value, "\n", "\\n")
		return "\"" + value + "\""
	}
	return value
}

// unescapeEnvValue unescapes environment variable values
func unescapeEnvValue(value string) string {
	// Remove surrounding quotes if present
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
		value = strings.ReplaceAll(value, "\\n", "\n")
		value = strings.ReplaceAll(value, "\\\"", "\"")
		value = strings.ReplaceAll(value, "\\\\", "\\")
	}
	return value
}

// shellQuote quotes a value for safe use in shell commands
func shellQuote(value string) string {
	// Use single quotes for shell safety, escape any single quotes in the value
	escaped := strings.ReplaceAll(value, "'", "'\"'\"'")
	return "'" + escaped + "'"
}
