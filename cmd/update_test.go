package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hugofrely/envswitch/internal/version"
)

func TestUpdateCommand(t *testing.T) {
	// Save original version
	oldVersion := version.Version
	defer func() { version.Version = oldVersion }()

	t.Run("dev version", func(t *testing.T) {
		version.Version = "dev"

		// Run command
		err := runUpdate(updateCmd, []string{})

		assert.NoError(t, err)
	})

	t.Run("command can be executed", func(t *testing.T) {
		version.Version = "dev"

		assert.NotPanics(t, func() {
			_ = runUpdate(updateCmd, []string{})
		})
	})
}

func TestUpdateCommandExists(t *testing.T) {
	// Verify update command is registered
	cmd, _, err := rootCmd.Find([]string{"update"})
	assert.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.Equal(t, "update", cmd.Name())
}

func TestUpdateCommandHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)

	// Test help flag
	updateCmd.SetArgs([]string{"--help"})
	err := updateCmd.Execute()

	// Help will cause a specific behavior, so we just check it doesn't panic
	assert.NoError(t, err)
}

func TestUpdateCommandShortDescription(t *testing.T) {
	assert.NotEmpty(t, updateCmd.Short)
	assert.Contains(t, strings.ToLower(updateCmd.Short), "update")
}

func TestUpdateCommandLongDescription(t *testing.T) {
	assert.NotEmpty(t, updateCmd.Long)
	assert.Contains(t, strings.ToLower(updateCmd.Long), "update")
	assert.Contains(t, strings.ToLower(updateCmd.Long), "version")
}
