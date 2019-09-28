// Package codec holds everything related to decoding paginated response
// and making subsequent pages calls
//
// WIP.
// TODO: add tests
package codec

import (
	"context"
	"errors"
	"sync"

	"github.com/uploadcare/uploadcare-go/ucare"
)

// Raw represents raw bytes data
type Raw []byte

// UnmarshalJSON implements json.Unmarshaler
func (v *Raw) UnmarshalJSON(data []byte) error {
	*v = Raw(data)
	return nil
}

// NextRawResulter abstracts reading raw results from paginated api response
type NextRawResulter interface {
	Next() bool
	ReadRawResult() (Raw, error)
}

// ResultBuf implements NextRawResulter
type ResultBuf struct {
	Ctx       context.Context
	ReqMethod string
	Client    ucare.Client

	sync.Mutex         // guards everything below
	NextPage   *string `json:"next"`
	Vals       []Raw   `json:"results"`
	at         int     // index to read from the Vals
}

// Next indicates if there is a result to read
func (b *ResultBuf) Next() bool {
	b.Lock()
	defer b.Unlock()
	return !(b.at >= len(b.Vals) && b.NextPage == nil)
}

// ErrEndOfResults denotes absence of results
var ErrEndOfResults = errors.New("No results are left to read")

// ReadRawResult reads returns next Raw result.
// It makes paginated requests when all results from the current page
// have been read.
func (b *ResultBuf) ReadRawResult() (Raw, error) {
	if !b.Next() {
		return nil, ErrEndOfResults
	}

	b.Lock()
	defer b.Unlock()

	var valsPrev []Raw
	if b.at >= len(b.Vals) && b.NextPage != nil {
		valsPrev, b.Vals = b.Vals, nil

		req, err := b.Client.NewRequest(
			b.Ctx,
			b.ReqMethod,
			*b.NextPage,
			nil,
		)
		if err != nil {
			return nil, err
		}

		err = b.Client.Do(req, b)
		if err != nil {
			return nil, err
		}

		b.Vals = append(valsPrev, b.Vals...)
	}

	res := b.Vals[b.at]
	b.at++

	return res, nil
}
