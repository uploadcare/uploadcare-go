package ucare

import (
	"context"
	"errors"
	"net/http"

	"github.com/uploadcare/uploadcare-go/v2/internal/config"
)

// Client describes API client behaviour
type Client interface {
	NewRequest(
		ctx context.Context,
		endpoint config.Endpoint,
		method string,
		requrl string,
		data ReqEncoder,
	) (*http.Request, error)
	Do(req *http.Request, resdata interface{}) error
}

// APICreds holds per project API credentials.
// You can find your credentials on the uploadcare dashboard.
type APICreds struct {
	SecretKey string
	PublicKey string
}

// Config holds configuration for the client
type Config struct {
	// HTTPClient allows you to set custom http client for the calls
	HTTPClient *http.Client
	// APIVersion specifies REST API version to be used
	APIVersion string
	// SignBasedAuthentication should be true if you want to use
	// signed uploads and signature based authentication for the
	// REST API calls.
	SignBasedAuthentication bool
	// UserAgent is appended to the default User-Agent string.
	// Use this to identify your application (e.g. "my-app/1.0.0").
	UserAgent string
	// Retry controls automatic retry of throttled (HTTP 429) requests.
	// When nil (the default), throttled requests fail immediately.
	// See RetryConfig for details on MaxRetries and MaxWaitSeconds.
	Retry *RetryConfig
	// CDNBase is the base URL for CDN file delivery.
	// When empty (default), it is automatically derived from the public key.
	// Set this to an absolute http(s) URL to override the automatic
	// per-project CDN domain.
	CDNBase string
}

// ReqEncoder exists to encode data into the prepared request.
// It may encode part of the data to the query string and other
// part into the request body. It may also set request headers for some
// payload types (multipart/form-data).
type ReqEncoder interface {
	EncodeReq(*http.Request) error
}

type client struct {
	backends   map[config.Endpoint]Client
	fallbackDo func(*http.Request, interface{}) error
	cdnBase    string
}

// NewClient initializes and configures new client for the high level API.
func NewClient(creds APICreds, conf *Config) (Client, error) {
	log.Infof("creating new client: %+v, %+v", creds, conf)

	if creds.SecretKey == "" || creds.PublicKey == "" {
		return nil, errors.New("uploadcare: invalid api creds provided")
	}

	conf, err := resolveConfig(conf, creds.PublicKey)
	if err != nil {
		return nil, err
	}

	c := client{
		backends: map[config.Endpoint]Client{
			config.RESTAPIEndpoint:   newRESTAPIClient(creds, conf),
			config.UploadAPIEndpoint: newUploadAPIClient(creds, conf),
		},
		fallbackDo: fallbackDoFunc(conf.HTTPClient),
		cdnBase:    conf.CDNBase,
	}

	return &c, nil
}

// CDNBase returns the CDN base URL resolved by NewClient. It is either the
// explicit Config.CDNBase (normalised) or the per-project URL derived from
// the public key. Services read this to rewrite API-returned CDN URLs, which
// otherwise always point at the legacy ucarecdn.com domain regardless of the
// project's configured CDN.
func (c *client) CDNBase() string { return c.cdnBase }

var errNoClient = errors.New("no client for such endpoint")

// NewRequests constructs new http request.
func (c *client) NewRequest(
	ctx context.Context,
	endpoint config.Endpoint,
	method string,
	requrl string,
	data ReqEncoder,
) (*http.Request, error) {
	b, ok := c.backends[endpoint]
	if !ok {
		return nil, errNoClient
	}
	return b.NewRequest(ctx, endpoint, method, requrl, data)
}

// Do performs the actual backend API call.
func (c *client) Do(req *http.Request, resdata interface{}) error {
	b, ok := c.backends[config.Endpoint(req.URL.Host)]
	if !ok {
		return c.fallbackDo(req, resdata)
	}
	return b.Do(req, resdata)
}

func resolveConfig(conf *Config, publicKey string) (*Config, error) {
	if conf == nil {
		conf = &Config{}
	} else {
		copied := *conf
		conf = &copied
	}
	if conf.APIVersion == "" {
		conf.APIVersion = defaultAPIVersion
	}
	if conf.HTTPClient == nil {
		conf.HTTPClient = http.DefaultClient
	}

	var err error
	conf.CDNBase, err = resolveCDNBase(conf.CDNBase, publicKey)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
