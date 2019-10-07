package file

import (
	"context"
	"net/http"

	"github.com/uploadcare/uploadcare-go/internal/codec"
)

type batchParams []string

func (d batchParams) EncodeReq(req *http.Request) error {
	return codec.EncodeReqBody(d, req)
}

// BatchStore is used to store multiple files in one go. Up to 100 files are
// supported per request.
func (s service) BatchStore(
	ctx context.Context,
	ids []string,
) (data BatchInfo, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPut,
		batchStorePathFormat,
		batchParams(ids),
		&data,
	)
	return
}

// BatchDelete is used to delete multiple files in one go. Up to 100 files are
// supported per request.
func (s service) BatchDelete(
	ctx context.Context,
	ids []string,
) (data BatchInfo, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodDelete,
		batchDeletePathFormat,
		batchParams(ids),
		&data,
	)
	return
}

// BatchInfo holds batch operation response data.
type BatchInfo struct {
	// Problems is a map of passed files IDs and problems associated problems
	Problems map[string]string `json:"problems"`

	// Results describes successfully operated files
	Results []Info `json:"result"`
}
