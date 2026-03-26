package ucare

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	c, err := NewBearerClient("test-token", NewBearerConfig())
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewBearerClient_RequiresConfig(t *testing.T) {
	t.Parallel()

	_, err := NewBearerClient("test-token", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NewBearerConfig")
}

func TestNewBearerClient_DefaultsHTTPClient(t *testing.T) {
	t.Parallel()

	conf := &Config{}
	c, err := NewBearerClient("test-token", conf)
	require.NoError(t, err)

	pClient := c.(*client).backends[config.RESTAPIEndpoint].(*projectAPIClient)
	assert.Same(t, http.DefaultClient, pClient.conn)
	assert.Nil(t, conf.HTTPClient)
}

func TestNewBearerConfig(t *testing.T) {
	t.Parallel()

	retry := &RetryConfig{MaxRetries: 3, MaxWaitSeconds: 10}
	conf := NewBearerConfig(WithRetry(retry), WithUserAgent("my-app/1.0"))

	assert.Same(t, http.DefaultClient, conf.HTTPClient)
	assert.Same(t, retry, conf.Retry)
	assert.Equal(t, "my-app/1.0", conf.UserAgent)
}

func TestProjectAPIClient_NewRequest(t *testing.T) {
	t.Parallel()

	client := newProjectAPIClient("my-bearer-token", NewBearerConfig())

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

	client := newProjectAPIClient("tok", NewBearerConfig())

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
		_, _ = w.Write([]byte(`{"pub_key":"abc","name":"Test"}`))
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
		_, _ = w.Write([]byte(`{"message":"Project not found.","code":"not_found_error"}`))
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
		_, _ = w.Write([]byte("Bad Gateway"))
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

func TestBearerClient_CrossHostNewRequest(t *testing.T) {
	t.Parallel()

	c, err := NewBearerClient("tok", NewBearerConfig())
	require.NoError(t, err)

	req, err := c.NewRequest(
		context.Background(),
		config.Endpoint("app.uploadcare.com"),
		http.MethodGet,
		"https://app.uploadcare.com/apps/api/project-api/v1/projects/?limit=2",
		nil,
	)
	assert.NoError(t, err)
	assert.Equal(t, "Bearer tok", req.Header.Get("Authorization"))
	assert.Equal(t, "https://app.uploadcare.com/apps/api/project-api/v1/projects/?limit=2", req.URL.String())
}

func TestBearerClient_CrossHostDo(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"next":null,"results":[{"pub_key":"pk3"}]}`))
	}))
	defer srv.Close()

	pClient := &projectAPIClient{conn: srv.Client(), token: "tok"}
	c := &client{
		backends: map[config.Endpoint]Client{
			config.RESTAPIEndpoint: pClient,
		},
		fallbackNewReq: func(ctx context.Context, endpoint config.Endpoint, method, requrl string, data ReqEncoder) (*http.Request, error) {
			return pClient.NewRequest(ctx, endpoint, method, requrl, data)
		},
		fallbackDo: func(req *http.Request, resdata interface{}) error {
			return pClient.Do(req, resdata)
		},
	}

	req, err := c.NewRequest(
		context.Background(),
		config.Endpoint(srv.Listener.Addr().String()),
		http.MethodGet,
		srv.URL+"/page2/",
		nil,
	)
	assert.NoError(t, err)

	var result struct {
		Next    *string         `json:"next"`
		Results json.RawMessage `json:"results"`
	}
	err = c.Do(req, &result)
	assert.NoError(t, err)
	assert.Nil(t, result.Next)
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
		_, _ = w.Write([]byte(`{"ok":true}`))
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

func TestProjectAPIClient_HandleThrottleBackoffExceedsMaxWait(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://api.uploadcare.com/projects/", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusTooManyRequests)
	resp := rec.Result()
	resp.Request = req

	client := &projectAPIClient{retry: &RetryConfig{MaxRetries: 10, MaxWaitSeconds: 3}}
	retry, err := client.handleResponse(resp, req, nil, 5)
	assert.False(t, retry)

	var throttleErr ThrottleError
	require.True(t, errors.As(err, &throttleErr))
	assert.Equal(t, 16, throttleErr.RetryAfter)
}
