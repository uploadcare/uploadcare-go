package upload

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/uploadcare/uploadcare-go/ucare"
)

// MultipartParams holds parameters for multipart upload
type MultipartParams struct {
	multipartAuthParams

	// FileName (required) is the original filename
	FileName string `form:"filename"`
	// Size (required) is a precise file size in bytes.
	// Should not exceed your project file size cap
	Size int64 `form:"size"`
	// ContentType (required) is the file MIME-type.
	ContentType string `form:"content_type"`

	// Data (required) reads the data to be uploaded
	Data io.ReadSeeker

	// ToStore sets the file storing behaviour
	// Valid values:
	//	upload.ToStoreTrue
	//	upload.ToStoreFalse
	//	upload.ToStoreAuto
	ToStore *string `form:"UPLOADCARE_STORE"`
}

type multipartAuthParams struct {
	PubKey string `form:"UPLOADCARE_PUB_KEY"`
	signatureExpire
}

// EncodeReq implements ucare.ReqEncoder
func (d *MultipartParams) EncodeReq(req *http.Request) error {
	d.PubKey, d.Signature, d.ExpiresAt = authFromContext(req.Context())()
	return encodeDataToForm(d, req)
}

// Multipart upload is useful when you are dealing with file larger than
// 100MB or explicitly want to use accelerated uploads.
// Another benefit is your file will go straight to AWS S3 bypassing our upload
// instances thus quickly becoming available for further use.
// Note, there also exists a minimum file size to use with Multipart Uploads, 10MB.
// Trying to use Multipart upload with a smaller file will result in an error.
func (s service) Multipart(
	ctx context.Context,
	params MultipartParams,
) (data MultipartData, err error) {
	if params.Data == nil {
		return nil, errors.New("nil data reader")
	}
	d := multipartData{
		ctx: ctx,

		data:        params.Data,
		contentType: params.ContentType,

		uploadPart:        s.uploadPart,
		completeMultipart: s.completeMultipart,

		done: make(chan FileInfo, 1),
		err:  make(chan error, 1),
	}

	if err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		multipartStartFormat,
		&params,
		&d,
	); err != nil {
		return nil, err
	}

	go d.uploadParts()

	return &d, err
}

// MultipartData holds response from Multipart upload request
type MultipartData interface {
	// Done is used to block and wait for upload to be finished
	Done() <-chan FileInfo
	// Error is used to listen for uploading error
	Error() <-chan error
}

// multipartData implements MultipartData
type multipartData struct {
	ctx context.Context

	data        io.ReadSeeker
	contentType string

	uploadPart        func(context.Context, string, ucare.ReqEncoder) error
	completeMultipart func(context.Context, string) (FileInfo, error)

	done chan FileInfo
	err  chan error

	ID    string   `json:"uuid"`
	Parts []string `json:"parts"`
}

// TODO: consider optimal values
const (
	concurrentUploads  = 5
	maxUploadPartTries = 2
	partSize           = 5242880 // 5MB
)

func (d *multipartData) uploadParts() {
	if d == nil || d.ID == "" {
		return
	}

	var wg sync.WaitGroup
	uploadSem := make(chan struct{}, concurrentUploads)

	for uploadSem != nil {
		select {
		case <-d.ctx.Done():
			err := d.ctx.Err()
			log.Errorf(
				"stopped uploading file: %s: %+v",
				d.ID,
				err,
			)
			if len(d.err) < cap(d.err) {
				d.err <- err
			}
			return
		case uploadSem <- struct{}{}:
			var partURL string
			if len(d.Parts) == 0 {
				uploadSem = nil
				continue
			}
			partURL, d.Parts = d.Parts[0], d.Parts[1:]

			part, err := d.partEncoder(partIndexFromURL(partURL))
			if err != nil {
				if len(d.err) < cap(d.err) {
					d.err <- err
				}
				return
			}

			wg.Add(1)
			go d.tryUploadPart(&wg, partURL, part)
		}
	}

	wg.Wait()

	fileInfo, err := d.completeMultipart(d.ctx, d.ID)
	if err != nil {
		log.Errorf("completing multipart upload: %s", err)
		if len(d.err) < cap(d.err) {
			d.err <- err
		}
		return
	}

	d.done <- fileInfo
	return
}

func (d *multipartData) tryUploadPart(
	wg *sync.WaitGroup,
	partURL string,
	part ucare.ReqEncoder,
) {
	defer wg.Done()

	tries := 0
try:
	tries++
	err := d.uploadPart(
		d.ctx,
		partURL,
		part,
	)
	if err != nil {
		if tries >= maxUploadPartTries {
			if len(d.err) < cap(d.err) {
				d.err <- err
			}
			return
		}
		goto try
	}
}

func (s service) uploadPart(
	ctx context.Context,
	url string,
	part ucare.ReqEncoder,
) error {
	return s.svc.ResourceOp(ctx, http.MethodPut, url, part, nil)
}

func (d *multipartData) partEncoder(partIndex int64) (ucare.ReqEncoder, error) {
	_, err := d.data.Seek(partIndex*partSize, 0)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, partSize)
	_, err = d.data.Read(buf)
	if err != nil {
		return nil, err
	}

	return partEncoder{buf, d.contentType}, nil
}

func partIndexFromURL(partURL string) int64 {
	u, _ := url.Parse(partURL)
	i, _ := strconv.ParseInt(u.Query().Get("partNumber"), 10, 64)
	if i < 1 {
		return 0
	}
	// first part is 1 but we need it to be 0
	return i - 1
}

type partEncoder struct {
	data        []byte
	contentType string
}

func (d partEncoder) EncodeReq(req *http.Request) error {
	req.Body = ioutil.NopCloser(bytes.NewReader(d.data))
	req.Header.Set("Content-Type", d.contentType)
	req.Header.Set("Content-Length", strconv.Itoa(len(d.data)))
	return nil
}

// Done is used to wait for uploading to be done. If uploading fails it will
// never receive FileInfo value:
//	select {
//	case fileinfo := <-res.Done():
//		// file info received
//	case err := <-res.Error():
//		// error received
//	}
func (d *multipartData) Done() <-chan FileInfo { return d.done }

// Error should be used to listen for errors inside of the select statement
func (d *multipartData) Error() <-chan error { return d.err }

type completeMultipartParams struct {
	PubKey string `form:"UPLOADCARE_PUB_KEY"`
	ID     string `form:"uuid"`
}

// EncodeReq implements ucare.ReqEncoder
func (d *completeMultipartParams) EncodeReq(req *http.Request) error {
	d.PubKey, _, _ = authFromContext(req.Context())()
	return encodeDataToForm(d, req)
}

func (s service) completeMultipart(
	ctx context.Context,
	id string,
) (data FileInfo, err error) {
	params := completeMultipartParams{ID: id}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		multipartCompleteFormat,
		&params,
		&data,
	)
	return
}
