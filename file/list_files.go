package file

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/uploadcare/uploadcare-go/ucare"
)

// ListParams holds all possible params to for the ListFiles method
type ListParams struct {
	// Removed is set to true if only include removed files in the response,
	// otherwise existing files are included. Defaults to false.
	Removed *bool `form:"removed"`

	// Stored is set to true if only include files that were stored.
	// Set to false to include only temporary files.
	// The default is unset: both stored and not stored files are returned
	Stored *bool `form:"stored"`

	// Limit specifies preferred amount of files in a list for a single
	// response. Defaults to 100, while the maximum is 1000
	Limit *int64 `form:"limit"`

	// Ordering specifies the way files are sorted in a returned list.
	// By default is set to datetime_uploaded.
	Ordering *string `form:"ordering"`

	// From specifies a starting point for filtering files.
	// The value depends on your ordering parameter value.
	From *string `form:"from"`
}

// EncodeRequest implements ucare.RequestEncoder
func (d *ListParams) EncodeRequest(req *http.Request) {
	t, v := reflect.TypeOf(d).Elem(), reflect.ValueOf(d).Elem()
	q := req.URL.Query()
	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		if f.IsNil() {
			continue
		}

		var val string
		switch valc := f.Interface().(type) {
		case *string:
			val = ucare.StringVal(valc)
		case *int64:
			val = strconv.FormatInt(ucare.Int64Val(valc), 10)
		case *bool:
			val = fmt.Sprintf("%t", ucare.BoolVal(valc))
		}

		q.Set(t.Field(i).Tag.Get("form"), val)
	}
	req.URL.RawQuery = q.Encode()
}

// FileList is a paginated list of files
type FileList struct {
	NextPage string     `json:"next"`
	PrevPage string     `json:"previous"`
	Total    int64      `json:"total"`
	PerPage  int64      `json:"per_page"`
	Results  []FileInfo `json:"results"`
}

// ListFiles returns a paginated list of files
func (s service) ListFiles(
	ctx context.Context,
	params *ListParams,
) (*FileList, error) {
	if params == nil {
		params = &ListParams{}
	}

	url := ucare.SingleSlashJoin(
		ucare.RESTAPIEndpoint,
		listFilesPathFormat,
	)

	req, err := s.client.NewRequest(http.MethodGet, url, params)
	if err != nil {
		return nil, err
	}

	var flist FileList
	err = s.client.Do(req, &flist)
	if err != nil {
		return nil, err
	}

	log.Debugf("received file list: %+v", flist)

	return &flist, nil
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

	// Hight is image height in pixels
	Hight int64 `json:"height"`

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
