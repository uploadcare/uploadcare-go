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
)

type testReqEncoder struct {
	body  string
	query string
}

func (t testReqEncoder) EncodeReq(r *http.Request) {
	r.URL.RawQuery = t.query
	r.Body = ioutil.NopCloser(strings.NewReader(t.body))
}

func TestClientNewRequest(t *testing.T) {
	t.Parallel()

	creds := APICreds{
		SecretKey: "testsecretkey",
		PublicKey: "testpublickey",
	}

	testurl := "http:/test.com/api/"
	ci, err := NewClient(creds)
	if err != nil {
		t.Fatal(err)
	}
	client, ok := ci.(*client)
	if !ok {
		t.Fatal("client had invalid underlying type")
	}
	if client.acceptHeader != "application/vnd.uploadcare-v0.5+json" {
		t.Fatal("accept header is wrong")
	}
	if client.userAgent != "UploadcareGo/0.1.0/testpublickey" {
		t.Fatal("user agent header is wrong")
	}

	cases := []struct {
		test string

		method   string
		fullpath string
		data     ReqEncoder

		checkReq func(*http.Request) error
	}{{
		test:     "simple case with no data",
		method:   http.MethodGet,
		fullpath: testurl,
		data:     nil,
		checkReq: func(r *http.Request) error {
			h := r.Header
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
		test:     "simple case with data",
		method:   http.MethodGet,
		fullpath: testurl,
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
				c.method,
				c.fullpath,
				c.data,
			)
			assert.Equal(t, nil, err)
			assert.Equal(t, nil, c.checkReq(req))
		})
	}

}
