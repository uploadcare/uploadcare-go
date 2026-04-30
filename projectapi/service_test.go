package projectapi

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

const (
	testPubKey   = "mypk"
	testSecretID = "sec-id"
)

func withProjectAPIService(t *testing.T, handler http.Handler, fn func(Service)) {
	t.Helper()
	uctest.WithHTTPServer(t, handler, func(t *testing.T, srv *httptest.Server) {
		fn(NewService(uctest.NewServerClient(srv)))
	})
}

func unexpectedRequestHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.RequestURI)
	})
}

func TestListProjects(t *testing.T) {
	t.Parallel()

	t.Run("with_limit", func(t *testing.T) {
		t.Parallel()

		withProjectAPIService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/projects/", r.URL.Path)
			assert.Equal(t, "10", r.URL.Query().Get("limit"))

			uctest.RespondJSON(t, w, map[string]any{
				"next": nil,
				"results": []Project{
					{PubKey: "pk1", Name: "Project 1"},
					{PubKey: "pk2", Name: "Project 2"},
				},
			})
		}), func(svc Service) {
			limit := uint64(10)
			list, err := svc.List(context.Background(), &ListParams{Limit: &limit})
			require.NoError(t, err)

			var projects []Project
			for list.Next() {
				p, err := list.ReadResult()
				require.NoError(t, err)
				projects = append(projects, *p)
			}

			require.Len(t, projects, 2)
			assert.Equal(t, "pk1", projects[0].PubKey)
			assert.Equal(t, "pk2", projects[1].PubKey)
		})
	})

	t.Run("nil_params", func(t *testing.T) {
		t.Parallel()

		withProjectAPIService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Empty(t, r.URL.RawQuery)
			uctest.RespondJSON(t, w, map[string]any{"next": nil, "results": []Project{}})
		}), func(svc Service) {
			list, err := svc.List(context.Background(), nil)
			require.NoError(t, err)
			assert.False(t, list.Next())
		})
	})
}

func TestCreateProject(t *testing.T) {
	t.Parallel()

	withProjectAPIService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/projects/", r.URL.Path)

		var params CreateProjectParams
		require.NoError(t, json.Unmarshal(uctest.ReadBody(t, r), &params))
		assert.Equal(t, "New Project", params.Name)

		uctest.RespondJSON(t, w, Project{PubKey: "newpk", Name: params.Name})
	}), func(svc Service) {
		data, err := svc.Create(context.Background(), CreateProjectParams{Name: "New Project"})
		require.NoError(t, err)
		assert.Equal(t, "newpk", data.PubKey)
		assert.Equal(t, "New Project", data.Name)
	})
}

func TestGetProject(t *testing.T) {
	t.Parallel()

	withProjectAPIService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/"+testPubKey+"/", r.URL.Path)

		uctest.RespondJSON(t, w, Project{PubKey: testPubKey, Name: "My Project"})
	}), func(svc Service) {
		data, err := svc.Get(context.Background(), testPubKey)
		require.NoError(t, err)
		assert.Equal(t, testPubKey, data.PubKey)
	})
}

func TestUpdateProject(t *testing.T) {
	t.Parallel()

	withProjectAPIService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/projects/"+testPubKey+"/", r.URL.Path)

		var params UpdateProjectParams
		require.NoError(t, json.Unmarshal(uctest.ReadBody(t, r), &params))
		require.NotNil(t, params.Name)
		assert.Equal(t, "Updated", *params.Name)

		uctest.RespondJSON(t, w, Project{PubKey: testPubKey, Name: *params.Name})
	}), func(svc Service) {
		name := "Updated"
		data, err := svc.Update(context.Background(), testPubKey, UpdateProjectParams{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "Updated", data.Name)
	})
}

func TestDeleteProject(t *testing.T) {
	t.Parallel()

	withProjectAPIService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/projects/"+testPubKey+"/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}), func(svc Service) {
		err := svc.Delete(context.Background(), testPubKey)
		require.NoError(t, err)
	})
}

func TestListSecrets(t *testing.T) {
	t.Parallel()

	withProjectAPIService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/"+testPubKey+"/secrets/", r.URL.Path)

		uctest.RespondJSON(t, w, map[string]any{
			"next":    nil,
			"results": []SecretListItem{{ID: testSecretID, Hint: "ea94"}},
		})
	}), func(svc Service) {
		list, err := svc.ListSecrets(context.Background(), testPubKey, nil)
		require.NoError(t, err)

		require.True(t, list.Next())
		s, err := list.ReadResult()
		require.NoError(t, err)
		assert.Equal(t, testSecretID, s.ID)
		assert.Equal(t, "ea94", s.Hint)
		assert.False(t, list.Next())
	})
}

