package conversion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildDocumentPath_FormatOnly(t *testing.T) {
	t.Parallel()

	path := BuildDocumentPath(DocumentPathOptions{UUID: "abc-123", Format: "pdf"})
	assert.Equal(t, "abc-123/document/-/format/pdf/", path)
}

func TestBuildDocumentPath_DefaultFormat(t *testing.T) {
	t.Parallel()

	path := BuildDocumentPath(DocumentPathOptions{UUID: "abc-123"})
	assert.Equal(t, "abc-123/document/-/format/pdf/", path)
}

func TestBuildDocumentPath_WithPage(t *testing.T) {
	t.Parallel()

	path := BuildDocumentPath(DocumentPathOptions{UUID: "abc-123", Format: "png", Page: 3})
	assert.Equal(t, "abc-123/document/-/format/png/-/page/3/", path)
}

func TestBuildVideoPath_Basic(t *testing.T) {
	t.Parallel()

	path := BuildVideoPath(VideoPathOptions{UUID: "abc-123", Format: "mp4"})
	assert.Equal(t, "abc-123/video/-/format/mp4/", path)
}

func TestBuildVideoPath_Full(t *testing.T) {
	t.Parallel()

	path := BuildVideoPath(VideoPathOptions{
		UUID:       "abc-123",
		Format:     "webm",
		Size:       "640x480",
		ResizeMode: "preserve_ratio",
		Quality:    "best",
		CutStart:   "000:00:05.000",
		CutLength:  "000:00:15.000",
		Thumbs:     10,
	})
	assert.Equal(t, "abc-123/video/-/format/webm/-/size/640x480/preserve_ratio/-/quality/best/-/cut/000:00:05.000/000:00:15.000/-/thumbs~10/", path)
}
