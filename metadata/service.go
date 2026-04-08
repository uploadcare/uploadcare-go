package metadata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/uploadcare/uploadcare-go/v2/internal/config"
	"github.com/uploadcare/uploadcare-go/v2/internal/svc"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

type Service interface {
	List(ctx context.Context, fileUUID string) (map[string]string, error)
	Get(ctx context.Context, fileUUID, key string) (string, error)
	Set(ctx context.Context, fileUUID, key, value string) (string, error)
	Delete(ctx context.Context, fileUUID, key string) error
}

type service struct {
	svc svc.Service
}

func NewService(client ucare.Client) Service {
	return service{svc.New(config.RESTAPIEndpoint, client, log)}
}

func (s service) List(
	ctx context.Context,
	fileUUID string,
) (data map[string]string, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		fmt.Sprintf("/files/%s/metadata/", fileUUID),
		nil,
		&data,
	)
	if err != nil {
		return
	}
	if len(data) > MaxKeysNumber {
		return nil, fmt.Errorf("%w: %d keys", ErrTooManyKeys, len(data))
	}
	return
}

func (s service) Get(
	ctx context.Context,
	fileUUID, key string,
) (data string, err error) {
	if err = validateKey(key); err != nil {
		return
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		metadataKeyPath(fileUUID, key),
		nil,
		&data,
	)
	return
}

func (s service) Set(
	ctx context.Context,
	fileUUID, key, value string,
) (data string, err error) {
	if err = validateKey(key); err != nil {
		return
	}
	if err = validateValue(value); err != nil {
		return
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPut,
		metadataKeyPath(fileUUID, key),
		stringBody(value),
		&data,
	)
	return
}

func (s service) Delete(
	ctx context.Context,
	fileUUID, key string,
) (err error) {
	if err = validateKey(key); err != nil {
		return
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodDelete,
		metadataKeyPath(fileUUID, key),
		nil,
		nil,
	)
	return
}

type stringBody string

func (s stringBody) EncodeReq(req *http.Request) error {
	raw, err := json.Marshal(string(s))
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(raw)
	req.Body = io.NopCloser(buf)
	req.ContentLength = int64(buf.Len())
	return nil
}

func metadataKeyPath(fileUUID, key string) string {
	return fmt.Sprintf(
		"/files/%s/metadata/%s/",
		fileUUID,
		escapeKeyPathSegment(key),
	)
}

func escapeKeyPathSegment(key string) string {
	switch key {
	case ".":
		return "%2E"
	case "..":
		return "%2E%2E"
	default:
		return url.PathEscape(key)
	}
}
