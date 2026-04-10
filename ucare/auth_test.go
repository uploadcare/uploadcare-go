package ucare

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	t.Parallel()

	t.Run("simple_rest_param", func(t *testing.T) {
		t.Parallel()

		creds := APICreds{SecretKey: "testsk", PublicKey: "testpk"}
		assert.Equal(t, "Uploadcare.Simple testpk:testsk", simpleRESTAPIAuthParam(creds))
	})

	t.Run("sign_based_rest_param", func(t *testing.T) {
		creds := APICreds{SecretKey: "demoprivatekey", PublicKey: "testpk"}
		req, _ := http.NewRequest(
			http.MethodGet,
			"/files/?limit=1&stored=true",
			nil,
		)

		// taken from https://uploadcare.com/docs/api_reference/rest/requests_auth/
		now := time.
			Unix(1541423681, 0).
			In(dateHeaderLocation).
			Format(dateHeaderFormat)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Date", now)

		expected := "Uploadcare testpk:3cbc4d2cf91f80c1ba162b926f8a975e8bec7995"
		assert.Equal(t, expected, signBasedRESTAPIAuthParam(creds, req))
	})

	t.Run("sign_based_upload_param", func(t *testing.T) {
		t.Parallel()

		secret := "project_secret_key"
		now := int64(1454903856)
		expected := "d39a461d41f607338abffee5f31da4d4e46535651c87346e76906bf75c064d47"

		got := signBasedUploadAPIAuthParam(secret, now)
		require.Equal(t, expected, got)
	})
}
