package file

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
	req.Header.Set("Accept", "application/vnd.uploadcare-v0.7+json")
	return req, nil
}

func (c *testClient) Do(req *http.Request, resdata interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

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

func TestInfo_WithIncludeAppdata(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/files/test-uuid/", r.URL.Path)
		assert.Equal(t, "appdata", r.URL.Query().Get("include"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Info{
			BasicFileInfo: BasicFileInfo{ID: "test-uuid"},
		})
	}))
	defer srv.Close()

	info, err := svc.Info(context.Background(), "test-uuid", &InfoParams{
		Include: ucare.String("appdata"),
	})
	assert.NoError(t, err)
	assert.Equal(t, "test-uuid", info.ID)
}

func TestInfo_NilParams(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/files/test-uuid/", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Info{
			BasicFileInfo: BasicFileInfo{ID: "test-uuid"},
		})
	}))
	defer srv.Close()

	info, err := svc.Info(context.Background(), "test-uuid", nil)
	assert.NoError(t, err)
	assert.Equal(t, "test-uuid", info.ID)
}

func TestListParams_WithInclude(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://example.test/files/", nil)
	assert.NoError(t, err)

	err = (&ListParams{Include: ucare.String("appdata")}).EncodeReq(req)
	assert.NoError(t, err)
	assert.Equal(t, "appdata", req.URL.Query().Get("include"))
}
