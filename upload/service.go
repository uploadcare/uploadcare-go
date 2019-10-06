// Package upload contains all upload related API stuff.
//
// Upload API is an addition to the REST API. It provides several ways of uploading
// files to the Uploadcare servers.
// Every uploaded file is temporary and subject to be deleted within a 24-hour
// period. To make any file permanent, you should store or copy it.
//
// The package provides uploading files by making requests with payload to
// the Uploadcare API endpoints. There are two basic upload types:
// - Direct uploads, a regular upload mode that suits most files less than 100MB
//   in size. You won’t be able to use this mode for larger files.
// - Multipart uploads, a more sophisticated upload mode supporting any files
//   larger than 10MB and implementing accelerated uploads through
//   a distributed network.
package upload

import (
	"context"

	"github.com/uploadcare/uploadcare-go/internal/config"
	"github.com/uploadcare/uploadcare-go/internal/svc"
	"github.com/uploadcare/uploadcare-go/ucare"
)

// Service describes all upload related API functionality
type Service interface {
	UploadFile(context.Context, *FileParams) (id string, err error)
	FromURL(context.Context, *FromURLParams) (FromURLData, error)
	FileInfo(ctx context.Context, id string) (*FileInfo, error)
	// CreateGroup
	// GroupInfo
}

type service struct{ svc svc.Service }

// NewService creates new upload service instance.
func NewService(client ucare.Client) Service {
	return service{svc.New(config.UploadAPIEndpoint, client, log)}
}

// Predefined file storing behaviour constants
const (
	ToStoreTrue  = "1"
	ToStoreFalse = "0"
	ToStoreAuto  = "auto"

	URLDuplicatesTrue  = "1"
	URLDuplicatesFalse = "0"

	uploadStatuSuccess     = "success"
	uploadStatusInProgress = "progress"
	uploadStatusError      = "error"
	uploadStatusWaiting    = "waiting"
	uploadStatusUnknown    = "unknown"
)

const (
	directUploadFormat  = "/base/"
	fromURLFormat       = "/from_url/"
	fromURLStatusFormat = "/from_url/status/?token=%s"
	fileInfoFormat      = "/info/"
)
