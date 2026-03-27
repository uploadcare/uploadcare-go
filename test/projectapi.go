package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uploadcare/uploadcare-go/v2/projectapi"
)

func projectAPIList(t *testing.T, svc projectapi.Service) {
	list, err := svc.List(context.Background(), nil)
	assert.NoError(t, err)

	assert.True(t, list.Next(), "expected at least one project")
	p, err := list.ReadResult()
	assert.NoError(t, err)
	assert.NotEmpty(t, p.PubKey)
	assert.NotEmpty(t, p.Name)
}

func projectAPIGet(t *testing.T, svc projectapi.Service, pubKey string) {
	data, err := svc.Get(context.Background(), pubKey)
	assert.NoError(t, err)
	assert.Equal(t, pubKey, data.PubKey)
	assert.NotEmpty(t, data.Name)
}

func projectAPIListSecrets(t *testing.T, svc projectapi.Service, pubKey string) {
	list, err := svc.ListSecrets(context.Background(), pubKey, nil)
	assert.NoError(t, err)

	if !list.Next() {
		t.Skip("no secret keys found, skipping assertions")
	}
	s, err := list.ReadResult()
	assert.NoError(t, err)
	assert.NotEmpty(t, s.ID)
	assert.NotEmpty(t, s.Hint)
}
