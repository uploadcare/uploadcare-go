package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uploadcare/uploadcare-go/v2/addon"
	"github.com/uploadcare/uploadcare-go/v2/test/testenv"
)

func addonClamAVExecute(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	result, err := r.Addon.Execute(ctx, addon.AddonClamAV, addon.ExecuteParams{
		Target: r.Artifacts.Files[0].ID,
		Params: addon.ClamAVParams{},
	})
	assert.Equal(t, nil, err)
	assert.NotEqual(t, "", result.RequestID)

	r.Artifacts.AddonRequestID = result.RequestID
}

func addonClamAVStatus(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	result, err := r.Addon.Status(ctx, addon.AddonClamAV, r.Artifacts.AddonRequestID)
	assert.Equal(t, nil, err)
	// Status should be one of the known values
	assert.Contains(t, []string{
		addon.StatusInProgress,
		addon.StatusDone,
		addon.StatusError,
		addon.StatusUnknown,
	}, result.Status)
}
