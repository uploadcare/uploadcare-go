package ucare

import (
	"context"
	"encoding/json"
	"fmt"
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

	if data != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set(userAgentHeaderKey, c.userAgent)
	req.Header.Set(authHeaderKey, "Bearer "+c.token)

	log.Debugf("created new project api request: %s %s", req.Method, req.URL)
	return req, nil
}

func (c *projectAPIClient) Do(req *http.Request, resdata interface{}) error {
	return doWithRetry(c.conn, c.retry, req, resdata, mapProjectAPIError)
}

func mapProjectAPIError(statusCode int, body []byte) error {
	apiErr := ProjectAPIError{StatusCode: statusCode}
	if json.Unmarshal(body, &apiErr) != nil || apiErr.Message == "" {
		apiErr.Message = stringOrStatus(body, statusCode)
	}
	switch statusCode {
	case http.StatusUnauthorized:
		return ProjectAuthError{ProjectAPIError: apiErr}
	case http.StatusForbidden:
		return ProjectForbiddenError{ProjectAPIError: apiErr}
	default:
		return apiErr
	}
}
