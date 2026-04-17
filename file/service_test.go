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

const testFileUUID = "test-uuid"

func TestInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		params    *InfoParams
		wantQuery string
	}{
		{"nil_params", nil, ""},
		{"include_appdata", &InfoParams{Include: ucare.String("appdata")}, "appdata"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/files/"+testFileUUID+"/", r.URL.Path)
				assert.Equal(t, tt.wantQuery, r.URL.Query().Get("include"))

				uctest.RespondJSON(t, w, Info{BasicFileInfo: BasicFileInfo{ID: testFileUUID}})
			}), func(t *testing.T, srv *httptest.Server) {
				svc := NewService(uctest.NewServerClient(srv))
				info, err := svc.Info(context.Background(), testFileUUID, tt.params)
				require.NoError(t, err)
				assert.Equal(t, testFileUUID, info.ID)
			})
		})
	}
}

func TestListParams_Include(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://example.test/files/", nil)
	require.NoError(t, err)

	require.NoError(t, (&ListParams{Include: ucare.String("appdata")}).EncodeReq(req))
	assert.Equal(t, "appdata", req.URL.Query().Get("include"))
}
