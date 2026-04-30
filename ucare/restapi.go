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

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", c.acceptHeader)
	req.Header.Set(userAgentHeaderKey, c.userAgent)
	req.Header.Set("Date", date)
	c.setAuthHeader(c.creds, req)

	log.Debugf("created new request: %s %s", req.Method, req.URL)
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

		log.Debugf("making %d request: %s %s", tries, req.Method, req.URL)

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
		apiErr := APIError{StatusCode: resp.StatusCode}
		if body, _ := io.ReadAll(resp.Body); json.Unmarshal(body, &apiErr) != nil {
			apiErr.Detail = stringOrStatus(body, resp.StatusCode)
		}
		return false, apiErr
	case 401:
		authErr := AuthError{APIError: APIError{StatusCode: 401}}
		if body, _ := io.ReadAll(resp.Body); json.Unmarshal(body, &authErr) != nil {
			authErr.Detail = stringOrStatus(body, 401)
		}
		return false, authErr
	case 403:
		forbiddenErr := ForbiddenError{APIError: APIError{StatusCode: 403}}
		if body, _ := io.ReadAll(resp.Body); json.Unmarshal(body, &forbiddenErr) != nil {
			forbiddenErr.Detail = stringOrStatus(body, 403)
		}
		return false, forbiddenErr
	case 406:
		return false, ErrInvalidVersion
	case 429:
		return handleThrottle(req.Context(), resp, c.retry, tries)
	default:
		if resp.StatusCode >= 400 {
			apiErr := APIError{StatusCode: resp.StatusCode}
			if body, _ := io.ReadAll(resp.Body); json.Unmarshal(body, &apiErr) != nil || apiErr.Detail == "" {
				apiErr.Detail = stringOrStatus(body, resp.StatusCode)
			}
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
