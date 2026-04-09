package projectapi

import (
	"context"
	"fmt"
	"net/http"
)

const (
	usagePathFmt       = "/projects/%s/usage/"
	usageMetricPathFmt = "/projects/%s/usage/%s/"
)

// GetUsage returns combined usage metrics for a project.
func (s service) GetUsage(
	ctx context.Context,
	pubKey string,
	params UsageDateRange,
) (data UsageMetricsCombined, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		fmt.Sprintf(usagePathFmt, pubKey),
		&params,
		&data,
	)
	return
}

// GetUsageMetric returns daily usage for a specific metric.
// Valid metric values: "traffic", "storage", "operations".
func (s service) GetUsageMetric(
	ctx context.Context,
	pubKey string,
	metric string,
	params UsageDateRange,
) (data UsageMetric, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		fmt.Sprintf(usageMetricPathFmt, pubKey, metric),
		&params,
		&data,
	)
	return
}
