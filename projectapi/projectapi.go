// Package projectapi provides bindings for the Uploadcare Project API.
//
// The Project API uses bearer token authentication. Create a client with
// ucare.NewBearerClient and pass it to NewService:
//
//	client, err := ucare.NewBearerClient("your-bearer-token", nil)
//	svc := projectapi.NewService(client)
//	list, err := svc.List(ctx, nil)
//	for list.Next() {
//		project, err := list.ReadResult()
//		// ...
//	}
package projectapi

import (
	"encoding/json"

	"github.com/uploadcare/uploadcare-go/v2/internal/codec"
)

// Project holds project information.
type Project struct {
	PubKey               string           `json:"pub_key,omitempty"`
	Name                 string           `json:"name,omitempty"`
	IsBlocked            *bool            `json:"is_blocked,omitempty"`
	IsSearchIndexAllowed *bool            `json:"is_search_index_allowed,omitempty"`
	IsSharedProject      bool             `json:"is_shared_project,omitempty"`
	Features             *ProjectFeatures `json:"features,omitempty"`
}

// ProjectFeatures holds per-project feature flags and settings.
type ProjectFeatures struct {
	GifToVideoConversion *FeatureToggle     `json:"gif_to_video_conversion,omitempty"`
	MimeTypeFiltering    *MimeTypeFiltering `json:"mime_type_filtering,omitempty"`
	TeamMembers          *TeamMembers       `json:"team_members,omitempty"`
	Uploads              *UploadSettings    `json:"uploads,omitempty"`
	VideoProcessing      *FeatureToggle     `json:"video_processing,omitempty"`
	MalwareProtection    *FeatureToggle     `json:"malware_protection,omitempty"`
	SVGValidation        *FeatureToggle     `json:"svg_validation,omitempty"`
}

// FeatureToggle is a simple feature with an is_enabled flag.
type FeatureToggle struct {
	IsEnabled *bool `json:"is_enabled,omitempty"`
}

// MimeTypeFiltering holds MIME-type upload filtering settings.
type MimeTypeFiltering struct {
	MimeTypes              []string `json:"mime_types,omitempty"`
	IsMimeFilteringEnabled *bool    `json:"is_mime_filtering_enabled,omitempty"`
}

// TeamMembers holds team membership info (read-only).
type TeamMembers struct {
	TeamSize int `json:"team_size,omitempty"`
}

// UploadSettings holds file upload settings for a project.
type UploadSettings struct {
	FilesizeLimit         *int64 `json:"filesize_limit"`
	Autostore             *bool  `json:"autostore,omitempty"`
	ImageResolutionLimit  *int64 `json:"image_resolution_limit"`
	IsSignedUploadEnabled *bool  `json:"is_signed_upload_enabled,omitempty"`
}

// ProjectList is a paginated iterator over projects.
// Use Next() and ReadResult() to iterate:
//
//	list, err := svc.List(ctx, nil)
//	for list.Next() {
//		project, err := list.ReadResult()
//		// ...
//	}
type ProjectList struct {
	raw codec.NextRawResulter
}

// Next indicates if there is a result to read.
func (v *ProjectList) Next() bool { return v.raw.Next() }

// ReadResult returns the next Project. If no results are left it
// returns codec.ErrEndOfResults.
func (v *ProjectList) ReadResult() (*Project, error) {
	raw, err := v.raw.ReadRawResult()
	if err != nil {
		return nil, err
	}
	var p Project
	err = json.Unmarshal(raw, &p)
	return &p, err
}

// SecretRevealed is returned when creating a new secret key.
// It contains the full secret value, which is only shown once.
type SecretRevealed struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

// SecretListItem is returned in secret key listings.
type SecretListItem struct {
	ID         string  `json:"id"`
	Hint       string  `json:"hint"`
	LastUsedAt *string `json:"last_used_at"`
}

// SecretList is a paginated iterator over secret keys.
type SecretList struct {
	raw codec.NextRawResulter
}

// Next indicates if there is a result to read.
func (v *SecretList) Next() bool { return v.raw.Next() }

// ReadResult returns the next SecretListItem. If no results are left it
// returns codec.ErrEndOfResults.
func (v *SecretList) ReadResult() (*SecretListItem, error) {
	raw, err := v.raw.ReadRawResult()
	if err != nil {
		return nil, err
	}
	var s SecretListItem
	err = json.Unmarshal(raw, &s)
	return &s, err
}

// UsageMetric holds daily usage data for a single metric type.
type UsageMetric struct {
	Metric string           `json:"metric"`
	Unit   string           `json:"unit"`
	Data   []UsageDataPoint `json:"data"`
}

// UsageDataPoint holds a single date/value pair in usage data.
type UsageDataPoint struct {
	Date  string `json:"date"`
	Value int64  `json:"value"`
}

// UsageMetricsCombined holds combined usage data for all metric types.
type UsageMetricsCombined struct {
	Units map[string]string        `json:"units"`
	Data  []CombinedUsageDataPoint `json:"data"`
}

// CombinedUsageDataPoint holds combined daily usage values.
type CombinedUsageDataPoint struct {
	Date       string `json:"date"`
	Traffic    int64  `json:"traffic"`
	Storage    int64  `json:"storage"`
	Operations int64  `json:"operations"`
}
