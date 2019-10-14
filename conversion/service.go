package conversion

import (
	"context"

	"github.com/uploadcare/uploadcare-go/internal/config"
	"github.com/uploadcare/uploadcare-go/internal/svc"
	"github.com/uploadcare/uploadcare-go/ucare"
)

// Service describes all conversion related API
type Service interface {
	Document(context.Context, Params) (Result, error)
	DocumentStatus(context.Context, string) (StatusResult, error)

	Video(context.Context, Params) (Result, error)
	VideoStatus(context.Context, string) (StatusResult, error)
}

type service struct {
	svc svc.Service
}

// NewService returns new instance of the Service
func NewService(client ucare.Client) Service {
	return service{svc.New(config.RESTAPIEndpoint, client, log)}
}

const (
	convertDocumentFormat       = "/convert/document/"
	convertDocumentStatusFormat = "/convert/document/status/%s/"
	convertVideoFormat          = "/convert/video/"
	convertVideoStatusFormat    = "/convert/video/status/%s/"
)

// Predefined constants
const (
	ToStoreTrue  = "1"
	ToStoreFalse = "0"
)
