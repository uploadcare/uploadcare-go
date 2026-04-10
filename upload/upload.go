package upload

import (
	"context"
	"io"
)

const DefaultMultipartThreshold int64 = 10 * 1024 * 1024 // 10MB

type UploadParams struct {
	Data               io.ReadSeeker
	Name               string
	ContentType        string
	Size               int64
	ToStore            *string
	MultipartThreshold *int64
}

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
