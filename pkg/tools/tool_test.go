package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangeType(t *testing.T) {
	t.Run("has correct change type constants", func(t *testing.T) {
		assert.Equal(t, ChangeType("added"), ChangeTypeAdded)
		assert.Equal(t, ChangeType("removed"), ChangeTypeRemoved)
		assert.Equal(t, ChangeType("modified"), ChangeTypeModified)
	})
}

func TestChange(t *testing.T) {
	t.Run("creates change object", func(t *testing.T) {
		change := Change{
			Type:     ChangeTypeAdded,
			Path:     "/path/to/file",
			OldValue: "",
			NewValue: "new content",
		}

		assert.Equal(t, ChangeTypeAdded, change.Type)
		assert.Equal(t, "/path/to/file", change.Path)
		assert.Empty(t, change.OldValue)
		assert.Equal(t, "new content", change.NewValue)
	})

	t.Run("creates modified change", func(t *testing.T) {
		change := Change{
			Type:     ChangeTypeModified,
			Path:     "/path/to/file",
			OldValue: "old content",
			NewValue: "new content",
		}

		assert.Equal(t, ChangeTypeModified, change.Type)
		assert.NotEmpty(t, change.OldValue)
		assert.NotEmpty(t, change.NewValue)
	})

	t.Run("creates removed change", func(t *testing.T) {
		change := Change{
			Type:     ChangeTypeRemoved,
			Path:     "/path/to/file",
			OldValue: "removed content",
			NewValue: "",
		}

		assert.Equal(t, ChangeTypeRemoved, change.Type)
		assert.NotEmpty(t, change.OldValue)
		assert.Empty(t, change.NewValue)
	})
}

// MockTool is a mock implementation of the Tool interface for testing
type MockTool struct {
	name         string
	installed    bool
	snapshotErr  error
	restoreErr   error
	metadata     map[string]interface{}
	metadataErr  error
	validateErr  error
	diffChanges  []Change
	diffErr      error
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) IsInstalled() bool {
	return m.installed
}

func (m *MockTool) Snapshot(snapshotPath string) error {
	return m.snapshotErr
}

func (m *MockTool) Restore(snapshotPath string) error {
	return m.restoreErr
}

func (m *MockTool) GetMetadata() (map[string]interface{}, error) {
	return m.metadata, m.metadataErr
}

func (m *MockTool) ValidateSnapshot(snapshotPath string) error {
	return m.validateErr
}

func (m *MockTool) Diff(snapshotPath string) ([]Change, error) {
	return m.diffChanges, m.diffErr
}

func TestToolInterface(t *testing.T) {
	t.Run("mock tool implements interface", func(t *testing.T) {
		var tool Tool = &MockTool{
			name:      "test-tool",
			installed: true,
			metadata:  map[string]interface{}{"version": "1.0"},
		}

		assert.Equal(t, "test-tool", tool.Name())
		assert.True(t, tool.IsInstalled())

		metadata, err := tool.GetMetadata()
		assert.NoError(t, err)
		assert.Equal(t, "1.0", metadata["version"])
	})

	t.Run("tool snapshot and restore", func(t *testing.T) {
		tool := &MockTool{
			name:      "test-tool",
			installed: true,
		}

		err := tool.Snapshot("/tmp/snapshot")
		assert.NoError(t, err)

		err = tool.Restore("/tmp/snapshot")
		assert.NoError(t, err)
	})

	t.Run("tool validation", func(t *testing.T) {
		tool := &MockTool{
			name:      "test-tool",
			installed: true,
		}

		err := tool.ValidateSnapshot("/tmp/snapshot")
		assert.NoError(t, err)
	})

	t.Run("tool diff", func(t *testing.T) {
		expectedChanges := []Change{
			{
				Type:     ChangeTypeAdded,
				Path:     "/config/new",
				NewValue: "value",
			},
			{
				Type:     ChangeTypeModified,
				Path:     "/config/existing",
				OldValue: "old",
				NewValue: "new",
			},
		}

		tool := &MockTool{
			name:        "test-tool",
			installed:   true,
			diffChanges: expectedChanges,
		}

		changes, err := tool.Diff("/tmp/snapshot")
		assert.NoError(t, err)
		assert.Len(t, changes, 2)
		assert.Equal(t, ChangeTypeAdded, changes[0].Type)
		assert.Equal(t, ChangeTypeModified, changes[1].Type)
	})
}
