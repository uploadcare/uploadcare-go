package ucare

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/config"
)

func TestNewBearerClient(t *testing.T) {
	t.Run("empty_token", func(t *testing.T) {
		t.Parallel()
		_, err := NewBearerClient("", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bearer token must not be empty")
	})
	t.Run("nil_config", func(t *testing.T) {
		t.Parallel()
		_, err := NewBearerClient("test-token", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "NewBearerConfig")
	})
	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		c, err := NewBearerClient("test-token", NewBearerConfig())
		require.NoError(t, err)
		assert.NotNil(t, c)
	})
	t.Run("defaults_http_client", func(t *testing.T) {
		t.Parallel()
		conf := &Config{}
		c, err := NewBearerClient("test-token", conf)
		require.NoError(t, err)

		pClient := c.(*client).backends[config.RESTAPIEndpoint].(*projectAPIClient)
		assert.Same(t, http.DefaultClient, pClient.conn)
		assert.Nil(t, conf.HTTPClient)
	})
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
	cases := []struct {
		test     string
		token    string
		method   string
		requrl   string
		data     ReqEncoder
		checkReq func(*http.Request) error
	}{
		{
			test:   "baseline_get",
			token:  "my-bearer-token",
			method: http.MethodGet,
			requrl: "/projects/",
			data:   nil,
			checkReq: func(r *http.Request) error {
				if g := r.Header.Get("Authorization"); g != "Bearer my-bearer-token" {
					return errors.New("wrong Authorization header")
				}
				if r.Header.Get("Content-Type") != "" {
					return errors.New("content-type should not be set on a bodyless request")
				}
				if !strings.Contains(r.Header.Get("User-Agent"), "UploadcareGo/") {
					return errors.New("expected User-Agent to contain UploadcareGo/")
				}
				if r.URL.String() != "https://api.uploadcare.com/projects/" {
					return errors.New("wrong request URL")
				}
				return nil
			},
		},
		{
			test:   "post_with_query_and_json",
			token:  "tok",
			method: http.MethodPost,
			requrl: "/projects/",
			data:   &testReqEncoder{body: `{"name":"test"}`, query: "limit=10"},
			checkReq: func(r *http.Request) error {
				if r.URL.RawQuery != "limit=10" {
					return errors.New("wrong RawQuery")
				}
				if r.Header.Get("Content-Type") != "application/json" {
					return errors.New("content-type must be application/json when a body is present")
				}
				return nil
			},
		},
		{
			test:   "preserves_encoder_content_type",
			token:  "tok",
			method: http.MethodPost,
			requrl: "/projects/",
			data:   &contentTypeEncoder{contentType: "multipart/form-data; boundary=xyz"},
			checkReq: func(r *http.Request) error {
				if r.Header.Get("Content-Type") != "multipart/form-data; boundary=xyz" {
					return errors.New("encoder Content-Type was not preserved")
				}
				return nil
			},
		},
	}

	for _, c := range cases {
		t.Run(c.test, func(t *testing.T) {
			t.Parallel()
			client := newProjectAPIClient(c.token, NewBearerConfig())
			req, err := client.NewRequest(
				context.Background(),
				config.RESTAPIEndpoint,
				c.method,
				c.requrl,
				c.data,
			)
			require.NoError(t, err)
			require.NoError(t, c.checkReq(req))
		})
	}
}

type contentTypeEncoder struct{ contentType string }

func (e contentTypeEncoder) EncodeReq(r *http.Request) error {
	r.Header.Set("Content-Type", e.contentType)
	return nil
}

func TestProjectAPIClient_Do_Success(t *testing.T) {
	t.Parallel()

	withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, map[string]string{"pub_key": "abc", "name": "Test"})
	}), func(t *testing.T, srv *httptest.Server) {
		client := &projectAPIClient{conn: srv.Client()}
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/projects/abc/", nil)
		require.NoError(t, err)

		var result struct {
			PubKey string `json:"pub_key"`
			Name   string `json:"name"`
		}
		err = client.Do(req, &result)
		assert.NoError(t, err)
		assert.Equal(t, "abc", result.PubKey)
		assert.Equal(t, "Test", result.Name)
	})
}

func TestProjectAPIClient_Do_NoContent(t *testing.T) {
	t.Parallel()

	withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), func(t *testing.T, srv *httptest.Server) {
		client := &projectAPIClient{conn: srv.Client()}
		req, err := http.NewRequest(http.MethodDelete, srv.URL+"/projects/abc/", nil)
		require.NoError(t, err)

		err = client.Do(req, nil)
		assert.NoError(t, err)
	})
}

func TestProjectAPIClient_Do_HTTPError(t *testing.T) {
	cases := []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus int
		wantMsg    string
		wantCode   string
	}{
		{
			name: "json_with_code",
			handler: func(w http.ResponseWriter, r *http.Request) {
				writeJSONStatus(w, http.StatusNotFound, map[string]string{
					"message": "Project not found.",
					"code":    "not_found_error",
				})
			},
			wantStatus: http.StatusNotFound,
			wantMsg:    "Project not found.",
			wantCode:   "not_found_error",
		},
		{
			name: "plain_body",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				_, _ = w.Write([]byte("Bad Gateway"))
			},
			wantStatus: http.StatusBadGateway,
			wantMsg:    "Bad Gateway",
			wantCode:   "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			withServer(t, tc.handler, func(t *testing.T, srv *httptest.Server) {
				client := &projectAPIClient{conn: srv.Client()}
				req, err := http.NewRequest(http.MethodGet, srv.URL+"/projects/abc/", nil)
				require.NoError(t, err)

				err = client.Do(req, nil)
				assert.Error(t, err)

				var apiErr ProjectAPIError
				assert.True(t, errors.As(err, &apiErr))
				assert.Equal(t, tc.wantStatus, apiErr.StatusCode)
				assert.Equal(t, tc.wantMsg, apiErr.Message)
				if tc.wantCode != "" {
					assert.Equal(t, tc.wantCode, apiErr.Code)
					assert.Contains(t, apiErr.Error(), tc.wantCode)
				}
			})
		})
	}
}

