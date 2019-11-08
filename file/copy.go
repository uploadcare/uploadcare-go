package file

import (
	"context"
	"net/http"

	"github.com/uploadcare/uploadcare-go/internal/codec"
	"github.com/uploadcare/uploadcare-go/ucare"
)

// CopyParams is used when copy original files or their modified
// versions to default storage. Source files MAY either be stored or just
// uploaded and MUST NOT be deleted
type CopyParams struct {
	LocalCopyParams

	// Target identifies a custom storage name related to your project.
	// Implies you are copying a file to a specified custom storage. Keep in
	// mind you can have multiple storages associated with a single S3
	// bucket.
	Target *string `json:"target"`

	// Pattern is used to specify file names Uploadcare passes to a custom
	// storage. In case the parameter is omitted, we use pattern of your
	// custom storage. Use any combination of allowed values:
	//	file.PatternDefault      = ${uuid}/${auto_filename}
	//	file.PatternAutoFileName = ${filename} ${effects} ${ext}
	//	file.PatternEffects      = processing operations put into a CDN URL
	//	file.PatternFileName     = original filename, no extension
	//	file.PatternID           = file UUID
	//	file.PatternExt          = file extension, leading dot, e.g. .jpg
	Pattern *string `json:"pattern"`
}

// LocalCopyParams is used when copy original files or their modified
// versions to default storage
type LocalCopyParams struct {
	// Source is a CDN URL or just ID (UUID) of a file subjected to copy
	Source string `json:"source"`
	// Store parameter only applies to the Uploadcare storage and MUST
	// be either true or false.
	// Valid values:
	//	file.StoreTrue
	//	file.StoreFalse
	Store *string `json:"store"`
	// MakePublic is applicable to custom storage only. MUST be either true or
	// false. True to make copied files available via public links, false to
	// reverse the behavior.
	// Valid values:
	//	file.MakePublicTrue
	//	file.MakePublicFalse
	MakePublic *string `json:"make_public"`
}

// EncodeReq implements ucare.ReqEncoder
func (d *LocalCopyParams) EncodeReq(req *http.Request) error {
	return codec.EncodeReqBody(d, req)
}

// Copy is the APIv05 version of the LocalCopy and RemoteCopy, use them instead
func (s service) Copy(
	ctx context.Context,
	params CopyParams,
) (data LocalCopyInfo, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		copyPathFormat,
		&params,
		&data,
	)
	return
}

// LocalCopy is used to copy original files or their modified versions to
// default storage. Source files MAY either be stored or just uploaded and MUST
// NOT be deleted
func (s service) LocalCopy(
	ctx context.Context,
	params LocalCopyParams,
) (data LocalCopyInfo, err error) {
	if params.Store == nil {
		params.Store = ucare.String(StoreFalse)
	}
	if params.MakePublic == nil {
		params.MakePublic = ucare.String(MakePublicTrue)
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		localCopyPathFormat,
		&params,
		&data,
	)
	return
}

// LocalCopyInfo holds LocalCopy response data
type LocalCopyInfo struct {
	Result Info `json:"result"`
}

// RemoteCopyParams is used when copy original files or their modified
// versions to a custom storage
type RemoteCopyParams struct {
	// Source is a CDN URL or just UUID of a file subjected to copy
	Source string `json:"source"`

	// Target identifies a custom storage name related to your project.
	// Implies you are copying a file to a specified custom storage. Keep in
	// mind you can have multiple storages associated with a single S3
	// bucket.
	Target string `json:"target"`

	// MakePublic is applicable to custom storage only. MUST be either true or
	// false. True to make copied files available via public links, false to
	// reverse the behavior.
	// Valid values:
	//	file.MakePublicTrue
	//	file.MakePublicFalse
	MakePublic *string `json:"make_public"`

	// Pattern is used to specify file names Uploadcare passes to a custom
	// storage. In case the parameter is omitted, we use pattern of your
	// custom storage. Use any combination of allowed values:
	//	file.PatternDefault      = ${uuid}/${auto_filename}
	//	file.PatternAutoFileName = ${filename} ${effects} ${ext}
	//	file.PatternEffects      = processing operations put into a CDN URL
	//	file.PatternFileName     = original filename, no extension
	//	file.PatternID           = file UUID
	//	file.PatternExt          = file extension, leading dot, e.g. .jpg
	Pattern *string `json:"pattern"`
}

// EncodeReq implements ucare.ReqEncoder
func (d RemoteCopyParams) EncodeReq(req *http.Request) error {
	return codec.EncodeReqBody(d, req)
}

// RemoteCopy is used to copy original files or their modified versions to a custom
// storage. Source files MAY either be stored or just uploaded and MUST NOT be
// deleted.
func (s service) RemoteCopy(
	ctx context.Context,
	params RemoteCopyParams,
) (data RemoteCopyInfo, err error) {
	if params.MakePublic == nil {
		params.MakePublic = ucare.String(MakePublicTrue)
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		remoteCopyPathFormat,
		params,
		&data,
	)
	if data.Result == nil {
		data.AlreadyExists = true
	}
	return
}

// RemoteCopyInfo holds RemoteCopy response data
type RemoteCopyInfo struct {
	// AlreadyExists is true if destination file with that name
	// already exists
	AlreadyExists bool
	// Result is a URL with the s3 scheme. Your bucket name is put
	//  as a host, and an s3 object path follows
	Result *string `json:"result"`
}
