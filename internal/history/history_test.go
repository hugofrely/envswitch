package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadHistory(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Setup HOME environment for test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create .envswitch directory
	envswitchDir := filepath.Join(tmpDir, ".envswitch")
	err := os.MkdirAll(envswitchDir, 0755)
	require.NoError(t, err)

	t.Run("returns empty history when file doesn't exist", func(t *testing.T) {
		history, err := LoadHistory()
		require.NoError(t, err)
		assert.NotNil(t, history)
		assert.Empty(t, history.Entries)
	})

	t.Run("loads existing history", func(t *testing.T) {
		// Create history file
		history := &History{
			Entries: []SwitchEntry{
				{
					Timestamp:  time.Now(),
					From:       "env1",
					To:         "env2",
					Success:    true,
					ToolsCount: 3,
					DurationMs: 1500,
				},
			},
		}
		err := history.Save()
		require.NoError(t, err)

		// Load it back
		loaded, err := LoadHistory()
		require.NoError(t, err)
		assert.Len(t, loaded.Entries, 1)
		assert.Equal(t, "env1", loaded.Entries[0].From)
		assert.Equal(t, "env2", loaded.Entries[0].To)
	})
}

func TestHistorySave(t *testing.T) {
	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	envswitchDir := filepath.Join(tmpDir, ".envswitch")
	err := os.MkdirAll(envswitchDir, 0755)
	require.NoError(t, err)

	history := &History{
		Entries: []SwitchEntry{
			{
				Timestamp:  time.Now(),
				From:       "test1",
				To:         "test2",
				Success:    true,
				ToolsCount: 2,
			},
		},
	}

	err = history.Save()
	require.NoError(t, err)

	// Verify file was created
	historyPath, err := GetHistoryPath()
	require.NoError(t, err)
	_, err = os.Stat(historyPath)
	assert.NoError(t, err)
}

func TestHistoryAddEntry(t *testing.T) {
	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	envswitchDir := filepath.Join(tmpDir, ".envswitch")
	err := os.MkdirAll(envswitchDir, 0755)
	require.NoError(t, err)

	history := &History{Entries: []SwitchEntry{}}

	entry := SwitchEntry{
		Timestamp:  time.Now(),
		From:       "env1",
		To:         "env2",
		Success:    true,
		ToolsCount: 1,
	}

	err = history.AddEntry(entry)
	require.NoError(t, err)

	// Load and verify
	loaded, err := LoadHistory()
	require.NoError(t, err)
	assert.Len(t, loaded.Entries, 1)
}

func TestHistoryGetLast(t *testing.T) {
	history := &History{
		Entries: []SwitchEntry{
			{From: "env1", To: "env2"},
			{From: "env2", To: "env3"},
			{From: "env3", To: "env4"},
			{From: "env4", To: "env5"},
		},
	}

	t.Run("gets last N entries", func(t *testing.T) {
		last2 := history.GetLast(2)
		assert.Len(t, last2, 2)
		assert.Equal(t, "env3", last2[0].From)
		assert.Equal(t, "env4", last2[1].From)
	})

	t.Run("handles N larger than size", func(t *testing.T) {
		last10 := history.GetLast(10)
		assert.Len(t, last10, 4)
	})

	t.Run("handles zero and negative N", func(t *testing.T) {
		assert.Empty(t, history.GetLast(0))
		assert.Empty(t, history.GetLast(-1))
	})
}

func TestHistoryGetLatest(t *testing.T) {
	t.Run("returns latest entry", func(t *testing.T) {
		history := &History{
			Entries: []SwitchEntry{
				{From: "env1", To: "env2"},
				{From: "env2", To: "env3"},
			},
		}

		latest := history.GetLatest()
		require.NotNil(t, latest)
		assert.Equal(t, "env2", latest.From)
		assert.Equal(t, "env3", latest.To)
	})

	t.Run("returns nil for empty history", func(t *testing.T) {
		history := &History{Entries: []SwitchEntry{}}
		assert.Nil(t, history.GetLatest())
	})
}
