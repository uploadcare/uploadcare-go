package conversion

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

func TestConversionParams_Encode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params Params
		want   string
	}{
		{
			name: "with_save_in_group",
			params: Params{
				Paths:       []string{"uuid/document/-/format/png/"},
				ToStore:     ucare.String(ToStoreTrue),
				SaveInGroup: "1",
			},
			want: `{
				"paths": ["uuid/document/-/format/png/"],
				"store": "1",
				"save_in_group": "1"
			}`,
		},
		{
			name: "without_save_in_group",
			params: Params{
				Paths:   []string{"uuid/document/-/format/pdf/"},
				ToStore: ucare.String(ToStoreFalse),
			},
			want: `{
				"paths": ["uuid/document/-/format/pdf/"],
				"store": "0"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodPost, "https://example.test/convert/document/", nil)
			require.NoError(t, err)
			require.NoError(t, tt.params.EncodeReq(req))

			body, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(body))
		})
	}
}
