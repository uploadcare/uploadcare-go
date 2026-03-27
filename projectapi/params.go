package projectapi

import (
	"net/http"

	"github.com/uploadcare/uploadcare-go/v2/internal/codec"
)

// ListParams holds optional parameters for listing projects or secrets.
type ListParams struct {
	Limit *uint64 `form:"limit"`
}

// EncodeReq implements ucare.ReqEncoder.
func (p *ListParams) EncodeReq(req *http.Request) error {
	return codec.EncodeReqQuery(p, req)
}

// CreateProjectParams holds parameters for creating a new project.
type CreateProjectParams struct {
	Name     string           `json:"name"`
	Features *ProjectFeatures `json:"features,omitempty"`
}

// EncodeReq implements ucare.ReqEncoder.
func (p CreateProjectParams) EncodeReq(req *http.Request) error {
	return codec.EncodeReqBody(p, req)
}

// UpdateProjectParams holds parameters for updating project settings.
// Only non-nil fields are sent.
type UpdateProjectParams struct {
	Name                 *string          `json:"name,omitempty"`
	IsBlocked            *bool            `json:"is_blocked,omitempty"`
	IsSearchIndexAllowed *bool            `json:"is_search_index_allowed,omitempty"`
	Features             *ProjectFeatures `json:"features,omitempty"`
}

// EncodeReq implements ucare.ReqEncoder.
func (p UpdateProjectParams) EncodeReq(req *http.Request) error {
	return codec.EncodeReqBody(p, req)
}

// UsageDateRange holds required date range parameters for usage queries.
// Dates must be in ISO 8601 format (YYYY-MM-DD). Maximum range is 90 days.
type UsageDateRange struct {
	From string `form:"from"`
	To   string `form:"to"`
}

// EncodeReq implements ucare.ReqEncoder.
func (p *UsageDateRange) EncodeReq(req *http.Request) error {
	return codec.EncodeReqQuery(p, req)
}
