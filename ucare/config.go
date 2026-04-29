package ucare

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	APIv07 = "v0.7"

	simpleAuthScheme    = "Uploadcare.Simple"
	signBasedAuthScheme = "Uploadcare"
	dateHeaderFormat    = time.RFC1123

	signedUploadTTL = 60 * time.Second

	defaultCDNDomain     = "ucarecd.net"
	cdnCNAMEPrefixLength = 10
)

var (
	defaultAPIVersion = APIv07

	authHeaderKey      = http.CanonicalHeaderKey("Authorization")
	userAgentHeaderKey = http.CanonicalHeaderKey("User-Agent")

	dateHeaderLocation = time.FixedZone("GMT", 0)

	errInvalidCDNBase = errors.New("uploadcare: invalid CDN base URL")
)

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

// Option configures a Config.
type Option func(*Config)

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Config) { c.HTTPClient = hc }
}

func WithAPIVersion(v string) Option {
	return func(c *Config) { c.APIVersion = v }
}

func WithSignBasedAuthentication() Option {
	return func(c *Config) { c.SignBasedAuthentication = true }
}

func WithUserAgent(ua string) Option {
	return func(c *Config) { c.UserAgent = ua }
}

func WithRetry(r *RetryConfig) Option {
	return func(c *Config) { c.Retry = r }
}

func WithCDNBase(url string) Option {
	return func(c *Config) { c.CDNBase = url }
}

// NewConfig builds the only Config shape NewClient accepts: defaults applied,
// CDNBase resolved against creds.PublicKey.
func NewConfig(creds APICreds, opts ...Option) (*Config, error) {
	if creds.PublicKey == "" {
		return nil, errors.New("uploadcare: invalid api creds: public key required")
	}
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.APIVersion == "" {
		cfg.APIVersion = defaultAPIVersion
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	cdnBase, err := resolveCDNBase(cfg.CDNBase, creds.PublicKey)
	if err != nil {
		return nil, err
	}
	cfg.CDNBase = cdnBase
	return cfg, nil
}

func cdnCNAMEPrefix(publicKey string) string {
	hash := sha256.Sum256([]byte(publicKey))
	prefix := new(big.Int).SetBytes(hash[:]).Text(36)
	if len(prefix) < cdnCNAMEPrefixLength {
		return prefix
	}
	return prefix[:cdnCNAMEPrefixLength]
}

func cdnBaseURL(publicKey string) string {
	return "https://" + cdnCNAMEPrefix(publicKey) + "." + defaultCDNDomain
}

func resolveCDNBase(raw, publicKey string) (string, error) {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if raw == "" {
		return cdnBaseURL(publicKey), nil
	}
	if !isValidCDNBase(raw) {
		return "", fmt.Errorf("%w: %q", errInvalidCDNBase, raw)
	}
	return raw, nil
}

func isValidCDNBase(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != "" && u.RawQuery == "" && u.Fragment == ""
}
