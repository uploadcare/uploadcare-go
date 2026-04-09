package projectapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

const (
	projectsPath   = "/projects/"
	projectPathFmt = "/projects/%s/"
)

// List returns a paginated iterator over accessible projects.
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

// Create creates a new project.
func (s service) Create(
	ctx context.Context,
	params CreateProjectParams,
) (data Project, err error) {
	err = s.svc.ResourceOp(ctx, http.MethodPost, projectsPath, params, &data)
	return
}

// Get returns project info by public key.
func (s service) Get(
	ctx context.Context,
	pubKey string,
) (data Project, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		fmt.Sprintf(projectPathFmt, pubKey),
		nil,
		&data,
	)
	return
}

// Update updates project settings.
func (s service) Update(
	ctx context.Context,
	pubKey string,
	params UpdateProjectParams,
) (data Project, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		fmt.Sprintf(projectPathFmt, pubKey),
		params,
		&data,
	)
	return
}

// Delete deletes a project.
func (s service) Delete(
	ctx context.Context,
	pubKey string,
) error {
	return s.svc.ResourceOp(
		ctx,
		http.MethodDelete,
		fmt.Sprintf(projectPathFmt, pubKey),
		nil,
		nil,
	)
}
