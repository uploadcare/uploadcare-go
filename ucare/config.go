package ucare

import (
	"net/http"
	"time"
)

// Public configuration constants
const (
	APIv05 = "v0.5"
	APIv06 = "v0.6"

	simpleAuthScheme    = "Uploadcare.Simple"
	signBasedAuthScheme = "Uploadcare"
	dateHeaderFormat    = time.RFC1123
)

var (
	supportedVersions = map[string]bool{
		APIv05: true,
		APIv06: true,
	}

	defaultAPIVersion = APIv05

	authHeaderKey      = http.CanonicalHeaderKey("Authorization")
	userAgentHeaderKey = http.CanonicalHeaderKey("X-UC-User-Agent")

	dateHeaderLocation = time.FixedZone("GMT", 0)
)
