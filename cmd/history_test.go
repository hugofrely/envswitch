package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/internal/history"
	"github.com/hugofrely/envswitch/pkg/environment"
)

func TestHistoryCommand(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Initialize envswitch directory
	envswitchDir := filepath.Join(tmpDir, ".envswitch")
	require.NoError(t, os.MkdirAll(envswitchDir, 0755))

	// Create test history
	hist := &history.History{
		Entries: []history.SwitchEntry{
			{
				Timestamp:  time.Now().Add(-2 * time.Hour),
				From:       "dev",
				To:         "prod",
				Success:    true,
				ToolsCount: 3,
				DurationMs: 1500,
			},
			{
				Timestamp:  time.Now().Add(-1 * time.Hour),
				From:       "prod",
				To:         "staging",
				Success:    true,
				ToolsCount: 4,
				DurationMs: 2000,
			},
			{
				Timestamp:  time.Now().Add(-30 * time.Minute),
				From:       "staging",
				To:         "dev",
				Success:    false,
				ErrorMsg:   "failed to restore kubectl",
				DurationMs: 500,
			},
		},
	}

	require.NoError(t, hist.Save())

	tests := []struct {
		name        string
		args        []string
		expectError bool
		validate    func(t *testing.T)
	}{
		{
			name:        "show default history",
			args:        []string{"history"},
			expectError: false,
		},
		{
			name:        "show history with limit",
			args:        []string{"history", "--limit", "2"},
			expectError: false,
		},
		{
			name:        "show all history",
			args:        []string{"history", "--all"},
			expectError: false,
		},
		{
			name:        "show detailed history",
			args:        []string{"history", "show"},
			expectError: false,
		},
		{
			name:        "clear history",
			args:        []string{"history", "clear"},
			expectError: false,
			validate: func(t *testing.T) {
				h, err := history.LoadHistory()
				require.NoError(t, err)
				assert.Len(t, h.Entries, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset root command
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.validate != nil {
				tt.validate(t)
			}

			// Reset for next test
			rootCmd.SetArgs([]string{})
		})
	}
}

func TestHistoryCommandEmpty(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Initialize envswitch directory (no history file)
	envswitchDir := filepath.Join(tmpDir, ".envswitch")
	require.NoError(t, os.MkdirAll(envswitchDir, 0755))

	rootCmd.SetArgs([]string{"history"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		ms       int64
		expected string
	}{
		{
			name:     "milliseconds",
			ms:       500,
			expected: "500ms",
		},
		{
			name:     "seconds",
			ms:       1500,
			expected: "1.50s",
		},
		{
			name:     "minutes",
			ms:       65000,
			expected: "1m5s",
		},
		{
			name:     "multiple minutes",
			ms:       125000,
			expected: "2m5s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.ms)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "no truncation needed",
			input:    "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "truncate long string",
			input:    "this is a very long string that needs truncation",
			maxLen:   20,
			expected: "this is a very lo...",
		},
		{
			name:     "exact length",
			input:    "exact",
			maxLen:   5,
			expected: "exact",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDisplayHistoryEntry(t *testing.T) {
	tests := []struct {
		name     string
		entry    history.SwitchEntry
		detailed bool
	}{
		{
			name: "successful switch compact",
			entry: history.SwitchEntry{
				Timestamp:  time.Now(),
				From:       "dev",
				To:         "prod",
				Success:    true,
				ToolsCount: 3,
				DurationMs: 1500,
			},
			detailed: false,
		},
		{
			name: "failed switch compact",
			entry: history.SwitchEntry{
				Timestamp:  time.Now(),
				From:       "prod",
				To:         "dev",
				Success:    false,
				ErrorMsg:   "restore failed",
				DurationMs: 500,
			},
			detailed: false,
		},
		{
			name: "successful switch detailed",
			entry: history.SwitchEntry{
				Timestamp:  time.Now(),
				From:       "dev",
				To:         "prod",
				Success:    true,
				ToolsCount: 3,
				DurationMs: 1500,
				BackupPath: "/path/to/backup.tar.gz",
			},
			detailed: true,
		},
		{
			name: "failed switch detailed",
			entry: history.SwitchEntry{
				Timestamp:  time.Now(),
				From:       "prod",
				To:         "dev",
				Success:    false,
				ErrorMsg:   "restore failed: kubectl error",
				DurationMs: 500,
			},
			detailed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test just ensures the function doesn't panic
			displayHistoryEntry(&tt.entry, tt.detailed)
		})
	}
}

func TestGetStatusText(t *testing.T) {
	assert.Equal(t, "Success", getStatusText(true))
	assert.Equal(t, "Failed", getStatusText(false))
}

func TestHistoryWithMetadata(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Initialize directories
	envswitchDir := filepath.Join(tmpDir, ".envswitch")
	require.NoError(t, os.MkdirAll(envswitchDir, 0755))
	envsDir := filepath.Join(envswitchDir, "environments")
	require.NoError(t, os.MkdirAll(envsDir, 0755))

	// Create a test environment
	envPath := filepath.Join(envsDir, "test-env")
	require.NoError(t, os.MkdirAll(envPath, 0755))

	env := &environment.Environment{
		Name:      "test-env",
		Path:      envPath,
		CreatedAt: time.Now(),
		LastUsed:  time.Now().Add(-1 * time.Hour),
		Tools:     make(map[string]environment.ToolConfig),
	}
	require.NoError(t, env.Save())

	// Create history with this environment
	hist := &history.History{
		Entries: []history.SwitchEntry{
			{
				Timestamp:  time.Now(),
				From:       "(none)",
				To:         "test-env",
				Success:    true,
				ToolsCount: 2,
				DurationMs: 1234,
			},
		},
	}
	require.NoError(t, hist.Save())

	// Test that history command works with this setup
	rootCmd.SetArgs([]string{"history"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}
