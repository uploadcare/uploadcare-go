package upload

import (
	"context"
	"io"
	"net/http"
)

// FileParams holds all possible params for the file upload
type FileParams struct {
	// File reads a file to be uploaded
	File io.Reader `form:"file"`

	// ToStore sets the file storing behaviour
	ToStore *string `form:"UPLOADCARE_STORE"`
}

func (d *FileParams) EncodeReq(req *http.Request) {
	// TODO: encode to body
	// - d itself
	// - UPLOADCARE_PUB_KEY
	// - signature
	// - expire
}

// UploadFile uploads a file and return its unique id (uuid).
// Comply with the RFC7578 standard.
func (s service) UploadFile(
	ctx context.Context,
	params *FileParams,
) (string, error) {
	panic("not implemented")
}
