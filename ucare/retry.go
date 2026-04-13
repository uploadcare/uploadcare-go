package ucare

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

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

	// Only bail out when the server explicitly asks for a wait that
	// exceeds our cap. Locally computed backoff is already bounded
	// by expBackoff's hardcoded ceiling and does not need capping.
	if retry.MaxWaitSeconds > 0 &&
		retryAfter > retry.MaxWaitSeconds {
		return false, ThrottleError{RetryAfter: retryAfter}
	}

	wait := retryAfter
	if wait <= 0 {
		wait = expBackoff(tries)
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
