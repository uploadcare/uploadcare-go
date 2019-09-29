// Package config holds all internal configurations
package config

import (
	"strings"
	"time"
)

// Internal configuration constants
const (
	RESTAPIEndpoint   = "https://api.uploadcare.com"
	UploadAPIEndpoint = "https://upload.uploadcare.com"

	ClientVersion   = "0.1.0"
	UserAgentPrefix = "UploadcareGo"

	AcceptHeaderFormat = "application/vnd.uploadcare-%s+json"

	MaxThrottleRetries = 3

	UCTimeLayout = "2006-01-02T15:04:05"
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
