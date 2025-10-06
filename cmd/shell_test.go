package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/internal/config"
)

func TestShellCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "shell", shellCmd.Use)
		assert.NotEmpty(t, shellCmd.Short)
		assert.NotEmpty(t, shellCmd.Long)
	})

	t.Run("has init subcommand", func(t *testing.T) {
		found := false
		for _, cmd := range shellCmd.Commands() {
			if cmd.Name() == "init" {
				found = true
				break
			}
		}
		assert.True(t, found, "shell init subcommand should exist")
	})

	t.Run("has install subcommand", func(t *testing.T) {
		found := false
		for _, cmd := range shellCmd.Commands() {
			if cmd.Name() == "install" {
				found = true
				break
			}
		}
		assert.True(t, found, "shell install subcommand should exist")
	})
}

func TestShellInitCommand(t *testing.T) {
	// Setup temp home directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("has correct metadata", func(t *testing.T) {
		assert.Contains(t, shellInitCmd.Use, "init")
		assert.NotEmpty(t, shellInitCmd.Short)
		assert.NotEmpty(t, shellInitCmd.Long)
	})

	t.Run("requires exactly one argument", func(t *testing.T) {
		err := shellInitCmd.Args(shellInitCmd, []string{"bash"})
		assert.NoError(t, err)

		err = shellInitCmd.Args(shellInitCmd, []string{})
		assert.Error(t, err)

		err = shellInitCmd.Args(shellInitCmd, []string{"bash", "extra"})
		assert.Error(t, err)
	})

	t.Run("has valid shell types as valid args", func(t *testing.T) {
		validArgs := shellInitCmd.ValidArgs
		assert.Contains(t, validArgs, "bash")
		assert.Contains(t, validArgs, "zsh")
		assert.Contains(t, validArgs, "fish")
	})

	t.Run("generates bash init script", func(t *testing.T) {
		// Initialize config
		cfg := config.DefaultConfig()
		err := cfg.Save()
		require.NoError(t, err)

		err = runShellInit(shellInitCmd, []string{"bash"})
		assert.NoError(t, err)
	})

	t.Run("generates zsh init script", func(t *testing.T) {
		// Initialize config
		cfg := config.DefaultConfig()
		err := cfg.Save()
		require.NoError(t, err)

		err = runShellInit(shellInitCmd, []string{"zsh"})
		assert.NoError(t, err)
	})

	t.Run("generates fish init script", func(t *testing.T) {
		// Initialize config
		cfg := config.DefaultConfig()
		err := cfg.Save()
		require.NoError(t, err)

		err = runShellInit(shellInitCmd, []string{"fish"})
		assert.NoError(t, err)
	})

	t.Run("uses default config when config file doesn't exist", func(t *testing.T) {
		// Use a fresh temp directory without config
		freshDir := filepath.Join(tempDir, "fresh")
		os.Setenv("HOME", freshDir)
		defer os.Setenv("HOME", tempDir)

		err := runShellInit(shellInitCmd, []string{"bash"})
		assert.NoError(t, err)
	})
}

