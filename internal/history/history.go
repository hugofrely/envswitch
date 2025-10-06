package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hugofrely/envswitch/pkg/environment"
)

// SwitchEntry represents a single switch operation in history
type SwitchEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	Success    bool      `json:"success"`
	ErrorMsg   string    `json:"error_msg,omitempty"`
	BackupPath string    `json:"backup_path,omitempty"`
	ToolsCount int       `json:"tools_count"`
	DurationMs int64     `json:"duration_ms"`
}

// History manages the switch history
type History struct {
	Entries []SwitchEntry `json:"entries"`
}

// GetHistoryPath returns the path to the history file
func GetHistoryPath() (string, error) {
	dir, err := environment.GetEnvswitchDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.json"), nil
}

// LoadHistory loads the switch history from disk
func LoadHistory() (*History, error) {
	historyPath, err := GetHistoryPath()
	if err != nil {
		return nil, err
	}

	// If history file doesn't exist, return empty history
	if _, statErr := os.Stat(historyPath); os.IsNotExist(statErr) {
		return &History{Entries: []SwitchEntry{}}, nil
	}

	data, err := os.ReadFile(historyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read history: %w", err)
	}

	var history History
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history: %w", err)
	}

	return &history, nil
}

// Save saves the history to disk
func (h *History) Save() error {
	historyPath, err := GetHistoryPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(historyPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write history: %w", err)
	}

	return nil
}

// AddEntry adds a new switch entry to the history
func (h *History) AddEntry(entry SwitchEntry) error {
	h.Entries = append(h.Entries, entry)
	return h.Save()
}

// GetLast returns the last N entries
func (h *History) GetLast(n int) []SwitchEntry {
	if n <= 0 {
		return []SwitchEntry{}
	}

	if n > len(h.Entries) {
		n = len(h.Entries)
	}

	return h.Entries[len(h.Entries)-n:]
}

// GetLatest returns the most recent switch entry, or nil if history is empty
func (h *History) GetLatest() *SwitchEntry {
	if len(h.Entries) == 0 {
		return nil
	}
	return &h.Entries[len(h.Entries)-1]
}
