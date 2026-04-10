package ucare

import (
	"net/http"
	"time"
)

const (
	APIv07 = "v0.7"

	simpleAuthScheme    = "Uploadcare.Simple"
	signBasedAuthScheme = "Uploadcare"
	dateHeaderFormat    = time.RFC1123

	signedUploadTTL = 60 * time.Second
)

var (
	defaultAPIVersion = APIv07

	authHeaderKey      = http.CanonicalHeaderKey("Authorization")
	userAgentHeaderKey = http.CanonicalHeaderKey("User-Agent")

	dateHeaderLocation = time.FixedZone("GMT", 0)
)