func TestCreateSecret(t *testing.T) {
	t.Parallel()

	withProjectAPIService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/projects/"+testPubKey+"/secrets/", r.URL.Path)

		uctest.RespondJSON(t, w, SecretRevealed{
			ID:     "new-sec-id",
			Secret: "ea9464b47affc143c22c",
		})
	}), func(svc Service) {
		data, err := svc.CreateSecret(context.Background(), testPubKey)
		require.NoError(t, err)
		assert.Equal(t, "new-sec-id", data.ID)
		assert.Equal(t, "ea9464b47affc143c22c", data.Secret)
	})
}

func TestDeleteSecret(t *testing.T) {
	t.Parallel()

	withProjectAPIService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/projects/"+testPubKey+"/secrets/"+testSecretID+"/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}), func(svc Service) {
		err := svc.DeleteSecret(context.Background(), testPubKey, testSecretID)
		require.NoError(t, err)
	})
}

func TestGetUsage(t *testing.T) {
	t.Parallel()

	withProjectAPIService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/"+testPubKey+"/usage/", r.URL.Path)
		assert.Equal(t, "2025-01-01", r.URL.Query().Get("from"))
		assert.Equal(t, "2025-01-31", r.URL.Query().Get("to"))

		uctest.RespondJSON(t, w, UsageMetricsCombined{
			Units: map[string]string{
				"traffic":    "bytes",
				"storage":    "bytes",
				"operations": "operations",
			},
			Data: []CombinedUsageDataPoint{
				{Date: "2025-01-01", Traffic: 100, Storage: 200, Operations: 5},
			},
		})
	}), func(svc Service) {
		data, err := svc.GetUsage(context.Background(), testPubKey, UsageDateRange{
			From: "2025-01-01",
			To:   "2025-01-31",
		})
		require.NoError(t, err)
		assert.Equal(t, "bytes", data.Units["traffic"])
		require.Len(t, data.Data, 1)
		assert.Equal(t, int64(100), data.Data[0].Traffic)
	})
}

func TestGetUsageMetric(t *testing.T) {
	t.Parallel()

	withProjectAPIService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/"+testPubKey+"/usage/traffic/", r.URL.Path)
		assert.Equal(t, "2025-01-01", r.URL.Query().Get("from"))
		assert.Equal(t, "2025-01-31", r.URL.Query().Get("to"))

		uctest.RespondJSON(t, w, UsageMetric{
			Metric: UsageMetricTraffic,
			Unit:   "bytes",
			Data:   []UsageDataPoint{{Date: "2025-01-01", Value: 12345}},
		})
	}), func(svc Service) {
		data, err := svc.GetUsageMetric(context.Background(), testPubKey, UsageMetricTraffic, UsageDateRange{
			From: "2025-01-01",
			To:   "2025-01-31",
		})
		require.NoError(t, err)
		assert.Equal(t, UsageMetricTraffic, data.Metric)
		assert.Equal(t, "bytes", data.Unit)
		require.Len(t, data.Data, 1)
		assert.Equal(t, int64(12345), data.Data[0].Value)
	})
}

func TestPathValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		call    func(Service) error
		wantErr error
	}{
		{
			name: "get_empty_pub_key",
			call: func(svc Service) error {
				_, err := svc.Get(context.Background(), "")
				return err
			},
			wantErr: ErrInvalidPubKey,
		},
		{
			name: "update_slash_pub_key",
			call: func(svc Service) error {
				_, err := svc.Update(context.Background(), "bad/key", UpdateProjectParams{})
				return err
			},
			wantErr: ErrInvalidPubKey,
		},
		{
			name:    "delete_empty_pub_key",
			call:    func(svc Service) error { return svc.Delete(context.Background(), "") },
			wantErr: ErrInvalidPubKey,
		},
		{
			name: "list_secrets_slash_pub_key",
			call: func(svc Service) error {
				_, err := svc.ListSecrets(context.Background(), "bad/key", nil)
				return err
			},
			wantErr: ErrInvalidPubKey,
		},
		{
			name: "create_secret_empty_pub_key",
			call: func(svc Service) error {
				_, err := svc.CreateSecret(context.Background(), "")
				return err
			},
			wantErr: ErrInvalidPubKey,
		},
		{
			name:    "delete_secret_empty_secret_id",
			call:    func(svc Service) error { return svc.DeleteSecret(context.Background(), testPubKey, "") },
			wantErr: ErrInvalidSecretID,
		},
		{
			name: "usage_slash_pub_key",
			call: func(svc Service) error {
				_, err := svc.GetUsage(context.Background(), "bad/key", UsageDateRange{})
				return err
			},
			wantErr: ErrInvalidPubKey,
		},
		{
			name: "invalid_usage_metric",
			call: func(svc Service) error {
				_, err := svc.GetUsageMetric(context.Background(), testPubKey, UsageMetricName("bandwidth"), UsageDateRange{})
				return err
			},
			wantErr: ErrInvalidUsageMetric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			withProjectAPIService(t, unexpectedRequestHandler(t), func(svc Service) {
				err := tt.call(svc)
				assert.ErrorIs(t, err, tt.wantErr)
			})
		})
	}
}
