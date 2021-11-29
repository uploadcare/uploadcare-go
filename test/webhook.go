package test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uploadcare/uploadcare-go/test/testenv"
	"github.com/uploadcare/uploadcare-go/ucare"
	"github.com/uploadcare/uploadcare-go/webhook"
)

var (
	webhookURL        string
	webhookSignSecret string
	webhookURLSuffix  int
)

func init() {
	rand.Seed(time.Now().UnixNano())
	webhookURLSuffix = rand.Intn(1000)

	webhookURL = fmt.Sprintf(
		"https://google.com/webhook_endpoint%d",
		webhookURLSuffix,
	)
	webhookSignSecret = "new_signing_secret"
}

func webhookCreate(t *testing.T, r *testenv.Runner) {
	params := webhook.Params{
		TargetURL: ucare.String(fmt.Sprintf(
			"https://duckduckgo.com/webhook_endpoint%d",
			webhookURLSuffix,
		)),
		SigningSecret: ucare.String("test_signing_secret"),
		IsActive:      ucare.Bool(true),
		Event:         ucare.String(webhook.EventFileUploaded),
	}
	info, err := r.Webhook.Create(context.Background(), params)
	assert.Equal(t, nil, err)

	r.Artifacts.Webhook = info
}

func webhookUpdate(t *testing.T, r *testenv.Runner) {
	params := webhook.Params{
		ID:            ucare.Int64(r.Artifacts.Webhook.ID),
		TargetURL:     ucare.String(webhookURL),
		SigningSecret: ucare.String(webhookSignSecret),
	}
	info, err := r.Webhook.Update(context.Background(), params)

	assert.Equal(t, nil, err)

	assert.Equal(t, webhookURL, info.TargetURL)
	assert.Equal(t, webhookSignSecret, *info.SigningSecret)
}

func webhookList(t *testing.T, r *testenv.Runner) {
	hooks, err := r.Webhook.List(context.Background())
	assert.Equal(t, nil, err)

	assert.True(t, len(hooks) >= 1)
}

func webhookDelete(t *testing.T, r *testenv.Runner) {
	err := r.Webhook.Delete(context.Background(), webhookURL)
	assert.Equal(t, nil, err)
}
