package conversion

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

func TestConversionParams_WithSaveInGroup(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodPost, "https://example.test/convert/document/", nil)
	assert.NoError(t, err)

	params := Params{
		Paths:       []string{"uuid/document/-/format/png/"},
		ToStore:     ucare.String(ToStoreTrue),
		SaveInGroup: ucare.String("1"),
	}

	err = params.EncodeReq(req)
	assert.NoError(t, err)

	body, err := io.ReadAll(req.Body)
	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"paths": ["uuid/document/-/format/png/"],
		"store": "1",
		"save_in_group": "1"
	}`, string(body))
}

func TestConversionParams_WithoutSaveInGroup(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodPost, "https://example.test/convert/document/", nil)
	assert.NoError(t, err)

	params := Params{
		Paths:   []string{"uuid/document/-/format/pdf/"},
		ToStore: ucare.String(ToStoreFalse),
	}

	err = params.EncodeReq(req)
	assert.NoError(t, err)

	body, err := io.ReadAll(req.Body)
	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"paths": ["uuid/document/-/format/pdf/"],
		"store": "0"
	}`, string(body))
	assert.NotContains(t, string(body), "save_in_group")
}
