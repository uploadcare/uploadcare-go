package ucare

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errReadCloser struct{ err error }

func (e errReadCloser) Read([]byte) (int, error) { return 0, e.err }
func (e errReadCloser) Close() error             { return nil }

func restAPIClientStubTransport(rt http.RoundTripper) *restAPIClient {
	return &restAPIClient{
		conn: &http.Client{Transport: rt},
	}
}

func TestProcessResponse_BodyReadErrorJoinedWithAPIError(t *testing.T) {
	t.Parallel()

	readErr := errors.New("simulated read failure")
	client := restAPIClientStubTransport(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Header:     make(http.Header),
			Body:       errReadCloser{err: readErr},
		}, nil
	}))
	req, err := http.NewRequest(http.MethodGet, "https://example.test/files/", nil)
	require.NoError(t, err)

	err = client.Do(req, nil)
	require.Error(t, err)

	var apiErr APIError
	require.True(t, errors.As(err, &apiErr), "typed API error must be preserved")
	assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)

	require.True(t, errors.Is(err, readErr), "read error must be reachable via errors.Is")
	assert.Contains(t, err.Error(), "simulated read failure")
}

type drainTrackedBody struct {
	r         io.Reader
	readToEnd *bool
	closed    *bool
}

func (d *drainTrackedBody) Read(p []byte) (int, error) {
	n, err := d.r.Read(p)
	if errors.Is(err, io.EOF) {
		*d.readToEnd = true
	}
	return n, err
}

func (d *drainTrackedBody) Close() error {
	*d.closed = true
	return nil
}

func TestProcessResponse_DrainsBody(t *testing.T) {
	t.Run("after_decode", func(t *testing.T) {
		t.Parallel()

		readToEnd, closed := false, false
		body := strings.NewReader(`{"ok":true}` + strings.Repeat("x", 256))

		client := restAPIClientStubTransport(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       &drainTrackedBody{r: body, readToEnd: &readToEnd, closed: &closed},
			}, nil
		}))
		req, err := http.NewRequest(http.MethodGet, "https://example.test/files/", nil)
		require.NoError(t, err)

		var result map[string]bool
		require.NoError(t, client.Do(req, &result))
		assert.True(t, result["ok"])
		assert.True(t, readToEnd, "body must be drained to EOF for connection reuse")
		assert.True(t, closed, "body must be closed")
	})
	t.Run("on_nil_resdata", func(t *testing.T) {
		t.Parallel()

		readToEnd, closed := false, false
		body := strings.NewReader(`{"ignored":"payload"}`)

		client := restAPIClientStubTransport(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Header:     make(http.Header),
				Body:       &drainTrackedBody{r: body, readToEnd: &readToEnd, closed: &closed},
			}, nil
		}))
		req, err := http.NewRequest(http.MethodDelete, "https://example.test/files/abc/", nil)
		require.NoError(t, err)

		require.NoError(t, client.Do(req, nil))
		assert.True(t, readToEnd, "body must be drained even when caller passes nil resdata")
		assert.True(t, closed)
	})
}
