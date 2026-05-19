package projectapi

import (
	"net/http"

	"github.com/uploadcare/uploadcare-go/v2/internal/codec"
)

type ListParams struct {
	Limit *uint64 `form:"limit"`
}

func (p *ListParams) EncodeReq(req *http.Request) error {
	return codec.EncodeReqQuery(p, req)
}

type CreateProjectParams struct {
	Name     string           `json:"name"`
	Features *ProjectFeatures `json:"features,omitempty"`
}

func (p CreateProjectParams) EncodeReq(req *http.Request) error {
	return codec.EncodeReqBody(p, req)
}

// Only non-nil fields are sent.
type UpdateProjectParams struct {
	Name                 *string          `json:"name,omitempty"`
	IsBlocked            *bool            `json:"is_blocked,omitempty"`
	IsSearchIndexAllowed *bool            `json:"is_search_index_allowed,omitempty"`
	Features             *ProjectFeatures `json:"features,omitempty"`
}

func (p UpdateProjectParams) EncodeReq(req *http.Request) error {
	return codec.EncodeReqBody(p, req)
}

// UsageDateRange dates use YYYY-MM-DD. The maximum range is 90 days.
type UsageDateRange struct {
	From string `form:"from"`
	To   string `form:"to"`
}

func (p *UsageDateRange) EncodeReq(req *http.Request) error {
	return codec.EncodeReqQuery(p, req)
}
