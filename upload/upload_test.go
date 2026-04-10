package upload

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/uctest"
)

func int64Ptr(v int64) *int64 { return &v }

func TestUpload(t *testing.T) {
	t.Parallel()

	t.Run("small_file_direct", func(t *testing.T) {
		t.Parallel()

		var directHit, multipartHit atomic.Int32
		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/base/":
				directHit.Add(1)
				uctest.RespondJSON(w, map[string]string{"file": "test-uuid-123"})
			case "/info/":
				uctest.RespondJSON(w, FileInfo{FileName: "small.txt"})
			case "/multipart/start/":
				multipartHit.Add(1)
				uctest.RespondJSON(w, map[string]interface{}{"uuid": "test-uuid-123", "parts": []string{}})
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewUploadServerClient(srv))

			data := bytes.NewReader(make([]byte, 1*1024*1024)) // 1MB
			info, err := svc.Upload(context.Background(), UploadParams{
				Data: data,
				Name: "small.txt",
				Size: 1 * 1024 * 1024,
			})

			require.NoError(t, err)
			assert.Equal(t, "small.txt", info.FileName)
			assert.Equal(t, int32(1), directHit.Load())
			assert.Equal(t, int32(0), multipartHit.Load())
		})
	})

	t.Run("exact_threshold_direct", func(t *testing.T) {
		t.Parallel()

		var directHit, multipartHit atomic.Int32
		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/base/":
				directHit.Add(1)
				uctest.RespondJSON(w, map[string]string{"file": "test-uuid-exact"})
			case "/info/":
				uctest.RespondJSON(w, FileInfo{FileName: "exact.bin"})
			case "/multipart/start/":
				multipartHit.Add(1)
				uctest.RespondJSON(w, map[string]interface{}{"uuid": "test-uuid-exact", "parts": []string{}})
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewUploadServerClient(srv))

			data := bytes.NewReader(make([]byte, DefaultMultipartThreshold))
			info, err := svc.Upload(context.Background(), UploadParams{
				Data: data,
				Name: "exact.bin",
				Size: DefaultMultipartThreshold,
			})

			require.NoError(t, err)
			assert.Equal(t, "exact.bin", info.FileName)
			assert.Equal(t, int32(1), directHit.Load())
			assert.Equal(t, int32(0), multipartHit.Load())
		})
	})

	t.Run("large_file_multipart", func(t *testing.T) {
		t.Parallel()

		var directHit, multipartHit, completeHit atomic.Int32
		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/base/":
				directHit.Add(1)
				uctest.RespondJSON(w, map[string]string{"file": "test-uuid-456"})
			case "/multipart/start/":
				multipartHit.Add(1)
				uctest.RespondJSON(w, map[string]interface{}{"uuid": "test-uuid-456", "parts": []string{}})
			case "/multipart/complete/":
				completeHit.Add(1)
				uctest.RespondJSON(w, FileInfo{FileName: "large.bin"})
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewUploadServerClient(srv))

			data := bytes.NewReader(make([]byte, 20*1024*1024)) // 20MB
			info, err := svc.Upload(context.Background(), UploadParams{
				Data:        data,
				Name:        "large.bin",
				ContentType: "application/octet-stream",
				Size:        20 * 1024 * 1024,
			})

			require.NoError(t, err)
			assert.Equal(t, "large.bin", info.FileName)
			assert.Equal(t, int32(0), directHit.Load())
			assert.Equal(t, int32(1), multipartHit.Load())
			assert.Equal(t, int32(1), completeHit.Load())
		})
	})

	t.Run("custom_threshold", func(t *testing.T) {
		t.Parallel()

		var directHit, multipartHit atomic.Int32
		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/base/":
				directHit.Add(1)
				uctest.RespondJSON(w, map[string]string{"file": "test-uuid-789"})
			case "/multipart/start/":
				multipartHit.Add(1)
				uctest.RespondJSON(w, map[string]interface{}{"uuid": "test-uuid-789", "parts": []string{}})
			case "/multipart/complete/":
				uctest.RespondJSON(w, FileInfo{FileName: "medium.bin"})
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewUploadServerClient(srv))

			data := bytes.NewReader(make([]byte, 5*1024*1024)) // 5MB
			threshold := int64(3 * 1024 * 1024)                // 3MB threshold
			info, err := svc.Upload(context.Background(), UploadParams{
				Data:               data,
				Name:               "medium.bin",
				ContentType:        "application/octet-stream",
				Size:               5 * 1024 * 1024,
				MultipartThreshold: &threshold,
			})

			require.NoError(t, err)
			assert.Equal(t, "medium.bin", info.FileName)
			assert.Equal(t, int32(0), directHit.Load())
			assert.Equal(t, int32(1), multipartHit.Load())
		})
	})

	t.Run("force_direct", func(t *testing.T) {
		t.Parallel()

		var directHit, multipartHit atomic.Int32
		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/base/":
				directHit.Add(1)
				uctest.RespondJSON(w, map[string]string{"file": "test-uuid-force-direct"})
			case "/info/":
				uctest.RespondJSON(w, FileInfo{FileName: "forced-direct.bin"})
			case "/multipart/start/":
				multipartHit.Add(1)
				uctest.RespondJSON(w, map[string]interface{}{"uuid": "test-uuid-force-direct", "parts": []string{}})
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewUploadServerClient(srv))

			data := bytes.NewReader(make([]byte, 20*1024*1024)) // 20MB, normally multipart
			info, err := svc.Upload(context.Background(), UploadParams{
				Data:               data,
				Name:               "forced-direct.bin",
				Size:               20 * 1024 * 1024,
				MultipartThreshold: int64Ptr(0),
			})

			require.NoError(t, err)
			assert.Equal(t, "forced-direct.bin", info.FileName)
			assert.Equal(t, int32(1), directHit.Load())
			assert.Equal(t, int32(0), multipartHit.Load())
		})
	})

	t.Run("force_multipart", func(t *testing.T) {
		t.Parallel()

		var directHit, multipartHit atomic.Int32
		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/base/":
				directHit.Add(1)
				uctest.RespondJSON(w, map[string]string{"file": "test-uuid-force-multi"})
			case "/multipart/start/":
				multipartHit.Add(1)
				uctest.RespondJSON(w, map[string]interface{}{"uuid": "test-uuid-force-multi", "parts": []string{}})
			case "/multipart/complete/":
				uctest.RespondJSON(w, FileInfo{FileName: "forced-multi.txt"})
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewUploadServerClient(srv))

			data := bytes.NewReader(make([]byte, 1024)) // 1KB, normally direct
			info, err := svc.Upload(context.Background(), UploadParams{
				Data:               data,
				Name:               "forced-multi.txt",
				ContentType:        "text/plain",
				Size:               1024,
				MultipartThreshold: int64Ptr(-1),
			})

			require.NoError(t, err)
			assert.Equal(t, "forced-multi.txt", info.FileName)
			assert.Equal(t, int32(0), directHit.Load())
			assert.Equal(t, int32(1), multipartHit.Load())
		})
	})
}
