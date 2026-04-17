package conversion

import (
	"context"

	"github.com/uploadcare/uploadcare-go/v2/internal/config"
	"github.com/uploadcare/uploadcare-go/v2/internal/svc"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

// Service describes all conversion related API
type Service interface {
	Document(context.Context, Params) (Result, error)
	DocumentStatus(context.Context, int64) (StatusResult, error)

	Video(context.Context, Params) (Result, error)
	VideoStatus(context.Context, int64) (StatusResult, error)
}

type service struct {
	svc svc.Service
}

// NewService returns new instance of the Service
func NewService(client ucare.Client) Service {
	return service{svc.New(config.RESTAPIEndpoint, client, log)}
}

const (
	convertDocumentFormat       = "/convert/document/"
	convertDocumentStatusFormat = "/convert/document/status/%d/"
	convertVideoFormat          = "/convert/video/"
	convertVideoStatusFormat    = "/convert/video/status/%d/"
)

// Predefined constants
const (
	ToStoreTrue  = "1"
	ToStoreFalse = "0"
)

// ResizeMode defines how a video is resized to fit the target dimensions.
type ResizeMode string

const (
	ResizeModePreserveRatio ResizeMode = "preserve_ratio"
	ResizeModeChangeRatio   ResizeMode = "change_ratio"
	ResizeModeScaleCrop     ResizeMode = "scale_crop"
	ResizeModeAddPadding    ResizeMode = "add_padding"
)

// Quality defines the output quality for video conversion.
type Quality string

const (
	QualityNormal  Quality = "normal"
	QualityBetter  Quality = "better"
	QualityBest    Quality = "best"
	QualityLighter Quality = "lighter"
	QualityLightest Quality = "lightest"
)
