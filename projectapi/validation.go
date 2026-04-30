package projectapi

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidPubKey      = errors.New("projectapi: public key must be a non-empty string without slashes")
	ErrInvalidSecretID    = errors.New("projectapi: secret ID must be a non-empty string without slashes")
	ErrInvalidUsageMetric = errors.New("projectapi: usage metric must be one of traffic, storage, operations")
)

func validatePubKey(pubKey string) error {
	return validatePathSegment(pubKey, ErrInvalidPubKey)
}

func validateSecretID(secretID string) error {
	return validatePathSegment(secretID, ErrInvalidSecretID)
}

func validateUsageMetric(metric UsageMetricName) error {
	switch metric {
	case UsageMetricTraffic, UsageMetricStorage, UsageMetricOperations:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidUsageMetric, metric)
	}
}

func validatePathSegment(value string, sentinel error) error {
	if value == "" || strings.Contains(value, "/") {
		return fmt.Errorf("%w: %q", sentinel, value)
	}
	return nil
}
