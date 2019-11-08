// Package file holds all primitives and logic around the file resource.
//
// The file resource is intended to handle user-uploaded files and
// is the main Uploadcare resource.
//
// Each of uploaded files has an ID (UUID) that is assigned once and never
// changes later.
package file

import (
	"context"

	"github.com/uploadcare/uploadcare-go/internal/config"
	"github.com/uploadcare/uploadcare-go/internal/svc"
	"github.com/uploadcare/uploadcare-go/ucare"
)

// Service describes all file related API
type Service interface {
	List(context.Context, ListParams) (*List, error)
	Info(ctx context.Context, id string) (Info, error)
	Store(ctx context.Context, id string) (Info, error)
	Delete(ctx context.Context, id string) (Info, error)
	BatchStore(ctx context.Context, ids []string) (BatchInfo, error)
	BatchDelete(ctx context.Context, ids []string) (BatchInfo, error)
	Copy(context.Context, CopyParams) (LocalCopyInfo, error)
	LocalCopy(context.Context, LocalCopyParams) (LocalCopyInfo, error)
	RemoteCopy(context.Context, RemoteCopyParams) (RemoteCopyInfo, error)
}

type service struct {
	svc svc.Service
}

const (
	listPathFormat   = "/files/"
	infoPathFormat   = "/files/%s/"
	deletePathFormat = "/files/%s/"

	storePathFormat       = "/files/%s/storage/"
	batchStorePathFormat  = "/files/storage/"
	batchDeletePathFormat = "/files/storage/"

	copyPathFormat       = "/files/"
	localCopyPathFormat  = "/files/local_copy/"
	remoteCopyPathFormat = "/files/remote_copy/"
)

// OrderBy predefined constants to be used in request params
const (
	OrderByUploadedAtAsc  = "datetime_uploaded"
	OrderByUploadedAtDesc = "-datetime_uploaded"
	OrderBySizeAsc        = "size"
	OrderBySizeDesc       = "-size"
)

// Copy file params constants
const (
	StoreTrue       = "true"
	StoreFalse      = "false"
	MakePublicTrue  = "true"
	MakePublicFalse = "false"
)

// Pattern remote copy param constants
const (
	PatternDefault      = "${default}"
	PatternAutoFileName = "${filename} ${effects} ${ext}"
	PatternEffects      = "${effects}"
	PatternFileName     = "${filename}"
	PatternID           = "${uuid}"
	PatternExt          = "${ext}"
)

// NewService returns new instance of the Service
func NewService(client ucare.Client) Service {
	return service{svc.New(config.RESTAPIEndpoint, client, log)}
}
