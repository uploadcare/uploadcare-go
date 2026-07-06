package file

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/uploadcare/uploadcare-go/v2/internal/codec"
	"github.com/uploadcare/uploadcare-go/v2/internal/svc"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

type SearchSort string

const (
	SortByScore                SearchSort = "score"
	SortByScoreDesc            SearchSort = "-score"
	SortByUploadedAt           SearchSort = "datetime_uploaded"
	SortByUploadedAtDesc       SearchSort = "-datetime_uploaded"
	SortBySize                 SearchSort = "size"
	SortBySizeDesc             SearchSort = "-size"
	SortByOriginalFilename     SearchSort = "original_filename"
	SortByOriginalFilenameDesc SearchSort = "-original_filename"
)

const (
	SearchIncludeAppData = "appdata"
)

const (
	SearchExactKeyUUID             = "uuid"
	SearchExactKeyDetectedMimeType = "detected_mime_type"
	SearchExactKeyOriginalFilename = "original_filename"
)

// SearchParams holds params for Search. At least one search condition must be set.
type SearchParams struct {
	// Defaults to 20; must be between 1 and 100.
	Limit *uint64 `form:"limit" json:"-"`

	// Offset plus limit must not exceed 1000.
	Offset *uint64 `form:"offset" json:"-"`

	// Valid value is SearchIncludeAppData.
	Include *string `form:"include" json:"-"`

	// Must be at least 4 characters.
	Query string `json:"query,omitempty"`

	// Each set value must be at least 4 characters.
	Phrase *SearchPhrase `json:"phrase,omitempty"`

	// Supported keys are SearchExactKeyUUID,
	// SearchExactKeyDetectedMimeType, SearchExactKeyOriginalFilename and the
	// "metadata[<key>]" syntax for metadata.
	Exact map[string][]string `json:"exact,omitempty"`

	DatetimeUploaded *SearchDatetime `json:"datetime_uploaded,omitempty"`
	Size             *SearchSize     `json:"size,omitempty"`
	IsImage          *bool           `json:"is_image,omitempty"`

	// Defaults to false. Enabling it significantly increases query latency.
	Fuzziness bool        `json:"fuzziness,omitempty"`
	Tags      *SearchTags `json:"tags,omitempty"`

	// Accepts 1 to 4 sort keys.
	Sort []SearchSort `json:"sort,omitempty"`
}

type SearchPhrase struct {
	OriginalFilename string `json:"original_filename,omitempty"`
	Metadata         string `json:"metadata,omitempty"`
	DetectedMimeType string `json:"detected_mime_type,omitempty"`
}

// Set at least one operator. Times are serialized as RFC 3339.
type SearchDatetime struct {
	Gt  *time.Time `json:"gt,omitempty"`
	Gte *time.Time `json:"gte,omitempty"`
	Lt  *time.Time `json:"lt,omitempty"`
	Lte *time.Time `json:"lte,omitempty"`
}

// Set at least one operator.
type SearchSize struct {
	Gt  *uint64 `json:"gt,omitempty"`
	Gte *uint64 `json:"gte,omitempty"`
	Lt  *uint64 `json:"lt,omitempty"`
	Lte *uint64 `json:"lte,omitempty"`
}

type SearchTags struct {
	Any  []string `json:"any,omitempty"`
	All  []string `json:"all,omitempty"`
	None []string `json:"none,omitempty"`
}

// SearchMatch is a single file match returned by Search.
type SearchMatch struct {
	ID               string                     `json:"uuid"`
	OriginalFileName string                     `json:"original_filename"`
	Size             uint64                     `json:"size"`
	Highlight        *SearchHighlight           `json:"highlight,omitempty"`
	AppData          map[string]json.RawMessage `json:"appdata,omitempty"`
}

// SearchHighlight holds matched fragments with the matched tokens wrapped in
// <em> tags.
type SearchHighlight struct {
	OriginalFileName []string          `json:"original_filename,omitempty"`
	DetectedMimeType []string          `json:"detected_mime_type,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// searchReq drives both the query string and the JSON body of the first request.
type searchReq SearchParams

func (p searchReq) EncodeReq(req *http.Request) error {
	if err := codec.EncodeReqQuery(&p, req); err != nil {
		return err
	}
	return codec.EncodeReqBody(p, req)
}

type searchNextReq struct {
	query string
	body  SearchParams
}

func (p searchNextReq) EncodeReq(req *http.Request) error {
	req.URL.RawQuery = p.query
	return codec.EncodeReqBody(p.body, req)
}

type searchPage struct {
	Results []SearchMatch `json:"results"`
	Total   uint64        `json:"total"`
	Next    *string       `json:"next"`
}

// SearchResult iterates over the matches of a search, fetching subsequent pages
// on demand.
type SearchResult struct {
	svc    svc.Service
	ctx    context.Context
	params SearchParams

	nextQuery *string // query string for the next page request; nil when exhausted
	results   []SearchMatch
	at        int
	total     uint64
}

// Total returns the total number of matches reported by the API.
func (r *SearchResult) Total() uint64 { return r.total }

// Next indicates if there is a result to read
func (r *SearchResult) Next() bool { return r.at < len(r.results) || r.nextQuery != nil }

// ReadResult returns next SearchMatch value. If no results are left to read it
// returns ucare.ErrEndOfResults.
func (r *SearchResult) ReadResult() (*SearchMatch, error) {
	// A page may come back empty yet still carry a next pointer (for example
	// when stale matches were filtered out server-side), so keep following
	// next until a page has results or there is no next.
	for r.at >= len(r.results) {
		if r.nextQuery == nil {
			return nil, ucare.ErrEndOfResults
		}
		if err := r.fetch(searchNextReq{query: *r.nextQuery, body: r.params}); err != nil {
			return nil, err
		}
	}

	m := r.results[r.at]
	r.at++

	log.Debugf("reading search result: %+v", m)

	return &m, nil
}

func (r *SearchResult) fetch(params ucare.ReqEncoder) error {
	var page searchPage
	err := r.svc.ResourceOp(
		r.ctx,
		http.MethodPost,
		searchPathFormat,
		params,
		&page,
	)
	if err != nil {
		return err
	}

	r.results = page.Results
	r.at = 0
	r.total = page.Total
	r.nextQuery = nil
	if page.Next != nil {
		u, perr := url.Parse(*page.Next)
		if perr != nil {
			return perr
		}
		q := u.RawQuery
		r.nextQuery = &q
	}

	return nil
}

// Search returns matches for the given search params.
//
// Example usage:
//
//	results, err := fileSvc.Search(ctx, params)
//	if err != nil {
//		// handle error
//	}
//	for results.Next() {
//		match, err := results.ReadResult()
//		...
//	}
func (s service) Search(
	ctx context.Context,
	params SearchParams,
) (res *SearchResult, err error) {
	if err = validateSearchParams(params); err != nil {
		return
	}

	res = &SearchResult{svc: s.svc, ctx: ctx, params: params}
	err = res.fetch(searchReq(params))
	return
}
