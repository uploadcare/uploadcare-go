package ucare

import (
	"net/http"
	"time"
)

// Public configuration constants
const (
	APIv05 = "v0.5"
	APIv06 = "v0.6"
	APIv07 = "v0.7"

	simpleAuthScheme    = "Uploadcare.Simple"
	signBasedAuthScheme = "Uploadcare"
	dateHeaderFormat    = time.RFC1123

	signedUploadTTL = 60 * time.Second
)

var (
	defaultAPIVersion = APIv07

	authHeaderKey      = http.CanonicalHeaderKey("Authorization")
	userAgentHeaderKey = http.CanonicalHeaderKey("X-UC-User-Agent")

	dateHeaderLocation = time.FixedZone("GMT", 0)
)
