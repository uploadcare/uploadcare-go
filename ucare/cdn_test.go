package ucare

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCDNCNAMEPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		publicKey string
		want      string
	}{
		{"demo_public_key", "demopublickey", "1s4oyld5dc"},
		{"another_key", "anotherkey", "4073mye3t0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, cdnCNAMEPrefix(tt.publicKey))
		})
	}
}

func TestCDNBaseURL(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		"https://1s4oyld5dc.ucarecd.net",
		cdnBaseURL("demopublickey"),
	)
}

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

func TestResolveConfig(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{}
	tests := []struct {
		name           string
		conf           *Config
		wantAPIVersion string
		wantHTTPClient *http.Client
		wantCDNBase    string
	}{
		{
			name:           "defaults",
			wantAPIVersion: defaultAPIVersion,
			wantHTTPClient: http.DefaultClient,
			wantCDNBase:    "https://1s4oyld5dc.ucarecd.net",
		},
		{
			name: "preserves_explicit_values",
			conf: &Config{
				APIVersion: "v0.6",
				HTTPClient: httpClient,
				CDNBase:    "https://cdn.example.com",
			},
			wantAPIVersion: "v0.6",
			wantHTTPClient: httpClient,
			wantCDNBase:    "https://cdn.example.com",
		},
		{
			name:           "normalizes_explicit_cdn_base",
			conf:           &Config{CDNBase: " https://cdn.example.com/ "},
			wantAPIVersion: defaultAPIVersion,
			wantHTTPClient: http.DefaultClient,
			wantCDNBase:    "https://cdn.example.com",
		},
		{
			name:           "defaults_blank_cdn_base",
			conf:           &Config{CDNBase: " \t\n "},
			wantAPIVersion: defaultAPIVersion,
			wantHTTPClient: http.DefaultClient,
			wantCDNBase:    "https://1s4oyld5dc.ucarecd.net",
		},
		{
			name:           "keeps_path_prefix",
			conf:           &Config{CDNBase: "https://cdn.example.com/media/"},
			wantAPIVersion: defaultAPIVersion,
			wantHTTPClient: http.DefaultClient,
			wantCDNBase:    "https://cdn.example.com/media",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conf, err := resolveConfig(tt.conf, "demopublickey")
			require.NoError(t, err)
			assert.Equal(t, tt.wantAPIVersion, conf.APIVersion)
			assert.Same(t, tt.wantHTTPClient, conf.HTTPClient)
			assert.Equal(t, tt.wantCDNBase, conf.CDNBase)
		})
	}

	t.Run("does_not_mutate_input", func(t *testing.T) {
		t.Parallel()
		original := &Config{CDNBase: " https://cdn.example.com/ "}
		conf, err := resolveConfig(original, "demopublickey")
		require.NoError(t, err)
		assert.NotSame(t, original, conf)
		assert.Empty(t, original.APIVersion)
		assert.Nil(t, original.HTTPClient)
		assert.Equal(t, " https://cdn.example.com/ ", original.CDNBase)
		assert.Equal(t, "https://cdn.example.com", conf.CDNBase)
	})
}

func TestResolveConfig_InvalidCDNBase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cdnBase string
	}{
		{"missing_scheme", "cdn.example.com"},
		{"unsupported_scheme", "ftp://cdn.example.com"},
		{"missing_host", "https:///media"},
		{"with_query", "https://cdn.example.com?x=1"},
		{"with_fragment", "https://cdn.example.com#files"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := resolveConfig(&Config{CDNBase: tt.cdnBase}, "demopublickey")
			assert.ErrorIs(t, err, errInvalidCDNBase)
		})
	}
}

func TestNewClient_InvalidCDNBase(t *testing.T) {
	t.Parallel()

	_, err := NewClient(testCreds(), &Config{CDNBase: "cdn.example.com"})
	assert.ErrorIs(t, err, errInvalidCDNBase)
}
