// Package codec holds everything related to decoding paginated response
// and making subsequent pages calls
//
// WIP.
// TODO: add tests
package codec

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/uploadcare/uploadcare-go/internal/config"
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

// EncodeReqQuery encodes data passed as an http.Request query string.
// NOTE: data must be a pointer to a struct type.
func EncodeReqQuery(data interface{}, req *http.Request) {
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return
	}
	t, v := reflect.TypeOf(data).Elem(), reflect.ValueOf(data).Elem()
	if t.Kind() != reflect.Struct {
		return
	}
	q := req.URL.Query()
	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		if f.IsNil() {
			continue
		}

		var val string
		switch valc := f.Interface().(type) {
		case *string:
			val = ucare.StringVal(valc)
		case *uint64:
			val = strconv.FormatUint(ucare.Uint64Val(valc), 10)
		case *bool:
			val = fmt.Sprintf("%t", ucare.BoolVal(valc))
		case *time.Time:
			val = valc.Format(config.UCTimeLayout)
		}

		q.Set(t.Field(i).Tag.Get("form"), val)
	}
	req.URL.RawQuery = q.Encode()
}
