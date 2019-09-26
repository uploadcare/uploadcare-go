package uploadcare

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	RESTAPIEndpoint   = "https://api.uploadcare.com/"
	UploadAPIEndpoint = "https://upload.uploadcare.com/"
)

// Client describes API client behaviour
type Client interface {
	NewRequest(method, url string, data RequestEncoder) (*http.Request, error)
	Do(req *http.Request, data RespBodyDecoder) error
}

type client struct {
	creds      APICreds
	apiVersion RESTAPIVersion

	userAgent     string
	acceptHeader  string
	setAuthHeader func(APICreds, *http.Request)

	conn *http.Client
}

// API Creds holds per project API credentials.
// You can find your credentials on the uploadcare dashboard.
type APICreds struct {
	SecretKey string
	PublicKey string
}

type optFunc func(*client) error

var (
	ErrInvalidCreds = errors.New("Invalid API credentials")

	ErrAuthForbidden = errors.New("Simple authentication over HTTP is " +
		"forbidden. Please, use HTTPS or signed requests instead")
	ErrInvalidAuthCreds = errors.New("Incorrect authentication credentials")
	ErrInvalidVersion   = errors.New("Could not satisfy the request Accept header")
)

// NewClient returns new API client with provided project credentials.
// Client is responsible for the underlying API calls.
// Opts are used for client configration.
func NewClient(creds APICreds, opts ...optFunc) (*client, error) {
	log.Infof("creating new uploadcare client with creds: %+v", creds)

	if creds.SecretKey == "" || creds.PublicKey == "" {
		return nil, ErrInvalidCreds
	}

	c := client{
		creds:      creds,
		apiVersion: DefaultAPIVersion,

		setAuthHeader: SimpleAuth,

		conn: http.DefaultClient,
	}

	for _, o := range opts {
		err := o(&c)
		if err != nil {
			return nil, err
		}
	}

	c.acceptHeader = fmt.Sprintf(acceptHeaderFormat, c.apiVersion)
	c.userAgent = fmt.Sprintf(
		"%s/%s/%s",
		userAgentPrefix,
		clientVersion,
		creds.PublicKey,
	)

	return &c, nil
}

func (c *client) NewRequest(
	method string,
	fullpath string,
	data RequestEncoder,
) (*http.Request, error) {
	req, err := http.NewRequest(method, fullpath, nil)
	if err != nil {
		return nil, err
	}

	data.EncodeRequest(req)

	date := time.Now().In(dateHeaderLocation).Format(dateHeaderFormat)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", c.acceptHeader)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Date", date)

	c.setAuthHeader(c.creds, req)

	log.Debugf("created new request: %+v", req)
	return req, nil
}

func (c *client) Do(req *http.Request, data RespBodyDecoder) error {
	tries := 0
try:
	tries += 1
	resp, err := c.conn.Do(req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 400:
		return ErrAuthForbidden
	case 401:
		return ErrInvalidAuthCreds
	case 406:
		return ErrInvalidVersion
	case 429:
		retryAfter, err := strconv.Atoi(
			resp.Header.Get("Retry-After"),
		)
		if err != nil {
			return fmt.Errorf("parse Retry-After: %w", err)
		}

		if tries > maxThrottleRetries {
			return ThrottleErr{retryAfter}
		}

		time.Sleep(time.Duration(retryAfter) * time.Second)
		goto try
	}

	err = data.DecodeRespBody(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// WithHTTPClient is used to provide your custom http client to the Client.
// Use it if you need custom transport configuration etc.
func WithHTTPClient(conn *http.Client) optFunc {
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
func WithAPIVersion(version RESTAPIVersion) optFunc {
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
func WithAuthentication(authFunc authFunc) optFunc {
	return func(client *client) (err error) {
		if authFunc == nil {
			err = errors.New("nil auth function provided")
		}
		client.setAuthHeader = authFunc
		return
	}
}

func SingleSlashJoin(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
