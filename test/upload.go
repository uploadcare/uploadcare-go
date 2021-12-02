package test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uploadcare/uploadcare-go/file"
	"github.com/uploadcare/uploadcare-go/test/testenv"
	"github.com/uploadcare/uploadcare-go/ucare"
	"github.com/uploadcare/uploadcare-go/upload"
)

func uploadFile(t *testing.T, r *testenv.Runner) {
	originalFileName := "integration_test_file"

	ctx := context.Background()
	params := upload.FileParams{
		Name:        originalFileName,
		Data:        strings.NewReader("test content"),
		ContentType: "text/plain",
	}
	id, err := r.Upload.File(ctx, params)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, "", id)

	r.Artifacts.Files = append(
		r.Artifacts.Files,
		&file.Info{BasicFileInfo: file.BasicFileInfo{ID: id}},
	)
}

func uploadFromURL(t *testing.T, r *testenv.Runner) {
	fileURL := "https://bit.ly/2LJ2xOf"

	ctx := context.Background()
	params := upload.FromURLParams{
		URL:  fileURL,
		Name: ucare.String("test_file_name"),
	}
	res, err := r.Upload.FromURL(ctx, params)
	if err != nil {
		t.Fatal(err)
	}

	info, ok := res.Info()
	if !ok {
		select {
		case info = <-res.Done():
		case err := <-res.Error():
			t.Error(err)
		}
	}
	assert.Equal(t, "photo_20190914_154427.jpg", info.OriginalFileName)
	r.Artifacts.Files = append(
		r.Artifacts.Files,
		&file.Info{BasicFileInfo: info.BasicFileInfo},
	)
}

func uploadFileInfo(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	info, err := r.Upload.FileInfo(ctx, r.Artifacts.Files[0].ID)
	assert.Equal(t, nil, err)
	assert.Equal(t, r.Artifacts.Files[0].ID, info.ID)

	r.Artifacts.Files[0] = &file.Info{BasicFileInfo: info.BasicFileInfo}
}

func uploadCreateGroup(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	var ids []string
	for _, r := range r.Artifacts.Files {
		ids = append(ids, r.ID + "/-/resize/x10/")
	}
	info, err := r.Upload.CreateGroup(ctx, ids)
	assert.Equal(t, nil, err)
	r.Artifacts.GroupIDs = append(r.Artifacts.GroupIDs, info.ID)
	for _, f := range info.Files {
		assert.Equal(t, "resize/x10/", f.DefaultEffects)
	}
}

func uploadGroupInfo(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	info, err := r.Upload.GroupInfo(ctx, r.Artifacts.GroupIDs[0])
	assert.Equal(t, nil, err)
	assert.Equal(t, r.Artifacts.GroupIDs[0], info.ID)
}

func uploadMultipart(t *testing.T, r *testenv.Runner) {
	originalFileName := "huge_test_file"
	size := int64(1024 * 1024 * 12) // 12MB

	ctx := context.Background()
	res, err := r.Upload.Multipart(ctx, upload.MultipartParams{
		FileName:    originalFileName,
		Size:        size,
		ContentType: "text/plain",
		Data:        &sizeReadSeeker{size: size},
	})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case info := <-res.Done():
		assert.Equal(t, info.OriginalFileName, info.OriginalFileName)
		r.Artifacts.Files = append(r.Artifacts.Files, &file.Info{
			BasicFileInfo: info.BasicFileInfo,
		})
	case err := <-res.Error():
		t.Error(err)
	}
}

type sizeReadSeeker struct {
	size   int64
	offset int64
}

func (s *sizeReadSeeker) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		p[i] = 'f'
	}
	if s.size-s.offset >= int64(len(p)) {
		return len(p), nil
	}
	return int(s.size - s.offset), nil
}

func (s *sizeReadSeeker) Seek(offset int64, whence int) (int64, error) {
	s.offset = offset
	return 0, nil
}
