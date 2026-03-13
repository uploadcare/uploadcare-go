// Package addon holds primitives and logic for the Uploadcare Addons API.
//
// Addons allow executing various processing tasks on uploaded files,
// such as background removal, virus scanning, and image recognition.
package addon

import "encoding/json"

// Addon name constants used in URL paths
const (
	AddonRemoveBG               = "remove_bg"
	AddonClamAV                 = "uc_clamav_virus_scan"
	AddonRekognitionLabels      = "aws_rekognition_detect_labels"
	AddonRekognitionModeration  = "aws_rekognition_detect_moderation_labels"
)

// Execution status constants
const (
	StatusInProgress = "in_progress"
	StatusDone       = "done"
	StatusError      = "error"
	StatusUnknown    = "unknown"
)

// ExecuteParams holds parameters for executing an addon
type ExecuteParams struct {
	// Target is the file UUID to process
	Target string `json:"target"`

	// Params holds addon-specific parameters.
	// Use RemoveBGParams, ClamAVParams, or nil depending on the addon.
	Params interface{} `json:"params,omitempty"`
}

// ExecuteResult holds the response from an addon execution request
type ExecuteResult struct {
	// RequestID is the unique identifier for this execution request
	RequestID string `json:"request_id"`
}

// StatusParams holds parameters for checking addon execution status
type StatusParams struct {
	// RequestID is the execution request ID returned by Execute
	RequestID string `json:"request_id"`
}

// StatusResult holds the response from an addon status check
type StatusResult struct {
	// Status is the current execution status
	Status string `json:"status"`

	// Result holds addon-specific result data.
	// The structure varies per addon.
	Result json.RawMessage `json:"result"`
}

// RemoveBGParams holds parameters for the remove.bg addon
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

// ClamAVParams holds parameters for the ClamAV virus scan addon
type ClamAVParams struct {
	PurgeInfected *bool `json:"purge_infected,omitempty"`
}
