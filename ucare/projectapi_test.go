package ucare

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	assert "github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/config"
)

func TestNewBearerClient_EmptyToken(t *testing.T) {
	t.Parallel()

	_, err := NewBearerClient("", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bearer token must not be empty")
}

func TestNewBearerClient_OK(t *testing.T) {
	t.Parallel()

	c, err := NewBearerClient("test-token", nil)
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestProjectAPIClient_NewRequest(t *testing.T) {
	t.Parallel()

	client := newProjectAPIClient("my-bearer-token", resolveBearerConfig(nil))

	req, err := client.NewRequest(
		context.Background(),
		config.RESTAPIEndpoint,
		http.MethodGet,
		"/projects/",
		nil,
	)
	assert.NoError(t, err)
	assert.Equal(t, "Bearer my-bearer-token", req.Header.Get("Authorization"))
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Contains(t, req.Header.Get("User-Agent"), "UploadcareGo/")
	assert.Equal(t, "https://api.uploadcare.com/projects/", req.URL.String())
}

func TestProjectAPIClient_NewRequest_WithData(t *testing.T) {
	t.Parallel()

	client := newProjectAPIClient("tok", resolveBearerConfig(nil))

	data := testReqEncoder{query: "limit=10"}
	req, err := client.NewRequest(
		context.Background(),
		config.RESTAPIEndpoint,
		http.MethodGet,
		"/projects/",
		&data,
	)
	assert.NoError(t, err)
	assert.Equal(t, "limit=10", req.URL.RawQuery)
}

func TestProjectAPIClient_Do_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"pub_key":"abc","name":"Test"}`))
	}))
	defer srv.Close()

	client := &projectAPIClient{conn: srv.Client()}
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/projects/abc/", nil)
	assert.NoError(t, err)

	var result struct {
		PubKey string `json:"pub_key"`
		Name   string `json:"name"`
	}
	err = client.Do(req, &result)
	assert.NoError(t, err)
	assert.Equal(t, "abc", result.PubKey)
	assert.Equal(t, "Test", result.Name)
}

func TestProjectAPIClient_Do_NoContent(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := &projectAPIClient{conn: srv.Client()}
	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/projects/abc/", nil)
	assert.NoError(t, err)

	err = client.Do(req, nil)
	assert.NoError(t, err)
}

func TestProjectAPIClient_Do_ErrorWithCode(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Project not found.","code":"not_found_error"}`))
	}))
	defer srv.Close()

	client := &projectAPIClient{conn: srv.Client()}
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/projects/abc/", nil)
	assert.NoError(t, err)

	err = client.Do(req, nil)
	assert.Error(t, err)

	var apiErr ProjectAPIError
	assert.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Equal(t, "Project not found.", apiErr.Message)
	assert.Equal(t, "not_found_error", apiErr.Code)
	assert.Contains(t, apiErr.Error(), "not_found_error")
}

func TestProjectAPIClient_Do_ErrorWithoutJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway"))
	}))
	defer srv.Close()

	client := &projectAPIClient{conn: srv.Client()}
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/projects/", nil)
	assert.NoError(t, err)

	err = client.Do(req, nil)
	assert.Error(t, err)

	var apiErr ProjectAPIError
	assert.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
	assert.Equal(t, "Bad Gateway", apiErr.Message)
}

func TestProjectAPIClient_Do_ThrottleNoRetry(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		w.Header().Set("Retry-After", "5")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := &projectAPIClient{conn: srv.Client()}
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/projects/", nil)
	assert.NoError(t, err)

	err = client.Do(req, nil)
	assert.Error(t, err)

	var throttleErr ThrottleError
	assert.True(t, errors.As(err, &throttleErr))
	assert.Equal(t, 5, throttleErr.RetryAfter)
	assert.Equal(t, int32(1), count.Load())
}

func TestProjectAPIClient_Do_ThrottleRetrySuccess(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := count.Add(1)
		if n < 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	client := &projectAPIClient{
		conn:  srv.Client(),
		retry: &RetryConfig{MaxRetries: 3},
	}
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/projects/", nil)
	assert.NoError(t, err)

	var result map[string]bool
	err = client.Do(req, &result)
	assert.NoError(t, err)
	assert.True(t, result["ok"])
	assert.Equal(t, int32(2), count.Load())
}
