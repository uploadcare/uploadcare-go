package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uploadcare/uploadcare-go/conversion"
	"github.com/uploadcare/uploadcare-go/test/testenv"
)

func conversionDocument(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()
	info, err := r.Conversion.Document(ctx, conversion.Params{
		Paths: []string{
			r.Artifacts.Files[0].ID + "/document/-/format/pdf/",
		},
	})
	assert.Equal(t, nil, err)
	if len(info.Jobs) == 0 {
		t.Fatal("job should be started")
	}

	r.Artifacts.ConversionJobs = append(
		r.Artifacts.ConversionJobs,
		info.Jobs[0],
	)
}

func conversionDocumentStatus(t *testing.T, r *testenv.Runner) {
	ctx := context.Background()

	job, err := r.Conversion.DocumentStatus(
		ctx,
		r.Artifacts.ConversionJobs[0].Token,
	)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, "", job.Status)
}
