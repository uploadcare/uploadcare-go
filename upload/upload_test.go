package upload

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/uctest"
)

func int64Ptr(v int64) *int64 { return &v }

func TestUpload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fileSize   int64
		threshold  *int64
		omitSize   bool
		wantDirect bool
	}{
		{
			name:       "small_file",
			fileSize:   1 * 1024 * 1024,
			wantDirect: true,
		},
		{
			name:       "exact_threshold",
			fileSize:   DefaultMultipartThreshold,
			wantDirect: true,
		},
		{
			name:     "above_threshold",
			fileSize: DefaultMultipartThreshold + 1,
		},
		{
			name:      "custom_threshold",
			fileSize:  5 * 1024 * 1024,
			threshold: int64Ptr(3 * 1024 * 1024),
		},
		{
			name:       "force_direct",
			fileSize:   DefaultMultipartThreshold + 1,
			threshold:  int64Ptr(0),
			wantDirect: true,
		},
		{
			name:      "force_multipart",
			fileSize:  1024,
			threshold: int64Ptr(-1),
		},
		{
			name:       "auto_size_direct",
			fileSize:   1024,
			omitSize:   true,
			wantDirect: true,
		},
		{
			name:     "auto_size_multipart",
			fileSize: DefaultMultipartThreshold + 1,
			omitSize: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var directHit, multipartHit atomic.Int32
			uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/base/":
					assert.Equal(t, http.MethodPost, r.Method)
					directHit.Add(1)
					uctest.RespondJSON(t, w, map[string]string{"file": "test-uuid"})
				case "/info/":
					uctest.RespondJSON(t, w, FileInfo{FileName: tt.name})
				case "/multipart/start/":
					assert.Equal(t, http.MethodPost, r.Method)
					multipartHit.Add(1)
					uctest.RespondJSON(t, w, map[string]any{
						"uuid": "test-uuid", "parts": []string{},
					})
				case "/multipart/complete/":
					assert.Equal(t, http.MethodPost, r.Method)
					uctest.RespondJSON(t, w, FileInfo{FileName: tt.name})
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}), func(t *testing.T, srv *httptest.Server) {
				svc := NewService(uctest.NewUploadServerClient(srv))

				params := UploadParams{
					Data:               bytes.NewReader(make([]byte, tt.fileSize)),
					Name:               tt.name,
					ContentType:        "application/octet-stream",
					MultipartThreshold: tt.threshold,
				}
				if !tt.omitSize {
					params.Size = tt.fileSize
				}

				info, err := svc.Upload(context.Background(), params)
				require.NoError(t, err)
				assert.Equal(t, tt.name, info.FileName)

				if tt.wantDirect {
					assert.Equal(t, int32(1), directHit.Load(), "expected direct upload")
				} else {
					assert.Equal(t, int32(1), multipartHit.Load(), "expected multipart upload")
				}
			})
		})
	}

	t.Run("multipart_empty_uuid", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/multipart/start/":
				assert.Equal(t, http.MethodPost, r.Method)
				uctest.RespondJSON(t, w, map[string]interface{}{"parts": []string{}})
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewUploadServerClient(srv))

			fileSize := DefaultMultipartThreshold + 1
			_, err := svc.Upload(context.Background(), UploadParams{
				Data:        bytes.NewReader(make([]byte, fileSize)),
				Name:        "bad-start.bin",
				ContentType: "application/octet-stream",
				Size:        fileSize,
			})

			require.Error(t, err)
			assert.Contains(t, err.Error(), "empty upload ID")
		})
	})
}

func TestUpload_MetadataPassThrough_DirectUpload(t *testing.T) {
	t.Parallel()

	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/base/":
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(body), `name="metadata[source]"`)
			assert.Contains(t, string(body), "cli")
			uctest.RespondJSON(t, w, map[string]string{"file": "test-uuid-metadata-direct"})
		case "/info/":
			uctest.RespondJSON(t, w, FileInfo{FileName: "meta.txt"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}), func(t *testing.T, srv *httptest.Server) {
		svc := NewService(uctest.NewUploadServerClient(srv))

		info, err := svc.Upload(context.Background(), UploadParams{
			Data:     strings.NewReader("hi"),
			Name:     "meta.txt",
			Size:     2,
			Metadata: map[string]string{"source": "cli"},
		})

		require.NoError(t, err)
		assert.Equal(t, "meta.txt", info.FileName)
	})
}

func TestUpload_MetadataPassThrough_MultipartUpload(t *testing.T) {
	t.Parallel()

	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/multipart/start/":
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(body), `name="metadata[source]"`)
			assert.Contains(t, string(body), "cli")
			uctest.RespondJSON(t, w, map[string]interface{}{
				"uuid":  "test-uuid-metadata-multipart",
				"parts": []string{},
			})
		case "/multipart/complete/":
			uctest.RespondJSON(t, w, FileInfo{FileName: "meta-multi.txt"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}), func(t *testing.T, srv *httptest.Server) {
		svc := NewService(uctest.NewUploadServerClient(srv))

		info, err := svc.Upload(context.Background(), UploadParams{
			Data:               bytes.NewReader([]byte("hello")),
			Name:               "meta-multi.txt",
			ContentType:        "text/plain",
			Size:               5,
			Metadata:           map[string]string{"source": "cli"},
			MultipartThreshold: int64Ptr(-1),
		})

		require.NoError(t, err)
		assert.Equal(t, "meta-multi.txt", info.FileName)
	})
}
