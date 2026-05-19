package projectapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

const (
	projectsPath   = "/projects/"
	projectPathFmt = "/projects/%s/"
)

func (s service) List(
	ctx context.Context,
	params *ListParams,
) (*ProjectList, error) {
	var enc ucare.ReqEncoder
	if params != nil {
		enc = params
	}
	resbuf, err := s.svc.List(ctx, projectsPath, enc)
	return &ProjectList{raw: resbuf}, err
}

func (s service) Create(
	ctx context.Context,
	params CreateProjectParams,
) (data Project, err error) {
	err = s.svc.ResourceOp(ctx, http.MethodPost, projectsPath, params, &data)
	return
}

func (s service) Get(
	ctx context.Context,
	pubKey string,
) (data Project, err error) {
	if err = validatePubKey(pubKey); err != nil {
		return
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		projectPath(pubKey),
		nil,
		&data,
	)
	return
}

func (s service) Update(
	ctx context.Context,
	pubKey string,
	params UpdateProjectParams,
) (data Project, err error) {
	if err = validatePubKey(pubKey); err != nil {
		return
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		projectPath(pubKey),
		params,
		&data,
	)
	return
}

func (s service) Delete(
	ctx context.Context,
	pubKey string,
) error {
	if err := validatePubKey(pubKey); err != nil {
		return err
	}
	return s.svc.ResourceOp(
		ctx,
		http.MethodDelete,
		projectPath(pubKey),
		nil,
		nil,
	)
}

func projectPath(pubKey string) string {
	return fmt.Sprintf(projectPathFmt, url.PathEscape(pubKey))
}
