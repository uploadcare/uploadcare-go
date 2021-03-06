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
	params := conversion.Params{
		Paths: []string{
			r.Artifacts.Files[1].ID + "/document/-/format/pdf/",
		},
	}
	info, err := r.Conversion.Document(ctx, params)
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

	if len(r.Artifacts.ConversionJobs) == 0 {
		t.Fatal("no conversion job token")
	}

	job, err := r.Conversion.DocumentStatus(
		ctx,
		r.Artifacts.ConversionJobs[0].Token,
	)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, "", job.Status)
}
