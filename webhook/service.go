// Package webhook holds all primitives and logic around the webhook resource.
package webhook

import (
	"context"

	"github.com/uploadcare/uploadcare-go/internal/config"
	"github.com/uploadcare/uploadcare-go/internal/svc"
	"github.com/uploadcare/uploadcare-go/ucare"
)

// Service describes all webhook related API
type Service interface {
	List(context.Context) ([]Info, error)
	Create(context.Context, Params) (Info, error)
	Update(context.Context, Params) (Info, error)
	Delete(ctx context.Context, targetURL string) error
}

type service struct {
	svc svc.Service
}

const (
	listPathFormat   = "/webhooks/"
	createPathFormat = "/webhooks/"
	updatePathFormat = "/webhooks/%d/"
	deletePathFormat = "/webhooks/unsubscribe/"
)

// Events
const (
	EventFileUploaded = "file.uploaded"
)

// NewService returns new instance of the Service
func NewService(client ucare.Client) Service {
	return service{svc.New(config.RESTAPIEndpoint, client, log)}
}