func TestProjectAPIClient_Do_AuthzErrors(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSONStatus(w, http.StatusUnauthorized, map[string]string{
				"message": "Invalid token.",
				"code":    "invalid_token",
			})
		}), func(t *testing.T, srv *httptest.Server) {
			client := &projectAPIClient{conn: srv.Client()}
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/projects/", nil)
			require.NoError(t, err)

			err = client.Do(req, nil)
			require.Error(t, err)

			var authErr ProjectAuthError
			require.True(t, errors.As(err, &authErr), "must surface as ProjectAuthError")
			assert.Equal(t, http.StatusUnauthorized, authErr.StatusCode)
			assert.Equal(t, "Invalid token.", authErr.Message)
			assert.Equal(t, "invalid_token", authErr.Code)
			assert.Contains(t, authErr.Error(), "authentication failed")

			var apiErr ProjectAPIError
			require.True(t, errors.As(err, &apiErr), "must remain reachable as ProjectAPIError via Unwrap")
			assert.Equal(t, "invalid_token", apiErr.Code)
		})
	})
	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSONStatus(w, http.StatusForbidden, map[string]string{
				"message": "No access to project.",
				"code":    "forbidden",
			})
		}), func(t *testing.T, srv *httptest.Server) {
			client := &projectAPIClient{conn: srv.Client()}
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/projects/abc/", nil)
			require.NoError(t, err)

			err = client.Do(req, nil)
			require.Error(t, err)

			var forbiddenErr ProjectForbiddenError
			require.True(t, errors.As(err, &forbiddenErr), "must surface as ProjectForbiddenError")
			assert.Equal(t, http.StatusForbidden, forbiddenErr.StatusCode)
			assert.Equal(t, "No access to project.", forbiddenErr.Message)
			assert.Contains(t, forbiddenErr.Error(), "forbidden")

			var apiErr ProjectAPIError
			require.True(t, errors.As(err, &apiErr), "must remain reachable as ProjectAPIError via Unwrap")
		})
	})
}

func TestProjectAPIClient_Do_Throttle(t *testing.T) {
	t.Run("no_retry", func(t *testing.T) {
		t.Parallel()
		var count atomic.Int32
		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count.Add(1)
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusTooManyRequests)
		}), func(t *testing.T, srv *httptest.Server) {
			client := &projectAPIClient{conn: srv.Client()}
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/projects/", nil)
			require.NoError(t, err)

			err = client.Do(req, nil)
			assert.Error(t, err)

			var throttleErr ThrottleError
			assert.True(t, errors.As(err, &throttleErr))
			assert.Equal(t, 5, throttleErr.RetryAfter)
			assert.Equal(t, int32(1), count.Load())
		})
	})
	t.Run("retry_then_success", func(t *testing.T) {
		t.Parallel()
		var count atomic.Int32
		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if count.Add(1) < 2 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			respondJSON(w, map[string]bool{"ok": true})
		}), func(t *testing.T, srv *httptest.Server) {
			client := &projectAPIClient{
				conn:  srv.Client(),
				retry: &RetryConfig{MaxRetries: 3},
			}
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/projects/", nil)
			require.NoError(t, err)

			var result map[string]bool
			err = client.Do(req, &result)
			assert.NoError(t, err)
			assert.True(t, result["ok"])
			assert.Equal(t, int32(2), count.Load())
		})
	})
}

func TestBearerClient_CrossHost(t *testing.T) {
	t.Run("NewRequest", func(t *testing.T) {
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
	})
	t.Run("Do", func(t *testing.T) {
		t.Parallel()
		withServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, map[string]json.RawMessage{
				"next":    json.RawMessage("null"),
				"results": json.RawMessage(`[{"pub_key":"pk3"}]`),
			})
		}), func(t *testing.T, srv *httptest.Server) {
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
			require.NoError(t, err)

			var result struct {
				Next    *string         `json:"next"`
				Results json.RawMessage `json:"results"`
			}
			err = c.Do(req, &result)
			assert.NoError(t, err)
			assert.Nil(t, result.Next)
		})
	})
}

func TestProjectAPIClient_HandleThrottleBackoffExceedsMaxWait(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://api.uploadcare.com/projects/", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusTooManyRequests)
	resp := rec.Result()
	resp.Request = req

	retry, err := processResponse(resp, req, nil, &RetryConfig{MaxRetries: 10, MaxWaitSeconds: 3}, 5, mapProjectAPIError)
	assert.False(t, retry)

	var throttleErr ThrottleError
	require.True(t, errors.As(err, &throttleErr))
	assert.Equal(t, 16, throttleErr.RetryAfter)
}

