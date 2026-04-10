// Package addon is the Uploadcare Addons API (execute processing tasks on files).
package addon

import "encoding/json"

type Name string

const (
	AddonRemoveBG              Name = "remove_bg"
	AddonClamAV                Name = "uc_clamav_virus_scan"
	AddonRekognitionLabels     Name = "aws_rekognition_detect_labels"
	AddonRekognitionModeration Name = "aws_rekognition_detect_moderation_labels"
)

const (
	StatusInProgress = "in_progress"
	StatusDone       = "done"
	StatusError      = "error"
	StatusUnknown    = "unknown"
)

type ExecuteParams struct {
	Target string `json:"target"`
	Params any    `json:"params,omitempty"`
}

type ExecuteResult struct {
	RequestID string `json:"request_id"`
}

type StatusParams struct {
	RequestID string `json:"request_id"`
}

type StatusResult struct {
	Status  string          `json:"status"`
	Result  json.RawMessage `json:"result"`
	Details json.RawMessage `json:"details,omitempty"`
}

type RemoveBGParams struct {
	Crop             *bool   `json:"crop,omitempty"`
	CropMargin       *string `json:"crop_margin,omitempty"`
	Scale            *string `json:"scale,omitempty"`
	AddShadow        *bool   `json:"add_shadow,omitempty"`
	TypeLevel        *string `json:"type_level,omitempty"`
	Type             *string `json:"type,omitempty"`
	Semitransparency *bool   `json:"semitransparency,omitempty"`
	Channels         *string `json:"channels,omitempty"`
	ROI              *string `json:"roi,omitempty"`
	Position         *string `json:"position,omitempty"`
}

type ClamAVParams struct {
	PurgeInfected *bool `json:"purge_infected,omitempty"`
}
