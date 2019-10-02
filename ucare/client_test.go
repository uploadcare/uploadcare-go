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

func TestClientNewRequest(t *testing.T) {
	t.Parallel()

	creds := APICreds{
		SecretKey: "testsecretkey",
		PublicKey: "testpublickey",
	}

	client, err := NewClient(creds, nil)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		test string

		endpoint config.Endpoint
		method   string
		requrl   string
		data     ReqEncoder

		checkReq func(*http.Request) error
	}{{
		test:     "rest api",
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
	}, {
		test:     "upload api get",
		endpoint: config.UploadAPIEndpoint,
		method:   http.MethodGet,
		requrl:   "/base/",
		data: testReqEncoder{
			body:  "formkey=formvalue",
			query: "qparam1=qparamvalue1&qparam2=qparamvalue2",
		},
		checkReq: func(r *http.Request) error {
			// check only data in this test case
			data, _ := ioutil.ReadAll(r.Body)
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
