// Package webhook holds all primitives and logic around the webhook resource.
package webhook

import (
	"context"

	"github.com/uploadcare/uploadcare-go/v2/internal/config"
	"github.com/uploadcare/uploadcare-go/v2/internal/svc"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
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

type Event string

const (
	EventFileUploaded    Event = "file.uploaded"
	EventFileStored      Event = "file.stored"
	EventFileDeleted     Event = "file.deleted"
	EventFileInfoUpdated Event = "file.info_updated"

	// EventFileInfected is deprecated. Use EventFileInfoUpdated instead.
	EventFileInfected Event = "file.infected"
)

func EventPtr(v Event) *Event { return &v }

// NewService returns new instance of the Service
func NewService(client ucare.Client) Service {
	return service{svc.New(config.RESTAPIEndpoint, client, log)}
}
