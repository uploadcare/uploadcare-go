package ucare

// RetryConfig controls automatic retry of throttled (HTTP 429) requests.
// When nil in Config (the default), throttled requests fail immediately.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts.
	// 0 means no retries even if RetryConfig is set.
	MaxRetries int
	// MaxWaitSeconds caps retry waits.
	// For REST requests, a positive Retry-After above this cap fails fast.
	// For upload requests, locally computed backoff is clamped to this cap.
	// 0 means no cap.
	MaxWaitSeconds int
}

// expBackoff returns exponential backoff wait: 1, 2, 4, 8... capped at 30s.
func expBackoff(attempt int) int {
	wait := 1 << (attempt - 1)
	return min(wait, 30)
}
