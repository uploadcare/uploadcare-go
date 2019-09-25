package uploadcare

import (
	"errors"
	"net/http"
)

// Client describes API client behaviour
type Client interface{}

type client struct {
	creds      APICreds
	apiVersion RESTAPIVersion

	authFunc func(APICreds, *http.Request)
	conn     *http.Client
}

// API Creds holds per project API credentials.
// You can find your credentials on the uploadcare dashboard.
type APICreds struct {
	SecretKey string
	PublicKey string
}

type optFunc func(*client)

// NewClient returns new API client with provided project credentials.
// Client is responsible for the underlying API calls.
// Opts are used for client configration.
func NewClient(creds APICreds, opts ...optFunc) (*client, error) {
	log.Infof("creating new uploadcare client with creds: %+v", creds)

	if creds.SecretKey == "" || creds.PublicKey == "" {
		return nil, errors.New("invalid creds provided")
	}

	c := client{
		creds:      creds,
		apiVersion: DefaultAPIVersion,

		authFunc: SimpleAuth,
		conn:     http.DefaultClient,
	}

	for _, o := range opts {
		o(&c)
	}

	return &c, nil
}

// WithHTTPClient is used to provide your custom http client to the Client.
// Use it if you need custom transport configuration etc.
func WithHTTPClient(conn *http.Client) optFunc {
	return func(client *client) { client.conn = conn }
}

// WithAPIVersion is used if you want to use version of the Uploadcare REST API
// different from the DefaultAPIVersion.
//
// If you're using functionality that is not supported by the selected
// version API you'll get ErrInvalidVersion.
func WithAPIVersion(version RESTAPIVersion) optFunc {
	return func(client *client) { client.apiVersion = version }
}

// WithAuthentication is used to change authentication mechanism.
//
// If you're using SignatureBasedAuth you need to enable it first in
// the Uploadcare dashboard
func WithAuthentication(authFunc authFunc) optFunc {
	return func(client *client) { client.authFunc = authFunc }
}
