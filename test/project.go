package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uploadcare/uploadcare-go/v2/test/testenv"
)

func projectInfo(t *testing.T, r *testenv.Runner) {
	info, err := r.Project.Info(context.Background())
	assert.Equal(t, nil, err)

	assert.True(t, info.Name != "")
}
