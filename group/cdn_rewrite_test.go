package group

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/uctest"
)

const (
	rewriteGroupID = "abc12345-6789-0000-1111-222233334444~3"
	rewriteCDN     = "https://abc1234567.ucarecd.net"
	legacyLink     = "https://ucarecdn.com/" + rewriteGroupID + "/"
)

func TestInfoCDNBase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cdn  string
		want string
	}{
		{"rewrites_cdn_link", rewriteCDN, rewriteCDN + "/" + rewriteGroupID + "/"},
		{"no_rewrite_when_cdn_base_empty", "", legacyLink},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/groups/"+rewriteGroupID+"/", r.URL.Path)
				uctest.RespondJSON(t, w, Info{ID: rewriteGroupID, FileCount: 3, CDNLink: legacyLink})
			}), func(t *testing.T, srv *httptest.Server) {
				c := uctest.NewServerClient(srv)
				c.CDN = tt.cdn
				svc := NewService(c)
				info, err := svc.Info(context.Background(), rewriteGroupID)
				require.NoError(t, err)
				assert.Equal(t, tt.want, info.CDNLink)
			})
		})
	}
}

func TestList_RewritesCDNLink(t *testing.T) {
	t.Parallel()
	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/groups/", r.URL.Path)
		resp := map[string]any{
			"next":    nil,
			"results": []Info{{ID: rewriteGroupID, FileCount: 3, CDNLink: legacyLink}},
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
		assert.Equal(t, rewriteCDN+"/"+rewriteGroupID+"/", info.CDNLink)
	})
}
