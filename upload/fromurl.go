package upload

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/uploadcare/uploadcare-go/ucare"
)

// FromURLParams holds parameters for upload from public URL link
type FromURLParams struct {
	fromURLAuthParams

	// URL is a file URL, which should be a public HTTP or HTTPS link
	URL string `form:"source_url"`

	// ToStore sets the file storing behaviour
	// Valid values are:
	//	upload.ToStoreTrue
	//	upload.ToStoreFalse
	//      upload.ToStoreAuto
	ToStore *string `form:"store"`

	// Name sets the name for a file uploaded from URL. If not defined, the
	// filename is obtained from either response headers or a source URL
	Name *string `form:"filename"`

	// CheckURLDuplicates runs the duplicate check and provides the
	// immediate-download behavior
	// Valid values:
	//	upload.URLDuplicatesTrue
	//	upload.URLDuplicatesFalse
	CheckURLDuplicates *string `form:"check_URL_duplicates"`

	// SaveURLDuplicates provides the save/update URL behavior. The
	// parameter can be used if you believe a source_url will be used more
	// than once. If you don’t explicitly defined, it is
	// by default set to the value of CheckURLDuplicates
	// Valid values:
	//	upload.URLDuplicatesTrue
	//	upload.URLDuplicatesFalse
	SaveURLDuplicates *string `form:"save_URL_duplicates"`
}

type fromURLAuthParams struct {
	PubKey string `form:"pub_key"`
	signatureExpire
}

// EncodeReq implements ucare.ReqEncoder
func (d *FromURLParams) EncodeReq(req *http.Request) error {
	d.PubKey, d.Signature, d.ExpiresAt = authFromContext(req.Context())()
	return encodeDataToForm(d, req)
}

// FromURL uploads file by its public URL.
//
// Usage example
//
//	params := &upload.FromURLParams{
//		URL: "https://bit.ly/2LJ2xOf",
//	}
//	res, err := uploadSvc.FromURL(ctx, params)
//	if err != nil {
//		// handle error
//	}
//
// Blocking goroutine while waiting for the file to be uploaded:
//
//	info, ok := res.Info()
//	if !ok {
//		// block here until done or error
//		select {
//		case info = <-res.Done():
//		case err = <-res.Error():
//		}
//	}
//	if err != nil {
//		// handle error uploading a file
//	}
//	fmt.Printf("file uploaded: %s", info.FileName)
//
// Not blocking waiting and tracking progress:
//
//	info, ok := res.FileInfo()
//	if !ok {
//		// separate waiting goroutine
//		go func() {
//			total := res.TotalSize()
//			for {
//				select {
//				case done := <-res.Progress():
//					fmt.Printf(
//						"uploaded %d/%d \n",
//						done,
//						total,
//					)
//				case info = <-res.Done():
//					return
//				case err = <-res.Error():
//					return
//				case <-ctx.Done():
//					// context cancelled or deadline reached
//					return
//				}
//			}
//		}()
//	}
//	...
func (s service) FromURL(
	ctx context.Context,
	params FromURLParams,
) (FromURLData, error) {
	data := fromURLData{
		ctx:           ctx,
		once:          &sync.Once{},
		fromURLStatus: s.fromURLStatus,
	}

	if params.ToStore == nil {
		params.ToStore = ucare.String(ToStoreAuto)
	}

	err := s.svc.ResourceOp(
		ctx,
		http.MethodPost,
		fromURLFormat,
		&params,
		&data,
	)
	return &data, err
}

// FromURLData is data type for handling uploads from url
type FromURLData interface {
	// Info returns uploaded file specific information.
	// If it was not uplaoded it returns (&FileInfo{}, false).
	// The caller is responsible for handling case when upload is not
	// done yet.
	Info() (FileInfo, bool)
	// Done is used to block and wait for upload to be finished
	Done() <-chan FileInfo
	// Progress is used to track uploading progress
	Progress() <-chan uint64
	// Error is used to listen for uploading error. If error is
	// received the background `wait` goroutine is terminated
	Error() <-chan error
	// TotalSize returns the total uploaded file size
	TotalSize() uint64
}

// fromURLData implements FromURLData
type fromURLData struct {
	ctx context.Context

	once     *sync.Once
	progress chan uint64
	done     chan FileInfo
	err      chan error

	fromURLStatus func(context.Context, string) (*fromURLStatusData, error)

	Type  *string `json:"type"`
	Token *string `json:"token"`
	*FileInfo
}

// FileInfo returns file info if uploading is done and otherwise nil
func (d *fromURLData) Info() (FileInfo, bool) {
	if d.Token != nil || d.FileInfo == nil {
		d.once.Do(func() {
			d.done = make(chan FileInfo, 1)
			d.progress = make(chan uint64, fromURLChanBuf)
			d.err = make(chan error, fromURLChanBuf)
			go d.wait()
		})
		return FileInfo{}, false
	}
	return *d.FileInfo, true
}

// Done is used to wait for uploading to be done. If uploading fails it will
// never receive FileInfo value:
//	select {
//	case fileinfo := <-res.Done():
//		// file info received
//	case err := <-res.Error():
//		// error received
//	}
func (d *fromURLData) Done() <-chan FileInfo { return d.done }

// Progress returns channel for tracking uploading progress
func (d *fromURLData) Progress() <-chan uint64 { return d.progress }

// Error should be used to listen for errors inside of the select statement
func (d *fromURLData) Error() <-chan error { return d.err }

// TODO: consider smaller buf size
const fromURLChanBuf = 10

func (d *fromURLData) wait() {
	if d == nil || d.Token == nil {
		return
	}
	for {
		select {
		case <-d.ctx.Done():
			err := d.ctx.Err()
			if len(d.err) < cap(d.err) {
				d.err <- err
			}
			log.Errorf(
				"stopped waiting for the file: %s: %+v",
				d.Token,
				err,
			)
			return
		case <-time.After(3 * time.Second):
			data, err := d.fromURLStatus(d.ctx, *d.Token)
			if err != nil {
				if len(d.err) < cap(d.err) {
					d.err <- err
				}
				return
			}

			log.Debugf(
				"checking file upload status: %s: %+v",
				*d.Token,
				data,
			)

			if data.Status == uploadStatusError {
				if len(d.err) < cap(d.err) {
					d.err <- errors.New(data.Error)
				}
				return
			}
			if data.Status == uploadStatusInProgress {
				if len(d.progress) < cap(d.progress) {
					d.progress <- data.Done
				}
				continue
			}
			d.done <- *data.FileInfo
			return
		}
	}
}

// TotalSize returns total file size to be uploaded
func (d *fromURLData) TotalSize() uint64 {
	if d.FileInfo == nil {
		return 0
	}
	return d.Total
}

type fromURLStatusData struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	*FileInfo
}

func (s service) fromURLStatus(
	ctx context.Context,
	token string,
) (data *fromURLStatusData, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		fmt.Sprintf(fromURLStatusFormat, token),
		nil,
		&data,
	)
	return
}
