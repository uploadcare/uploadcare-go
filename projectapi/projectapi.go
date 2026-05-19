// Package projectapi is the Uploadcare Project API (projects, secret keys, and usage metrics).
package projectapi

import (
	"encoding/json"

	"github.com/uploadcare/uploadcare-go/v2/internal/codec"
)

type Project struct {
	PubKey               string           `json:"pub_key,omitempty"`
	Name                 string           `json:"name,omitempty"`
	IsBlocked            *bool            `json:"is_blocked,omitempty"`
	IsSearchIndexAllowed *bool            `json:"is_search_index_allowed,omitempty"`
	IsSharedProject      bool             `json:"is_shared_project,omitempty"`
	Features             *ProjectFeatures `json:"features,omitempty"`
}

type ProjectFeatures struct {
	GifToVideoConversion *FeatureToggle     `json:"gif_to_video_conversion,omitempty"`
	MimeTypeFiltering    *MimeTypeFiltering `json:"mime_type_filtering,omitempty"`
	TeamMembers          *TeamMembers       `json:"team_members,omitempty"`
	Uploads              *UploadSettings    `json:"uploads,omitempty"`
	VideoProcessing      *FeatureToggle     `json:"video_processing,omitempty"`
	MalwareProtection    *FeatureToggle     `json:"malware_protection,omitempty"`
	SVGValidation        *FeatureToggle     `json:"svg_validation,omitempty"`
}

type FeatureToggle struct {
	IsEnabled *bool `json:"is_enabled,omitempty"`
}

type MimeTypeFiltering struct {
	MimeTypes              []string `json:"mime_types,omitempty"`
	IsMimeFilteringEnabled *bool    `json:"is_mime_filtering_enabled,omitempty"`
}

type TeamMembers struct {
	TeamSize int `json:"team_size,omitempty"`
}

type UploadSettings struct {
	FilesizeLimit         *int64 `json:"filesize_limit,omitempty"`
	Autostore             *bool  `json:"autostore,omitempty"`
	ImageResolutionLimit  *int64 `json:"image_resolution_limit,omitempty"`
	IsSignedUploadEnabled *bool  `json:"is_signed_upload_enabled,omitempty"`
}

type ProjectList struct {
	raw codec.NextRawResulter
}

func (v *ProjectList) Next() bool { return v.raw.Next() }

func (v *ProjectList) ReadResult() (*Project, error) {
	raw, err := v.raw.ReadRawResult()
	if err != nil {
		return nil, err
	}
	var p Project
	err = json.Unmarshal(raw, &p)
	return &p, err
}

// SecretRevealed contains the full secret value, which is only shown once.
type SecretRevealed struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

type SecretListItem struct {
	ID         string  `json:"id"`
	Hint       string  `json:"hint"`
	LastUsedAt *string `json:"last_used_at"`
}

type SecretList struct {
	raw codec.NextRawResulter
}

func (v *SecretList) Next() bool { return v.raw.Next() }

func (v *SecretList) ReadResult() (*SecretListItem, error) {
	raw, err := v.raw.ReadRawResult()
	if err != nil {
		return nil, err
	}
	var s SecretListItem
	err = json.Unmarshal(raw, &s)
	return &s, err
}

type UsageMetric struct {
	Metric UsageMetricName  `json:"metric"`
	Unit   string           `json:"unit"`
	Data   []UsageDataPoint `json:"data"`
}

type UsageDataPoint struct {
	Date  string `json:"date"`
	Value int64  `json:"value"`
}

type UsageMetricsCombined struct {
	Units map[string]string        `json:"units"`
	Data  []CombinedUsageDataPoint `json:"data"`
}

type CombinedUsageDataPoint struct {
	Date       string `json:"date"`
	Traffic    int64  `json:"traffic"`
	Storage    int64  `json:"storage"`
	Operations int64  `json:"operations"`
}
