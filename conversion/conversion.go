package conversion

import (
	"context"
	"fmt"
	"net/http"

	"github.com/uploadcare/uploadcare-go/internal/codec"
)

// Params holds conversion job params
type Params struct {
	// Paths is an array of IDs (UUIDs) of your source documents to convert
	// together with the specified target format.
	// Here is how it should be specified:
	//	:uuid/document/-/format/:target-format/
	//
	// You can also provide a complete CDN URL. It can then be used as an
	// alias to your converted file ID (UUID):
	//	https://ucarecdn.com/:uuid/document/-/format/:target-format/
	//
	// :uuid identifies the source file you want to convert, it should be
	// followed by /document/, otherwise, your request will return an error.
	// /-/ is a necessary delimiter that helps our API tell file identifiers
	// from processing operations.
	//
	// The following operations are available during conversion:
	//	/format/:target-format/ defines the target format you want a source
	// file converted to. The supported values for :target-format are: doc,
	// docx, xls, xlsx, odt, ods, rtf, txt, pdf (default), jpg, png. In case
	// the /format/ operation was not found, your input document will be
	// converted to pdf. Note, when converting multi-page documents to image
	// formats (jpg or png), your output will be a zip archive holding a number
	// of images corresponding to the input page count.
	//	/page/:number/ converts a single page of a multi-paged document to
	// either jpg or png. The method will not work for any other target
	// formats. :number stands for the one-based number of a page to convert.
	Paths []string `json:"paths"`

	// ToStore is the flag indicating if we should store your outputs.
	// Valid values:
	//	conversion.ToStoreTrue
	//	conversion.ToStoreFalse
	ToStore *string `json:"store"`
}

// EncodeReq implements ucare.ReqEncoder
func (d *Params) EncodeReq(req *http.Request) error {
	return codec.EncodeReqBody(d, req)
}

// Document starts document conversion job
func (s service) Document(
	ctx context.Context,
	params Params,
) (data Result, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		convertDocumentFormat,
		&params,
		&data,
	)
	return
}

// Result holds conversion job request result
type Result struct {
	// Problems related to your processing job, if any. Key is the
	// path you requested
	Problems map[string]string `json:"problems"`
	// Jobs holds result for each requested path, in case of no
	// errors for that path.
	Jobs []Job `json:"result"`
}

// Job holds conversion job info
type Job struct {
	// OriginalSource is a source file identifier including a target format,
	// if present
	OriginalSource string `json:"original_source"`
	// ID is a UUID of your converted document
	ID string `json:"uuid"`
	// Token is a conversion job token that can be used to get a job status
	Token int64 `json:"token"`
	// ThumbnailsGroupID is a UUID of a file group with thumbnails
	// for an output video, based on the `thumbs` operation parameters
	ThumbnailsGroupID *string `json:"thumbnails_group_id"`
}

// DocumentStatus gets document conversion job status
func (s service) DocumentStatus(
	ctx context.Context,
	token int64,
) (data StatusResult, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		fmt.Sprintf(convertDocumentStatusFormat, token),
		nil,
		&data,
	)
	return
}

// StatusResult holds conversion job status request result
type StatusResult struct {
	// Status holds conversion job status, can be one of the following:
	// - pending — a source file is being prepared for conversion.
	// - processing — conversion is in progress.
	// - finished — the conversion is finished.
	// - failed — we failed to convert the source, see error for details.
	// - canceled — the conversion was canceled.
	Status string `json:"status"`
	// Error holds a conversion error if we were unable to handle your file
	Error *string `json:"error"`
	// Result repeats the contents of your processing output
	Result struct {
		// ID is the uuid of a converted target file
		ID string `json:"uuid"`
		// ThumbnailsGroupID is a UUID of a file group with thumbnails
		// for an output video, based on the `thumbs` operation parameters
		ThumbnailsGroupID *string `json:"thumbnails_group_id"`
	} `json:"result"`
}

// Video starts video conversion job
func (s service) Video(
	ctx context.Context,
	params Params,
) (data Result, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		convertVideoFormat,
		&params,
		&data,
	)
	return
}

// VideoStatus gets video conversion job status
func (s service) VideoStatus(
	ctx context.Context,
	token int64,
) (data StatusResult, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		fmt.Sprintf(convertVideoStatusFormat, token),
		nil,
		&data,
	)
	return
}
