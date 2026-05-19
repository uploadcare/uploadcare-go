package ucare

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/uploadcare/uploadcare-go/v2/internal/config"
)

type uploadAPIClient struct {
	authFunc UploadAPIAuthFunc

	conn  *http.Client
	retry *RetryConfig
}

func newUploadAPIClient(creds APICreds, conf *Config) Client {
	c := uploadAPIClient{
		authFunc: simpleUploadAPIAuthFunc(creds),
		conn:     conf.HTTPClient,
		retry:    conf.Retry,
	}

	if conf.SignBasedAuthentication {
		c.authFunc = signBasedUploadAPIAuthFunc(creds)
	}

	return &c
}

func (c *uploadAPIClient) NewRequest(
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
	ctx = context.WithValue(ctx, config.CtxAuthFuncKey, c.authFunc)
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

	log.Debugf(
		"created new request: %s %+v %+v",
		req.Method,
		req.URL,
		req.Header,
	)

	return req, nil
}

func (c *uploadAPIClient) Do(
	req *http.Request,
	resdata interface{},
) error {
	return doWithRetry(c.conn, c.retry, req, resdata, mapUploadError)
}

func mapUploadError(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusBadRequest:
		return ValidationError{APIError{StatusCode: http.StatusBadRequest, Detail: string(body)}}
	case http.StatusForbidden:
		return ForbiddenError{APIError{StatusCode: http.StatusForbidden, Detail: string(body)}}
	case http.StatusRequestEntityTooLarge:
		return ErrFileTooLarge
	default:
		apiErr := APIError{StatusCode: statusCode}
		if json.Unmarshal(body, &apiErr) != nil || apiErr.Detail == "" {
			detail := string(body)
			if detail == "" {
				detail = http.StatusText(statusCode)
			}
			apiErr.Detail = detail
		}
		return apiErr
	}
}
