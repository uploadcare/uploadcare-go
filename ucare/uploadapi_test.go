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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		test:     "form_data",
		endpoint: config.UploadAPIEndpoint,
		method:   http.MethodPost,
		requrl:   "/base/",
		data: testReqEncoder{
			body:  "formkey=formvalue",
			query: "qparam1=qparamvalue1&qparam2=qparamvalue2",
		},
		checkReq: func(r *http.Request) error {
			data, _ := io.ReadAll(r.Body)
			if string(data) != "formkey=formvalue" {
				return errors.New("invalid req body data")
			}
			if r.URL.RawQuery != "qparam1=qparamvalue1&qparam2=qparamvalue2" {
				return errors.New("invalid req query")
			}
			return nil
		},
	}}

	for _, c := range cases {
		t.Run(c.test, func(t *testing.T) {
			t.Parallel()

			req, err := client.NewRequest(
				context.Background(),
				c.endpoint,
				c.method,
				c.requrl,
				c.data,
			)
			require.NoError(t, err)
			require.NoError(t, c.checkReq(req))
		})
	}
}

func TestUploadDo(t *testing.T) {
	t.Parallel()

	t.Run("throttle_no_retry_by_default", func(t *testing.T) {
		t.Parallel()

		var count atomic.Int32
		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count.Add(1)
			w.WriteHeader(http.StatusTooManyRequests)
		}), func(t *testing.T, srv *httptest.Server) {
			client := &uploadAPIClient{conn: srv.Client()}
			req, err := http.NewRequest(http.MethodPost, srv.URL+"/base/", nil)
			require.NoError(t, err)

			err = client.Do(req, nil)

			require.Error(t, err)
			var throttleErr ThrottleError
			require.True(t, errors.As(err, &throttleErr))
			assert.Equal(t, int32(1), count.Load())
		})
	})

	t.Run("throttle_retry_success", func(t *testing.T) {
		t.Parallel()

		var count atomic.Int32
		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			n := count.Add(1)
			if n < 2 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			respondJSON(w, map[string]string{"file": "test-id"})
		}), func(t *testing.T, srv *httptest.Server) {
			client := &uploadAPIClient{
				conn:  srv.Client(),
				retry: &RetryConfig{MaxRetries: 3},
			}
			req, err := http.NewRequest(http.MethodPost, srv.URL+"/base/", nil)
			require.NoError(t, err)

			var result map[string]string
			err = client.Do(req, &result)

			require.NoError(t, err)
			assert.Equal(t, "test-id", result["file"])
			assert.Equal(t, int32(2), count.Load())
		})
	})

	t.Run("requires_pointer_resdata", func(t *testing.T) {
		t.Parallel()

		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, map[string]string{"file": "test-id"})
		}), func(t *testing.T, srv *httptest.Server) {
			client := &uploadAPIClient{conn: srv.Client()}
			req, err := http.NewRequest(http.MethodPost, srv.URL+"/base/", nil)
			require.NoError(t, err)

			result := map[string]string{}
			err = client.Do(req, result)

			require.Error(t, err)
			var invalidUnmarshal *json.InvalidUnmarshalError
			require.True(t, errors.As(err, &invalidUnmarshal))
		})
	})

	t.Run("unhandled_status_plain_text", func(t *testing.T) {
		t.Parallel()

		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("upstream connect error"))
		}), func(t *testing.T, srv *httptest.Server) {
			client := &uploadAPIClient{conn: srv.Client()}
			req, err := http.NewRequest(http.MethodPost, srv.URL+"/base/", nil)
			require.NoError(t, err)

			err = client.Do(req, nil)

			require.Error(t, err)
			var apiErr APIError
			require.True(t, errors.As(err, &apiErr))
			assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
			assert.Equal(t, "upstream connect error", apiErr.Detail)
		})
	})
}
