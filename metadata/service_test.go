package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/uctest"
)

const testFileUUID = "test-uuid"

func testPathList() string {
	return "/files/" + testFileUUID + "/metadata/"
}

func testPathKey(key string) string {
	return testPathList() + key + "/"
}

func unexpectedRequestHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.RequestURI)
	})
}

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("returns_map", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, testPathList(), r.URL.Path)
			uctest.RespondJSON(w, map[string]string{"key1": "value1", "key2": "value2"})
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			data, err := svc.List(context.Background(), testFileUUID)
			require.NoError(t, err)
			assert.Equal(t, map[string]string{"key1": "value1", "key2": "value2"}, data)
		})
	})

	t.Run("too_many_keys", func(t *testing.T) {
		t.Parallel()

		payload := make(map[string]string)
		for i := range MaxKeysNumber + 1 {
			payload[fmt.Sprintf("k%d", i)] = "v"
		}

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uctest.RespondJSON(w, payload)
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			_, err := svc.List(context.Background(), testFileUUID)
			assert.ErrorIs(t, err, ErrTooManyKeys)
		})
	})
}

func TestGet(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, testPathKey("mykey"), r.URL.Path)
			uctest.RespondJSON(w, "my-value")
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			val, err := svc.Get(context.Background(), testFileUUID, "mykey")
			require.NoError(t, err)
			assert.Equal(t, "my-value", val)
		})
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"Not found."}`))
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			_, err := svc.Get(context.Background(), testFileUUID, "nokey")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "404")
		})
	})
}

func TestSet(t *testing.T) {
	t.Parallel()

	t.Run("round_trip", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			assert.Equal(t, testPathKey("mykey"), r.URL.Path)

			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var got string
			require.NoError(t, json.Unmarshal(body, &got))
			assert.Equal(t, "new-value", got)

			uctest.RespondJSON(w, "new-value")
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			val, err := svc.Set(context.Background(), testFileUUID, "mykey", "new-value")
			require.NoError(t, err)
			assert.Equal(t, "new-value", val)
		})
	})

	t.Run("max_value_length", func(t *testing.T) {
		t.Parallel()

		val := strings.Repeat("a", MaxValueLength)
		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			uctest.RespondJSON(w, val)
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			got, err := svc.Set(context.Background(), testFileUUID, "k", val)
			require.NoError(t, err)
			assert.Equal(t, val, got)
		})
	})

	t.Run("value_too_long", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			value string
		}{
			{"ascii", strings.Repeat("a", MaxValueLength+1)},
			{"unicode", strings.Repeat("☃", MaxValueLength+1)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				uctest.WithHTTPServer(t, unexpectedRequestHandler(t), func(t *testing.T, srv *httptest.Server) {
					svc := NewService(uctest.NewServerClient(srv))
					_, err := svc.Set(context.Background(), testFileUUID, "k", tt.value)
					assert.ErrorIs(t, err, ErrValueTooLong)
				})
			})
		}
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, testPathKey("mykey"), r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}), func(t *testing.T, srv *httptest.Server) {
		svc := NewService(uctest.NewServerClient(srv))
		err := svc.Delete(context.Background(), testFileUUID, "mykey")
		require.NoError(t, err)
	})
}

func TestKeyValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(Service) error
	}{
		{
			name: "get rejects slash",
			call: func(svc Service) error {
				_, err := svc.Get(context.Background(), testFileUUID, "a/b")
				return err
			},
		},
		{
			name: "set rejects empty",
			call: func(svc Service) error {
				_, err := svc.Set(context.Background(), testFileUUID, "", "value")
				return err
			},
		},
		{
			name: "delete rejects too long",
			call: func(svc Service) error {
				err := svc.Delete(context.Background(), testFileUUID, strings.Repeat("a", 65))
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uctest.WithHTTPServer(t, unexpectedRequestHandler(t), func(t *testing.T, srv *httptest.Server) {
				svc := NewService(uctest.NewServerClient(srv))
				err := tt.call(svc)
				assert.ErrorIs(t, err, ErrInvalidKey)
			})
		})
	}
}

func TestDotSegmentKeysAreEscaped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		key        string
		wantURI    string
		wantBody   string
		call       func(Service, string) error
		statusCode int
	}{
		{
			name:       "get dot key",
			method:     http.MethodGet,
			key:        ".",
			wantURI:    testPathList() + "%2E/",
			statusCode: http.StatusOK,
			call: func(svc Service, key string) error {
				_, err := svc.Get(context.Background(), testFileUUID, key)
				return err
			},
		},
		{
			name:       "set dotdot key",
			method:     http.MethodPut,
			key:        "..",
			wantURI:    testPathList() + "%2E%2E/",
			wantBody:   `"value"`,
			statusCode: http.StatusOK,
			call: func(svc Service, key string) error {
				_, err := svc.Set(context.Background(), testFileUUID, key, "value")
				return err
			},
		},
		{
			name:       "delete dot key",
			method:     http.MethodDelete,
			key:        ".",
			wantURI:    testPathList() + "%2E/",
			statusCode: http.StatusNoContent,
			call: func(svc Service, key string) error {
				return svc.Delete(context.Background(), testFileUUID, key)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, tt.wantURI, r.RequestURI)

				if tt.wantBody != "" {
					body, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					assert.Equal(t, tt.wantBody, string(body))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode != http.StatusNoContent {
					_ = json.NewEncoder(w).Encode("ok")
				}
			}), func(t *testing.T, srv *httptest.Server) {
				svc := NewService(uctest.NewServerClient(srv))
				err := tt.call(svc, tt.key)
				require.NoError(t, err)
			})
		})
	}
}
