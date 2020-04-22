package ucare

import (
	"net/http"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

func TestSimpleAuthRESTAPIParam(t *testing.T) {
	t.Parallel()

	creds := APICreds{
		SecretKey: "testsk",
		PublicKey: "testpk",
	}

	expectedParam := "Uploadcare.Simple testpk:testsk"
	authParam := simpleRESTAPIAuthParam(creds)

	assert.Equal(t, expectedParam, authParam)
}

func TestSignBasedRESTAPIAuthParam(t *testing.T) {
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

func TestSignBasedUploadAPIAuthParam(t *testing.T) {
	secret := "project_secret_key"
	now := int64(1454903856)
	expected := "d39a461d41f607338abffee5f31da4d4e46535651c87346e76906bf75c064d47"

	assert.Equal(t, expected, signBasedUploadAPIAuthParam(secret, now))
}
