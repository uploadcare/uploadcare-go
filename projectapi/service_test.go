package projectapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uploadcare/uploadcare-go/v2/internal/config"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

type testClient struct {
	httpClient *http.Client
	baseURL    string
}

func (c *testClient) NewRequest(
	ctx context.Context,
	_ config.Endpoint,
	method, requrl string,
	data ucare.ReqEncoder,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+requrl, nil)
	if err != nil {
		return nil, err
	}
	if data != nil {
		if err = data.EncodeReq(req); err != nil {
			return nil, err
		}
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *testClient) Do(req *http.Request, resdata interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}

	if resdata == nil || reflect.ValueOf(resdata).IsNil() {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(resdata)
}

func newTestService(handler http.Handler) (Service, *httptest.Server) {
	srv := httptest.NewServer(handler)
	client := &testClient{httpClient: srv.Client(), baseURL: srv.URL}
	return NewService(client), srv
}

func TestListProjects(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/", r.URL.Path)
		assert.Equal(t, "10", r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"next":null,"results":[{"pub_key":"pk1","name":"Project 1"},{"pub_key":"pk2","name":"Project 2"}]}`)
	}))
	defer srv.Close()

	limit := uint64(10)
	list, err := svc.List(context.Background(), &ListParams{Limit: &limit})
	assert.NoError(t, err)

	var projects []Project
	for list.Next() {
		p, err := list.ReadResult()
		assert.NoError(t, err)
		projects = append(projects, *p)
	}
	assert.Len(t, projects, 2)
	assert.Equal(t, "pk1", projects[0].PubKey)
	assert.Equal(t, "pk2", projects[1].PubKey)
}

func TestListProjects_NilParams(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "", r.URL.RawQuery)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"next":null,"results":[]}`)
	}))
	defer srv.Close()

	list, err := svc.List(context.Background(), nil)
	assert.NoError(t, err)
	assert.False(t, list.Next())
}

func TestCreateProject(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/projects/", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		var params CreateProjectParams
		assert.NoError(t, json.Unmarshal(body, &params))
		assert.Equal(t, "New Project", params.Name)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Project{
			PubKey: "newpk",
			Name:   "New Project",
		})
	}))
	defer srv.Close()

	data, err := svc.Create(context.Background(), CreateProjectParams{
		Name: "New Project",
	})
	assert.NoError(t, err)
	assert.Equal(t, "newpk", data.PubKey)
	assert.Equal(t, "New Project", data.Name)
}

func TestGetProject(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/mypk/", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Project{
			PubKey: "mypk",
			Name:   "My Project",
		})
	}))
	defer srv.Close()

	data, err := svc.Get(context.Background(), "mypk")
	assert.NoError(t, err)
	assert.Equal(t, "mypk", data.PubKey)
}

func TestUpdateProject(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/projects/mypk/", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		var raw map[string]interface{}
		assert.NoError(t, json.Unmarshal(body, &raw))
		assert.Equal(t, "Updated", raw["name"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Project{
			PubKey: "mypk",
			Name:   "Updated",
		})
	}))
	defer srv.Close()

	name := "Updated"
	data, err := svc.Update(context.Background(), "mypk", UpdateProjectParams{
		Name: &name,
	})
	assert.NoError(t, err)
	assert.Equal(t, "Updated", data.Name)
}

func TestDeleteProject(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/projects/mypk/", r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := svc.Delete(context.Background(), "mypk")
	assert.NoError(t, err)
}

func TestListSecrets(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/mypk/secrets/", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"next":null,"results":[{"id":"sec-id","hint":"ea94"}]}`)
	}))
	defer srv.Close()

	list, err := svc.ListSecrets(context.Background(), "mypk", nil)
	assert.NoError(t, err)

	assert.True(t, list.Next())
	s, err := list.ReadResult()
	assert.NoError(t, err)
	assert.Equal(t, "sec-id", s.ID)
	assert.Equal(t, "ea94", s.Hint)
	assert.False(t, list.Next())
}

func TestCreateSecret(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/projects/mypk/secrets/", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SecretRevealed{
			ID:     "new-sec-id",
			Secret: "ea9464b47affc143c22c",
		})
	}))
	defer srv.Close()

	data, err := svc.CreateSecret(context.Background(), "mypk")
	assert.NoError(t, err)
	assert.Equal(t, "new-sec-id", data.ID)
	assert.Equal(t, "ea9464b47affc143c22c", data.Secret)
}

func TestDeleteSecret(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/projects/mypk/secrets/sec-id/", r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := svc.DeleteSecret(context.Background(), "mypk", "sec-id")
	assert.NoError(t, err)
}

func TestGetUsage(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/mypk/usage/", r.URL.Path)
		assert.Equal(t, "2025-01-01", r.URL.Query().Get("from"))
		assert.Equal(t, "2025-01-31", r.URL.Query().Get("to"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UsageMetricsCombined{
			Units: map[string]string{
				"traffic":    "bytes",
				"storage":    "bytes",
				"operations": "operations",
			},
			Data: []CombinedUsageDataPoint{
				{Date: "2025-01-01", Traffic: 100, Storage: 200, Operations: 5},
			},
		})
	}))
	defer srv.Close()

	data, err := svc.GetUsage(context.Background(), "mypk", UsageDateRange{
		From: "2025-01-01",
		To:   "2025-01-31",
	})
	assert.NoError(t, err)
	assert.Equal(t, "bytes", data.Units["traffic"])
	assert.Len(t, data.Data, 1)
	assert.Equal(t, int64(100), data.Data[0].Traffic)
}

func TestGetUsageMetric(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/mypk/usage/traffic/", r.URL.Path)
		assert.Equal(t, "2025-01-01", r.URL.Query().Get("from"))
		assert.Equal(t, "2025-01-31", r.URL.Query().Get("to"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UsageMetric{
			Metric: "traffic",
			Unit:   "bytes",
			Data: []UsageDataPoint{
				{Date: "2025-01-01", Value: 12345},
			},
		})
	}))
	defer srv.Close()

	data, err := svc.GetUsageMetric(context.Background(), "mypk", "traffic", UsageDateRange{
		From: "2025-01-01",
		To:   "2025-01-31",
	})
	assert.NoError(t, err)
	assert.Equal(t, "traffic", data.Metric)
	assert.Equal(t, "bytes", data.Unit)
	assert.Len(t, data.Data, 1)
	assert.Equal(t, int64(12345), data.Data[0].Value)
}
