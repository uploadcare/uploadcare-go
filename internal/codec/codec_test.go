package codec_test

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/file"
	"github.com/uploadcare/uploadcare-go/v2/internal/codec"
	"github.com/uploadcare/uploadcare-go/v2/internal/config"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
	"github.com/uploadcare/uploadcare-go/v2/upload"
)

func TestEncodeReqQuery(t *testing.T) {
	t.Parallel()

	now, _ := time.Parse(config.UCTimeLayout, "2015-04-02T10:00:00")

	cases := []struct {
		test string

		params        interface{}
		expectedQuery url.Values
		wantErr       bool
	}{{
		test: "full list of params",
		params: &file.ListParams{
			Removed:      ucare.Bool(true),
			Stored:       ucare.Bool(false),
			Limit:        ucare.Uint64(500),
			OrderBy:      ucare.String(file.OrderByUploadedAtAsc),
			StartingFrom: ucare.Time(now.AddDate(0, -3, 0)),
		},
		expectedQuery: url.Values{
			"removed":  []string{"true"},
			"stored":   []string{"false"},
			"limit":    []string{"500"},
			"ordering": []string{"datetime_uploaded"},
			"from":     []string{"2015-01-02T10:00:00"},
		},
	}, {
		test:          "empty list",
		params:        &file.ListParams{},
		expectedQuery: url.Values{}}, {
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
		wantErr:       true,
	}, {
		test: "not struct pointer params type",
		params: &map[string]*bool{
			"stored": ucare.Bool(false),
		},
		expectedQuery: url.Values{},
		wantErr:       true,
	}}

	for _, c := range cases {
		c := c
		t.Run(c.test, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequest("GET", "", nil)
			err := codec.EncodeReqQuery(c.params, req)
			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
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

func TestEncodeReqBody(t *testing.T) {
	t.Parallel()

	cases := []struct {
		test string

		params           interface{}
		expectedBodyData string
	}{{
		test: "slice data",

		params:           []string{"test1", "test2"},
		expectedBodyData: "[\"test1\",\"test2\"]",
	}}
	for _, c := range cases {
		c := c
		t.Run(c.test, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequest("PUT", "", nil)
			require.NoError(t, codec.EncodeReqBody(c.params, req))

			data, err := io.ReadAll(req.Body)
			assert.Equal(t, nil, err)
			assert.Equal(t, c.expectedBodyData, string(data))
		})
	}
}

func TestEncodeReqFormData(t *testing.T) {
	t.Parallel()

	cases := []struct {
		test string

		data interface{}
		// not validating the form data request
		// just checking if stuff is there
		testReq func(written string) error
		err     bool
	}{{
		test: "simple file upload payload",
		data: &upload.FileParams{
			Data: strings.NewReader("test hello world"),
			Name: "test_file_name",
		},
		testReq: func(written string) error {
			if !strings.Contains(written, "test hello world") ||
				!strings.Contains(written, "test_file_name") {
				return errors.New("file is not written")
			}
			return nil
		},
		err: false,
	}, {
		test: "nil data",
		data: &upload.FileParams{
			Data: nil,
			Name: "test_file_name",
		},
		testReq: nil,
		err:     true,
	}, {
		test: "empty file name",
		data: &upload.FileParams{
			Data: strings.NewReader("test"),
			Name: "",
		},
		testReq: nil,
		err:     true,
	}, {
		test: "random data with form value",
		data: &struct {
			TestField string `form:"test_field"`
		}{
			TestField: "testdata",
		},
		testReq: func(written string) error {
			if !strings.Contains(written, "test_field") ||
				!strings.Contains(written, "testdata") {
				return errors.New("did not work at all")
			}
			return nil
		},
		err: false,
	}, {
		test: "random data holding map",
		data: &struct {
			M map[string]string
		}{
			M: map[string]string{"key": "testdata"},
		},
		testReq: func(written string) error {
			if !strings.Contains(written, "key") ||
				!strings.Contains(written, "testdata") {
				return errors.New("did not work at all")
			}
			return nil
		},
		err: false,
	}}

	for _, c := range cases {
		c := c
		t.Run(c.test, func(t *testing.T) {
			t.Parallel()

			body, _, err := codec.EncodeReqFormData(c.data)
			if err != nil && !c.err {
				t.Fatal(err)
			} else if err == nil && c.err {
				t.Fatal("error must be returned")
			}

			if body == nil {
				return
			}

			data, _ := io.ReadAll(body)

			assert.Equal(t, nil, err)
			assert.Equal(t, nil, c.testReq(string(data)))
		})
	}
}
