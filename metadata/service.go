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

// Service describes all file metadata related API
type Service interface {
	// List returns all metadata key-value pairs for a file
	List(ctx context.Context, fileUUID string) (map[string]string, error)

	// Get returns a single metadata value by key
	Get(ctx context.Context, fileUUID, key string) (string, error)

	// Set creates or updates a metadata key-value pair
	Set(ctx context.Context, fileUUID, key, value string) (string, error)

	// Delete removes a metadata key
	Delete(ctx context.Context, fileUUID, key string) error
}

type service struct {
	svc svc.Service
}

// NewService returns new instance of the Service
func NewService(client ucare.Client) Service {
	return service{svc.New(config.RESTAPIEndpoint, client, log)}
}

// List returns all metadata key-value pairs for a file
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
	return
}

// Get returns a single metadata value by key
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

// Set creates or updates a metadata key-value pair
func (s service) Set(
	ctx context.Context,
	fileUUID, key, value string,
) (data string, err error) {
	if err = validateKey(key); err != nil {
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

// Delete removes a metadata key
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

// stringBody is a ReqEncoder that writes a JSON-encoded string as the
// request body. This is needed because the metadata Set endpoint expects
// a plain JSON string (e.g. "value") rather than a JSON object.
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
