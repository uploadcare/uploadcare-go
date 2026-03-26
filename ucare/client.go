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

// ReqEncoder exists to encode data into the prepared request.
// It may encode part of the data to the query string and other
// part into the request body. It may also set request headers for some
// payload types (multipart/form-data).
type ReqEncoder interface {
	EncodeReq(*http.Request) error
}

type client struct {
	backends       map[config.Endpoint]Client
	fallbackDo     func(*http.Request, interface{}) error
	fallbackNewReq func(context.Context, config.Endpoint, string, string, ReqEncoder) (*http.Request, error)
	cdnBase        string
}

// NewClient initializes and configures new client for the high level API.
func NewClient(creds APICreds, conf *Config) (Client, error) {
	log.Infof("creating new client: %+v, %+v", creds, conf)

	if creds.SecretKey == "" || creds.PublicKey == "" {
		return nil, errors.New("uploadcare: invalid api creds provided")
	}
	if conf == nil {
		return nil, errors.New("uploadcare: config required, build via NewConfig")
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
		if c.fallbackNewReq != nil {
			return c.fallbackNewReq(ctx, endpoint, method, requrl, data)
		}
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
	if conf == nil {
		return nil, errors.New("uploadcare: config required, build via NewBearerConfig")
	}
	conf = resolveBearerConfig(conf)

	pClient := newProjectAPIClient(token, conf)

	// Pagination next/previous URLs may point to a different host
	// (e.g. app.uploadcare.com). Route those through the same
	// bearer-auth client so auth headers and error handling apply.
	c := client{
		backends: map[config.Endpoint]Client{
			config.RESTAPIEndpoint: pClient,
		},
		fallbackNewReq: func(ctx context.Context, endpoint config.Endpoint, method, requrl string, data ReqEncoder) (*http.Request, error) {
			return pClient.NewRequest(ctx, endpoint, method, requrl, data)
		},
		fallbackDo: func(req *http.Request, resdata interface{}) error {
			return pClient.Do(req, resdata)
		},
	}

	return &c, nil
}
