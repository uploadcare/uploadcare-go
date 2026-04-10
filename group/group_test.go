package group

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/uctest"
)

const testGroupID = "test-group-id~3"

func TestInfo(t *testing.T) {
	t.Parallel()

	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/groups/"+testGroupID+"/", r.URL.Path)

		uctest.RespondJSON(t, w, Info{
			ID:        testGroupID,
			FileCount: 3,
			CDNLink:   "https://ucarecdn.com/" + testGroupID + "/",
		})
	}), func(t *testing.T, srv *httptest.Server) {
		svc := NewService(uctest.NewServerClient(srv))
		info, err := svc.Info(context.Background(), testGroupID)
		require.NoError(t, err)
		assert.Equal(t, testGroupID, info.ID)
		assert.Equal(t, uint64(3), info.FileCount)
		assert.Equal(t, "https://ucarecdn.com/"+testGroupID+"/", info.CDNLink)
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/groups/"+testGroupID+"/", r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	}), func(t *testing.T, srv *httptest.Server) {
		svc := NewService(uctest.NewServerClient(srv))

		err := svc.Delete(context.Background(), testGroupID)
		require.NoError(t, err)
	})
}

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("single_page", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/groups/", r.URL.Path)

			resp := map[string]any{
				"next":    nil,
				"results": []Info{{ID: "g1~1", FileCount: 1}, {ID: "g2~2", FileCount: 2}},
			}
			uctest.RespondJSON(t, w, resp)
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			list, err := svc.List(context.Background(), ListParams{})
			require.NoError(t, err)

			var results []Info
			for list.Next() {
				info, err := list.ReadResult()
				require.NoError(t, err)
				results = append(results, *info)
			}

			require.Len(t, results, 2)
			assert.Equal(t, "g1~1", results[0].ID)
			assert.Equal(t, "g2~2", results[1].ID)
		})
	})

	t.Run("with_query_params", func(t *testing.T) {
		t.Parallel()

		limit := uint64(10)
		ordering := OrderByCreatedAtDesc

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "10", r.URL.Query().Get("limit"))
			assert.Equal(t, OrderByCreatedAtDesc, r.URL.Query().Get("ordering"))

			resp := map[string]any{
				"next":    nil,
				"results": []Info{},
			}
			uctest.RespondJSON(t, w, resp)
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			list, err := svc.List(context.Background(), ListParams{
				Limit:   &limit,
				OrderBy: &ordering,
			})
			require.NoError(t, err)
			assert.False(t, list.Next())
		})
	})

	t.Run("empty_results", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]any{
				"next":    nil,
				"results": []Info{},
			}
			uctest.RespondJSON(t, w, resp)
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			list, err := svc.List(context.Background(), ListParams{})
			require.NoError(t, err)
			assert.False(t, list.Next())
		})
	})

	t.Run("full_unmarshal", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := `{
				"next": null,
				"results": [
					{
						"id": "abc~2",
						"files_count": 2,
						"cdn_url": "https://ucarecdn.com/abc~2/",
						"datetime_created": "2024-01-15T10:30:00Z",
						"datetime_stored": null
					}
				]
			}`
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(raw))
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			list, err := svc.List(context.Background(), ListParams{})
			require.NoError(t, err)

			require.True(t, list.Next())
			info, err := list.ReadResult()
			require.NoError(t, err)

			assert.Equal(t, "abc~2", info.ID)
			assert.Equal(t, uint64(2), info.FileCount)
			assert.Equal(t, "https://ucarecdn.com/abc~2/", info.CDNLink)
			require.NotNil(t, info.CreatedAt)
			assert.Nil(t, info.StoredAt)

			assert.Equal(t, 2024, info.CreatedAt.Year())
			assert.Equal(t, time.January, info.CreatedAt.Month())
			assert.Equal(t, 15, info.CreatedAt.Day())
		})
	})
}
