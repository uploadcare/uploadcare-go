package tag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/uctest"
)

const testFileUUID = "test-uuid"

func testTagsPath() string {
	return "/files/" + testFileUUID + "/tags/"
}

func unexpectedRequestHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.RequestURI)
	})
}

func TestList(t *testing.T) {
	t.Parallel()

	uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, testTagsPath(), r.URL.Path)
		uctest.RespondJSON(t, w, map[string][]string{"tags": {"cat", "animal"}})
	}), func(t *testing.T, srv *httptest.Server) {
		svc := NewService(uctest.NewServerClient(srv))
		tags, err := svc.List(context.Background(), testFileUUID)
		require.NoError(t, err)
		assert.Equal(t, []string{"cat", "animal"}, tags)
	})
}

func TestReplace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tags []string
		want []string
	}{
		{
			name: "normalizes tags",
			tags: []string{" Cat ", "ANIMAL", "cat"},
			want: []string{"cat", "animal"},
		},
		{
			name: "nil clears with empty array",
			tags: nil,
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, testTagsPath(), r.URL.Path)

				var body replaceParams
				require.NoError(t, json.Unmarshal(uctest.ReadBody(t, r), &body))
				assert.Equal(t, tt.want, body.Tags)
				assert.NotNil(t, body.Tags)

				uctest.RespondJSON(t, w, Result{
					Tags:  body.Tags,
					Added: body.Tags,
				})
			}), func(t *testing.T, srv *httptest.Server) {
				svc := NewService(uctest.NewServerClient(srv))
				result, err := svc.Replace(context.Background(), testFileUUID, tt.tags)
				require.NoError(t, err)
				assert.Equal(t, tt.want, result.Tags)
			})
		})
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		params     UpdateParams
		wantBody   map[string][]string
		wantResult Result
	}{
		{
			name: "normalizes add and delete",
			params: UpdateParams{
				Add:    []string{" Summer ", "FEATURED", "summer"},
				Delete: []string{"DRAFT"},
			},
			wantBody: map[string][]string{
				"add":    {"summer", "featured"},
				"delete": {"draft"},
			},
			wantResult: Result{
				Tags:    []string{"summer", "featured"},
				Added:   []string{"summer", "featured"},
				Deleted: []string{"draft"},
			},
		},
		{
			name:     "empty is no-op body",
			params:   UpdateParams{},
			wantBody: map[string][]string{},
			wantResult: Result{
				Tags:    []string{},
				Added:   []string{},
				Deleted: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPatch, r.Method)
				assert.Equal(t, testTagsPath(), r.URL.Path)

				var body map[string][]string
				require.NoError(t, json.Unmarshal(uctest.ReadBody(t, r), &body))
				assert.Equal(t, tt.wantBody, body)
				uctest.RespondJSON(t, w, tt.wantResult)
			}), func(t *testing.T, srv *httptest.Server) {
				svc := NewService(uctest.NewServerClient(srv))
				result, err := svc.Update(context.Background(), testFileUUID, tt.params)
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			})
		})
	}
}

func TestValidationShortCircuitsRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		call    func(Service) error
		wantErr error
	}{
		{
			name: "list rejects invalid file UUID",
			call: func(svc Service) error {
				_, err := svc.List(context.Background(), "../bad")
				return err
			},
			wantErr: ErrInvalidFileUUID,
		},
		{
			name: "replace rejects invalid tag",
			call: func(svc Service) error {
				_, err := svc.Replace(context.Background(), testFileUUID, []string{"bad tag"})
				return err
			},
			wantErr: ErrInvalidCharacters,
		},
		{
			name: "update rejects invalid added tag",
			call: func(svc Service) error {
				_, err := svc.Update(context.Background(), testFileUUID, UpdateParams{
					Add: []string{"bad tag"},
				})
				return err
			},
			wantErr: ErrInvalidCharacters,
		},
		{
			name: "update rejects invalid deleted tag",
			call: func(svc Service) error {
				_, err := svc.Update(context.Background(), testFileUUID, UpdateParams{
					Delete: []string{"bad tag"},
				})
				return err
			},
			wantErr: ErrInvalidCharacters,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uctest.WithHTTPServer(t, unexpectedRequestHandler(t), func(t *testing.T, srv *httptest.Server) {
				svc := NewService(uctest.NewServerClient(srv))
				assert.ErrorIs(t, tt.call(svc), tt.wantErr)
			})
		})
	}
}
