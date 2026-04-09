package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	assert "github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/config"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

type mockUploadClient struct {
	baseURL string
	conn    *http.Client
}

func (c *mockUploadClient) NewRequest(
	ctx context.Context,
	_ config.Endpoint,
	method string,
	requrl string,
	data ucare.ReqEncoder,
) (*http.Request, error) {
	fullURL := c.baseURL + requrl
	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, config.CtxAuthFuncKey, ucare.UploadAPIAuthFunc(
		func() (string, *string, *int64) {
			return "testpubkey", nil, nil
		},
	))
	req = req.WithContext(ctx)
	if data != nil {
		if err := data.EncodeReq(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

func (c *mockUploadClient) Do(req *http.Request, resdata interface{}) error {
	resp, err := c.conn.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resdata != nil {
		return json.NewDecoder(resp.Body).Decode(resdata)
	}
	return nil
}

func int64Ptr(v int64) *int64 { return &v }

func TestUpload_SmallFile_DirectUpload(t *testing.T) {
	t.Parallel()

	var directHit, multipartHit atomic.Int32
	fileID := "test-uuid-123"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/base/":
			directHit.Add(1)
			json.NewEncoder(w).Encode(map[string]string{"file": fileID})
		case "/info/":
			json.NewEncoder(w).Encode(FileInfo{FileName: "small.txt"})
		case "/multipart/start/":
			multipartHit.Add(1)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"uuid": fileID, "parts": []string{}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := &mockUploadClient{baseURL: srv.URL, conn: srv.Client()}
	svc := NewService(client)

	data := bytes.NewReader(make([]byte, 1*1024*1024)) // 1MB
	info, err := svc.Upload(context.Background(), UploadParams{
		Data: data,
		Name: "small.txt",
		Size: 1 * 1024 * 1024,
	})

	assert.NoError(t, err)
	assert.Equal(t, "small.txt", info.FileName)
	assert.Equal(t, int32(1), directHit.Load())
	assert.Equal(t, int32(0), multipartHit.Load())
}

func TestUpload_ExactThreshold_DirectUpload(t *testing.T) {
	t.Parallel()

	var directHit, multipartHit atomic.Int32
	fileID := "test-uuid-exact"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/base/":
			directHit.Add(1)
			json.NewEncoder(w).Encode(map[string]string{"file": fileID})
		case "/info/":
			json.NewEncoder(w).Encode(FileInfo{FileName: "exact.bin"})
		case "/multipart/start/":
			multipartHit.Add(1)
			json.NewEncoder(w).Encode(map[string]interface{}{"uuid": fileID, "parts": []string{}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := &mockUploadClient{baseURL: srv.URL, conn: srv.Client()}
	svc := NewService(client)

	// File size exactly at DefaultMultipartThreshold → should use direct upload
	data := bytes.NewReader(make([]byte, DefaultMultipartThreshold))
	info, err := svc.Upload(context.Background(), UploadParams{
		Data: data,
		Name: "exact.bin",
		Size: DefaultMultipartThreshold,
	})

	assert.NoError(t, err)
	assert.Equal(t, "exact.bin", info.FileName)
	assert.Equal(t, int32(1), directHit.Load())
	assert.Equal(t, int32(0), multipartHit.Load())
}

func TestUpload_LargeFile_MultipartUpload(t *testing.T) {
	t.Parallel()

	var directHit, multipartHit, completeHit atomic.Int32
	fileID := "test-uuid-456"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/base/":
			directHit.Add(1)
			json.NewEncoder(w).Encode(map[string]string{"file": fileID})
		case "/multipart/start/":
			multipartHit.Add(1)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"uuid":  fileID,
				"parts": []string{},
			})
		case "/multipart/complete/":
			completeHit.Add(1)
			json.NewEncoder(w).Encode(FileInfo{FileName: "large.bin"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := &mockUploadClient{baseURL: srv.URL, conn: srv.Client()}
	svc := NewService(client)

	data := bytes.NewReader(make([]byte, 20*1024*1024)) // 20MB
	info, err := svc.Upload(context.Background(), UploadParams{
		Data:        data,
		Name:        "large.bin",
		ContentType: "application/octet-stream",
		Size:        20 * 1024 * 1024,
	})

	assert.NoError(t, err)
	assert.Equal(t, "large.bin", info.FileName)
	assert.Equal(t, int32(0), directHit.Load())
	assert.Equal(t, int32(1), multipartHit.Load())
	assert.Equal(t, int32(1), completeHit.Load())
}

func TestUpload_CustomThreshold(t *testing.T) {
	t.Parallel()

	var directHit, multipartHit atomic.Int32
	fileID := "test-uuid-789"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/base/":
			directHit.Add(1)
			json.NewEncoder(w).Encode(map[string]string{"file": fileID})
		case "/multipart/start/":
			multipartHit.Add(1)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"uuid":  fileID,
				"parts": []string{},
			})
		case "/multipart/complete/":
			json.NewEncoder(w).Encode(FileInfo{FileName: "medium.bin"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := &mockUploadClient{baseURL: srv.URL, conn: srv.Client()}
	svc := NewService(client)

	data := bytes.NewReader(make([]byte, 5*1024*1024)) // 5MB
	threshold := int64(3 * 1024 * 1024)                // 3MB threshold
	info, err := svc.Upload(context.Background(), UploadParams{
		Data:               data,
		Name:               "medium.bin",
		ContentType:        "application/octet-stream",
		Size:               5 * 1024 * 1024,
		MultipartThreshold: &threshold,
	})

	assert.NoError(t, err)
	assert.Equal(t, "medium.bin", info.FileName)
	assert.Equal(t, int32(0), directHit.Load())
	assert.Equal(t, int32(1), multipartHit.Load())
}

func TestUpload_ForceDirectUpload(t *testing.T) {
	t.Parallel()

	var directHit, multipartHit atomic.Int32
	fileID := "test-uuid-force-direct"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/base/":
			directHit.Add(1)
			json.NewEncoder(w).Encode(map[string]string{"file": fileID})
		case "/info/":
			json.NewEncoder(w).Encode(FileInfo{FileName: "forced-direct.bin"})
		case "/multipart/start/":
			multipartHit.Add(1)
			json.NewEncoder(w).Encode(map[string]interface{}{"uuid": fileID, "parts": []string{}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := &mockUploadClient{baseURL: srv.URL, conn: srv.Client()}
	svc := NewService(client)

	data := bytes.NewReader(make([]byte, 20*1024*1024)) // 20MB, normally multipart
	info, err := svc.Upload(context.Background(), UploadParams{
		Data:               data,
		Name:               "forced-direct.bin",
		Size:               20 * 1024 * 1024,
		MultipartThreshold: int64Ptr(0), // force direct
	})

	assert.NoError(t, err)
	assert.Equal(t, "forced-direct.bin", info.FileName)
	assert.Equal(t, int32(1), directHit.Load())
	assert.Equal(t, int32(0), multipartHit.Load())
}

func TestUpload_ForceMultipartUpload(t *testing.T) {
	t.Parallel()

	var directHit, multipartHit atomic.Int32
	fileID := "test-uuid-force-multi"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/base/":
			directHit.Add(1)
			json.NewEncoder(w).Encode(map[string]string{"file": fileID})
		case "/multipart/start/":
			multipartHit.Add(1)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"uuid":  fileID,
				"parts": []string{},
			})
		case "/multipart/complete/":
			json.NewEncoder(w).Encode(FileInfo{FileName: "forced-multi.txt"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := &mockUploadClient{baseURL: srv.URL, conn: srv.Client()}
	svc := NewService(client)

	data := bytes.NewReader(make([]byte, 1024)) // 1KB, normally direct
	info, err := svc.Upload(context.Background(), UploadParams{
		Data:               data,
		Name:               "forced-multi.txt",
		ContentType:        "text/plain",
		Size:               1024,
		MultipartThreshold: int64Ptr(-1), // force multipart
	})

	assert.NoError(t, err)
	assert.Equal(t, "forced-multi.txt", info.FileName)
	assert.Equal(t, int32(0), directHit.Load())
	assert.Equal(t, int32(1), multipartHit.Load())
}

func TestUpload_MetadataPassThrough_DirectUpload(t *testing.T) {
	t.Parallel()

	fileID := "test-uuid-metadata-direct"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/base/":
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(body), `name="metadata[source]"`)
			assert.Contains(t, string(body), "cli")
			json.NewEncoder(w).Encode(map[string]string{"file": fileID})
		case "/info/":
			json.NewEncoder(w).Encode(FileInfo{FileName: "meta.txt"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := &mockUploadClient{baseURL: srv.URL, conn: srv.Client()}
	svc := NewService(client)

	info, err := svc.Upload(context.Background(), UploadParams{
		Data:     strings.NewReader("hi"),
		Name:     "meta.txt",
		Size:     2,
		Metadata: map[string]string{"source": "cli"},
	})

	assert.NoError(t, err)
	assert.Equal(t, "meta.txt", info.FileName)
}

func TestUpload_MetadataPassThrough_MultipartUpload(t *testing.T) {
	t.Parallel()

	fileID := "test-uuid-metadata-multipart"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/multipart/start/":
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(body), `name="metadata[source]"`)
			assert.Contains(t, string(body), "cli")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"uuid":  fileID,
				"parts": []string{},
			})
		case "/multipart/complete/":
			json.NewEncoder(w).Encode(FileInfo{FileName: "meta-multi.txt"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := &mockUploadClient{baseURL: srv.URL, conn: srv.Client()}
	svc := NewService(client)

	info, err := svc.Upload(context.Background(), UploadParams{
		Data:               bytes.NewReader([]byte("hello")),
		Name:               "meta-multi.txt",
		ContentType:        "text/plain",
		Size:               5,
		Metadata:           map[string]string{"source": "cli"},
		MultipartThreshold: int64Ptr(-1),
	})

	assert.NoError(t, err)
	assert.Equal(t, "meta-multi.txt", info.FileName)
}
