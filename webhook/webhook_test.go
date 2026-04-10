package webhook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventConstants(t *testing.T) {
	t.Parallel()

	t.Run("uploaded", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "file.uploaded", EventFileUploaded)
	})
	t.Run("stored", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "file.stored", EventFileStored)
	})
	t.Run("deleted", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "file.deleted", EventFileDeleted)
	})
	t.Run("info_updated", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "file.info_updated", EventFileInfoUpdated)
	})
	t.Run("infected", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "file.infected", EventFileInfected)
	})
}
