package uctest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"

	"github.com/uploadcare/uploadcare-go/v2/internal/config"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

type Client struct {
	HTTP    *http.Client
	BaseURL string
}

func NewServerClient(srv *httptest.Server) *Client {
	return &Client{
		HTTP:    srv.Client(),
		BaseURL: srv.URL,
	}
}

func (c *Client) NewRequest(
	ctx context.Context,
	_ config.Endpoint,
	method, requrl string,
	data ucare.ReqEncoder,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+requrl, nil)
	if err != nil {
		return nil, err
	}
	if data != nil {
		if err = data.EncodeReq(req); err != nil {
			return nil, err
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", fmt.Sprintf(config.AcceptHeaderFormat, ucare.APIv07))
	return req, nil
}

func (c *Client) Do(req *http.Request, resdata interface{}) error {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}

	if resdata == nil || reflect.ValueOf(resdata).IsNil() {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(resdata)
}
