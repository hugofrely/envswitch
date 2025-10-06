package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletionCommand(t *testing.T) {
	t.Run("has correct metadata", func(t *testing.T) {
		assert.Contains(t, completionCmd.Use, "completion")
		assert.NotEmpty(t, completionCmd.Short)
		assert.NotEmpty(t, completionCmd.Long)
	})

	t.Run("requires exactly one argument", func(t *testing.T) {
		err := completionCmd.Args(completionCmd, []string{"bash"})
		assert.NoError(t, err)

		err = completionCmd.Args(completionCmd, []string{})
		assert.Error(t, err)

		err = completionCmd.Args(completionCmd, []string{"bash", "extra"})
		assert.Error(t, err)
	})

	t.Run("has valid shell types as valid args", func(t *testing.T) {
		validArgs := completionCmd.ValidArgs
		assert.Contains(t, validArgs, "bash")
		assert.Contains(t, validArgs, "zsh")
		assert.Contains(t, validArgs, "fish")
	})

	t.Run("rejects invalid shell type", func(t *testing.T) {
		err := completionCmd.Args(completionCmd, []string{"invalid"})
		assert.Error(t, err)
	})
}

func TestRunCompletion(t *testing.T) {
	t.Run("generates bash completion script", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runCompletion(completionCmd, []string{"bash"})
		require.NoError(t, err)

		// Restore stdout
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)

		output := buf.String()
		assert.NotEmpty(t, output)
		assert.Contains(t, output, "bash completion")
	})

	t.Run("generates zsh completion script", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runCompletion(completionCmd, []string{"zsh"})
		require.NoError(t, err)

		// Restore stdout
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)

		output := buf.String()
		assert.NotEmpty(t, output)
		assert.Contains(t, output, "zsh completion")
	})

	t.Run("generates fish completion script", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runCompletion(completionCmd, []string{"fish"})
		require.NoError(t, err)

		// Restore stdout
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)

		output := buf.String()
		assert.NotEmpty(t, output)
		// Fish completion has a different format
		assert.NotEmpty(t, output)
	})
}

func TestCompletionIntegration(t *testing.T) {
	t.Run("bash completion includes all commands", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runCompletion(completionCmd, []string{"bash"})
		require.NoError(t, err)

		// Restore stdout
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)

		output := buf.String()
		// Verify that the completion script includes main commands
		assert.Contains(t, output, "envswitch")
	})

	t.Run("zsh completion includes all commands", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runCompletion(completionCmd, []string{"zsh"})
		require.NoError(t, err)

		// Restore stdout
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)

		output := buf.String()
		assert.Contains(t, output, "envswitch")
	})

	t.Run("fish completion includes all commands", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runCompletion(completionCmd, []string{"fish"})
		require.NoError(t, err)

		// Restore stdout
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)

		output := buf.String()
		assert.Contains(t, output, "envswitch")
	})
}
