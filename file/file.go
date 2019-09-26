// Package file holds all primitives and logic around file entity
package file

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/uploadcare/uploadcare-go/uploadcare"
)

// Service describes all file related API
type Service interface {
	ListFiles(context.Context, *ListParams) (*FileList, error)
	//FileInfo(context.Context, ID) (FileInfo, error)
	//DeleteFile(context.Context, ID) (FileInfo, error)
	//StoreFile(context.Context, ID) (FileInfo, error)
}

type service struct {
	client uploadcare.Client
}

const (
	listFilesPathFormat  = "/files/"
	fileInfoPathFormat   = "/files/%s/"
	deleteFilePathFormat = "/files/%s/"
	storeFilePathFormat  = "/files/%s/storage/"
)

// New return new instance of the FileService
func New(client uploadcare.Client) Service { return service{client} }

const (
	OrderByUploadedAtAsc  = "datetime_uploaded"
	OrderByUploadedAtDesc = "-datetime_uploaded"
	OrderBySizeAsc        = "size"
	OrderBySizeDesc       = "-size"
)

// ListParams holds all possible params to for the ListFiles method
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

// EncodeRequest is uploadcare.RequestEncoder implementation
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
			val = uploadcare.StringVal(valc)
		case *int64:
			val = strconv.FormatInt(uploadcare.Int64Val(valc), 10)
		case *bool:
			val = fmt.Sprintf("%t", uploadcare.BoolVal(valc))
		}

		q.Set(t.Field(i).Tag.Get("form"), val)
	}
	req.URL.RawQuery = q.Encode()
}

type FileList struct {
	NextPage string     `json:"next"`
	PrevPage string     `json:"previous"`
	Total    int64      `json:"total"`
	PerPage  int64      `json:"per_page"`
	Results  []FileInfo `json:"results"`
}

func (d *FileList) DecodeRespBody(body io.Reader) error {
	return json.NewDecoder(body).Decode(d)
}

// ListFiles returns a paginated list of files
func (s service) ListFiles(
	ctx context.Context,
	params *ListParams,
) (*FileList, error) {
	if params == nil {
		params = &ListParams{}
	}

	url := uploadcare.SingleSlashJoin(
		uploadcare.RESTAPIEndpoint,
		listFilesPathFormat,
	)

	req, err := s.client.NewRequest(http.MethodGet, url, params)
	if err != nil {
		return nil, err
	}

	var flist FileList
	err = s.client.Do(req, &flist)

	return &flist, err
}

type FileInfo struct {
	RemovedAt        *time.Time `json:"datetime_removed"`
	StoredAt         *time.Time `json:"datetime_stored"`
	UploadedAt       *time.Time `json:"datetime_uploaded"`
	ImageInfo        *ImageInfo `json:"image_info"`
	MimeType         string     `json:"mime_type"`
	OriginalFileURL  string     `json:"original_file_url"`
	OriginalFileName string     `json:"original_filename"`
	URI              string     `json:"uri"`
	ID               string     `json:"uuid"`
	Size             int64      `json:"size"`
	IsImage          bool       `json:"is_image"`
	IsReady          bool       `json:"is_ready"`
}

type ImageInfo struct {
}

//// FileInfo is used to acquire some file-specific info
//func (s service) FileInfo(ctx context.Context, id ID) (FileInfo, error) {
//	panic("not implemented")
//}
//
//// DeleteFile is used to remove individual files
//func (s service) DeleteFile(ctx context.Context, id ID) (FileInfo, error) {
//	panic("not implemented")
//}
//
//// StoreFile is used to store a single file by ID
//func (s service) StoreFile(ctx context.Context, id ID) (FileInfo, error) {
//	panic("not implemented")
//}
