package uploadcare

import (
	"strings"
	"time"
)

type RESTAPIVersion string

const (
	APIv05 RESTAPIVersion = "v0.5"
	APIV06 RESTAPIVersion = "v0.6"

	clientVersion   = "0.1.0"
	userAgentPrefix = "UploadcareGo"

	acceptHeaderFormat = "application/vnd.uploadcare-%s+json"

	maxThrottleRetries = 3

	ucTimeLayout = "2006-01-02T15:04:05"
)

var (
	supportedVersions = map[RESTAPIVersion]bool{
		APIv05: true,
		APIV06: true,
	}

	DefaultAPIVersion = APIv05
)

type Time struct{ time.Time }

func (t *Time) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		t = nil
		return nil
	}
	// time is returned in a different format every time
	// so we need to normalize it in order to parse it
	dotInd := strings.Index(s, ".")
	if dotInd > -1 {
		s = s[:dotInd]
	}
	t.Time, err = time.Parse(ucTimeLayout, s)
	return err
}
