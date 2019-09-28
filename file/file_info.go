package file

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/uploadcare/uploadcare-go/ucare"
)

// FileInfo acquires some file-specific info
func (s service) FileInfo(ctx context.Context, fileID string) (FileInfo, error) {
	if fileID == "" {
		return FileInfo{}, errors.New("empty file id provided")
	}

	method := http.MethodGet
	path := fmt.Sprintf(fileInfoPathFormat, fileID)
	url := ucare.RESTAPIEndpoint + path

	req, err := s.client.NewRequest(ctx, method, url, nil)
	if err != nil {
		return FileInfo{}, err
	}

	var finfo FileInfo
	err = s.client.Do(req, &finfo)

	log.Debugf("received file info: %+v", finfo)

	return finfo, err
}

type FileInfo struct {
	// RemovedAt is date and time when a file was removed, if any
	RemovedAt *ucare.Time `json:"datetime_removed"`

	// StoredAt is date and time of the last store request, if any
	StoredAt *ucare.Time `json:"datetime_stored"`

	// UploadedAt is a date and time when a file was uploaded
	UploadedAt *ucare.Time `json:"datetime_uploaded"`

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
	Size int64 `json:"size"`

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

type ImageInfo struct {
	// ColorMode is image color mode
	ColorMode string `json:"color_mode"`

	// Format specifies image format
	Format string `json:"format"`

	// Height is image height in pixels
	Height int64 `json:"height"`

	// Width is image width in pixels
	Width int64 `json:"width"`

	// Orientation is image orientation from EXIF
	Orientation *int64 `json:"orientation"`

	// DPI specifies image DPI for two dimensions
	DPI []int64 `json:"dpi"`

	// GeoLocation is geo-location of image from EXIF
	GeoLocation *Location `json:"geo_location"`

	// DateTimeOriginal is image date and time from EXIF
	DateTimeOriginal *ucare.Time `json:"datetime_original"`

	// Sequence denotes if image is sequence image (GIF for example)
	Sequence bool `json:"sequence"`
}

type Location struct {
	Latitude  int64 `json:"latitude"`
	Longitude int64 `json:"longitude"`
}
