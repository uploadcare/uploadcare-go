package upload

import (
	"context"
	"net/http"

	"github.com/uploadcare/uploadcare-go/file"
	"github.com/uploadcare/uploadcare-go/internal/codec"
)

// FileInfo holds file info (in the context of uploading)
type FileInfo struct {
	file.BasicFileInfo

	// IsStored is true if file is stored
	IsStored bool `json:"is_stored"`

	// Done denotes currently uploaded file size in bytes
	Done uint64 `json:"done"`
	// Total is same as size
	Total uint64 `json:"total"`

	// Filename holds sanitized OriginalFileName
	FileName string `json:"filename"`

	// S3Bucket is your custom user bucket on which file are stored. Only
	// available of you setup foreign storage bucket for your project
	S3Bucket string `json:"s3_bucket"`

	// DefaultEffects holds CDN media transformations applied to the file
	// when its group was created
	DefaultEffects string `json:"default_effects"`
}

type fileInfoParams struct {
	FileID string `form:"file_id"`
	PubKey string `form:"pub_key"`
}

// EncodeReqQuery implements ucare.ReqEncoder
func (d *fileInfoParams) EncodeReq(req *http.Request) error {
	d.PubKey, _, _ = authFromContext(req.Context())()
	return codec.EncodeReqQuery(d, req)
}

// FileInfo returns file info in the context of uploading
func (s service) FileInfo(
	ctx context.Context,
	fileID string,
) (data FileInfo, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		fileInfoFormat,
		&fileInfoParams{FileID: fileID},
		&data,
	)
	return
}
