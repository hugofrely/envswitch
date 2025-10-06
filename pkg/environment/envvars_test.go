package environment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCaptureEnvVars(t *testing.T) {
	t.Run("captures existing environment variables", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("TEST_VAR_1", "value1")
		os.Setenv("TEST_VAR_2", "value2")
		defer os.Unsetenv("TEST_VAR_1")
		defer os.Unsetenv("TEST_VAR_2")

		envVars, err := CaptureEnvVars([]string{"TEST_VAR_1", "TEST_VAR_2"})

		require.NoError(t, err)
		assert.Len(t, envVars, 2)
		assert.Equal(t, "TEST_VAR_1", envVars[0].Key)
		assert.Equal(t, "value1", envVars[0].Value)
		assert.Equal(t, "TEST_VAR_2", envVars[1].Key)
		assert.Equal(t, "value2", envVars[1].Value)
	})

	t.Run("skips unset environment variables", func(t *testing.T) {
		os.Setenv("TEST_VAR_EXISTS", "exists")
		defer os.Unsetenv("TEST_VAR_EXISTS")

		envVars, err := CaptureEnvVars([]string{"TEST_VAR_EXISTS", "TEST_VAR_DOES_NOT_EXIST"})

		require.NoError(t, err)
		assert.Len(t, envVars, 1)
		assert.Equal(t, "TEST_VAR_EXISTS", envVars[0].Key)
	})

	t.Run("handles empty variable list", func(t *testing.T) {
		envVars, err := CaptureEnvVars([]string{})

		require.NoError(t, err)
		assert.Len(t, envVars, 0)
	})
}

