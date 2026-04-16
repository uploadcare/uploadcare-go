package ucare

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// RetryConfig controls automatic retry of throttled (HTTP 429) requests.
//
// MaxRetries limits how many times a request is retried.
// MaxWaitSeconds caps the per-retry wait time. When the effective wait
// (either the server's Retry-After value or the computed exponential
// backoff) exceeds this cap, the request fails immediately with a
// ThrottleError instead of sleeping. Set to 0 to disable the cap.
type RetryConfig struct {
	MaxRetries     int
	MaxWaitSeconds int
}

// handleThrottle decides whether to retry a 429 response. It returns
// (true, nil) after sleeping when a retry should be attempted, or
// (false, ThrottleError) when retries are exhausted or the wait
// exceeds the configured cap.
func handleThrottle(
	ctx context.Context,
	resp *http.Response,
	retry *RetryConfig,
	tries int,
) (bool, error) {
	retryAfter, err := strconv.Atoi(
		resp.Header.Get("Retry-After"),
	)
	if err != nil || retryAfter < 0 {
		retryAfter = 0
	}

	if retry == nil || tries > retry.MaxRetries {
		return false, ThrottleError{RetryAfter: retryAfter}
	}

	wait := retryAfter
	if wait <= 0 {
		wait = expBackoff(tries)
	}

	if retry.MaxWaitSeconds > 0 && wait > retry.MaxWaitSeconds {
		return false, ThrottleError{RetryAfter: wait}
	}

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(time.Duration(wait) * time.Second):
	}
	return true, nil
}

func expBackoff(attempt int) int {
	wait := 1 << (attempt - 1)
	return min(wait, 30)
}
