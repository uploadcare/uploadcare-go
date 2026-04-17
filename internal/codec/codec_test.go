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
		test: "include appdata",
		params: &file.ListParams{
			Include: ucare.String("appdata"),
		},
		expectedQuery: url.Values{
			"include": []string{"appdata"},
		},
	}, {
		test: "metadata bracket keys",
		params: &struct {
			Metadata map[string]string `form:"metadata"`
		}{
			Metadata: map[string]string{"key1": "val1", "key2": "val2"},
		},
		expectedQuery: url.Values{
			"metadata[key1]": []string{"val1"},
			"metadata[key2]": []string{"val2"},
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
		t.Run(c.test, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequest(http.MethodGet, "https://example.test", nil)
			err := codec.EncodeReqQuery(c.params, req)
			if c.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			q := req.URL.Query()

			if len(c.expectedQuery) == 0 && req.URL.RawQuery != "" {
				t.Error("should have no qparams")
			}
			for k, v := range c.expectedQuery {
				assert.Equal(t, v, q[k])
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
		test:             "slice data",
		params:           []string{"test1", "test2"},
		expectedBodyData: "[\"test1\",\"test2\"]",
	}}

	for _, c := range cases {
		t.Run(c.test, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequest(http.MethodPut, "https://example.test", nil)
			require.NoError(t, codec.EncodeReqBody(c.params, req))

			data, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			assert.Equal(t, c.expectedBodyData, string(data))
		})
	}
}

func TestEncodeReqFormData(t *testing.T) {
	t.Parallel()

	cases := []struct {
		test string

		data    interface{}
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
	}, {
		test: "nil data",
		data: &upload.FileParams{
			Data: nil,
			Name: "test_file_name",
		},
		err: true,
	}, {
		test: "empty file name",
		data: &upload.FileParams{
			Data: strings.NewReader("test"),
			Name: "",
		},
		err: true,
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
	}, {
		test: "metadata bracket keys",
		data: &struct {
			Metadata map[string]string `form:"metadata"`
		}{
			Metadata: map[string]string{"key1": "val1", "key2": "val2"},
		},
		testReq: func(written string) error {
			if !strings.Contains(written, `name="metadata[key1]"`) ||
				!strings.Contains(written, "val1") ||
				!strings.Contains(written, `name="metadata[key2]"`) ||
				!strings.Contains(written, "val2") {
				return errors.New("metadata bracket keys not written")
			}
			return nil
		},
	}, {
		test: "nil metadata omitted",
		data: &struct {
			Metadata map[string]string `form:"metadata"`
		}{},
		testReq: func(written string) error {
			if strings.Contains(written, "metadata[") {
				return errors.New("nil metadata should not be written")
			}
			return nil
		},
	}, {
		test: "empty metadata omitted",
		data: &struct {
			Metadata map[string]string `form:"metadata"`
		}{
			Metadata: map[string]string{},
		},
		testReq: func(written string) error {
			if strings.Contains(written, "metadata[") {
				return errors.New("empty metadata should not be written")
			}
			return nil
		},
	}}

	for _, c := range cases {
		t.Run(c.test, func(t *testing.T) {
			t.Parallel()

			body, _, err := codec.EncodeReqFormData(c.data)
			if c.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			data, err := io.ReadAll(body)
			require.NoError(t, err)
			require.NoError(t, c.testReq(string(data)))
		})
	}
}
