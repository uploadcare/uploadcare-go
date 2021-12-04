package webhook

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/uploadcare/uploadcare-go/internal/codec"
	"github.com/uploadcare/uploadcare-go/internal/config"
)

// List of project webhooks
func (s service) List(
	ctx context.Context,
) (data []Info, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		listPathFormat,
		nil,
		&data,
	)
	return
}

// Create and subscribe to webhook
func (s service) Create(
	ctx context.Context,
	params Params,
) (data Info, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		createPathFormat,
		params,
		&data,
	)
	return
}

// Update webhook attributes
func (s service) Update(
	ctx context.Context,
	params Params,
) (data Info, err error) {
	if params.ID == nil {
		return Info{}, errors.New("params.ID is required")
	}

	err = s.svc.ResourceOp(
		ctx,
		http.MethodPut,
		fmt.Sprintf(updatePathFormat, *params.ID),
		params,
		&data,
	)
	return
}

// Unsubscribe and delete webhoo
func (s service) Delete(
	ctx context.Context,
	targetURL string,
) (err error) {
	var params = deleteParams{targetURL}

	err = s.svc.ResourceOp(
		ctx,
		http.MethodDelete,
		deletePathFormat,
		params,
		nil,
	)
	return
}

type deleteParams struct {
	TargetURL string `json:"target_url"`
}

func (p deleteParams) EncodeReq(req *http.Request) error {
	return codec.EncodeReqBody(p, req)
}

// Info holds webhook related information
type Info struct {
	// Webhook ID
	ID int64 `json:"id"`

	// Webhook creation date-time.
	CreatedAt *config.Time `json:"created"`

	// Webhook update date-time.
	UpdatedAt *config.Time `json:"updated"`

	// Webhook event.
	Event string `json:"event"`

	// Where webhook data will be posted.
	TargetURL string `json:"target_url"`

	// Signing secret (optional)
	SigningSecret *string `json:"signing_secret,omitempty"`

	// Webhook project ID.
	Project int64 `json:"project"`

	// Whether webhook is active
	IsActive bool `json:"is_active"`
}

// Params is for creating and updating webhook
type Params struct {
	ID *int64 `json:"id,omitempty"`
	// A URL that is triggered by an event, for example, a file upload.
	// A target URL MUST be unique for each project — event type combination.
	// Will not be changed if set to nil.
	TargetURL *string `json:"target_url,omitempty"`
	// Signing secret can be added when creating or updating a webhook
	SigningSecret *string `json:"signing_secret"`
	// An event you subscribe to. Presently, we only support the EventFileUploaded event.
	// Will not be changed if set to nil.
	Event *string `json:"event,omitempty"`
	// Marks a subscription as either active or not, defaults to true, otherwise false.
	// Will not be changed if set to nil.
	IsActive *bool `json:"is_active,omitempty"`
}

// EncodeReq implements ucare.ReqEncoder
func (p Params) EncodeReq(req *http.Request) error {
	return codec.EncodeReqBody(p, req)
}
