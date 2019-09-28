package file

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/uploadcare/uploadcare-go/internal/codec"
	"github.com/uploadcare/uploadcare-go/internal/config"
	"github.com/uploadcare/uploadcare-go/ucare"
)

// ListParams holds all possible params to for the List method
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
	Limit *int64 `form:"limit"`

	// Ordering specifies the way files are sorted in a returned list.
	// By default is set to datetime_uploaded.
	Ordering *string `form:"ordering"`

	// From specifies a starting point for filtering files.
	// The value depends on your ordering parameter value.
	From *string `form:"from"`
}

// EncodeRequest implements ucare.RequestEncoder
func (d *ListParams) EncodeRequest(req *http.Request) {
	t, v := reflect.TypeOf(d).Elem(), reflect.ValueOf(d).Elem()
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
		case *int64:
			val = strconv.FormatInt(ucare.Int64Val(valc), 10)
		case *bool:
			val = fmt.Sprintf("%t", ucare.BoolVal(valc))
		}

		q.Set(t.Field(i).Tag.Get("form"), val)
	}
	req.URL.RawQuery = q.Encode()
}

// List is a paginated list of files
type List struct{ codec.NextRawResulter }

// ReadResult returns next Info value. If no results are left to read it
// returns ucare.ErrEndOfResults.
// Example usage:
//	for fileList.Next() {
//		info, err := fileList.ReadResult()
//		...
//	}
func (v *List) ReadResult() (*Info, error) {
	raw, err := v.ReadRawResult()
	if err != nil {
		return nil, err
	}

	var fi Info
	err = json.Unmarshal(raw, &fi)

	log.Debugf("reading file list result: %+v", fi)

	return &fi, err
}

// List returns a paginated list of files
func (s service) List(
	ctx context.Context,
	params *ListParams,
) (*List, error) {
	if params == nil {
		params = &ListParams{}
	}

	method := http.MethodGet
	url := config.RESTAPIEndpoint + listPathFormat

	req, err := s.client.NewRequest(ctx, method, url, params)
	if err != nil {
		return nil, err
	}

	resbuf := &codec.ResultBuf{
		Ctx:       ctx,
		ReqMethod: method,
		Client:    s.client,
	}
	err = s.client.Do(req, &resbuf)
	if err != nil {
		return nil, err
	}

	return &List{resbuf}, nil
}
