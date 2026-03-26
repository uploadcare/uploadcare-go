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
	// See RetryConfig for REST vs. Upload API differences in MaxWaitSeconds.
	Retry *RetryConfig
	// CDNBase is the base URL for CDN file delivery.
	// When empty (default), it is automatically derived from the public key.
	// Set this to override the automatic per-project CDN domain.
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
}

// NewClient initializes and configures new client for the high level API.
func NewClient(creds APICreds, conf *Config) (Client, error) {
	log.Infof("creating new client: %+v, %+v", creds, conf)

	if creds.SecretKey == "" || creds.PublicKey == "" {
		return nil, errors.New("uploadcare: invalid api creds provided")
	}

	conf = resolveConfig(conf, creds)

	c := client{
		backends: map[config.Endpoint]Client{
			config.RESTAPIEndpoint:   newRESTAPIClient(creds, conf),
			config.UploadAPIEndpoint: newUploadAPIClient(creds, conf),
		},
		fallbackDo: fallbackDoFunc(conf.HTTPClient),
	}

	return &c, nil
}

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

// NewBearerClient initializes a client that authenticates with a bearer token.
// Use this for the Project API, which requires token-based authentication
// instead of the pub/secret key credentials used by NewClient.
func NewBearerClient(token string, conf *Config) (Client, error) {
	if token == "" {
		return nil, errors.New("uploadcare: bearer token must not be empty")
	}

	conf = resolveBearerConfig(conf)

	pClient := newProjectAPIClient(token, conf)

	c := client{
		backends: map[config.Endpoint]Client{
			config.RESTAPIEndpoint: pClient,
		},
		// Pagination next/previous URLs may point to a different host
		// (e.g. app.uploadcare.com). Route those through the same
		// bearer-auth client so auth headers and error handling apply.
		fallbackDo: func(req *http.Request, resdata interface{}) error {
			return pClient.Do(req, resdata)
		},
	}

	return &c, nil
}

func resolveBearerConfig(conf *Config) *Config {
	if conf == nil {
		conf = &Config{}
	} else {
		copied := *conf
		conf = &copied
	}
	if conf.HTTPClient == nil {
		conf.HTTPClient = http.DefaultClient
	}
	return conf
}

func resolveConfig(conf *Config, creds APICreds) *Config {
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
	if conf.CDNBase == "" {
		conf.CDNBase = CDNBaseURL(creds.PublicKey)
	}
	return conf
}
