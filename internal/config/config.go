// Package config holds all internal configurations
package config

import (
	"strings"
	"time"
)

// Endpoint represents backend endpoint
type Endpoint string

// Internal configuration constants
const (
	RESTAPIEndpoint   Endpoint = "api.uploadcare.com"
	UploadAPIEndpoint Endpoint = "upload.uploadcare.com"

	ClientVersion   = "0.1.0"
	UserAgentPrefix = "UploadcareGo"

	AcceptHeaderFormat = "application/vnd.uploadcare-%s+json"

	MaxThrottleRetries = 3

	UCTimeLayout = "2006-01-02T15:04:05"
)

// For reflection based payload encoding purpose.
const (
	FileFieldName     = "Data"
	FilenameFieldName = "Name"
)

// Time is needed just to parse custom formated time string
// returned from the Uploadcare API
type Time struct{ time.Time }

// UnmarshalJSON implements json.Unmarshaler
func (t *Time) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		t = nil
		return
	}
	// time is returned in a different format every time
	// so we need to normalize it in order to parse it
	dotInd := strings.Index(s, ".")
	if dotInd > -1 {
		s = s[:dotInd]
	}
	t.Time, err = time.Parse(UCTimeLayout, s)
	return
}

// CtxAuthFunc is a type for the auth func context key
type CtxAuthFunc struct{}

// CtxAuthFuncKey is a context key for passing auth func through the context
var CtxAuthFuncKey CtxAuthFunc
