package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/uctest"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

func TestList(t *testing.T) {
	t.Parallel()

	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/webhooks/", r.URL.Path)

		uctest.RespondJSON(t, w, []Info{
			{ID: 1, TargetURL: "https://example.com/hook1", Event: EventFileUploaded, IsActive: true},
			{ID: 2, TargetURL: "https://example.com/hook2", Event: EventFileStored, IsActive: false},
		})
	}), func(t *testing.T, srv *httptest.Server) {
		svc := NewService(uctest.NewServerClient(srv))
		hooks, err := svc.List(context.Background())
		require.NoError(t, err)

		require.Len(t, hooks, 2)
		assert.Equal(t, int64(1), hooks[0].ID)
		assert.Equal(t, "https://example.com/hook1", hooks[0].TargetURL)
		assert.Equal(t, EventFileUploaded, hooks[0].Event)
		assert.True(t, hooks[0].IsActive)
		assert.Equal(t, EventFileStored, hooks[1].Event)
		assert.False(t, hooks[1].IsActive)
	})
}

func TestCreate(t *testing.T) {
	t.Parallel()

	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/webhooks/", r.URL.Path)

		var params Params
		require.NoError(t, json.Unmarshal(uctest.ReadBody(t, r), &params))
		assert.Equal(t, "https://example.com/hook", *params.TargetURL)
		assert.Equal(t, EventFileUploaded, *params.Event)

		uctest.RespondJSON(t, w, Info{
			ID:        42,
			TargetURL: *params.TargetURL,
			Event:     *params.Event,
			IsActive:  true,
		})
	}), func(t *testing.T, srv *httptest.Server) {
		svc := NewService(uctest.NewServerClient(srv))
		info, err := svc.Create(context.Background(), Params{
			TargetURL: ucare.String("https://example.com/hook"),
			Event:     EventPtr(EventFileUploaded),
			IsActive:  ucare.Bool(true),
		})
		require.NoError(t, err)
		assert.Equal(t, int64(42), info.ID)
		assert.Equal(t, "https://example.com/hook", info.TargetURL)
		assert.Equal(t, EventFileUploaded, info.Event)
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			assert.Equal(t, "/webhooks/42/", r.URL.Path)

			var params Params
			require.NoError(t, json.Unmarshal(uctest.ReadBody(t, r), &params))
			assert.Equal(t, "https://example.com/updated", *params.TargetURL)

			uctest.RespondJSON(t, w, Info{
				ID:        42,
				TargetURL: *params.TargetURL,
				Event:     EventFileUploaded,
				IsActive:  true,
			})
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			info, err := svc.Update(context.Background(), Params{
				ID:        ucare.Int64(42),
				TargetURL: ucare.String("https://example.com/updated"),
			})
			require.NoError(t, err)
			assert.Equal(t, "https://example.com/updated", info.TargetURL)
		})
	})

	t.Run("nil_id", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("unexpected request: ID is nil, should not reach server")
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			_, err := svc.Update(context.Background(), Params{
				TargetURL: ucare.String("https://example.com/hook"),
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "params.ID is required")
		})
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/webhooks/unsubscribe/", r.URL.Path)

		var body deleteParams
		require.NoError(t, json.Unmarshal(uctest.ReadBody(t, r), &body))
		assert.Equal(t, "https://example.com/hook", body.TargetURL)

		w.WriteHeader(http.StatusNoContent)
	}), func(t *testing.T, srv *httptest.Server) {
		svc := NewService(uctest.NewServerClient(srv))
		err := svc.Delete(context.Background(), "https://example.com/hook")
		require.NoError(t, err)
	})
}
