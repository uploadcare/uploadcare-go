package conversion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildDocumentPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts DocumentPathOptions
		want string
	}{
		{
			name: "format_only",
			opts: DocumentPathOptions{UUID: "abc-123", Format: "pdf"},
			want: "abc-123/document/-/format/pdf/",
		},
		{
			name: "default_format",
			opts: DocumentPathOptions{UUID: "abc-123"},
			want: "abc-123/document/-/format/pdf/",
		},
		{
			name: "default_format_with_page",
			opts: DocumentPathOptions{UUID: "abc-123", Page: 1},
			want: "abc-123/document/-/format/png/-/page/1/",
		},
		{
			name: "with_page",
			opts: DocumentPathOptions{UUID: "abc-123", Format: "png", Page: 3},
			want: "abc-123/document/-/format/png/-/page/3/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, BuildDocumentPath(tt.opts))
		})
	}
}

func TestBuildVideoPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts VideoPathOptions
		want string
	}{
		{
			name: "basic",
			opts: VideoPathOptions{UUID: "abc-123", Format: "mp4"},
			want: "abc-123/video/-/format/mp4/",
		},
		{
			name: "default_format",
			opts: VideoPathOptions{UUID: "abc-123"},
			want: "abc-123/video/-/format/mp4/",
		},
		{
			name: "full",
			opts: VideoPathOptions{
				UUID:       "abc-123",
				Format:     "webm",
				Size:       "640x480",
				ResizeMode: "preserve_ratio",
				Quality:    "best",
				CutStart:   "000:00:05.000",
				CutLength:  "000:00:15.000",
				Thumbs:     10,
			},
			want: "abc-123/video/-/format/webm/-/size/640x480/preserve_ratio/-/quality/best/-/cut/000:00:05.000/000:00:15.000/-/thumbs~10/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, BuildVideoPath(tt.opts))
		})
	}
}
