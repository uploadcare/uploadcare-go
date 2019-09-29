package codec_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uploadcare/uploadcare-go/file"
	"github.com/uploadcare/uploadcare-go/internal/codec"
	"github.com/uploadcare/uploadcare-go/internal/config"
	"github.com/uploadcare/uploadcare-go/ucare"
)

func TestEncodeReqQuery(t *testing.T) {
	t.Parallel()

	now, _ := time.Parse(config.UCTimeLayout, "2015-04-02T10:00:00")

	cases := []struct {
		test string

		params        interface{}
		expectedQuery url.Values
	}{{
		test: "full list of params",
		params: &file.ListParams{
			Removed:      ucare.Bool(true),
			Stored:       ucare.Bool(false),
			Limit:        ucare.Uint64(500),
			OrderBy:      ucare.String(file.OrderBySizeAsc),
			StartingFrom: ucare.Time(now.AddDate(0, -3, 0)),
		},
		expectedQuery: url.Values{
			"removed":  []string{"true"},
			"stored":   []string{"false"},
			"limit":    []string{"500"},
			"ordering": []string{"size"},
			"from":     []string{"2015-01-02T10:00:00"},
		},
	}, {
		test:          "empty list",
		params:        &file.ListParams{},
		expectedQuery: url.Values{},
	}, {
		test: "part of the params are filled",
		params: &file.ListParams{
			Stored:  ucare.Bool(false),
			OrderBy: ucare.String(file.OrderByUploadedAtDesc),
		},
		expectedQuery: url.Values{
			"stored":   []string{"false"},
			"ordering": []string{"-datetime_uploaded"},
		},
	}, {
		test: "not struct pointer params type",
		params: map[string]*bool{
			"stored": ucare.Bool(false),
		},
		expectedQuery: url.Values{},
	}, {
		test: "not struct pointer params type",
		params: &map[string]*bool{
			"stored": ucare.Bool(false),
		},
		expectedQuery: url.Values{},
	}}

	for _, c := range cases {
		c := c
		t.Run(c.test, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequest("GET", "", nil)
			codec.EncodeReqQuery(c.params, req)
			q := req.URL.Query()

			if len(c.expectedQuery) == 0 && req.URL.RawQuery != "" {
				t.Error("should have no qparams")
			}
			for k, v := range c.expectedQuery {
				assert.Equal(t, q[k], v)
			}
		})
	}
}
