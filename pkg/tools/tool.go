package tools

// Tool is the interface that all tool integrations must implement
type Tool interface {
	// Name returns the name of the tool
	Name() string

	// IsInstalled checks if the tool is installed on the system
	IsInstalled() bool

	// Snapshot captures the current state of the tool into the snapshot directory
	Snapshot(snapshotPath string) error

	// Restore restores the tool's state from the snapshot directory
	Restore(snapshotPath string) error

	// GetMetadata returns metadata about the current state of the tool
	GetMetadata() (map[string]interface{}, error)

	// ValidateSnapshot validates that a snapshot is valid and complete
	ValidateSnapshot(snapshotPath string) error

	// Diff compares the current state with a snapshot and returns the differences
	Diff(snapshotPath string) ([]Change, error)
}

// Change represents a difference between two states
type Change struct {
	Type     ChangeType `json:"type"`
	Path     string     `json:"path"`
	OldValue string     `json:"old_value,omitempty"`
	NewValue string     `json:"new_value,omitempty"`
}

// ChangeType represents the type of change
type ChangeType string

const (
	ChangeTypeAdded    ChangeType = "added"
	ChangeTypeRemoved  ChangeType = "removed"
	ChangeTypeModified ChangeType = "modified"
)
