package addon

import (
	"context"
	"fmt"
	"net/http"

	"github.com/uploadcare/uploadcare-go/v2/internal/codec"
	"github.com/uploadcare/uploadcare-go/v2/internal/config"
	"github.com/uploadcare/uploadcare-go/v2/internal/svc"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

// Service describes all addon related API
type Service interface {
	// Execute starts an addon execution on a file
	Execute(ctx context.Context, addonName Name, params ExecuteParams) (ExecuteResult, error)

	// Status checks the execution status of an addon request
	Status(ctx context.Context, addonName Name, requestID string) (StatusResult, error)
}

type service struct {
	svc svc.Service
}

// NewService returns new instance of the Service
func NewService(client ucare.Client) Service {
	return service{svc.New(config.RESTAPIEndpoint, client, log)}
}

// Execute starts an addon execution on a file
func (s service) Execute(
	ctx context.Context,
	addonName Name,
	params ExecuteParams,
) (data ExecuteResult, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		fmt.Sprintf("/addons/%s/execute/", addonName),
		executeBody(params),
		&data,
	)
	return
}

// Status checks the execution status of an addon request
func (s service) Status(
	ctx context.Context,
	addonName Name,
	requestID string,
) (data StatusResult, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		fmt.Sprintf("/addons/%s/execute/status/", addonName),
		statusQuery{RequestID: requestID},
		&data,
	)
	return
}

type executeBody ExecuteParams

func (p executeBody) EncodeReq(req *http.Request) error {
	return codec.EncodeReqBody(p, req)
}

type statusQuery struct {
	RequestID string `form:"request_id"`
}

func (p statusQuery) EncodeReq(req *http.Request) error {
	return codec.EncodeReqQuery(&p, req)
}
