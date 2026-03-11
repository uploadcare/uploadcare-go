package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	assert "github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/config"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

// testClient implements ucare.Client for test purposes, pointing at an
// httptest.Server instead of the real Uploadcare API.
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
	req.Header.Set("Accept", "application/vnd.uploadcare-v0.7+json")
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

func TestList(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/files/test-uuid/metadata/", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"key1": "value1",
			"key2": "value2",
		})
	}))
	defer srv.Close()

	data, err := svc.List(context.Background(), "test-uuid")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"key1": "value1", "key2": "value2"}, data)
}

func TestGet(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/files/test-uuid/metadata/mykey/", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("my-value")
	}))
	defer srv.Close()

	val, err := svc.Get(context.Background(), "test-uuid", "mykey")
	assert.NoError(t, err)
	assert.Equal(t, "my-value", val)
}

func TestSet(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/files/test-uuid/metadata/mykey/", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		var got string
		assert.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "new-value", got)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("new-value")
	}))
	defer srv.Close()

	val, err := svc.Set(context.Background(), "test-uuid", "mykey", "new-value")
	assert.NoError(t, err)
	assert.Equal(t, "new-value", val)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/files/test-uuid/metadata/mykey/", r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := svc.Delete(context.Background(), "test-uuid", "mykey")
	assert.NoError(t, err)
}

func TestGet_NotFound(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"detail":"Not found."}`))
	}))
	defer srv.Close()

	_, err := svc.Get(context.Background(), "test-uuid", "nokey")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}
