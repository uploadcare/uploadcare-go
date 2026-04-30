package projectapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const (
	usagePathFmt       = "/projects/%s/usage/"
	usageMetricPathFmt = "/projects/%s/usage/%s/"
)

type UsageMetricName string

const (
	UsageMetricTraffic    UsageMetricName = "traffic"
	UsageMetricStorage    UsageMetricName = "storage"
	UsageMetricOperations UsageMetricName = "operations"
)

func (s service) GetUsage(
	ctx context.Context,
	pubKey string,
	params UsageDateRange,
) (data UsageMetricsCombined, err error) {
	if err = validatePubKey(pubKey); err != nil {
		return
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		usagePath(pubKey),
		&params,
		&data,
	)
	return
}

func (s service) GetUsageMetric(
	ctx context.Context,
	pubKey string,
	metric UsageMetricName,
	params UsageDateRange,
) (data UsageMetric, err error) {
	if err = validatePubKey(pubKey); err != nil {
		return
	}
	if err = validateUsageMetric(metric); err != nil {
		return
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		usageMetricPath(pubKey, metric),
		&params,
		&data,
	)
	return
}

func usagePath(pubKey string) string {
	return fmt.Sprintf(usagePathFmt, url.PathEscape(pubKey))
}

func usageMetricPath(pubKey string, metric UsageMetricName) string {
	return fmt.Sprintf(usageMetricPathFmt, url.PathEscape(pubKey), url.PathEscape(string(metric)))
}
