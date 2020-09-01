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

	"github.com/uploadcare/uploadcare-go/internal/config"
)

type restAPIClient struct {
	creds      APICreds
	apiVersion string

	userAgent     string
	acceptHeader  string
	setAuthHeader restAPIAuthFunc

	conn *http.Client
}

func newRESTAPIClient(creds APICreds, conf *Config) Client {
	c := restAPIClient{
		creds:      creds,
		apiVersion: conf.APIVersion,

		setAuthHeader: simpleRESTAPIAuth,

		conn: conf.HTTPClient,
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
	req, err := http.NewRequest(method, requrl, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
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
	tries := 0

try:
	tries++

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
	if req.Body != nil {
		defer req.Body.Close()
	}

	log.Debugf("received response: %+v", resp)

	switch resp.StatusCode {
	case 400, 404:
		var err respErr
		if e := json.NewDecoder(resp.Body).Decode(&err); e != nil {
			return e
		}
		resp.Body.Close()
		return err
	case 401:
		var err authErr
		if e := json.NewDecoder(resp.Body).Decode(&err); e != nil {
			return e
		}
		resp.Body.Close()
		return err
	case 406:
		return ErrInvalidVersion
	case 429:
		retryAfter, err := strconv.Atoi(
			resp.Header.Get("Retry-After"),
		)
		if err != nil {
			return fmt.Errorf("invalid Retry-After: %w", err)
		}

		if tries > config.MaxThrottleRetries {
			return throttleErr{retryAfter}
		}

		time.Sleep(time.Duration(retryAfter) * time.Second)
		goto try
	default:
	}

	if resdata == nil || reflect.ValueOf(resdata).IsNil() {
		return nil
	}

	err = json.NewDecoder(resp.Body).Decode(&resdata)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
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
