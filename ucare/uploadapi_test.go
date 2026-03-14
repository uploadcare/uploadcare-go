package ucare

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	assert "github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/config"
)

func TestUploadAPIClient(t *testing.T) {
	t.Parallel()

	client := newUploadAPIClient(testCreds(), resolveConfig(nil))

	cases := []struct {
		test string

		endpoint config.Endpoint
		method   string
		requrl   string
		data     ReqEncoder

		checkReq func(*http.Request) error
	}{{
		test:     "simple case ",
		endpoint: config.UploadAPIEndpoint,
		method:   http.MethodPost,
		requrl:   "/base/",
		data: testReqEncoder{
			body:  "formkey=formvalue",
			query: "qparam1=qparamvalue1&qparam2=qparamvalue2",
		},
		checkReq: func(r *http.Request) error {
			// check only data in this test case
			data, _ := io.ReadAll(r.Body)
			if string(data) != "formkey=formvalue" {
				return errors.New("invalid req body data")
			}
			qr := r.URL.RawQuery
			if qr != "qparam1=qparamvalue1&qparam2=qparamvalue2" {
				return errors.New("invlid req query")
			}
			return nil
		},
	}}

	for _, c := range cases {
		c := c
		t.Run(c.test, func(t *testing.T) {
			t.Parallel()

			req, err := client.NewRequest(
				context.Background(),
				c.endpoint,
				c.method,
				c.requrl,
				c.data,
			)
			assert.Equal(t, nil, err)
			assert.Equal(t, nil, c.checkReq(req))
		})
	}
}

func TestUploadDo_ThrottleNoRetryByDefault(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := &uploadAPIClient{conn: srv.Client()}
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/base/", nil)
	assert.NoError(t, err)

	err = client.Do(req, nil)

	assert.Error(t, err)
	var throttleErr ThrottleError
	assert.True(t, errors.As(err, &throttleErr))
	assert.Equal(t, int32(1), count.Load())
}

func TestUploadDo_ThrottleRetrySuccess(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := count.Add(1)
		if n < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"file":"test-id"}`))
	}))
	defer srv.Close()

	client := &uploadAPIClient{
		conn:  srv.Client(),
		retry: &RetryConfig{MaxRetries: 3},
	}
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/base/", nil)
	assert.NoError(t, err)

	var result map[string]string
	err = client.Do(req, &result)

	assert.NoError(t, err)
	assert.Equal(t, "test-id", result["file"])
	assert.Equal(t, int32(2), count.Load())
}

func TestUploadDo_SuccessRequiresPointerResdata(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"file":"test-id"}`))
	}))
	defer srv.Close()

	client := &uploadAPIClient{conn: srv.Client()}
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/base/", nil)
	assert.NoError(t, err)

	result := map[string]string{}
	err = client.Do(req, result)

	assert.Error(t, err)
	var invalidUnmarshal *json.InvalidUnmarshalError
	assert.True(t, errors.As(err, &invalidUnmarshal))
}

func TestUploadDo_UnhandledStatusPlainTextBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("upstream connect error"))
	}))
	defer srv.Close()

	client := &uploadAPIClient{conn: srv.Client()}
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/base/", nil)
	assert.NoError(t, err)

	err = client.Do(req, nil)

	assert.Error(t, err)
	var apiErr APIError
	assert.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
	assert.Equal(t, "upstream connect error", apiErr.Detail)
}
