package upload

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/group"
	"github.com/uploadcare/uploadcare-go/v2/internal/uctest"
)

const (
	rewriteGroupID = "abc12345-6789-0000-1111-222233334444~2"
	rewriteCDN     = "https://abc1234567.ucarecd.net"
	legacyLink     = "https://ucarecdn.com/" + rewriteGroupID + "/"
)

func TestGroupInfoCDNBase(t *testing.T) {
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
				assert.Equal(t, "/group/info/", r.URL.Path)
				uctest.RespondJSON(t, w, GroupInfo{
					Info:    group.Info{ID: rewriteGroupID},
					CDNLink: legacyLink,
				})
			}), func(t *testing.T, srv *httptest.Server) {
				c := uctest.NewUploadServerClient(srv)
				c.CDN = tt.cdn
				svc := NewService(c)
				info, err := svc.GroupInfo(context.Background(), rewriteGroupID)
				require.NoError(t, err)
				assert.Equal(t, tt.want, info.CDNLink)
			})
		})
	}
}

func TestCreateGroup_RewritesCDNLink(t *testing.T) {
	t.Parallel()

	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/group/", r.URL.Path)
		uctest.RespondJSON(t, w, GroupInfo{
			Info:    group.Info{ID: rewriteGroupID},
			CDNLink: legacyLink,
		})
	}), func(t *testing.T, srv *httptest.Server) {
		c := uctest.NewUploadServerClient(srv)
		c.CDN = rewriteCDN
		svc := NewService(c)
		info, err := svc.CreateGroup(context.Background(), []string{"file-uuid"})
		require.NoError(t, err)
		assert.Equal(t, rewriteCDN+"/"+rewriteGroupID+"/", info.CDNLink)
	})
}
