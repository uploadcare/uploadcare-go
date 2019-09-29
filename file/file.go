package file

import (
	"context"
	"net/http"

	"github.com/uploadcare/uploadcare-go/internal/config"
)

// Info acquires some file-specific info
func (s service) Info(
	ctx context.Context,
	fileID string,
) (data Info, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		infoPathFormat,
		fileID,
		&data,
	)
	return
}

// Store a single file by its id
func (s service) Store(
	ctx context.Context,
	fileID string,
) (data Info, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPut,
		storePathFormat,
		fileID,
		&data,
	)
	return
}

// Delete removes file by its id
func (s service) Delete(
	ctx context.Context,
	fileID string,
) (data Info, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodDelete,
		deletePathFormat,
		fileID,
		&data,
	)
	return
}

// Info holds file specific information
type Info struct {
	// RemovedAt is date and time when a file was removed, if any
	RemovedAt *config.Time `json:"datetime_removed"`

	// StoredAt is date and time of the last store request, if any
	StoredAt *config.Time `json:"datetime_stored"`

	// UploadedAt is a date and time when a file was uploaded
	UploadedAt *config.Time `json:"datetime_uploaded"`

	// ImageInfo holds image metadata
	ImageInfo *ImageInfo `json:"image_info"`

	// MimeType specifies file MIME-type
	MimeType string `json:"mime_type"`

	// OriginalFileURL is a publicly available file CDN URL.
	// Available if a file is not deleted
	OriginalFileURL string `json:"original_file_url"`

	// OriginalFileName is a file name taken from uploaded file
	OriginalFileName string `json:"original_filename"`

	// URI is a API resource URL for a file
	URI string `json:"uri"`

	// ID is a file unique id (UUID)
	ID string `json:"uuid"`

	// Size denotes file size in bytes
	Size uint64 `json:"size"`

	// IsImage denotes if a file is an image
	IsImage bool `json:"is_image"`

	// IsReady denotes if file is ready to be used after upload
	IsReady bool `json:"is_ready"`
}

// Image color mode contants
const (
	ImageColorModeRGB   = "RGB"
	ImageColorModeRGBA  = "RGBA"
	ImageColorModeRGBa  = "RGBa"
	ImageColorModeRGBX  = "RGBX"
	ImageColorModeL     = "L"
	ImageColorModeLA    = "LA"
	ImageColorModeLa    = "La"
	ImageColorModeP     = "P"
	ImageColorModePA    = "PA"
	ImageColorModeCMYK  = "CMYK"
	ImageColorModeYCbCr = "YCbCr"
	ImageColorModeHSV   = "HSV"
	ImageColorModeLAB   = "LAB"
)

// ImageInfo holds image-specific information
type ImageInfo struct {
	// ColorMode is image color mode
	ColorMode string `json:"color_mode"`

	// Format specifies image format
	Format string `json:"format"`

	// Height is image height in pixels
	Height uint64 `json:"height"`

	// Width is image width in pixels
	Width uint64 `json:"width"`

	// Orientation is image orientation from EXIF
	Orientation *int64 `json:"orientation"`

	// DPI specifies image DPI for two dimensions
	DPI []int64 `json:"dpi"`

	// GeoLocation is geo-location of image from EXIF
	GeoLocation *Location `json:"geo_location"`

	// DateTimeOriginal is image date and time from EXIF
	DateTimeOriginal *config.Time `json:"datetime_original"`

	// Sequence denotes if image is sequence image (GIF for example)
	Sequence bool `json:"sequence"`
}

// Location holds location coordinates
type Location struct {
	Latitude  int64 `json:"latitude"`
	Longitude int64 `json:"longitude"`
}
