package projectapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

const (
	secretsPathFmt = "/projects/%s/secrets/"
	secretPathFmt  = "/projects/%s/secrets/%s/"
)

// ListSecrets returns secret keys for a project.
func (s service) ListSecrets(
	ctx context.Context,
	pubKey string,
	params *ListParams,
) (data SecretList, err error) {
	var enc ucare.ReqEncoder
	if params != nil {
		enc = params
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		fmt.Sprintf(secretsPathFmt, pubKey),
		enc,
		&data,
	)
	return
}

// CreateSecret creates a new secret key for a project.
func (s service) CreateSecret(
	ctx context.Context,
	pubKey string,
) (data SecretRevealed, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		fmt.Sprintf(secretsPathFmt, pubKey),
		nil,
		&data,
	)
	return
}

// DeleteSecret deletes a secret key.
func (s service) DeleteSecret(
	ctx context.Context,
	pubKey string,
	secretID string,
) error {
	return s.svc.ResourceOp(
		ctx,
		http.MethodDelete,
		fmt.Sprintf(secretPathFmt, pubKey, secretID),
		nil,
		nil,
	)
}
