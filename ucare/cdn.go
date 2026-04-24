package ucare

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
)

const (
	defaultCDNDomain     = "ucarecd.net"
	cdnCNAMEPrefixLength = 10
)

var errInvalidCDNBase = errors.New("uploadcare: invalid CDN base URL")

type cdnBaseProvider interface {
	CDNBase() string
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

	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("%w: %q", errInvalidCDNBase, raw)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("%w: %q", errInvalidCDNBase, raw)
	}

	return raw, nil
}

// ClientCDNBase returns the CDN base URL associated with the client, if any.
// Returns the empty string for Client implementations that do not expose a
// CDN base (e.g. test doubles); callers should treat that as "do not rewrite".
func ClientCDNBase(c Client) string {
	if p, ok := c.(cdnBaseProvider); ok {
		return p.CDNBase()
	}
	return ""
}

// RewriteCDNURL returns originalURL with its scheme and host replaced by those
// of cdnBase, preserving the original path (including any trailing filename
// segment like /{uuid}/pineapple.jpg) and query. Returns originalURL unchanged
// if cdnBase is empty or either URL fails to parse as absolute.
func RewriteCDNURL(originalURL, cdnBase string) string {
	if cdnBase == "" || originalURL == "" {
		return originalURL
	}
	orig, err := url.Parse(originalURL)
	if err != nil || orig.Scheme == "" || orig.Host == "" {
		return originalURL
	}
	base, err := url.Parse(cdnBase)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return originalURL
	}
	orig.Scheme = base.Scheme
	orig.User = base.User
	orig.Host = base.Host
	if p := strings.TrimRight(base.Path, "/"); p != "" {
		orig.Path = p + orig.Path
		orig.RawPath = ""
	}
	return orig.String()
}