func TestSaveEnvVars(t *testing.T) {
	t.Run("saves environment variables to file", func(t *testing.T) {
		tempDir := t.TempDir()
		env := &Environment{
			Name: "test-env",
			Path: tempDir,
		}

		envVars := []EnvVar{
			{Key: "VAR1", Value: "value1"},
			{Key: "VAR2", Value: "value2"},
		}

		err := env.SaveEnvVars(envVars)
		require.NoError(t, err)

		// Verify file was created
		envFilePath := filepath.Join(tempDir, "snapshots", envVarsFileName)
		assert.FileExists(t, envFilePath)

		// Verify file contents
		content, err := os.ReadFile(envFilePath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "VAR1=value1")
		assert.Contains(t, string(content), "VAR2=value2")
	})

	t.Run("escapes special characters in values", func(t *testing.T) {
		tempDir := t.TempDir()
		env := &Environment{
			Name: "test-env",
			Path: tempDir,
		}

		envVars := []EnvVar{
			{Key: "PATH", Value: "/usr/bin:/usr/local/bin"},
			{Key: "MESSAGE", Value: "hello world"},
			{Key: "MULTILINE", Value: "line1\nline2"},
			{Key: "QUOTED", Value: "value with \"quotes\""},
		}

		err := env.SaveEnvVars(envVars)
		require.NoError(t, err)

		// Load and verify
		loaded, err := env.LoadEnvVars()
		require.NoError(t, err)
		assert.Len(t, loaded, 4)

		// Verify values are preserved correctly
		for i, original := range envVars {
			assert.Equal(t, original.Key, loaded[i].Key)
			assert.Equal(t, original.Value, loaded[i].Value)
		}
	})

	t.Run("handles empty env vars list", func(t *testing.T) {
		tempDir := t.TempDir()
		env := &Environment{
			Name: "test-env",
			Path: tempDir,
		}

		err := env.SaveEnvVars([]EnvVar{})
		require.NoError(t, err)

		// File should not be created for empty list
		envFilePath := filepath.Join(tempDir, "snapshots", envVarsFileName)
		_, err = os.Stat(envFilePath)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestLoadEnvVars(t *testing.T) {
	t.Run("loads environment variables from file", func(t *testing.T) {
		tempDir := t.TempDir()
		env := &Environment{
			Name: "test-env",
			Path: tempDir,
		}

		// Create env vars file
		snapshotsDir := filepath.Join(tempDir, "snapshots")
		os.MkdirAll(snapshotsDir, 0755)
		envFilePath := filepath.Join(snapshotsDir, envVarsFileName)

		content := `VAR1=value1
VAR2=value2
PATH=/usr/bin:/usr/local/bin
`
		err := os.WriteFile(envFilePath, []byte(content), 0644)
		require.NoError(t, err)

		envVars, err := env.LoadEnvVars()
		require.NoError(t, err)
		assert.Len(t, envVars, 3)
		assert.Equal(t, "VAR1", envVars[0].Key)
		assert.Equal(t, "value1", envVars[0].Value)
		assert.Equal(t, "VAR2", envVars[1].Key)
		assert.Equal(t, "value2", envVars[1].Value)
		assert.Equal(t, "PATH", envVars[2].Key)
		assert.Equal(t, "/usr/bin:/usr/local/bin", envVars[2].Value)
	})

	t.Run("skips comments and empty lines", func(t *testing.T) {
		tempDir := t.TempDir()
		env := &Environment{
			Name: "test-env",
			Path: tempDir,
		}

		snapshotsDir := filepath.Join(tempDir, "snapshots")
		os.MkdirAll(snapshotsDir, 0755)
		envFilePath := filepath.Join(snapshotsDir, envVarsFileName)

		content := `# Comment line
VAR1=value1

# Another comment
VAR2=value2

`
		err := os.WriteFile(envFilePath, []byte(content), 0644)
		require.NoError(t, err)

		envVars, err := env.LoadEnvVars()
		require.NoError(t, err)
		assert.Len(t, envVars, 2)
	})

	t.Run("returns empty slice when file does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		env := &Environment{
			Name: "test-env",
			Path: tempDir,
		}

		envVars, err := env.LoadEnvVars()
		require.NoError(t, err)
		assert.Len(t, envVars, 0)
	})

	t.Run("handles malformed lines gracefully", func(t *testing.T) {
		tempDir := t.TempDir()
		env := &Environment{
			Name: "test-env",
			Path: tempDir,
		}

		snapshotsDir := filepath.Join(tempDir, "snapshots")
		os.MkdirAll(snapshotsDir, 0755)
		envFilePath := filepath.Join(snapshotsDir, envVarsFileName)

		content := `VAR1=value1
MALFORMED_LINE_WITHOUT_EQUALS
VAR2=value2
ANOTHER=MALFORMED=LINE=WITH=MULTIPLE=EQUALS
`
		err := os.WriteFile(envFilePath, []byte(content), 0644)
		require.NoError(t, err)

		envVars, err := env.LoadEnvVars()
		require.NoError(t, err)
		// Should load valid lines and skip malformed ones
		assert.GreaterOrEqual(t, len(envVars), 2)
	})
}

func TestRestoreEnvVars(t *testing.T) {
	t.Run("sets environment variables in current process", func(t *testing.T) {
		envVars := []EnvVar{
			{Key: "TEST_RESTORE_1", Value: "restore_value1"},
			{Key: "TEST_RESTORE_2", Value: "restore_value2"},
		}

		err := RestoreEnvVars(envVars)
		require.NoError(t, err)

		assert.Equal(t, "restore_value1", os.Getenv("TEST_RESTORE_1"))
		assert.Equal(t, "restore_value2", os.Getenv("TEST_RESTORE_2"))

		// Cleanup
		os.Unsetenv("TEST_RESTORE_1")
		os.Unsetenv("TEST_RESTORE_2")
	})

	t.Run("handles empty env vars list", func(t *testing.T) {
		err := RestoreEnvVars([]EnvVar{})
		require.NoError(t, err)
	})
}

func TestGenerateShellExports(t *testing.T) {
	t.Run("generates export commands", func(t *testing.T) {
		envVars := []EnvVar{
			{Key: "VAR1", Value: "value1"},
			{Key: "VAR2", Value: "value2"},
		}

		exports := GenerateShellExports(envVars)

		assert.Contains(t, exports, "export VAR1='value1'")
		assert.Contains(t, exports, "export VAR2='value2'")
	})

	t.Run("properly quotes values with special characters", func(t *testing.T) {
		envVars := []EnvVar{
			{Key: "PATH", Value: "/usr/bin:/usr/local/bin"},
			{Key: "MESSAGE", Value: "hello world"},
			{Key: "APOSTROPHE", Value: "it's working"},
		}

		exports := GenerateShellExports(envVars)

		assert.Contains(t, exports, "export PATH='/usr/bin:/usr/local/bin'")
		assert.Contains(t, exports, "export MESSAGE='hello world'")
		// Apostrophe should be escaped
		assert.Contains(t, exports, "APOSTROPHE=")
	})

	t.Run("handles empty env vars list", func(t *testing.T) {
		exports := GenerateShellExports([]EnvVar{})
		assert.Equal(t, "", exports)
	})
}

func TestEscapeUnescapeEnvValue(t *testing.T) {
	testCases := []struct {
		name     string
		original string
	}{
		{"simple value", "simple"},
		{"value with spaces", "hello world"},
		{"value with newline", "line1\nline2"},
		{"value with quotes", "value with \"quotes\""},
		{"value with backslash", "path\\to\\file"},
		{"complex value", "complex \"value\" with\nnewlines and\\backslashes"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			escaped := escapeEnvValue(tc.original)
			unescaped := unescapeEnvValue(escaped)
			assert.Equal(t, tc.original, unescaped)
		})
	}
}

