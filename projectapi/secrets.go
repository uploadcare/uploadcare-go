package projectapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

const (
	secretsPathFmt = "/projects/%s/secrets/"
	secretPathFmt  = "/projects/%s/secrets/%s/"
)

func (s service) ListSecrets(
	ctx context.Context,
	pubKey string,
	params *ListParams,
) (*SecretList, error) {
	if err := validatePubKey(pubKey); err != nil {
		return nil, err
	}
	var enc ucare.ReqEncoder
	if params != nil {
		enc = params
	}
	resbuf, err := s.svc.List(ctx, secretsPath(pubKey), enc)
	return &SecretList{raw: resbuf}, err
}

func (s service) CreateSecret(
	ctx context.Context,
	pubKey string,
) (data SecretRevealed, err error) {
	if err = validatePubKey(pubKey); err != nil {
		return
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		secretsPath(pubKey),
		nil,
		&data,
	)
	return
}

func (s service) DeleteSecret(
	ctx context.Context,
	pubKey string,
	secretID string,
) error {
	if err := validatePubKey(pubKey); err != nil {
		return err
	}
	if err := validateSecretID(secretID); err != nil {
		return err
	}
	return s.svc.ResourceOp(
		ctx,
		http.MethodDelete,
		secretPath(pubKey, secretID),
		nil,
		nil,
	)
}

func secretsPath(pubKey string) string {
	return fmt.Sprintf(secretsPathFmt, url.PathEscape(pubKey))
}

func secretPath(pubKey, secretID string) string {
	return fmt.Sprintf(secretPathFmt, url.PathEscape(pubKey), url.PathEscape(secretID))
}
