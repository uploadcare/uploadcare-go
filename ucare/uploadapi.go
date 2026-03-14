package ucare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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
	for tries := 1; ; tries++ {
		if tries > 1 && req.GetBody != nil {
			var err error
			req.Body, err = req.GetBody()
			if err != nil {
				return err
			}
		}

		log.Debugf("making %d request: %s %+v", tries, req.Method, req.URL)

		resp, err := c.conn.Do(req)
		if err != nil {
			return err
		}

		retry, err := c.handleResponse(resp, resdata, tries)
		if err != nil || !retry {
			return err
		}
	}
}

func (c *uploadAPIClient) handleResponse(
	resp *http.Response,
	resdata interface{},
	tries int,
) (bool, error) {
	defer resp.Body.Close()

	log.Debugf("received response: %+v", resp)

	switch resp.StatusCode {
	case 400:
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		return false, ValidationError{APIError{StatusCode: 400, Detail: string(data)}}
	case 403:
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		return false, ForbiddenError{APIError{StatusCode: 403, Detail: string(data)}}
	case 413:
		return false, ErrFileTooLarge
	case 429:
		if c.retry == nil || tries > c.retry.MaxRetries {
			return false, ThrottleError{}
		}
		wait := expBackoff(tries)
		if c.retry.MaxWaitSeconds > 0 && wait > c.retry.MaxWaitSeconds {
			wait = c.retry.MaxWaitSeconds
		}
		select {
		case <-resp.Request.Context().Done():
			return false, resp.Request.Context().Err()
		case <-time.After(time.Duration(wait) * time.Second):
		}
		return true, nil
	default:
		if resp.StatusCode >= 400 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return false, err
			}
			var apiErr APIError
			if json.Unmarshal(body, &apiErr) != nil || apiErr.Detail == "" {
				detail := string(body)
				if detail == "" {
					detail = http.StatusText(resp.StatusCode)
				}
				apiErr.Detail = detail
			}
			apiErr.StatusCode = resp.StatusCode
			return false, apiErr
		}
	}

	if isNilResponseData(resdata) {
		return false, nil
	}

	if err := json.NewDecoder(resp.Body).Decode(resdata); err != nil {
		return false, err
	}

	return false, nil
}
