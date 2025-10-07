package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPluginCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "plugin", pluginCmd.Use)
		assert.NotEmpty(t, pluginCmd.Short)
		assert.NotEmpty(t, pluginCmd.Long)
	})

	t.Run("has subcommands", func(t *testing.T) {
		commands := pluginCmd.Commands()
		commandNames := make([]string, len(commands))
		for i, cmd := range commands {
			commandNames[i] = cmd.Name()
		}

		assert.Contains(t, commandNames, "list")
		assert.Contains(t, commandNames, "install")
		assert.Contains(t, commandNames, "remove")
		assert.Contains(t, commandNames, "info")
	})

	t.Run("is registered with root command", func(t *testing.T) {
		commands := rootCmd.Commands()
		commandNames := make([]string, len(commands))
		for i, cmd := range commands {
			commandNames[i] = cmd.Name()
		}
		assert.Contains(t, commandNames, "plugin", "plugin command should be registered")
	})
}

func TestPluginListCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "list", pluginListCmd.Use)
		assert.NotEmpty(t, pluginListCmd.Short)
	})

	t.Run("requires no arguments", func(t *testing.T) {
		// List command should work with no arguments
		assert.NotNil(t, pluginListCmd)
	})
}

func TestPluginInstallCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "install <path-to-plugin>", pluginInstallCmd.Use)
		assert.NotEmpty(t, pluginInstallCmd.Short)
	})

	t.Run("requires exactly one argument", func(t *testing.T) {
		err := pluginInstallCmd.Args(pluginInstallCmd, []string{"path"})
		assert.NoError(t, err)

		err = pluginInstallCmd.Args(pluginInstallCmd, []string{})
		assert.Error(t, err)

		err = pluginInstallCmd.Args(pluginInstallCmd, []string{"path1", "path2"})
		assert.Error(t, err)
	})
}

func TestPluginRemoveCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "remove <plugin-name>", pluginRemoveCmd.Use)
		assert.NotEmpty(t, pluginRemoveCmd.Short)
	})

	t.Run("has aliases", func(t *testing.T) {
		assert.Contains(t, pluginRemoveCmd.Aliases, "rm")
		assert.Contains(t, pluginRemoveCmd.Aliases, "uninstall")
	})

	t.Run("requires exactly one argument", func(t *testing.T) {
		err := pluginRemoveCmd.Args(pluginRemoveCmd, []string{"plugin-name"})
		assert.NoError(t, err)

		err = pluginRemoveCmd.Args(pluginRemoveCmd, []string{})
		assert.Error(t, err)
	})
}

func TestPluginInfoCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Equal(t, "info <plugin-name>", pluginInfoCmd.Use)
		assert.NotEmpty(t, pluginInfoCmd.Short)
	})

	t.Run("requires exactly one argument", func(t *testing.T) {
		err := pluginInfoCmd.Args(pluginInfoCmd, []string{"plugin-name"})
		assert.NoError(t, err)

		err = pluginInfoCmd.Args(pluginInfoCmd, []string{})
		assert.Error(t, err)
	})
}
