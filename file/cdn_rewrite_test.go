package file

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/uctest"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

const (
	rewriteUUID       = "11111111-2222-3333-4444-555555555555"
	rewriteCDN        = "https://abc1234567.ucarecd.net"
	legacyURL         = "https://ucarecdn.com/" + rewriteUUID + "/pineapple.jpg"
	expectedRewritten = "https://abc1234567.ucarecd.net/" + rewriteUUID + "/pineapple.jpg"
)

func legacyInfo(rawURL string) Info {
	url := rawURL
	return Info{
		BasicFileInfo:   BasicFileInfo{ID: rewriteUUID},
		OriginalFileURL: &url,
	}
}

func serveFileInfo(t *testing.T, cdn string, info Info, fn func(*testing.T, Service)) {
	t.Helper()
	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/files/"+rewriteUUID+"/", r.URL.Path)
		uctest.RespondJSON(t, w, info)
	}), func(t *testing.T, srv *httptest.Server) {
		c := uctest.NewServerClient(srv)
		c.CDN = cdn
		fn(t, NewService(c))
	})
}

func TestInfoCDNBase(t *testing.T) {
	t.Parallel()

	bareURL := "https://ucarecdn.com/" + rewriteUUID + "/"
	tests := []struct {
		name    string
		cdn     string
		info    Info
		wantURL *string
	}{
		{"rewrites_original_file_url", rewriteCDN, legacyInfo(legacyURL), ucare.String(expectedRewritten)},
		{"no_rewrite_when_cdn_base_empty", "", legacyInfo(legacyURL), ucare.String(legacyURL)},
		{"no_rewrite_when_original_url_nil", rewriteCDN, Info{BasicFileInfo: BasicFileInfo{ID: rewriteUUID}}, nil},
		{"preserves_trailing_slash_only_path", rewriteCDN, legacyInfo(bareURL), ucare.String(rewriteCDN + "/" + rewriteUUID + "/")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			serveFileInfo(t, tt.cdn, tt.info, func(t *testing.T, svc Service) {
				info, err := svc.Info(context.Background(), rewriteUUID, nil)
				require.NoError(t, err)
				if tt.wantURL == nil {
					assert.Nil(t, info.OriginalFileURL)
					return
				}
				require.NotNil(t, info.OriginalFileURL)
				assert.Equal(t, *tt.wantURL, *info.OriginalFileURL)
			})
		})
	}
}

func TestBatchStore_RewritesResults(t *testing.T) {
	t.Parallel()
	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/files/storage/", r.URL.Path)
		uctest.RespondJSON(t, w, BatchInfo{Results: []Info{legacyInfo(legacyURL), legacyInfo(legacyURL)}})
	}), func(t *testing.T, srv *httptest.Server) {
		c := uctest.NewServerClient(srv)
		c.CDN = rewriteCDN
		svc := NewService(c)
		b, err := svc.BatchStore(context.Background(), []string{rewriteUUID, rewriteUUID})
		require.NoError(t, err)
		require.Len(t, b.Results, 2)
		for _, r := range b.Results {
			require.NotNil(t, r.OriginalFileURL)
			assert.Equal(t, expectedRewritten, *r.OriginalFileURL)
		}
	})
}

func TestList_RewritesResults(t *testing.T) {
	t.Parallel()
	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/files/", r.URL.Path)
		resp := map[string]any{
			"next":    nil,
			"results": []Info{legacyInfo(legacyURL)},
		}
		uctest.RespondJSON(t, w, resp)
	}), func(t *testing.T, srv *httptest.Server) {
		c := uctest.NewServerClient(srv)
		c.CDN = rewriteCDN
		svc := NewService(c)
		list, err := svc.List(context.Background(), ListParams{})
		require.NoError(t, err)
		require.True(t, list.Next())
		info, err := list.ReadResult()
		require.NoError(t, err)
		require.NotNil(t, info.OriginalFileURL)
		assert.Equal(t, expectedRewritten, *info.OriginalFileURL)
	})
}

func TestLocalCopy_RewritesResult(t *testing.T) {
	t.Parallel()
	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/files/local_copy/", r.URL.Path)
		uctest.RespondJSON(t, w, LocalCopyInfo{Result: legacyInfo(legacyURL)})
	}), func(t *testing.T, srv *httptest.Server) {
		c := uctest.NewServerClient(srv)
		c.CDN = rewriteCDN
		svc := NewService(c)
		res, err := svc.LocalCopy(context.Background(), LocalCopyParams{Source: rewriteUUID, Store: ucare.String(StoreFalse)})
		require.NoError(t, err)
		require.NotNil(t, res.Result.OriginalFileURL)
		assert.Equal(t, expectedRewritten, *res.Result.OriginalFileURL)
	})
}
