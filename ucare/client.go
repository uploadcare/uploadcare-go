package ucare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/uploadcare/uploadcare-go/internal/config"
)

// Client describes API client behaviour
type Client interface {
	NewRequest(
		ctx context.Context,
		method string,
		url string,
		data ReqEncoder,
	) (*http.Request, error)
	Do(req *http.Request, resdata interface{}) error
}

type client struct {
	creds      APICreds
	apiVersion string

	userAgent     string
	acceptHeader  string
	setAuthHeader func(APICreds, *http.Request)

	conn *http.Client
}

// APICreds holds per project API credentials.
// You can find your credentials on the uploadcare dashboard.
type APICreds struct {
	SecretKey string
	PublicKey string
}

// OptFunc is a function that does some configuration on the passed client
type OptFunc func(*client) error

// NewClient returns new API client with provided project credentials.
// Client is responsible for the underlying API calls.
// Opts are used for client configration.
func NewClient(creds APICreds, opts ...OptFunc) (Client, error) {
	log.Infof("creating new uploadcare client with creds: %+v", creds)

	if creds.SecretKey == "" || creds.PublicKey == "" {
		return nil, ErrInvalidAuthCreds
	}

	c := client{
		creds:      creds,
		apiVersion: defaultAPIVersion,

		setAuthHeader: SimpleAuth,

		conn: http.DefaultClient,
	}

	for _, o := range opts {
		err := o(&c)
		if err != nil {
			return nil, err
		}
	}

	c.acceptHeader = fmt.Sprintf(config.AcceptHeaderFormat, c.apiVersion)
	c.userAgent = fmt.Sprintf(
		"%s/%s/%s",
		config.UserAgentPrefix,
		config.ClientVersion,
		creds.PublicKey,
	)

	return &c, nil
}

// ReqEncoder exists to encode data into prepared request.
// It may encode part of the data to the query string and other
// part into the request body
type ReqEncoder interface {
	EncodeReq(*http.Request)
}

func (c *client) NewRequest(
	ctx context.Context,
	method string,
	fullpath string,
	data ReqEncoder,
) (*http.Request, error) {
	req, err := http.NewRequest(method, fullpath, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	if data != nil {
		data.EncodeReq(req)
	}

	date := time.Now().In(dateHeaderLocation).Format(dateHeaderFormat)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", c.acceptHeader)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Date", date)

	c.setAuthHeader(c.creds, req)

	log.Debugf("created new request: %+v", req)
	return req, nil
}

func (c *client) Do(req *http.Request, resdata interface{}) error {
	tries := 0
try:
	tries++

	log.Debugf("making %d request: %+v", tries, req)

	resp, err := c.conn.Do(req)
	if err != nil {
		return err
	}

	log.Debugf("received response: %+v", resp)

	switch resp.StatusCode {
	case 400:
		return ErrAuthForbidden
	case 401:
		var err authErr
		if e := json.NewDecoder(resp.Body).Decode(&err); e != nil {
			return e
		}
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
	}

	err = json.NewDecoder(resp.Body).Decode(&resdata)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// WithHTTPClient is used to provide your custom http client to the Client.
// Use it if you need custom transport configuration etc.
func WithHTTPClient(conn *http.Client) OptFunc {
	return func(client *client) (err error) {
		if conn == nil {
			err = errors.New("nil http client provided")
		}
		client.conn = conn
		return
	}
}

// WithAPIVersion is used if you want to use version of the Uploadcare REST API
// different from the DefaultAPIVersion.
//
// If you're using functionality that is not supported by the selected
// version API you'll get ErrInvalidVersion.
func WithAPIVersion(version string) OptFunc {
	return func(client *client) (err error) {
		if _, ok := supportedVersions[version]; !ok {
			err = errors.New("unsupported API version provided")
		}
		client.apiVersion = version
		return
	}

}

// WithAuthentication is used to change authentication mechanism.
//
// If you're using SignatureBasedAuth you need to enable it first in
// the Uploadcare dashboard
func WithAuthentication(authFunc authFunc) OptFunc {
	return func(client *client) (err error) {
		if authFunc == nil {
			err = errors.New("nil auth function provided")
		}
		client.setAuthHeader = authFunc
		return
	}
}
