package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uploadcare/uploadcare-go/v2/test/testenv"
)

func metadataSet(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	val, err := r.Metadata.Set(ctx, r.Artifacts.Files[0].ID, "test_key", "test_value")
	assert.Equal(t, nil, err)
	assert.Equal(t, "test_value", val)
}

func metadataGet(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	val, err := r.Metadata.Get(ctx, r.Artifacts.Files[0].ID, "test_key")
	assert.Equal(t, nil, err)
	assert.Equal(t, "test_value", val)
}

func metadataList(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	data, err := r.Metadata.List(ctx, r.Artifacts.Files[0].ID)
	assert.Equal(t, nil, err)
	assert.Equal(t, "test_value", data["test_key"])
}

func metadataDelete(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	err := r.Metadata.Delete(ctx, r.Artifacts.Files[0].ID, "test_key")
	assert.Equal(t, nil, err)

	// verify it's gone
	_, err = r.Metadata.Get(ctx, r.Artifacts.Files[0].ID, "test_key")
	assert.NotNil(t, err)
}
