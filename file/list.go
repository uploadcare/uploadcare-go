package file

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/uploadcare/uploadcare-go/internal/codec"
)

// ListParams holds all possible params for for the List method
type ListParams struct {
	// Removed is set to true if only include removed files in the response,
	// otherwise existing files are included. Defaults to false.
	Removed *bool `form:"removed"`

	// Stored is set to true if only include files that were stored.
	// Set to false to include only temporary files.
	// The default is unset: both stored and not stored files are returned
	Stored *bool `form:"stored"`

	// Limit specifies preferred amount of files in a list for a single
	// response. Defaults to 100, while the maximum is 1000
	Limit *uint64 `form:"limit"`

	// OrderBy specifies the way files are sorted in a returned list.
	// By default is set to datetime_uploaded.
	OrderBy *string `form:"ordering"`

	// StartingFrom specifies a starting point for filtering files.
	// The value depends on your ordering parameter value.
	StartingFrom *time.Time `form:"from"`
}

// EncodeReq implements ucare.ReqEncoder
func (d *ListParams) EncodeReq(req *http.Request) {
	codec.EncodeReqQuery(d, req)
}

// List holds a list of files
type List struct {
	raw codec.NextRawResulter
}

// Next indicates if there is a result to read
func (v *List) Next() bool { return v.raw.Next() }

// ReadResult returns next Info value. If no results are left to read it
// returns ucare.ErrEndOfResults.
func (v *List) ReadResult() (*Info, error) {
	raw, err := v.raw.ReadRawResult()
	if err != nil {
		return nil, err
	}

	var fi Info
	err = json.Unmarshal(raw, &fi)

	log.Debugf("reading file list result: %+v", fi)

	return &fi, err
}

// List returns a list of files.
//
// Example usage:
//	fileList, err := fileSvc.List(ctx, params)
//	if err != nil {
//		// handle error
//	}
//	for fileList.Next() {
//		info, err := fileList.ReadResult()
//		...
//	}
func (s service) List(ctx context.Context, params *ListParams) (*List, error) {
	resbuf, err := s.svc.List(ctx, listPathFormat, params)
	return &List{raw: resbuf}, err
}
