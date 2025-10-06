package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGcloudTool(t *testing.T) {
	gcloud := &GCloudTool{}

	t.Run("has correct name", func(t *testing.T) {
		assert.Equal(t, "gcloud", gcloud.Name())
	})

	t.Run("checks installation", func(t *testing.T) {
		// This will depend on whether gcloud is actually installed
		// We just verify the method is callable
		_ = gcloud.IsInstalled()
	})
}
