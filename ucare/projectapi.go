package ucare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/uploadcare/uploadcare-go/v2/internal/config"
)

type projectAPIClient struct {
	token     string
	userAgent string
	conn      *http.Client
	retry     *RetryConfig
}

func newProjectAPIClient(token string, conf *Config) Client {
	c := projectAPIClient{
		token: token,
		conn:  conf.HTTPClient,
		retry: conf.Retry,
	}

	c.userAgent = fmt.Sprintf(
		"%s/%s",
		config.UserAgentPrefix,
		config.ClientVersion,
	)
	if conf.UserAgent != "" {
		c.userAgent += " " + conf.UserAgent
	}

	return &c
}

func (c *projectAPIClient) NewRequest(
	ctx context.Context,
	endpoint config.Endpoint,
	method string,
	requrl string,
	data ReqEncoder,
) (*http.Request, error) {
	requrl, err := resolveReqURL(endpoint, requrl)
	if err != nil {
		return nil, fmt.Errorf("resolving req url: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, requrl, nil)
	if err != nil {
		return nil, err
	}
	if data != nil {
		req.GetBody = getBodyBuilder(req, data)
		req.Body, err = req.GetBody()
		if err != nil {
			return nil, err
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(userAgentHeaderKey, c.userAgent)
	req.Header.Set(authHeaderKey, "Bearer "+c.token)

	log.Debugf("created new project api request: %s %s", req.Method, req.URL)
	return req, nil
}

func (c *projectAPIClient) Do(req *http.Request, resdata interface{}) error {
	for tries := 1; ; tries++ {
		if tries > 1 && req.GetBody != nil {
			var err error
			req.Body, err = req.GetBody()
			if err != nil {
				return err
			}
		}

		log.Debugf("making %d project api request: %s %s", tries, req.Method, req.URL)

		resp, err := c.conn.Do(req)
		if err != nil {
			return err
		}

		retry, err := c.handleResponse(resp, req, resdata, tries)
		if err != nil || !retry {
			return err
		}
	}
}

func (c *projectAPIClient) handleResponse(
	resp *http.Response,
	req *http.Request,
	resdata interface{},
	tries int,
) (bool, error) {
	defer func() { _ = resp.Body.Close() }()

	log.Debugf("received project api response: %+v", resp)

	if resp.StatusCode == 429 {
		return handleThrottle(req.Context(), resp, c.retry, tries)
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		var apiErr ProjectAPIError
		apiErr.StatusCode = resp.StatusCode
		if json.Unmarshal(body, &apiErr) != nil || apiErr.Message == "" {
			apiErr.Message = stringOrStatus(body, resp.StatusCode)
		}
		return false, apiErr
	}

	if isNilResponseData(resdata) {
		return false, nil
	}

	if err := json.NewDecoder(resp.Body).Decode(resdata); err != nil {
		return false, err
	}

	return false, nil
}
