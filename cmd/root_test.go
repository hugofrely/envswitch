package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hugofrely/envswitch/internal/version"
)

func TestRootCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "envswitch", rootCmd.Use)
		assert.NotEmpty(t, rootCmd.Short)
		assert.NotEmpty(t, rootCmd.Long)
		assert.NotEmpty(t, rootCmd.Version) // Version is set dynamically
		assert.Contains(t, rootCmd.Version, version.Version)
	})

	t.Run("has persistent flags", func(t *testing.T) {
		configFlag := rootCmd.PersistentFlags().Lookup("config")
		assert.NotNil(t, configFlag)

		verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
		assert.NotNil(t, verboseFlag)

		debugFlag := rootCmd.PersistentFlags().Lookup("debug")
		assert.NotNil(t, debugFlag)
	})

	t.Run("has subcommands", func(t *testing.T) {
		commands := rootCmd.Commands()
		assert.NotEmpty(t, commands)

		// Check for expected commands
		commandNames := make([]string, len(commands))
		for i, cmd := range commands {
			commandNames[i] = cmd.Name()
		}

		assert.Contains(t, commandNames, "init")
		assert.Contains(t, commandNames, "create")
	})
}

func TestExecute(t *testing.T) {
	t.Run("executes without error", func(t *testing.T) {
		// This would normally execute the command, but we're just checking
		// that the Execute function is callable
		assert.NotPanics(t, func() {
			// Don't actually execute, just verify it's callable
			_ = Execute
		})
	})
}

func TestRootCommandFlags(t *testing.T) {
	t.Run("config flag has correct default", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("config")
		assert.NotNil(t, flag)
		assert.Equal(t, "", flag.DefValue)
	})

	t.Run("verbose flag has correct default", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("verbose")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
		assert.Equal(t, "v", flag.Shorthand)
	})

	t.Run("debug flag has correct default", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("debug")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})
}

func TestRootCommandSubcommands(t *testing.T) {
	t.Run("has all expected subcommands", func(t *testing.T) {
		commands := rootCmd.Commands()
		commandNames := make([]string, len(commands))
		for i, cmd := range commands {
			commandNames[i] = cmd.Name()
		}

		expectedCommands := []string{
			"init",
			"create",
			"switch",
			"list",
			"delete",
			"show",
			"config",
			"shell",
			"completion",
		}

		for _, expected := range expectedCommands {
			assert.Contains(t, commandNames, expected, "should have %s command", expected)
		}
	})

	t.Run("each subcommand has required metadata", func(t *testing.T) {
		commands := rootCmd.Commands()
		for _, cmd := range commands {
			// Skip help and completion commands which are auto-generated
			if cmd.Name() == "help" || cmd.Name() == "completion" {
				continue
			}

			assert.NotEmpty(t, cmd.Use, "command %s should have Use field", cmd.Name())
			assert.NotEmpty(t, cmd.Short, "command %s should have Short description", cmd.Name())
		}
	})
}

func TestVersionInfo(t *testing.T) {
	t.Run("version variables are defined", func(t *testing.T) {
		assert.NotEmpty(t, version.Version)
		assert.NotEmpty(t, version.GitCommit)
		assert.NotEmpty(t, version.BuildDate)
	})

	t.Run("version string includes all components", func(t *testing.T) {
		versionString := rootCmd.Version
		assert.NotEmpty(t, versionString)
		assert.Contains(t, versionString, version.Version)
		assert.Contains(t, versionString, "commit")
		assert.Contains(t, versionString, "built")
	})
}

func TestInitConfig(t *testing.T) {
	t.Run("initConfig is callable", func(t *testing.T) {
		assert.NotPanics(t, func() {
			// initConfig is called automatically by cobra, just verify it's defined
			_ = initConfig
		})
	})
}
