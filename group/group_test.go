package group

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/uctest"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/groups/test-group-id~3/", r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	svc := NewService(uctest.NewServerClient(srv))

	err := svc.Delete(context.Background(), "test-group-id~3")
	require.NoError(t, err)
}