func TestShellQuote(t *testing.T) {
	t.Run("simple value", func(t *testing.T) {
		quoted := shellQuote("simple")
		assert.Equal(t, "'simple'", quoted)
	})

	t.Run("value with spaces", func(t *testing.T) {
		quoted := shellQuote("hello world")
		assert.Equal(t, "'hello world'", quoted)
	})

	t.Run("value with apostrophe", func(t *testing.T) {
		quoted := shellQuote("it's working")
		assert.Equal(t, "'it'\"'\"'s working'", quoted)
	})

	t.Run("value with special characters", func(t *testing.T) {
		quoted := shellQuote("/usr/bin:/usr/local/bin")
		assert.Equal(t, "'/usr/bin:/usr/local/bin'", quoted)
	})
}

func TestEnvVarsIntegration(t *testing.T) {
	t.Run("full workflow: capture, save, load, restore", func(t *testing.T) {
		tempDir := t.TempDir()
		env := &Environment{
			Name: "test-env",
			Path: tempDir,
		}

		// Set test environment variables
		os.Setenv("INTEGRATION_TEST_1", "integration_value1")
		os.Setenv("INTEGRATION_TEST_2", "integration_value2")
		defer os.Unsetenv("INTEGRATION_TEST_1")
		defer os.Unsetenv("INTEGRATION_TEST_2")

		// Capture
		captured, err := CaptureEnvVars([]string{"INTEGRATION_TEST_1", "INTEGRATION_TEST_2"})
		require.NoError(t, err)
		assert.Len(t, captured, 2)

		// Save
		err = env.SaveEnvVars(captured)
		require.NoError(t, err)

		// Unset variables
		os.Unsetenv("INTEGRATION_TEST_1")
		os.Unsetenv("INTEGRATION_TEST_2")

		// Load
		loaded, err := env.LoadEnvVars()
		require.NoError(t, err)
		assert.Len(t, loaded, 2)

		// Restore
		err = RestoreEnvVars(loaded)
		require.NoError(t, err)

		// Verify
		assert.Equal(t, "integration_value1", os.Getenv("INTEGRATION_TEST_1"))
		assert.Equal(t, "integration_value2", os.Getenv("INTEGRATION_TEST_2"))

		// Cleanup
		os.Unsetenv("INTEGRATION_TEST_1")
		os.Unsetenv("INTEGRATION_TEST_2")
	})
}
