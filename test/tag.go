package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/tag"
	"github.com/uploadcare/uploadcare-go/v2/test/testenv"
)

func tagListUploaded(t *testing.T, r *testenv.Runner) {
	tags, err := r.Tag.List(context.Background(), r.Artifacts.Files[0].ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"sdk-integration", "go"}, tags)
}

func tagReplace(t *testing.T, r *testenv.Runner) {
	result, err := r.Tag.Replace(
		context.Background(),
		r.Artifacts.Files[0].ID,
		[]string{"approved", "Go"},
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"approved", "go"}, result.Tags)
	assert.ElementsMatch(t, []string{"approved"}, result.Added)
	assert.ElementsMatch(t, []string{"sdk-integration"}, result.Deleted)
}

func tagUpdate(t *testing.T, r *testenv.Runner) {
	result, err := r.Tag.Update(
		context.Background(),
		r.Artifacts.Files[0].ID,
		tag.UpdateParams{
			Add:    []string{"featured"},
			Delete: []string{"approved"},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"go", "featured"}, result.Tags)
	assert.Equal(t, []string{"featured"}, result.Added)
	assert.Equal(t, []string{"approved"}, result.Deleted)
}

func tagClear(t *testing.T, r *testenv.Runner) {
	result, err := r.Tag.Replace(
		context.Background(),
		r.Artifacts.Files[0].ID,
		nil,
	)
	require.NoError(t, err)
	assert.Empty(t, result.Tags)
}
