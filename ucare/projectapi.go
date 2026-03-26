package ucare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

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

	log.Debugf("created new project api request: %+v", req)
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

		log.Debugf("making %d project api request: %+v", tries, req)

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
	defer resp.Body.Close()

	log.Debugf("received project api response: %+v", resp)

	if resp.StatusCode == 429 {
		retryAfter, err := strconv.Atoi(
			resp.Header.Get("Retry-After"),
		)
		if err != nil {
			retryAfter = 0
		}
		if c.retry == nil || tries > c.retry.MaxRetries {
			return false, ThrottleError{RetryAfter: retryAfter}
		}
		if c.retry.MaxWaitSeconds > 0 &&
			retryAfter > c.retry.MaxWaitSeconds {
			return false, ThrottleError{RetryAfter: retryAfter}
		}
		wait := retryAfter
		if wait <= 0 {
			wait = expBackoff(tries)
		}
		select {
		case <-req.Context().Done():
			return false, req.Context().Err()
		case <-time.After(time.Duration(wait) * time.Second):
		}
		return true, nil
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		var apiErr ProjectAPIError
		apiErr.StatusCode = resp.StatusCode
		if json.Unmarshal(body, &apiErr) != nil || apiErr.Message == "" {
			apiErr.Message = http.StatusText(resp.StatusCode)
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
