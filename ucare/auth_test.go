package ucare

import (
	"net/http"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

func testSimpleAuthRESTAPIParam(t *testing.T) {
	t.Parallel()

	creds := APICreds{
		SecretKey: "testsk",
		PublicKey: "testpk",
	}

	expectedParam := "Uploadcare.Simple testpk:testsk"
	authParam := simpleRESTAPIAuthParam(creds)

	assert.Equal(t, expectedParam, authParam)
}

func testSignBasedRESTAPIAuthParam(t *testing.T) {
	creds := APICreds{
		SecretKey: "demoprivatekey",
		PublicKey: "testpk",
	}
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
	authParam := signBasedRESTAPIAuthParam(creds, req)

	assert.Equal(t, expected, authParam)
}

func testSignBasedUploadAPIAuthParam(t *testing.T) {
	secret := "project_secret_key"
	now := int64(1454903856)
	expected := "46f70d2b4fb6196daeb2c16bf44a7f1e"

	assert.Equal(t, expected, signBasedUploadAPIAuthParam(secret, now))
}
