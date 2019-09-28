// Package file holds all primitives and logic around file entity
package file

import (
	"context"

	"github.com/uploadcare/uploadcare-go/ucare"
)

// Service describes all file related API
type Service interface {
	List(context.Context, *ListParams) (*List, error)
	Info(ctx context.Context, id string) (Info, error)
	Delete(ctx context.Context, id string) (Info, error)
	Store(ctx context.Context, id string) (Info, error)
}

type service struct {
	client ucare.Client
}

const (
	listPathFormat   = "/files/"
	infoPathFormat   = "/files/%s/"
	deletePathFormat = "/files/%s/"
	storePathFormat  = "/files/%s/storage/"
)

// NewService return new instance of the Service
func NewService(client ucare.Client) Service { return service{client} }

// OrderBy predefined constants to be used in request params
const (
	OrderByUploadedAtAsc  = "datetime_uploaded"
	OrderByUploadedAtDesc = "-datetime_uploaded"
	OrderBySizeAsc        = "size"
	OrderBySizeDesc       = "-size"
)
