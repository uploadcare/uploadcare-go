// Package codec holds encoding decoding reading writing stuff
package codec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
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
			"",
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
func EncodeReqQuery(data interface{}, req *http.Request) error {
	if err := checkDataType(data); err != nil {
		return err
	}

	q := req.URL.Query()
	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		if f.Kind() == reflect.Ptr && f.IsNil() {
			continue
		}

		q.Set(t.Field(i).Tag.Get("form"), fieldValue(f))
	}
	req.URL.RawQuery = q.Encode()
	return nil
}

// EncodeReqFormData encodes data passed as a form data.
// NOTE: data must be a pointer to a struct type.
func EncodeReqFormData(data interface{}) (io.ReadCloser, string, error) {
	if err := checkDataType(data); err != nil {
		return nil, "", err
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	fileField := config.FileFieldName

	// writing file to the form
	fileFieldName, ok := reflect.TypeOf(data).Elem().FieldByName(fileField)
	if ok {
		f, ok := reflect.ValueOf(data).Elem().
			FieldByName(fileField).
			Interface().(*os.File)
		if !ok {
			return nil, "", errors.New("File field must be on *os.File")
		}
		formFieldName := fileFieldName.Tag.Get("form")
		part, err := writer.CreateFormFile(formFieldName, f.Name())
		if err != nil {
			return nil, "", err
		}
		if _, err = io.Copy(part, f); err != nil {
			return nil, "", err
		}
	}

	for i := 0; i < t.NumField(); i++ {
		tf, vf := t.Field(i), v.Field(i)

		if vf.Kind() == reflect.Struct {
			// packing embedded struct fields
			for i := 0; i < vf.NumField(); i++ {
				f := vf.Field(i)
				if f.Kind() == reflect.Ptr && f.IsNil() {
					continue
				}

				formKey := tf.Type.Field(k).Tag.Get("form")
				if formKey == "" {
					continue
				}

				_ = writer.WriteField(formkey, fieldValue(f))
			}
			continue
		}
		if vf.Kind() != reflect.Struct && vf.IsNil() {
			continue
		}

		formTag := t.Field(i).Tag.Get("form")
		if formTag == strings.ToLower(fileField) || formTag == "" {
			continue
		}
		_ = writer.WriteField(formTag, fieldValue(vf))
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return ioutil.NopCloser(body), writer.FormDataContentType(), nil
}

func fieldValue(v reflect.Value) (val string) {
	switch valc := f.Interface().(type) {
	case string:
		val = valc
	case *string:
		val = ucare.StringVal(valc)
	case *uint64:
		val = strconv.FormatUint(ucare.Uint64Val(valc), 10)
	case *int64:
		val = strconv.FormatInt(ucare.Int64Val(valc), 10)
	case *bool:
		val = fmt.Sprintf("%t", ucare.BoolVal(valc))
	case *time.Time:
		val = valc.Format(config.UCTimeLayout)
	}
	return
}

func checkDataType(d interface{}) error {
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return errors.New("data is not a pointer type")
	}
	t, v := reflect.TypeOf(data).Elem(), reflect.ValueOf(data).Elem()
	if t.Kind() != reflect.Struct {
		return errors.New("data is not a struct type")
	}
	return nil
}
