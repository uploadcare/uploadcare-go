package conversion

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{
			name: "negative_page_ignored",
			opts: DocumentPathOptions{UUID: "abc-123", Page: -1},
			want: "abc-123/document/-/format/pdf/",
		},
		{
			name: "zero_page_ignored",
			opts: DocumentPathOptions{UUID: "abc-123", Format: "png", Page: 0},
			want: "abc-123/document/-/format/png/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := BuildDocumentPath(tt.opts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildDocumentPath_EmptyUUID(t *testing.T) {
	t.Parallel()
	_, err := BuildDocumentPath(DocumentPathOptions{})
	assert.ErrorIs(t, err, errEmptyUUID)
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
				ResizeMode: ResizeModePreserveRatio,
				Quality:    QualityBest,
				CutStart:   "000:00:05.000",
				CutLength:  "000:00:15.000",
				Thumbs:     10,
			},
			want: "abc-123/video/-/format/webm/-/size/640x480/preserve_ratio/-/quality/best/-/cut/000:00:05.000/000:00:15.000/-/thumbs~10/",
		},
		{
			name: "size_without_resize_mode",
			opts: VideoPathOptions{UUID: "abc-123", Size: "1920x1080"},
			want: "abc-123/video/-/format/mp4/-/size/1920x1080/",
		},
		{
			name: "negative_thumbs_ignored",
			opts: VideoPathOptions{UUID: "abc-123", Thumbs: -5},
			want: "abc-123/video/-/format/mp4/",
		},
		{
			name: "only_quality",
			opts: VideoPathOptions{UUID: "abc-123", Quality: QualityLighter},
			want: "abc-123/video/-/format/mp4/-/quality/lighter/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := BuildVideoPath(tt.opts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildVideoPath_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    VideoPathOptions
		wantErr error
	}{
		{"empty_uuid", VideoPathOptions{}, errEmptyUUID},
		{"resize_mode_without_size", VideoPathOptions{UUID: "abc-123", ResizeMode: "preserve_ratio"}, errResizeModeNoSize},
		{"cut_start_only", VideoPathOptions{UUID: "abc-123", CutStart: "000:00:05.000"}, errIncompleteCut},
		{"cut_length_only", VideoPathOptions{UUID: "abc-123", CutLength: "000:00:15.000"}, errIncompleteCut},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := BuildVideoPath(tt.opts)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
