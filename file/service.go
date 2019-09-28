// Package file holds all primitives and logic around file entity
package file

import (
	"context"

	"github.com/uploadcare/uploadcare-go/ucare"
)

// Service describes all file related API
type Service interface {
	ListFiles(context.Context, *ListParams) (*FileList, error)
	FileInfo(ctx context.Context, id string) (FileInfo, error)
	//DeleteFile(context.Context, ID) (FileInfo, error)
	//StoreFile(context.Context, ID) (FileInfo, error)
}

type service struct {
	client ucare.Client
}

const (
	listFilesPathFormat  = "/files/"
	fileInfoPathFormat   = "/files/%s/"
	deleteFilePathFormat = "/files/%s/"
	storeFilePathFormat  = "/files/%s/storage/"
)

// New return new instance of the Service
func New(client ucare.Client) Service { return service{client} }

const (
	OrderByUploadedAtAsc  = "datetime_uploaded"
	OrderByUploadedAtDesc = "-datetime_uploaded"
	OrderBySizeAsc        = "size"
	OrderBySizeDesc       = "-size"
)
