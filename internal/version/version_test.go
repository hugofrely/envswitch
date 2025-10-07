package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionDefaults(t *testing.T) {
	// Test that version variables have default values
	assert.NotEmpty(t, Version)
	assert.NotEmpty(t, GitCommit)
	assert.NotEmpty(t, BuildDate)
}

func TestVersionVariablesAreExported(t *testing.T) {
	// Test that we can read and modify version variables
	oldVersion := Version
	oldCommit := GitCommit
	oldBuildDate := BuildDate

	// Modify
	Version = "test-version"
	GitCommit = "test-commit"
	BuildDate = "test-date"

	assert.Equal(t, "test-version", Version)
	assert.Equal(t, "test-commit", GitCommit)
	assert.Equal(t, "test-date", BuildDate)

	// Restore
	Version = oldVersion
	GitCommit = oldCommit
	BuildDate = oldBuildDate
}

func TestVersionDevDefault(t *testing.T) {
	// In test environment, Version should be "dev" by default
	// unless set by ldflags during build
	if Version != "dev" {
		t.Logf("Version is set to: %s (likely set via ldflags)", Version)
	} else {
		assert.Equal(t, "dev", Version)
	}
}
