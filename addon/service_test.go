package addon

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

func TestExecute(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/addons/remove_bg/execute/", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		var params ExecuteParams
		assert.NoError(t, json.Unmarshal(body, &params))
		assert.Equal(t, "file-uuid-123", params.Target)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ExecuteResult{RequestID: "req-456"})
	}))
	defer srv.Close()

	result, err := svc.Execute(context.Background(), AddonRemoveBG, ExecuteParams{
		Target: "file-uuid-123",
	})
	assert.NoError(t, err)
	assert.Equal(t, "req-456", result.RequestID)
}

func TestExecute_WithRemoveBGParams(t *testing.T) {
	t.Parallel()

	crop := true
	typeLevel := "2"

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		var raw map[string]json.RawMessage
		assert.NoError(t, json.Unmarshal(body, &raw))

		var params RemoveBGParams
		assert.NoError(t, json.Unmarshal(raw["params"], &params))
		assert.Equal(t, &crop, params.Crop)
		assert.Equal(t, &typeLevel, params.TypeLevel)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ExecuteResult{RequestID: "req-789"})
	}))
	defer srv.Close()

	result, err := svc.Execute(context.Background(), AddonRemoveBG, ExecuteParams{
		Target: "file-uuid",
		Params: RemoveBGParams{
			Crop:      &crop,
			TypeLevel: &typeLevel,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "req-789", result.RequestID)
}

func TestExecute_WithClamAVParams(t *testing.T) {
	t.Parallel()

	purge := true

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		var raw map[string]json.RawMessage
		assert.NoError(t, json.Unmarshal(body, &raw))

		var params ClamAVParams
		assert.NoError(t, json.Unmarshal(raw["params"], &params))
		assert.Equal(t, &purge, params.PurgeInfected)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ExecuteResult{RequestID: "req-clam"})
	}))
	defer srv.Close()

	result, err := svc.Execute(context.Background(), AddonClamAV, ExecuteParams{
		Target: "file-uuid",
		Params: ClamAVParams{PurgeInfected: &purge},
	})
	assert.NoError(t, err)
	assert.Equal(t, "req-clam", result.RequestID)
}

func TestExecute_NoParams(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		var raw map[string]json.RawMessage
		assert.NoError(t, json.Unmarshal(body, &raw))

		// params should be absent when nil
		_, hasParams := raw["params"]
		assert.False(t, hasParams)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ExecuteResult{RequestID: "req-rek"})
	}))
	defer srv.Close()

	result, err := svc.Execute(context.Background(), AddonRekognitionLabels, ExecuteParams{
		Target: "file-uuid",
	})
	assert.NoError(t, err)
	assert.Equal(t, "req-rek", result.RequestID)
}

func TestStatus(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/addons/remove_bg/execute/status/", r.URL.Path)
		assert.Equal(t, "req-456", r.URL.Query().Get("request_id"))

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"done","result":{"foreground_type":"person"}}`))
	}))
	defer srv.Close()

	result, err := svc.Status(context.Background(), AddonRemoveBG, "req-456")
	assert.NoError(t, err)
	assert.Equal(t, StatusDone, result.Status)
	assert.Contains(t, string(result.Result), "foreground_type")
}

func TestStatus_InProgress(t *testing.T) {
	t.Parallel()

	svc, srv := newTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"in_progress"}`))
	}))
	defer srv.Close()

	result, err := svc.Status(context.Background(), AddonClamAV, "req-pending")
	assert.NoError(t, err)
	assert.Equal(t, StatusInProgress, result.Status)
}
