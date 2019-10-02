package ucare

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/internal/config"
)

type testReqEncoder struct {
	body  string
	query string
}

func (t testReqEncoder) EncodeReq(r *http.Request) error {
	r.URL.RawQuery = t.query
	r.Body = ioutil.NopCloser(strings.NewReader(t.body))
	return nil
}

func testCreds() APICreds {
	return APICreds{
		SecretKey: "testsecretkey",
		PublicKey: "testpublickey",
	}
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
				"application/vnd.uploadcare-v0.5+json" {
				return errors.New("wrong accept header")
			}
			if h.Get("X-UC-User-Agent") !=
				"UploadcareGo/0.1.0/testpublickey" {
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
