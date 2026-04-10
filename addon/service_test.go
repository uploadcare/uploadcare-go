package addon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/uctest"
)

func execPath(addon Name) string {
	return "/addons/" + string(addon) + "/execute/"
}

func statusPath(addon Name) string {
	return "/addons/" + string(addon) + "/execute/status/"
}

func TestExecute(t *testing.T) {
	t.Parallel()

	t.Run("remove_bg", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, execPath(AddonRemoveBG), r.URL.Path)

			var params ExecuteParams
			require.NoError(t, json.Unmarshal(uctest.ReadBody(t, r), &params))
			assert.Equal(t, "file-uuid-123", params.Target)

			uctest.RespondJSON(t, w, ExecuteResult{RequestID: "req-456"})
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			result, err := svc.Execute(context.Background(), AddonRemoveBG, ExecuteParams{
				Target: "file-uuid-123",
			})
			require.NoError(t, err)
			assert.Equal(t, "req-456", result.RequestID)
		})
	})

	t.Run("remove_bg_params", func(t *testing.T) {
		t.Parallel()

		crop := true
		typeLevel := "2"

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := uctest.ParseJSONMap(t, uctest.ReadBody(t, r))
			var params RemoveBGParams
			require.NoError(t, json.Unmarshal(raw["params"], &params))
			assert.Equal(t, &crop, params.Crop)
			assert.Equal(t, &typeLevel, params.TypeLevel)

			uctest.RespondJSON(t, w, ExecuteResult{RequestID: "req-789"})
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			result, err := svc.Execute(context.Background(), AddonRemoveBG, ExecuteParams{
				Target: "file-uuid",
				Params: RemoveBGParams{Crop: &crop, TypeLevel: &typeLevel},
			})
			require.NoError(t, err)
			assert.Equal(t, "req-789", result.RequestID)
		})
	})

	t.Run("clamav_params", func(t *testing.T) {
		t.Parallel()

		purge := true

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := uctest.ParseJSONMap(t, uctest.ReadBody(t, r))
			var params ClamAVParams
			require.NoError(t, json.Unmarshal(raw["params"], &params))
			assert.Equal(t, &purge, params.PurgeInfected)

			uctest.RespondJSON(t, w, ExecuteResult{RequestID: "req-clam"})
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			result, err := svc.Execute(context.Background(), AddonClamAV, ExecuteParams{
				Target: "file-uuid",
				Params: ClamAVParams{PurgeInfected: &purge},
			})
			require.NoError(t, err)
			assert.Equal(t, "req-clam", result.RequestID)
		})
	})

	t.Run("no_params_omitted", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := uctest.ParseJSONMap(t, uctest.ReadBody(t, r))
			_, hasParams := raw["params"]
			assert.False(t, hasParams)

			uctest.RespondJSON(t, w, ExecuteResult{RequestID: "req-rek"})
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			result, err := svc.Execute(context.Background(), AddonRekognitionLabels, ExecuteParams{
				Target: "file-uuid",
			})
			require.NoError(t, err)
			assert.Equal(t, "req-rek", result.RequestID)
		})
	})
}

func TestStatus(t *testing.T) {
	t.Parallel()

	t.Run("done", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, statusPath(AddonRemoveBG), r.URL.Path)
			assert.Equal(t, "req-456", r.URL.Query().Get("request_id"))

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"done","result":{"foreground_type":"person"}}`))
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			result, err := svc.Status(context.Background(), AddonRemoveBG, "req-456")
			require.NoError(t, err)
			assert.Equal(t, StatusDone, result.Status)
			assert.Contains(t, string(result.Result), "foreground_type")
		})
	})

	t.Run("in_progress", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"in_progress"}`))
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			result, err := svc.Status(context.Background(), AddonClamAV, "req-pending")
			require.NoError(t, err)
			assert.Equal(t, StatusInProgress, result.Status)
		})
	})

	t.Run("error_details", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"error","details":["scan failed","timeout"]}`))
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			result, err := svc.Status(context.Background(), AddonClamAV, "req-err")
			require.NoError(t, err)
			assert.Equal(t, StatusError, result.Status)
			assert.JSONEq(t, `["scan failed","timeout"]`, string(result.Details))
		})
	})
}
