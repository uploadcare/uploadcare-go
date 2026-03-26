package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uploadcare/uploadcare-go/v2/projectapi"
)

func projectAPIList(t *testing.T, svc projectapi.Service) {
	data, err := svc.List(context.Background(), nil)
	assert.NoError(t, err)
	assert.True(t, data.Total > 0, "expected at least one project")
	assert.True(t, len(data.Results) > 0)
	assert.NotEmpty(t, data.Results[0].PubKey)
	assert.NotEmpty(t, data.Results[0].Name)
}

func projectAPIGet(t *testing.T, svc projectapi.Service, pubKey string) {
	data, err := svc.Get(context.Background(), pubKey)
	assert.NoError(t, err)
	assert.Equal(t, pubKey, data.PubKey)
	assert.NotEmpty(t, data.Name)
}

func projectAPIListSecrets(t *testing.T, svc projectapi.Service, pubKey string) {
	data, err := svc.ListSecrets(context.Background(), pubKey, nil)
	assert.NoError(t, err)
	if len(data.Results) == 0 {
		t.Skip("no secret keys found, skipping assertions")
	}
	assert.True(t, data.Total > 0, "expected at least one secret key")
	assert.NotEmpty(t, data.Results[0].ID)
	assert.NotEmpty(t, data.Results[0].Hint)
}
