package ucare

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	assert "github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/internal/config"
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
