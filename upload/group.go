package upload

import (
	"context"
	"fmt"
	"net/http"

	"github.com/uploadcare/uploadcare-go/group"
	"github.com/uploadcare/uploadcare-go/internal/codec"
)

type createGroupParams struct {
	groupAuthParams

	FileIDs map[string]string
}

type groupAuthParams struct {
	PubKey string `form:"pub_key"`
	signatureExpire
}

// EncodeReq implementes ucare.ReqEncoder
func (d *createGroupParams) EncodeReq(req *http.Request) error {
	d.PubKey, d.Signature, d.ExpiresAt = authFromContext(req.Context())()
	return encodeDataToForm(d, req)
}

// CreateGroup creates files group from a set of files by using their IDs with
// or without applied CDN media processing operations.
//
// Example:
//	[
//		"d6d34fa9-addd-472c-868d-2e5c105f9fcd",
//		"b1026315-8116-4632-8364-607e64fca723/-/resize/x800/",
//	]
func (s service) CreateGroup(
	ctx context.Context,
	ids []string,
) (info GroupInfo, err error) {
	idmap := make(map[string]string, len(ids))
	for i, id := range ids {
		idmap[fmt.Sprintf("files[%d]", i)] = id
	}
	params := createGroupParams{FileIDs: idmap}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		createGroupFormat,
		&params,
		&info,
	)

	log.Debugf("created group: %+v", info)

	return
}

// GroupInfo holds group specific info
type GroupInfo struct {
	group.Info

	// CDNLink is a CDN url of the group
	CDNLink string `json:"cdn_url"`

	// APILink is the API url to get this info
	APILink string `json:"uri"`

	// Files are objects holding your file info.
	//
	// CDN transformations that were present in the request params
	// can be found in the DefaultEffects field.
	Files []FileInfo `json:"files"`
}

type groupInfoParams struct {
	PubKey string `form:"pub_key"`

	// ID is a group ID. It look like UUID~N
	ID string `form:"group_id"`
}

// EncodeReq implementes ucare.ReqEncoder
func (d *groupInfoParams) EncodeReq(req *http.Request) error {
	d.PubKey, _, _ = authFromContext(req.Context())()
	return codec.EncodeReqQuery(d, req)
}

// GroupInfo returns group specific info.
//
// GroupID look like UUID~N, for example:
//	"d52d7136-a2e5-4338-9f45-affbf83b857d~2"
func (s service) GroupInfo(
	ctx context.Context,
	groupID string,
) (info GroupInfo, err error) {
	params := groupInfoParams{
		ID: groupID,
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		groupInfoFormat,
		&params,
		&info,
	)
	return
}