func TestShellInstallCommand(t *testing.T) {
	// Setup temp home directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("has correct metadata", func(t *testing.T) {
		assert.Contains(t, shellInstallCmd.Use, "install")
		assert.NotEmpty(t, shellInstallCmd.Short)
		assert.NotEmpty(t, shellInstallCmd.Long)
	})

	t.Run("requires exactly one argument", func(t *testing.T) {
		err := shellInstallCmd.Args(shellInstallCmd, []string{"bash"})
		assert.NoError(t, err)

		err = shellInstallCmd.Args(shellInstallCmd, []string{})
		assert.Error(t, err)

		err = shellInstallCmd.Args(shellInstallCmd, []string{"bash", "extra"})
		assert.Error(t, err)
	})

	t.Run("has valid shell types as valid args", func(t *testing.T) {
		validArgs := shellInstallCmd.ValidArgs
		assert.Contains(t, validArgs, "bash")
		assert.Contains(t, validArgs, "zsh")
		assert.Contains(t, validArgs, "fish")
	})

	t.Run("installs bash integration", func(t *testing.T) {
		// Initialize config
		cfg := config.DefaultConfig()
		err := cfg.Save()
		require.NoError(t, err)

		// Create .bashrc file
		bashrc := filepath.Join(tempDir, ".bashrc")
		err = os.WriteFile(bashrc, []byte("# existing bashrc\n"), 0644)
		require.NoError(t, err)

		err = runShellInstall(shellInstallCmd, []string{"bash"})
		assert.NoError(t, err)

		// Verify file was modified
		content, err := os.ReadFile(bashrc)
		require.NoError(t, err)
		assert.Contains(t, string(content), "envswitch")
	})

	t.Run("installs zsh integration", func(t *testing.T) {
		// Initialize config
		cfg := config.DefaultConfig()
		err := cfg.Save()
		require.NoError(t, err)

		// Create .zshrc file
		zshrc := filepath.Join(tempDir, ".zshrc")
		err = os.WriteFile(zshrc, []byte("# existing zshrc\n"), 0644)
		require.NoError(t, err)

		err = runShellInstall(shellInstallCmd, []string{"zsh"})
		assert.NoError(t, err)

		// Verify file was modified
		content, err := os.ReadFile(zshrc)
		require.NoError(t, err)
		assert.Contains(t, string(content), "envswitch")
	})

	t.Run("installs fish integration", func(t *testing.T) {
		// Initialize config
		cfg := config.DefaultConfig()
		err := cfg.Save()
		require.NoError(t, err)

		// Create fish config directory and file
		fishConfigDir := filepath.Join(tempDir, ".config", "fish")
		err = os.MkdirAll(fishConfigDir, 0755)
		require.NoError(t, err)

		fishConfig := filepath.Join(fishConfigDir, "config.fish")
		err = os.WriteFile(fishConfig, []byte("# existing fish config\n"), 0644)
		require.NoError(t, err)

		err = runShellInstall(shellInstallCmd, []string{"fish"})
		assert.NoError(t, err)

		// Verify file was modified
		content, err := os.ReadFile(fishConfig)
		require.NoError(t, err)
		assert.Contains(t, string(content), "envswitch")
	})

	t.Run("creates bash config file if it doesn't exist", func(t *testing.T) {
		// Use a fresh temp directory
		freshDir := filepath.Join(tempDir, "fresh-bash")
		os.Setenv("HOME", freshDir)
		defer os.Setenv("HOME", tempDir)

		// Initialize config
		cfg := config.DefaultConfig()
		err := cfg.Save()
		require.NoError(t, err)

		err = runShellInstall(shellInstallCmd, []string{"bash"})
		assert.NoError(t, err)

		// Verify file was created (could be .bashrc or .bash_profile)
		bashrc := filepath.Join(freshDir, ".bashrc")
		bashProfile := filepath.Join(freshDir, ".bash_profile")

		_, errBashrc := os.Stat(bashrc)
		_, errBashProfile := os.Stat(bashProfile)

		// At least one should exist
		assert.True(t, errBashrc == nil || errBashProfile == nil, "either .bashrc or .bash_profile should be created")
	})

	t.Run("creates zsh config file if it doesn't exist", func(t *testing.T) {
		// Use a fresh temp directory
		freshDir := filepath.Join(tempDir, "fresh-zsh")
		os.Setenv("HOME", freshDir)
		defer os.Setenv("HOME", tempDir)

		// Initialize config
		cfg := config.DefaultConfig()
		err := cfg.Save()
		require.NoError(t, err)

		err = runShellInstall(shellInstallCmd, []string{"zsh"})
		assert.NoError(t, err)

		// Verify file was created
		zshrc := filepath.Join(freshDir, ".zshrc")
		_, err = os.Stat(zshrc)
		assert.NoError(t, err, "zshrc should be created")
	})

	t.Run("creates fish config file if it doesn't exist", func(t *testing.T) {
		// Use a fresh temp directory
		freshDir := filepath.Join(tempDir, "fresh-fish")
		os.Setenv("HOME", freshDir)
		defer os.Setenv("HOME", tempDir)

		// Initialize config
		cfg := config.DefaultConfig()
		err := cfg.Save()
		require.NoError(t, err)

		err = runShellInstall(shellInstallCmd, []string{"fish"})
		assert.NoError(t, err)

		// Verify file was created
		fishConfig := filepath.Join(freshDir, ".config", "fish", "config.fish")
		_, err = os.Stat(fishConfig)
		assert.NoError(t, err, "fish config should be created")
	})

	t.Run("uses default config when config file doesn't exist", func(t *testing.T) {
		// Use a fresh temp directory without config
		freshDir := filepath.Join(tempDir, "fresh-no-config")
		err := os.MkdirAll(freshDir, 0755)
		require.NoError(t, err)

		os.Setenv("HOME", freshDir)
		defer os.Setenv("HOME", tempDir)

		// Create at least one bash config file for the install to work
		bashProfile := filepath.Join(freshDir, ".bash_profile")
		err = os.WriteFile(bashProfile, []byte("# test\n"), 0644)
		require.NoError(t, err)

		err = runShellInstall(shellInstallCmd, []string{"bash"})
		assert.NoError(t, err)
	})
}
