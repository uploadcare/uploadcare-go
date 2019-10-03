package upload

import (
	"context"
	"errors"
	"net/http"

	"github.com/uploadcare/uploadcare-go/file"
	"github.com/uploadcare/uploadcare-go/internal/codec"
	"github.com/uploadcare/uploadcare-go/internal/config"
	"github.com/uploadcare/uploadcare-go/ucare"
)

// FileInfo holds file info (in the context of uploading)
type FileInfo struct {
	file.BasicFileInfo

	// IsStored is true if file is stored
	IsStored bool `json:"is_stored"`

	// Done denotes currently uploaded file size in bytes
	Done int64 `json:"done"`
	// Total is same as size
	Total uint64 `json:"total"`

	// Filename holds sanitized OriginalFileName
	FileName string `json:"filename"`
}

type fileInfoParams struct {
	FileID string `form:"file_id"`
	PubKey string `form:"pub_key"`
}

// EncodeReqQuery implements ucare.ReqEncoder
func (d *fileInfoParams) EncodeReq(req *http.Request) error {
	authFuncI := req.Context().Value(config.CtxAuthFuncKey)
	authFunc, ok := authFuncI.(ucare.UploadAPIAuthFunc)
	if !ok {
		return errors.New("auth func has a wrong signature")
	}
	d.PubKey, _, _ = authFunc()
	return codec.EncodeReqQuery(d, req)
}

// FileInfo returns file info in the context of uploading
func (s service) FileInfo(
	ctx context.Context,
	fileID string,
) (data *FileInfo, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		fileInfoPathFormat,
		&fileInfoParams{FileID: fileID},
		&data,
	)
	return
}
