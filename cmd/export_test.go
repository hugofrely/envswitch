package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExportCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "export [environment-name...]", exportCmd.Use)
		assert.NotEmpty(t, exportCmd.Short)
		assert.NotEmpty(t, exportCmd.Long)
	})

	t.Run("has output flag", func(t *testing.T) {
		flag := exportCmd.Flags().Lookup("output")
		assert.NotNil(t, flag)
		assert.Equal(t, "o", flag.Shorthand)
	})

	t.Run("has all flag", func(t *testing.T) {
		flag := exportCmd.Flags().Lookup("all")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})

	t.Run("is registered with root command", func(t *testing.T) {
		commands := rootCmd.Commands()
		commandNames := make([]string, len(commands))
		for i, cmd := range commands {
			commandNames[i] = cmd.Name()
		}
		assert.Contains(t, commandNames, "export", "export command should be registered")
	})
}
