package tag

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/uploadcare/uploadcare-go/v2/internal/codec"
	"github.com/uploadcare/uploadcare-go/v2/internal/config"
	"github.com/uploadcare/uploadcare-go/v2/internal/filetag"
	"github.com/uploadcare/uploadcare-go/v2/internal/svc"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

type Service interface {
	List(ctx context.Context, fileUUID string) ([]string, error)
	Replace(ctx context.Context, fileUUID string, tags []string) (Result, error)
	Update(ctx context.Context, fileUUID string, params UpdateParams) (Result, error)
}

type service struct {
	svc svc.Service
}

func NewService(client ucare.Client) Service {
	return service{svc.New(config.RESTAPIEndpoint, client, log)}
}

type listResponse struct {
	Tags []string `json:"tags"`
}

type replaceParams struct {
	Tags []string `json:"tags"`
}

func (p replaceParams) EncodeReq(req *http.Request) error {
	return codec.EncodeReqBody(p, req)
}

type updateBody UpdateParams

func (p updateBody) EncodeReq(req *http.Request) error {
	return codec.EncodeReqBody(p, req)
}

func (s service) List(
	ctx context.Context,
	fileUUID string,
) (tags []string, err error) {
	if err = validateFileUUID(fileUUID); err != nil {
		return
	}

	var response listResponse
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		tagsPath(fileUUID),
		nil,
		&response,
	)
	return response.Tags, err
}

func (s service) Replace(
	ctx context.Context,
	fileUUID string,
	tags []string,
) (result Result, err error) {
	if err = validateFileUUID(fileUUID); err != nil {
		return
	}
	tags, err = filetag.Normalize(tags, filetag.MaxCount)
	if err != nil {
		return
	}

	err = s.svc.ResourceOp(
		ctx,
		http.MethodPut,
		tagsPath(fileUUID),
		replaceParams{Tags: tags},
		&result,
	)
	return
}

func (s service) Update(
	ctx context.Context,
	fileUUID string,
	params UpdateParams,
) (result Result, err error) {
	if err = validateFileUUID(fileUUID); err != nil {
		return
	}
	// Every unique tag in Add is present in the merged result, so the API's
	// final MaxCount limit also bounds Add. Other merged-set overflows cannot
	// be validated without knowing the file's current tags.
	params.Add, err = filetag.Normalize(params.Add, filetag.MaxCount)
	if err != nil {
		return
	}
	// Delete has no request-level count limit; absent tags are no-ops.
	params.Delete, err = filetag.Normalize(params.Delete, 0)
	if err != nil {
		return
	}

	err = s.svc.ResourceOp(
		ctx,
		http.MethodPatch,
		tagsPath(fileUUID),
		updateBody(params),
		&result,
	)
	return
}

func tagsPath(fileUUID string) string {
	return fmt.Sprintf("/files/%s/tags/", url.PathEscape(fileUUID))
}
