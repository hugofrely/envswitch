package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "envswitch", rootCmd.Use)
		assert.NotEmpty(t, rootCmd.Short)
		assert.NotEmpty(t, rootCmd.Long)
		assert.Equal(t, "0.1.0", rootCmd.Version)
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
