package upload

import (
	"context"
	"io"
)

// DefaultMultipartThreshold is the default file size threshold for switching
// from direct upload to multipart upload. Files larger than this value
// will use multipart upload.
const DefaultMultipartThreshold int64 = 10 * 1024 * 1024 // 10MB

// UploadParams holds parameters for the unified Upload method.
type UploadParams struct {
	// Data (required) reads the data to be uploaded.
	Data io.ReadSeeker
	// Name (required) is the original filename.
	Name string
	// ContentType is the file MIME-type.
	ContentType string
	// Size (required) is the precise file size in bytes.
	Size int64
	// ToStore sets the file storing behaviour.
	ToStore *string
	// Metadata stores user-defined key-value pairs with the uploaded file.
	Metadata map[string]string
	// MultipartThreshold controls the upload method selection:
	//   nil    → use DefaultMultipartThreshold (10MB)
	//   > 0   → use as custom threshold
	//   0     → force direct upload
	//   < 0   → force multipart upload
	MultipartThreshold *int64
}

// Upload automatically selects direct or multipart upload based on file size
// and the configured threshold. It returns a FileInfo for the uploaded file.
func (s service) Upload(ctx context.Context, params UploadParams) (FileInfo, error) {
	threshold := DefaultMultipartThreshold
	if params.MultipartThreshold != nil {
		threshold = *params.MultipartThreshold
	}

	useMultipart := false
	if threshold < 0 {
		useMultipart = true
	} else if threshold == 0 {
		useMultipart = false
	} else {
		useMultipart = params.Size > threshold
	}

	if useMultipart {
		return s.uploadMultipart(ctx, params)
	}
	return s.uploadDirect(ctx, params)
}

func (s service) uploadDirect(ctx context.Context, params UploadParams) (FileInfo, error) {
	id, err := s.File(ctx, FileParams{
		Data:        params.Data,
		Name:        params.Name,
		ContentType: params.ContentType,
		ToStore:     params.ToStore,
		Metadata:    params.Metadata,
	})
	if err != nil {
		return FileInfo{}, err
	}
	return s.FileInfo(ctx, id)
}

func (s service) uploadMultipart(ctx context.Context, params UploadParams) (FileInfo, error) {
	data, err := s.Multipart(ctx, MultipartParams{
		Data:        params.Data,
		FileName:    params.Name,
		ContentType: params.ContentType,
		Size:        params.Size,
		ToStore:     params.ToStore,
		Metadata:    params.Metadata,
	})
	if err != nil {
		return FileInfo{}, err
	}

	select {
	case info := <-data.Done():
		return info, nil
	case err := <-data.Error():
		return FileInfo{}, err
	case <-ctx.Done():
		return FileInfo{}, ctx.Err()
	}
}
