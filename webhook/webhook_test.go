package webhook

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestEventConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "file.uploaded", EventFileUploaded)
	assert.Equal(t, "file.stored", EventFileStored)
	assert.Equal(t, "file.deleted", EventFileDeleted)
	assert.Equal(t, "file.info_updated", EventFileInfoUpdated)
	assert.Equal(t, "file.infected", EventFileInfected)
}
