// Package file holds all primitives and logic around file entity
package file

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strconv"

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
func New(client uploadcare.Client) Service {
	return service{client}
}

// ID represents unique file id (uuid)
type ID string

// ListParams holds all possible params to for the ListFiles method
type ListParams struct {
	Removed  *string `form:"removed"`
	Stored   *string `form:"stored"`
	Limit    *int    `form:"limit"`
	Ordering *string `form:"ordering"`
	// From field
	// AddFields *string `form:"add_fields"`
}

// EncodeRequest is uploadcare.RequestEncoder implementation
func (d *ListParams) EncodeRequest(req *http.Request) {
	t, v := reflect.TypeOf(d), reflect.ValueOf(d)
	for i := 0; i < t.NumField(); i++ {
		if v.Field(i).Interface() == nil {
			continue
		}
		valPtr := v.Field(i).Interface()
		if valPtr == nil {
			continue
		}
		var val string
		switch valc := valPtr.(type) {
		case *string:
			val = *valc
		case *int:
			val = strconv.Itoa(*valc)
		}
		req.URL.Query().Set(
			t.Field(i).Tag.Get("form"),
			val,
		)
	}
}

type FileList struct {
}

func (d *FileList) DecodeRespBody(body io.ReadCloser) error {
	panic("not implemented")
	return nil
}

// ListFiles returns a paginated list of files
func (s service) ListFiles(
	ctx context.Context,
	params *ListParams,
) (*FileList, error) {
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
