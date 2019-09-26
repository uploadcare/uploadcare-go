package uploadcare

import (
	"net/http"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

func TestSimpleAuthParam(t *testing.T) {
	t.Parallel()

	creds := APICreds{
		SecretKey: "testsk",
		PublicKey: "testpk",
	}

	expectedParam := "Uploadcare.Simple testpk:testsk"
	authParam := simpleAuthParam(creds)

	assert.Equal(t, expectedParam, authParam)
}

func TestSignBasedAuthParam(t *testing.T) {
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
	authParam := signBasedAuthParam(creds, req)

	assert.Equal(t, expected, authParam)
}
