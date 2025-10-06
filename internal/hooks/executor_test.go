package hooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/pkg/environment"
)

func TestExecuteHooks(t *testing.T) {
	t.Run("executes single hook successfully", func(t *testing.T) {
		hooks := []environment.Hook{
			{
				Command:     "echo 'test'",
				Description: "Test hook",
			},
		}

		err := ExecuteHooks(hooks, "test-env")
		assert.NoError(t, err)
	})

	t.Run("executes multiple hooks in order", func(t *testing.T) {
		hooks := []environment.Hook{
			{Command: "echo 'first'"},
			{Command: "echo 'second'"},
			{Command: "echo 'third'"},
		}

		err := ExecuteHooks(hooks, "test-env")
		assert.NoError(t, err)
	})

	t.Run("fails on hook error", func(t *testing.T) {
		hooks := []environment.Hook{
			{Command: "exit 1", Description: "Failing hook"},
		}

		err := ExecuteHooks(hooks, "test-env")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "hook failed")
	})

	t.Run("stops on first failure", func(t *testing.T) {
		hooks := []environment.Hook{
			{Command: "echo 'first'"},
			{Command: "exit 1"},
			{Command: "echo 'should not run'"},
		}

		err := ExecuteHooks(hooks, "test-env")
		assert.Error(t, err)
	})

	t.Run("executes script instead of command", func(t *testing.T) {
		hooks := []environment.Hook{
			{
				Script:      "echo 'script test'\nexit 0",
				Description: "Script hook",
			},
		}

		err := ExecuteHooks(hooks, "test-env")
		assert.NoError(t, err)
	})

	t.Run("fails if hook has neither command nor script", func(t *testing.T) {
		hooks := []environment.Hook{
			{Description: "Invalid hook"},
		}

		err := ExecuteHooks(hooks, "test-env")
		assert.Error(t, err)
	})

	t.Run("handles empty hooks list", func(t *testing.T) {
		err := ExecuteHooks([]environment.Hook{}, "test-env")
		assert.NoError(t, err)
	})
}

func TestExecuteHook(t *testing.T) {
	t.Run("sets ENVSWITCH_ENV variable", func(t *testing.T) {
		hook := environment.Hook{
			Command: "test \"$ENVSWITCH_ENV\" = \"my-env\"",
		}

		err := executeHook(hook, "my-env", 1, 1)
		require.NoError(t, err)
	})

	t.Run("uses description when provided", func(t *testing.T) {
		hook := environment.Hook{
			Command:     "echo 'test'",
			Description: "Custom description",
		}

		err := executeHook(hook, "test-env", 1, 1)
		assert.NoError(t, err)
	})

	t.Run("uses command as description when not provided", func(t *testing.T) {
		hook := environment.Hook{
			Command: "echo 'test'",
		}

		err := executeHook(hook, "test-env", 1, 1)
		assert.NoError(t, err)
	})
}
