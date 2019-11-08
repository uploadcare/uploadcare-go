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
	count := 0
	for list.Next() && count < 10 {
		res, err := list.ReadResult()
		if err != nil {
			t.Fatal(err)
		}
		r.Artifacts.GroupIDs = append(r.Artifacts.GroupIDs, res.ID)
		count++
	}
}

func groupInfo(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	info, err := r.Group.Info(ctx, r.Artifacts.GroupIDs[0])
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, r.Artifacts.GroupIDs[0], info.ID)
}

func groupStore(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	info, err := r.Group.Store(ctx, r.Artifacts.GroupIDs[0])
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, info.StoredAt)
}
