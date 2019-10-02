package group

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/uploadcare/uploadcare-go/internal/codec"
)

// ListParams holds all possible params for the List method
type ListParams struct {
	// Limit specifies preferred amount of groups in a list for a single
	// response. Defaults to 100, while the maximum is 1000
	Limit *uint64 `form:"limit"`

	// OrderBy specifies the way groups are sorted in a returned list.
	// By default is set to datetime_created.
	OrderBy *string `form:"ordering"`

	// StartingFrom is a starting point for filtering group lists.
	StartingFrom *time.Time `form:"from"`
}

// EncodeReq implements ucare.ReqEncoder
func (d *ListParams) EncodeReq(req *http.Request) error {
	return codec.EncodeReqQuery(d, req)
}

// List holds a list of files
type List struct {
	raw codec.NextRawResulter
}

// Next indicates if there is a result to read
func (v *List) Next() bool { return v.raw.Next() }

// ReadResult returns next Info value. If no results are left to read it
// returns ucare.ErrEndOfResults.
// Example usage:
//	for groupList.Next() {
//		info, err := groupList.ReadResult()
//		...
//	}
func (v *List) ReadResult() (*Info, error) {
	raw, err := v.raw.ReadRawResult()
	if err != nil {
		return nil, err
	}

	var gi Info
	err = json.Unmarshal(raw, &gi)

	log.Debugf("reading group list result: %+v", gi)

	return &gi, err
}

// List returns a list of groups.
//
// Example usage:
//	groupList, err := groupSvc.List(ctx, params)
//	if err != nil {
//		// handle error
//	}
//	for groupList.Next() {
//		info, err := groupList.ReadResult()
//		...
//	}
func (s service) List(ctx context.Context, params *ListParams) (*List, error) {
	resbuf, err := s.svc.List(ctx, listPathFormat, params)
	return &List{raw: resbuf}, err
}
