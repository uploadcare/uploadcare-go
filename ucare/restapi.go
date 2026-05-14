package ucare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
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

	if data != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", c.acceptHeader)
	req.Header.Set(userAgentHeaderKey, c.userAgent)
	req.Header.Set("Date", date)
	c.setAuthHeader(c.creds, req)

	log.Debugf("created new request: %s %s", req.Method, req.URL)
	return req, nil
}

func (c *restAPIClient) Do(req *http.Request, resdata interface{}) error {
	return doWithRetry(c.conn, c.retry, req, resdata, mapRESTError)
}

func mapRESTError(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusBadRequest, http.StatusNotFound:
		apiErr := APIError{StatusCode: statusCode}
		if json.Unmarshal(body, &apiErr) != nil {
			apiErr.Detail = stringOrStatus(body, statusCode)
		}
		return apiErr
	case http.StatusUnauthorized:
		authErr := AuthError{APIError: APIError{StatusCode: http.StatusUnauthorized}}
		if json.Unmarshal(body, &authErr) != nil {
			authErr.Detail = stringOrStatus(body, http.StatusUnauthorized)
		}
		return authErr
	case http.StatusForbidden:
		forbiddenErr := ForbiddenError{APIError: APIError{StatusCode: http.StatusForbidden}}
		if json.Unmarshal(body, &forbiddenErr) != nil {
			forbiddenErr.Detail = stringOrStatus(body, http.StatusForbidden)
		}
		return forbiddenErr
	case http.StatusNotAcceptable:
		return ErrInvalidVersion
	default:
		apiErr := APIError{StatusCode: statusCode}
		if json.Unmarshal(body, &apiErr) != nil || apiErr.Detail == "" {
			apiErr.Detail = stringOrStatus(body, statusCode)
		}
		return apiErr
	}
}

func stringOrStatus(body []byte, statusCode int) string {
	if s := strings.TrimSpace(string(body)); s != "" {
		return s
	}
	return http.StatusText(statusCode)
}

func isNilResponseData(resdata interface{}) bool {
	if resdata == nil {
		return true
	}

	v := reflect.ValueOf(resdata)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
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
