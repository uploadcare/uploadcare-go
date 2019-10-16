package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uploadcare/uploadcare-go/group"
	"github.com/uploadcare/uploadcare-go/test/testenv"
)

func groupList(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	list, err := r.Group.List(ctx, group.ListParams{})
	assert.Equal(t, nil, err)
	for list.Next() {
		res, err := list.ReadResult()
		assert.Equal(t, nil, err)
		r.Artifacts.GroupIDs = append(r.Artifacts.GroupIDs, res.ID)
	}
}

func groupInfo(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	info, err := r.Group.Info(ctx, r.Artifacts.GroupIDs[0])
	assert.Equal(t, nil, err)
	assert.Equal(t, r.Artifacts.GroupIDs[0], info.ID)
}

func groupStore(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	info, err := r.Group.Store(ctx, r.Artifacts.GroupIDs[0])
	assert.Equal(t, nil, err)
	assert.NotNil(t, info.StoredAt)
}
