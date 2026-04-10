package ucare

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/config"
)

type testReqEncoder struct {
	body  string
	query string
}

func (t testReqEncoder) EncodeReq(r *http.Request) error {
	r.URL.RawQuery = t.query
	r.Body = io.NopCloser(strings.NewReader(t.body))
	return nil
}

func testCreds() APICreds {
	return APICreds{
		SecretKey: "testsecretkey",
		PublicKey: "testpublickey",
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type trackedReadCloser struct {
	io.ReadCloser
	closed *bool
}

func (t trackedReadCloser) Close() error {
	*t.closed = true
	return t.ReadCloser.Close()
}

func respondJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func withServer(t *testing.T, handler http.Handler, fn func(*testing.T, *httptest.Server)) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	fn(t, srv)
}

func TestRESTAPIClient(t *testing.T) {
	t.Parallel()

	client := newRESTAPIClient(testCreds(), resolveConfig(nil))

	cases := []struct {
		test string

		endpoint config.Endpoint
		method   string
		requrl   string
		data     ReqEncoder

		checkReq func(*http.Request) error
	}{{
		test:     "simple case",
		endpoint: config.RESTAPIEndpoint,
		method:   http.MethodGet,
		requrl:   "/files/",
		data:     nil,
		checkReq: func(r *http.Request) error {
			h := r.Header
			if h.Get("Accept") !=
				"application/vnd.uploadcare-v0.7+json" {
				return errors.New("wrong accept header")
			}
			if h.Get("User-Agent") !=
				"UploadcareGo/2.0.0/testpublickey" {
				return errors.New("wrong user-agent header")
			}
			if h.Get("Content-Type") != "application/json" {
				return errors.New("wrong content-type header")
			}
			_, err := time.Parse(dateHeaderFormat, h.Get("Date"))
			if err != nil {
				return err
			}
			if h.Get("Authorization") == "" {
				return errors.New("auth header is not set")
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

func TestDo(t *testing.T) {
	t.Parallel()

	t.Run("unhandled_status_with_detail", func(t *testing.T) {
		t.Parallel()

		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"detail":"Addon is already running for this file."}`))
		}), func(t *testing.T, srv *httptest.Server) {
			client := &restAPIClient{conn: srv.Client()}
			req, err := http.NewRequest(http.MethodPost, srv.URL+"/addons/uc_clamav_virus_scan/execute/", nil)
			require.NoError(t, err)

			var result struct {
				RequestID string `json:"request_id"`
			}
			err = client.Do(req, &result)

			require.Error(t, err)
			var apiErr APIError
			assert.True(t, errors.As(err, &apiErr))
			assert.Equal(t, http.StatusConflict, apiErr.StatusCode)
			assert.Equal(t, "Addon is already running for this file.", apiErr.Detail)
			assert.Equal(t, "", result.RequestID)
		})
	})

	t.Run("unhandled_status_without_detail", func(t *testing.T) {
		t.Parallel()

		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("Bad Gateway"))
		}), func(t *testing.T, srv *httptest.Server) {
			client := &restAPIClient{conn: srv.Client()}
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/files/", nil)
			require.NoError(t, err)

			var result map[string]string
			err = client.Do(req, &result)

			require.Error(t, err)
			var apiErr APIError
			assert.True(t, errors.As(err, &apiErr))
			assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
			assert.Equal(t, "Bad Gateway", apiErr.Detail)
		})
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()

		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"detail":"Account is inactive."}`))
		}), func(t *testing.T, srv *httptest.Server) {
			client := &restAPIClient{conn: srv.Client()}
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/files/", nil)
			require.NoError(t, err)

			err = client.Do(req, nil)

			require.Error(t, err)
			var forbiddenErr ForbiddenError
			assert.True(t, errors.As(err, &forbiddenErr))
			assert.Equal(t, 403, forbiddenErr.StatusCode)
			assert.Equal(t, "Account is inactive.", forbiddenErr.Detail)
		})
	})

	t.Run("unhandled_status_nil_resdata", func(t *testing.T) {
		t.Parallel()

		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"detail":"Conflict"}`))
		}), func(t *testing.T, srv *httptest.Server) {
			client := &restAPIClient{conn: srv.Client()}
			req, err := http.NewRequest(http.MethodDelete, srv.URL+"/groups/abc~1/", nil)
			require.NoError(t, err)

			err = client.Do(req, nil)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "Conflict")
		})
	})

	t.Run("success_nil_resdata_closes_body", func(t *testing.T) {
		t.Parallel()

		closed := false
		client := &restAPIClient{
			conn: &http.Client{
				Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusNoContent,
						Header:     make(http.Header),
						Body: trackedReadCloser{
							ReadCloser: io.NopCloser(strings.NewReader("")),
							closed:     &closed,
						},
					}, nil
				}),
			},
		}
		req, err := http.NewRequest(http.MethodDelete, "https://example.test/groups/abc~1/", nil)
		require.NoError(t, err)

		err = client.Do(req, nil)

		require.NoError(t, err)
		assert.True(t, closed)
	})
}

func TestDoThrottle(t *testing.T) {
	t.Parallel()

	t.Run("no_retry_by_default", func(t *testing.T) {
		t.Parallel()

		var count atomic.Int32
		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count.Add(1)
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
		}), func(t *testing.T, srv *httptest.Server) {
			client := &restAPIClient{conn: srv.Client()}
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/files/", nil)
			require.NoError(t, err)

			err = client.Do(req, nil)

			require.Error(t, err)
			var throttleErr ThrottleError
			assert.True(t, errors.As(err, &throttleErr))
			assert.Equal(t, 1, throttleErr.RetryAfter)
			assert.Equal(t, int32(1), count.Load())
		})
	})

	t.Run("retry_success", func(t *testing.T) {
		t.Parallel()

		var count atomic.Int32
		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			n := count.Add(1)
			if n < 3 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			respondJSON(w, map[string]bool{"ok": true})
		}), func(t *testing.T, srv *httptest.Server) {
			client := &restAPIClient{
				conn:  srv.Client(),
				retry: &RetryConfig{MaxRetries: 3},
			}
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/files/", nil)
			require.NoError(t, err)

			var result map[string]bool
			err = client.Do(req, &result)

			require.NoError(t, err)
			assert.True(t, result["ok"])
			assert.Equal(t, int32(3), count.Load())
		})
	})

	t.Run("retries_exhausted", func(t *testing.T) {
		t.Parallel()

		var count atomic.Int32
		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count.Add(1)
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
		}), func(t *testing.T, srv *httptest.Server) {
			client := &restAPIClient{
				conn:  srv.Client(),
				retry: &RetryConfig{MaxRetries: 2},
			}
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/files/", nil)
			require.NoError(t, err)

			err = client.Do(req, nil)

			require.Error(t, err)
			var throttleErr ThrottleError
			assert.True(t, errors.As(err, &throttleErr))
			// 1st request + 2 retries = 3 total, then on 3rd retry (tries=3) tries > MaxRetries(2)
			assert.Equal(t, int32(3), count.Load())
		})
	})

	t.Run("max_wait_exceeded", func(t *testing.T) {
		t.Parallel()

		var count atomic.Int32
		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count.Add(1)
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
		}), func(t *testing.T, srv *httptest.Server) {
			client := &restAPIClient{
				conn:  srv.Client(),
				retry: &RetryConfig{MaxRetries: 3, MaxWaitSeconds: 1},
			}
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/files/", nil)
			require.NoError(t, err)

			start := time.Now()
			err = client.Do(req, nil)

			require.Error(t, err)
			var throttleErr ThrottleError
			assert.True(t, errors.As(err, &throttleErr))
			assert.Equal(t, 60, throttleErr.RetryAfter)
			assert.Equal(t, int32(1), count.Load())
			assert.Less(t, time.Since(start).Seconds(), 5.0)
		})
	})

	t.Run("context_cancelled", func(t *testing.T) {
		t.Parallel()

		var count atomic.Int32
		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count.Add(1)
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
		}), func(t *testing.T, srv *httptest.Server) {
			client := &restAPIClient{
				conn:  srv.Client(),
				retry: &RetryConfig{MaxRetries: 3},
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/files/", nil)
			require.NoError(t, err)

			err = client.Do(req, nil)

			require.Error(t, err)
		})
	})
}
