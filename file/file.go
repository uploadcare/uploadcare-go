package file

import (
	"context"
	"fmt"
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
		fmt.Sprintf(infoPathFormat, fileID),
		nil,
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
		fmt.Sprintf(storePathFormat, fileID),
		nil,
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
		fmt.Sprintf(deletePathFormat, fileID),
		nil,
		&data,
	)
	return
}

// BasicFileInfo holds common file information no matter what is the context
type BasicFileInfo struct {
	// ID is a file unique id (UUID)
	ID string `json:"uuid"`

	// ImageInfo holds image metadata
	ImageInfo *ImageInfo `json:"image_info"`

	// VideoMeta holds video metadata
	VideoMeta *VideoMeta `json:"video_info"`

	// MimeType specifies file MIME-type
	MimeType string `json:"mime_type"`

	// OriginalFileName is a file name taken from uploaded file
	OriginalFileName string `json:"original_filename"`

	// Size denotes file size in bytes
	Size uint64 `json:"size"`

	// IsImage denotes if a file is an image
	IsImage bool `json:"is_image"`

	// IsReady denotes if file is ready to be used after upload
	IsReady bool `json:"is_ready"`
}

// Info holds file specific information
type Info struct {
	BasicFileInfo

	// RemovedAt is date and time when a file was removed, if any
	RemovedAt *config.Time `json:"datetime_removed"`

	// StoredAt is date and time of the last store request, if any
	StoredAt *config.Time `json:"datetime_stored"`

	// UploadedAt is a date and time when a file was uploaded
	UploadedAt *config.Time `json:"datetime_uploaded"`

	// OriginalFileURL is a publicly available file CDN URL.
	// Available if a file is not deleted
	OriginalFileURL *string `json:"original_file_url"`

	// URL is a API resource URL for a file
	URL string `json:"url"`

	// Source is a file upload source. This field contains information
	// about from where file was uploaded, for example: facebook, gdrive,
	// gphotos, etc
	Source *string `json:"source"`

	// Variatios is a dictionary of other files that has been created using
	// this file as source. Used for video, document and etc. conversion
	Variations *map[string]string `json:"variations"`

	// RecognitionInfo is a dictionary of file categories with it"s
	// confidence
	RecognitionInfo map[string]float64 `json:"rekognition_info"`
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
	// could be *string or *int64
	Orientation interface{} `json:"orientation"`

	// DPI specifies image DPI for two dimensions
	DPI []float64 `json:"dpi"`

	// GeoLocation is geo-location of image from EXIF
	GeoLocation *Location `json:"geo_location"`

	// DateTimeOriginal is image date and time from EXIF
	DateTimeOriginal *string `json:"datetime_original"`

	// Sequence denotes if image is sequence image (GIF for example)
	Sequence bool `json:"sequence"`
}

// VideoMeta holds video metadata
type VideoMeta struct {
	// Duration is a video duration in milliseconds
	Duration uint64 `json:"duration"`

	// Format is a video format(MP4 for example)
	Format string `json:"format"`

	// Bitrate is a video bitrate
	Bitrate uint64 `json:"bitrate"`

	// Audio holds audio stream metadata
	Audio *AudioStreamMeta `json:"audio"`

	// Video holds video stream metadata
	Video *VideoStreamMeta `json:"video"`
}

// AudioStreamMeta holds audio stream metadata
type AudioStreamMeta struct {
	// Bitrate holds audio bitrate
	Bitrate *uint64 `json:"bitrate"`

	// Codec holds audio stream codec
	Codec *string `json:"codec"`

	// SampleRate holds audio stream sample rate
	SampleRate *uint64 `json:"sample_rate"`

	// Channels holds audio stream number of channels
	Channels *string `json:"channels"`
}

// VideoStreamMeta holds video stream metadata
type VideoStreamMeta struct {
	// Height is video stream image height
	Height uint64 `json:"height"`

	// Width is a video stream image width
	Width uint64 `json:"width"`

	// FrameRate is a video stream frame rate
	FrameRate float64 `json:"frame_rate"`

	// Bitrate holds video bitrate
	Bitrate *uint64 `json:"bitrate"`

	// Codec holds video stream codec
	Codec *string `json:"codec"`
}

// Location holds location coordinates
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
