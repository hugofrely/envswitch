package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/internal/config"
)

func TestGenerateInitScript(t *testing.T) {
	cfg := &config.Config{
		EnablePromptIntegration: true,
		PromptFormat:            "({env}) ",
		PromptColor:             "green",
	}

	t.Run("bash script generation", func(t *testing.T) {
		script, err := GenerateInitScript("bash", cfg)
		require.NoError(t, err)
		assert.Contains(t, script, "__envswitch_prompt")
		assert.Contains(t, script, "cat ~/.envswitch/current.lock")
		assert.Contains(t, script, "PS1")
		assert.Contains(t, script, "32") // green color code
	})

	t.Run("zsh script generation", func(t *testing.T) {
		script, err := GenerateInitScript("zsh", cfg)
		require.NoError(t, err)
		assert.Contains(t, script, "__envswitch_prompt")
		assert.Contains(t, script, "cat ~/.envswitch/current.lock")
		assert.Contains(t, script, "PROMPT")
		assert.Contains(t, script, "green")
	})

	t.Run("fish script generation", func(t *testing.T) {
		script, err := GenerateInitScript("fish", cfg)
		require.NoError(t, err)
		assert.Contains(t, script, "__envswitch_prompt")
		assert.Contains(t, script, "cat ~/.envswitch/current.lock")
		assert.Contains(t, script, "fish_prompt")
		assert.Contains(t, script, "green")
	})

	t.Run("unsupported shell returns error", func(t *testing.T) {
		_, err := GenerateInitScript("powershell", cfg)
		assert.Error(t, err)
	})

	t.Run("disabled prompt integration", func(t *testing.T) {
		disabledCfg := &config.Config{
			EnablePromptIntegration: false,
		}
		script, err := GenerateInitScript("bash", disabledCfg)
		require.NoError(t, err)
		assert.Contains(t, script, "disabled")
	})
}

func TestParsePromptFormat(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"", "(%s) "},
		{"({env}) ", "(%s) "},
		{"[{env}] ", "[%s] "},
		{"{env}> ", "%s> "},
		{"env:{env} ", "env:%s "},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := parsePromptFormat(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParsePromptColor(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"black", "30"},
		{"red", "31"},
		{"green", "32"},
		{"yellow", "33"},
		{"blue", "34"},
		{"magenta", "35"},
		{"cyan", "36"},
		{"white", "37"},
		{"default", ""},
		{"unknown", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := parsePromptColor(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseZshColor(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"green", "green"},
		{"blue", "blue"},
		{"default", ""},
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := parseZshColor(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseFishColor(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"green", "green"},
		{"blue", "blue"},
		{"default", "normal"},
		{"", "normal"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := parseFishColor(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetShellConfigFile(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	t.Run("bash config file", func(t *testing.T) {
		configFile, err := getShellConfigFile("bash")
		require.NoError(t, err)
		// Should be either .bashrc or .bash_profile
		assert.True(t,
			strings.HasSuffix(configFile, ".bashrc") ||
				strings.HasSuffix(configFile, ".bash_profile"))
	})

	t.Run("zsh config file", func(t *testing.T) {
		configFile, err := getShellConfigFile("zsh")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".zshrc"), configFile)
	})

	t.Run("fish config file", func(t *testing.T) {
		configFile, err := getShellConfigFile("fish")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".config", "fish", "config.fish"), configFile)
	})

	t.Run("unsupported shell", func(t *testing.T) {
		_, err := getShellConfigFile("unknown")
		assert.Error(t, err)
	})
}

func TestIsAlreadyInstalled(t *testing.T) {
	t.Run("not installed", func(t *testing.T) {
		tempFile := filepath.Join(t.TempDir(), "test_config")
		os.WriteFile(tempFile, []byte("# some config\n"), 0644)

		result := isAlreadyInstalled(tempFile)
		assert.False(t, result)
	})

	t.Run("already installed", func(t *testing.T) {
		tempFile := filepath.Join(t.TempDir(), "test_config")
		content := `# some config
# envswitch shell integration - DO NOT EDIT
# some script
`
		os.WriteFile(tempFile, []byte(content), 0644)

		result := isAlreadyInstalled(tempFile)
		assert.True(t, result)
	})

	t.Run("file does not exist", func(t *testing.T) {
		result := isAlreadyInstalled("/nonexistent/file")
		assert.False(t, result)
	})
}

func TestInstallShellIntegration(t *testing.T) {
	t.Run("generates valid script for installation", func(t *testing.T) {
		cfg := &config.Config{
			EnablePromptIntegration: true,
			PromptFormat:            "({env}) ",
			PromptColor:             "green",
		}

		// Just test script generation, not actual file modification
		// to avoid modifying user's actual shell config
		script, err := GenerateInitScript("bash", cfg)
		require.NoError(t, err)
		assert.Contains(t, script, "__envswitch_prompt")
		assert.Contains(t, script, "PS1")
	})
}

func TestScriptIntegration(t *testing.T) {
	t.Run("generated scripts are valid", func(t *testing.T) {
		cfg := &config.Config{
			EnablePromptIntegration: true,
			PromptFormat:            "({env}) ",
			PromptColor:             "cyan",
		}

		shells := []string{"bash", "zsh", "fish"}

		for _, shell := range shells {
			t.Run(shell, func(t *testing.T) {
				script, err := GenerateInitScript(shell, cfg)
				require.NoError(t, err)
				assert.NotEmpty(t, script)

				// Verify essential components
				assert.Contains(t, script, "__envswitch_prompt")
				assert.Contains(t, script, "current.lock")
				assert.Contains(t, script, "env-vars.env")
			})
		}
	})
}
