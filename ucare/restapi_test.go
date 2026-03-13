package ucare

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
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

func TestDo_UnhandledStatusWithDetail(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"detail":"Addon is already running for this file."}`))
	}))
	defer srv.Close()

	client := &restAPIClient{conn: srv.Client()}
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/addons/uc_clamav_virus_scan/execute/", nil)
	assert.NoError(t, err)

	var result struct {
		RequestID string `json:"request_id"`
	}
	err = client.Do(req, &result)

	assert.Error(t, err)
	var apiErr respErr
	assert.True(t, errors.As(err, &apiErr))
	assert.Equal(t, "Addon is already running for this file.", apiErr.Details)
	assert.Equal(t, "", result.RequestID)
}

func TestDo_UnhandledStatusWithoutDetail(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway"))
	}))
	defer srv.Close()

	client := &restAPIClient{conn: srv.Client()}
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/files/", nil)
	assert.NoError(t, err)

	var result map[string]string
	err = client.Do(req, &result)

	assert.Error(t, err)
	var statusErr unexpectedStatusErr
	assert.True(t, errors.As(err, &statusErr))
	assert.Equal(t, http.StatusBadGateway, statusErr.StatusCode)
}

func TestDo_UnhandledStatusNilResdata(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"detail":"Conflict"}`))
	}))
	defer srv.Close()

	client := &restAPIClient{conn: srv.Client()}
	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/groups/abc~1/", nil)
	assert.NoError(t, err)

	err = client.Do(req, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Conflict")
}

func TestDo_SuccessNilResdataClosesBody(t *testing.T) {
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
	assert.NoError(t, err)

	err = client.Do(req, nil)

	assert.NoError(t, err)
	assert.True(t, closed)
}
