package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "import <archive-path>", importCmd.Use)
		assert.NotEmpty(t, importCmd.Short)
		assert.NotEmpty(t, importCmd.Long)
	})

	t.Run("has name flag", func(t *testing.T) {
		flag := importCmd.Flags().Lookup("name")
		assert.NotNil(t, flag)
		assert.Equal(t, "n", flag.Shorthand)
	})

	t.Run("has force flag", func(t *testing.T) {
		flag := importCmd.Flags().Lookup("force")
		assert.NotNil(t, flag)
		assert.Equal(t, "f", flag.Shorthand)
		assert.Equal(t, "false", flag.DefValue)
	})

	t.Run("has all flag", func(t *testing.T) {
		flag := importCmd.Flags().Lookup("all")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})

	t.Run("requires exactly one argument", func(t *testing.T) {
		err := importCmd.Args(importCmd, []string{"archive.tar.gz"})
		assert.NoError(t, err)

		err = importCmd.Args(importCmd, []string{})
		assert.Error(t, err)

		err = importCmd.Args(importCmd, []string{"archive1.tar.gz", "archive2.tar.gz"})
		assert.Error(t, err)
	})

	t.Run("is registered with root command", func(t *testing.T) {
		commands := rootCmd.Commands()
		commandNames := make([]string, len(commands))
		for i, cmd := range commands {
			commandNames[i] = cmd.Name()
		}
		assert.Contains(t, commandNames, "import", "import command should be registered")
	})
}
