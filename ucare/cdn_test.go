package ucare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRewriteCDNURL(t *testing.T) {
	t.Parallel()

	uuid := "11111111-2222-3333-4444-555555555555"
	cdn := "https://abc1234567.ucarecd.net"

	tests := []struct {
		name     string
		original string
		cdnBase  string
		want     string
	}{
		{
			name:     "preserves_filename",
			original: "https://ucarecdn.com/" + uuid + "/pineapple.jpg",
			cdnBase:  cdn,
			want:     cdn + "/" + uuid + "/pineapple.jpg",
		},
		{
			name:     "trailing_slash_only",
			original: "https://ucarecdn.com/" + uuid + "/",
			cdnBase:  cdn,
			want:     cdn + "/" + uuid + "/",
		},
		{
			name:     "preserves_nested_effects_path",
			original: "https://ucarecdn.com/" + uuid + "/-/resize/800x/image.jpg",
			cdnBase:  cdn,
			want:     cdn + "/" + uuid + "/-/resize/800x/image.jpg",
		},
		{
			name:     "preserves_query",
			original: "https://ucarecdn.com/" + uuid + "/pineapple.jpg?v=2",
			cdnBase:  cdn,
			want:     cdn + "/" + uuid + "/pineapple.jpg?v=2",
		},
		{
			name:     "cdn_base_with_path_prefix",
			original: "https://ucarecdn.com/" + uuid + "/pineapple.jpg",
			cdnBase:  "https://cdn.example.com/media",
			want:     "https://cdn.example.com/media/" + uuid + "/pineapple.jpg",
		},
		{
			name:     "empty_cdn_base_returns_original",
			original: "https://ucarecdn.com/" + uuid + "/file.jpg",
			cdnBase:  "",
			want:     "https://ucarecdn.com/" + uuid + "/file.jpg",
		},
		{
			name:     "empty_original_returns_empty",
			original: "",
			cdnBase:  cdn,
			want:     "",
		},
		{
			name:     "unparseable_cdn_base_returns_original",
			original: "https://ucarecdn.com/" + uuid + "/file.jpg",
			cdnBase:  "://broken",
			want:     "https://ucarecdn.com/" + uuid + "/file.jpg",
		},
		{
			name:     "scheme_relative_cdn_base_returns_original",
			original: "https://ucarecdn.com/" + uuid + "/file.jpg",
			cdnBase:  "//cdn.example.com",
			want:     "https://ucarecdn.com/" + uuid + "/file.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, RewriteCDNURL(tt.original, tt.cdnBase))
		})
	}
}
