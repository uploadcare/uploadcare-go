package ucare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/uploadcare/uploadcare-go/v2/internal/config"
)

type restAPIClient struct {
	creds      APICreds
	apiVersion string

	userAgent     string
	acceptHeader  string
	setAuthHeader restAPIAuthFunc

	conn  *http.Client
	retry *RetryConfig
}

func newRESTAPIClient(creds APICreds, conf *Config) Client {
	c := restAPIClient{
		creds:      creds,
		apiVersion: conf.APIVersion,

		setAuthHeader: simpleRESTAPIAuth,

		conn:  conf.HTTPClient,
		retry: conf.Retry,
	}

	if conf.SignBasedAuthentication {
		c.setAuthHeader = signBasedRESTAPIAuth
	}

	c.acceptHeader = fmt.Sprintf(config.AcceptHeaderFormat, c.apiVersion)
	c.userAgent = fmt.Sprintf(
		"%s/%s/%s",
		config.UserAgentPrefix,
		config.ClientVersion,
		creds.PublicKey,
	)
	if conf.UserAgent != "" {
		c.userAgent += " " + conf.UserAgent
	}

	return &c
}

func getBodyBuilder(req *http.Request, data ReqEncoder) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		err := data.EncodeReq(req)
		return req.Body, err
	}
}

func (c *restAPIClient) NewRequest(
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

	date := time.Now().In(dateHeaderLocation).Format(dateHeaderFormat)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", c.acceptHeader)
	req.Header.Set(userAgentHeaderKey, c.userAgent)
	req.Header.Set("Date", date)
	c.setAuthHeader(c.creds, req)

	log.Debugf("created new request: %+v", req)
	return req, nil
}

func (c *restAPIClient) Do(req *http.Request, resdata interface{}) error {
	for tries := 1; ; tries++ {
		if tries > 1 && req.GetBody != nil {
			var err error
			req.Body, err = req.GetBody()
			if err != nil {
				return err
			}
		}

		log.Debugf("making %d request: %+v", tries, req)

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

func (c *restAPIClient) handleResponse(
	resp *http.Response,
	req *http.Request,
	resdata interface{},
	tries int,
) (bool, error) {
	defer func() { _ = resp.Body.Close() }()

	log.Debugf("received response: %+v", resp)

	switch resp.StatusCode {
	case 400, 404:
		var apiErr APIError
		if e := json.NewDecoder(resp.Body).Decode(&apiErr); e != nil {
			return false, e
		}
		apiErr.StatusCode = resp.StatusCode
		return false, apiErr
	case 401:
		var authErr AuthError
		if e := json.NewDecoder(resp.Body).Decode(&authErr); e != nil {
			return false, e
		}
		authErr.StatusCode = 401
		return false, authErr
	case 403:
		var forbiddenErr ForbiddenError
		if e := json.NewDecoder(resp.Body).Decode(&forbiddenErr); e != nil {
			return false, e
		}
		forbiddenErr.StatusCode = 403
		return false, forbiddenErr
	case 406:
		return false, ErrInvalidVersion
	case 429:
		retryAfter, err := strconv.Atoi(
			resp.Header.Get("Retry-After"),
		)
		if err != nil || retryAfter < 0 {
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
			// Without a usable Retry-After from the server, REST retries fall
			// back to exponential backoff instead of applying MaxWaitSeconds.
			wait = expBackoff(tries)
		}
		select {
		case <-req.Context().Done():
			return false, req.Context().Err()
		case <-time.After(time.Duration(wait) * time.Second):
		}
		return true, nil
	default:
		if resp.StatusCode >= 400 {
			var apiErr APIError
			if e := json.NewDecoder(resp.Body).Decode(&apiErr); e != nil || apiErr.Detail == "" {
				apiErr.Detail = http.StatusText(resp.StatusCode)
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

func isNilResponseData(resdata interface{}) bool {
	if resdata == nil {
		return true
	}

	v := reflect.ValueOf(resdata)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func resolveReqURL(endpoint config.Endpoint, requrl string) (string, error) {
	u, err := url.Parse(requrl)
	if err != nil {
		return "", err
	}
	base, err := url.Parse("https://" + string(endpoint))
	if err != nil {
		return "", err
	}
	return base.ResolveReference(u).String(), nil
}
